package runtime

import (
	stdcontext "context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type continuousPageExecution struct {
	queryKey       string
	entity         string
	direction      core.SortDirection
	pageSize       uint64
	originalOffset uint64
	ttlSeconds     uint64
	optimized      bool
	seekCursorID   string
}

type idSetBuildLockEntry struct {
	lock *sync.Mutex
	refs int
}

var idSetBuildLocks = struct {
	sync.Mutex
	entries map[string]*idSetBuildLockEntry
}{entries: make(map[string]*idSetBuildLockEntry)}

func acquireIDSetBuildLock(queryKey string) (*sync.Mutex, func()) {
	idSetBuildLocks.Lock()
	entry := idSetBuildLocks.entries[queryKey]
	if entry == nil {
		entry = &idSetBuildLockEntry{lock: &sync.Mutex{}}
		idSetBuildLocks.entries[queryKey] = entry
	}
	entry.refs++
	idSetBuildLocks.Unlock()
	return entry.lock, func() {
		idSetBuildLocks.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(idSetBuildLocks.entries, queryKey)
		}
		idSetBuildLocks.Unlock()
	}
}

type RuntimeDataService struct {
	metadata MetadataStore
	executor data_service.DataServiceExecutor
}

func NewRuntimeDataService(metadata MetadataStore, executor data_service.DataServiceExecutor) *RuntimeDataService {
	return &RuntimeDataService{
		metadata: metadata,
		executor: executor,
	}
}

func (s *RuntimeDataService) FetchAll(context stdcontext.Context, query *core.SelectQuery) (rows []core.Record, err error) {
	userCtx, _ := UserContextFrom(context)
	telemetry := RuntimeTelemetry(NoopRuntimeTelemetry{})
	if userCtx != nil {
		telemetry = userCtx.RuntimeTelemetry()
	}
	context, scope := StartRuntimeOperation(context, telemetry, NewRuntimeOperation("query", query.Entity+".list", map[string]RuntimeAttributeValue{
		"teaql.entity.type": query.Entity,
	}))
	defer func() {
		if err != nil {
			scope.Failure(RuntimeErrorType(err))
		} else {
			scope.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": len(rows)})
		}
	}()
	prepared := cloneSelectQuery(query, query.Entity)
	if err = prepared.PrepareForList(); err != nil {
		return nil, err
	}
	prepared, idSetOrder, idSetEmpty := s.prepareIDSetPage(context, prepared)
	if idSetEmpty {
		return []core.Record{}, nil
	}
	executionQuery, continuous := s.prepareContinuousPage(context, prepared)
	rows, err = s.fetchRows(context, executionQuery)
	if err != nil {
		return nil, err
	}
	if err := s.enhanceRelations(context, rows, executionQuery); err != nil {
		return nil, err
	}
	if err := s.enhanceRelationAggregates(context, rows, executionQuery); err != nil {
		return nil, err
	}
	if len(idSetOrder) > 0 {
		rows = restoreIDSetOrder(rows, idSetOrder)
	}
	s.registerContinuousPage(context, continuous, rows)
	return rows, nil
}

func (s *RuntimeDataService) prepareIDSetPage(context stdcontext.Context, query *core.SelectQuery) (*core.SelectQuery, []uint64, bool) {
	userCtx, ok := UserContextFrom(context)
	if !ok || query.IDSetPagination == nil {
		if ok {
			userCtx.ObserveIDSet("ID_SET_DISABLED", "UNKNOWN", 0)
		}
		return query, nil, false
	}
	options := query.IDSetPagination
	if query.Slice == nil || query.Slice.Limit == nil || *query.Slice.Limit == 0 ||
		query.PartitionBy != nil || len(query.Aggregates) > 0 || len(query.GroupBy) > 0 || query.RawSql != nil {
		userCtx.ObserveIDSet("ID_SET_FALLBACK_UNSUPPORTED_SHAPE", "UNKNOWN", 0)
		return query, nil, false
	}
	for _, order := range query.OrderBy {
		if order.Expr != nil || order.Field == "" {
			userCtx.ObserveIDSet("ID_SET_FALLBACK_NON_DETERMINISTIC_ORDER", "UNKNOWN", 0)
			return query, nil, false
		}
	}
	stable := query.Clone()
	hasIDOrder := false
	for _, order := range stable.OrderBy {
		if order.Field == "id" {
			hasIDOrder = true
			break
		}
	}
	if !hasIDOrder {
		stable.OrderAsc("id")
	}
	queryKey := idSetQueryKey(userCtx, stable, options.Namespace)
	retained, err := userCtx.IDSetStore().GetIDSet(context, queryKey)
	if err != nil {
		userCtx.ObserveIDSet("ID_SET_FALLBACK_STORE_UNAVAILABLE", "UNKNOWN", 0)
		return query, nil, false
	}
	plan := "ID_SET_HIT"
	if retained == nil {
		lock, release := acquireIDSetBuildLock(queryKey)
		defer release()
		lock.Lock()
		defer lock.Unlock()
		retained, err = userCtx.IDSetStore().GetIDSet(context, queryKey)
		if err != nil {
			userCtx.ObserveIDSet("ID_SET_FALLBACK_STORE_UNAVAILABLE", "UNKNOWN", 0)
			return query, nil, false
		}
		if retained == nil {
			idQuery := stable.Clone()
			idQuery.Projection = []string{"id"}
			idQuery.ExprProjection = nil
			idQuery.Relations = nil
			idQuery.RelationAggregates = nil
			idQuery.ChildEnhancements = nil
			idQuery.Slice = nil
			idQuery.Limit(options.MaxIDs + 1)
			idQuery.IDSetPagination = nil
			idRows, fetchErr := s.fetchRows(context, idQuery)
			if fetchErr != nil {
				userCtx.ObserveIDSet("ID_SET_FALLBACK_STORE_UNAVAILABLE", "UNKNOWN", 0)
				return query, nil, false
			}
			ids := make([]uint64, 0, len(idRows))
			for _, row := range idRows {
				id, present := row["id"]
				parsed, valid := id.TryU64()
				if !present || !valid {
					userCtx.ObserveIDSet("ID_SET_FALLBACK_UNSUPPORTED_SHAPE", "UNKNOWN", 0)
					return query, nil, false
				}
				ids = append(ids, parsed)
			}
			if uint64(len(ids)) > options.MaxIDs {
				userCtx.ObserveIDSet("ID_SET_FALLBACK_LIMIT_EXCEEDED", "LOWER_BOUND", uint64(len(ids)))
				return query, nil, false
			}
			retained = &RetainedIDSet{QueryKey: queryKey, IDs: ids, ExpiresAt: time.Now().Add(time.Duration(options.TTLSeconds) * time.Second)}
			if err = userCtx.IDSetStore().PutIDSet(context, retained); err != nil {
				userCtx.ObserveIDSet("ID_SET_FALLBACK_STORE_UNAVAILABLE", "UNKNOWN", 0)
				return query, nil, false
			}
			plan = "ID_SET_BUILD"
		}
	}
	userCtx.ObserveIDSet(plan, "EXACT", uint64(len(retained.IDs)))
	start := query.Slice.Offset
	if start >= uint64(len(retained.IDs)) {
		return query, nil, true
	}
	end := start + *query.Slice.Limit
	if end > uint64(len(retained.IDs)) {
		end = uint64(len(retained.IDs))
	}
	pageIDs := append([]uint64(nil), retained.IDs[start:end]...)
	values := make([]core.Value, len(pageIDs))
	for i, id := range pageIDs {
		values[i] = core.ValU64(id)
	}
	page := query.Clone()
	page.Slice = nil
	page.IDSetPagination = nil
	page.AndFilter(core.ExprInList("id", values))
	return page, pageIDs, false
}

func idSetQueryKey(context *UserContext, query *core.SelectQuery, namespace string) string {
	normalized := query.Clone()
	normalized.Slice = nil
	normalized.Projection = nil
	normalized.ExprProjection = nil
	normalized.Relations = nil
	normalized.RelationAggregates = nil
	normalized.CommentText = nil
	normalized.TraceChain = nil
	normalized.IDSetPagination = nil
	payload, _ := json.Marshal(normalized)
	scope := fmt.Sprintf("%s|%s|%T:%v|%T:%v|%v|", namespace, context.userIdentifier,
		context.GetResource("db"), context.GetResource("db"), context.requestPolicy,
		context.requestPolicy, context.GetResource("activeRoot"))
	digest := sha256.Sum256(append([]byte(scope), payload...))
	return "teaql:id-set:v1:" + hex.EncodeToString(digest[:])
}

func restoreIDSetOrder(rows []core.Record, ids []uint64) []core.Record {
	byID := make(map[uint64]core.Record, len(rows))
	for _, row := range rows {
		if value, ok := row["id"]; ok {
			if id, valid := value.TryU64(); valid {
				byID[id] = row
			}
		}
	}
	ordered := make([]core.Record, 0, len(rows))
	for _, id := range ids {
		if row, ok := byID[id]; ok {
			ordered = append(ordered, row)
		}
	}
	return ordered
}

func (s *RuntimeDataService) prepareContinuousPage(context stdcontext.Context, query *core.SelectQuery) (*core.SelectQuery, *continuousPageExecution) {
	userCtx, ok := UserContextFrom(context)
	if !ok || query.ContinuousPageFetch == nil {
		if ok {
			userCtx.ObserveContinuousPage("DISABLED", "")
		}
		return query, nil
	}
	options := query.ContinuousPageFetch
	if query.Slice == nil || query.Slice.Limit == nil || *query.Slice.Limit == 0 {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:INVALID_SLICE", "")
		return query, nil
	}
	if query.PartitionBy != nil || len(query.Aggregates) > 0 || len(query.GroupBy) > 0 {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:UNSUPPORTED_QUERY_SHAPE", "")
		return query, nil
	}
	if len(query.OrderBy) != 1 || query.OrderBy[0].Field != "id" || query.OrderBy[0].Expr != nil {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:ORDER_NOT_SEEKABLE_ID", "")
		return query, nil
	}
	queryKey := continuousPageQueryKey(userCtx, query, options.Namespace)
	execution := &continuousPageExecution{
		queryKey: queryKey, entity: query.Entity, direction: query.OrderBy[0].Direction,
		pageSize: *query.Slice.Limit, originalOffset: query.Slice.Offset,
		ttlSeconds: options.TTLSeconds,
	}
	if query.Slice.Offset == 0 {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:FIRST_PAGE", "")
		return query, execution
	}
	cacheContext, cacheScope := StartRuntimeOperation(context, userCtx.RuntimeTelemetry(), NewRuntimeOperation("cache", "continuous-page.get", map[string]RuntimeAttributeValue{"teaql.cache.operation": "get"}))
	cursor, err := userCtx.ContinuousPageCursorStore().GetContinuousPageCursor(cacheContext, queryKey, query.Slice.Offset)
	if err != nil {
		cacheScope.Failure(RuntimeErrorType(err))
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:STORE_UNAVAILABLE", "")
		return query, execution
	}
	if cursor == nil {
		cacheScope.Success(map[string]RuntimeAttributeValue{"teaql.cache.result": "miss"})
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:CACHE_MISS", "")
		return query, execution
	}
	cacheScope.Success(map[string]RuntimeAttributeValue{"teaql.cache.result": "hit"})
	if cursor.Entity != query.Entity || cursor.Direction != execution.direction ||
		cursor.PageSize != execution.pageSize || cursor.NextOffset != execution.originalOffset ||
		!cursor.ExpiresAt.After(time.Now()) {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:CURSOR_INVALID", "")
		return query, execution
	}
	query.Slice.Offset = 0
	seek := core.ExprGt("id", cursor.Boundary)
	if execution.direction == core.SortDesc {
		seek = core.ExprLt("id", cursor.Boundary)
	}
	if query.Filter == nil {
		query.Filter = seek
	} else {
		query.Filter = core.ExprAndNode(query.Filter, seek)
	}
	execution.optimized, execution.seekCursorID = true, cursor.CursorID
	userCtx.ObserveContinuousPage("CURSOR_SEEK", cursor.CursorID)
	return query, execution
}

func (s *RuntimeDataService) registerContinuousPage(context stdcontext.Context, execution *continuousPageExecution, rows []core.Record) {
	userCtx, ok := UserContextFrom(context)
	if !ok || execution == nil || uint64(len(rows)) != execution.pageSize || len(rows) == 0 {
		return
	}
	boundary, ok := rows[len(rows)-1]["id"]
	if !ok {
		return
	}
	cursor := &ContinuousPageCursor{
		CursorID: fmt.Sprintf("cpg_%x", time.Now().UnixNano()), QueryKey: execution.queryKey,
		Entity: execution.entity, Direction: execution.direction, Boundary: boundary,
		PageSize: execution.pageSize, NextOffset: execution.originalOffset + uint64(len(rows)),
		ExpiresAt: time.Now().Add(time.Duration(execution.ttlSeconds) * time.Second),
	}
	cacheContext, cacheScope := StartRuntimeOperation(context, userCtx.RuntimeTelemetry(), NewRuntimeOperation("cache", "continuous-page.put", map[string]RuntimeAttributeValue{"teaql.cache.operation": "put"}))
	if err := userCtx.ContinuousPageCursorStore().PutContinuousPageCursor(cacheContext, cursor); err != nil {
		cacheScope.Failure(RuntimeErrorType(err))
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:STORE_UNAVAILABLE", "")
	} else if execution.optimized {
		cacheScope.Success(map[string]RuntimeAttributeValue{"teaql.cache.result": "stored"})
		userCtx.ObserveContinuousPage("CURSOR_SEEK", execution.seekCursorID)
	} else if execution.originalOffset == 0 {
		cacheScope.Success(map[string]RuntimeAttributeValue{"teaql.cache.result": "stored"})
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:FIRST_PAGE", "")
	} else {
		cacheScope.Success(map[string]RuntimeAttributeValue{"teaql.cache.result": "stored"})
	}
}

func continuousPageQueryKey(context *UserContext, query *core.SelectQuery, namespace string) string {
	normalized := cloneSelectQuery(query, query.Entity)
	normalized.Slice.Offset = 0
	normalized.CommentText = nil
	normalized.TraceChain = nil
	payload, _ := json.Marshal(normalized)
	digest := sha256.Sum256(append([]byte(namespace+"|"+context.userIdentifier+"|"), payload...))
	return "teaql:continuous-page:v1:" + hex.EncodeToString(digest[:])
}

func (s *RuntimeDataService) fetchRows(context stdcontext.Context, query *core.SelectQuery) (rows []core.Record, err error) {
	userCtx, _ := UserContextFrom(context)
	telemetry := RuntimeTelemetry(NoopRuntimeTelemetry{})
	if userCtx != nil {
		telemetry = userCtx.RuntimeTelemetry()
	}
	context, scope := StartRuntimeOperation(context, telemetry, NewRuntimeOperation("provider", "data-service.query", map[string]RuntimeAttributeValue{
		"teaql.entity.type": query.Entity,
	}))
	defer func() {
		if err != nil {
			scope.Failure(RuntimeErrorType(err))
		} else {
			scope.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": len(rows)})
		}
	}()
	qExec, ok := s.executor.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("executor does not support Query")
	}

	req := &data_service.QueryRequest{
		Query:      query,
		TraceChain: query.TraceChain,
		Comment:    query.CommentText,
	}

	res, err := qExec.Query(context, req)
	if err != nil {
		return nil, &DataServiceError{Type: "Executor", ExecutorError: err}
	}

	rows = res.Rows
	return rows, nil
}

func (s *RuntimeDataService) enhanceRelations(context stdcontext.Context, parents []core.Record, query *core.SelectQuery) error {
	if len(parents) == 0 || len(query.Relations) == 0 {
		return nil
	}
	parentDescriptor := s.metadata.Entity(query.Entity)
	if parentDescriptor == nil {
		return fmt.Errorf("unknown entity %s", query.Entity)
	}
	for _, load := range query.Relations {
		userCtx, _ := UserContextFrom(context)
		telemetry := RuntimeTelemetry(NoopRuntimeTelemetry{})
		if userCtx != nil {
			telemetry = userCtx.RuntimeTelemetry()
		}
		relation := parentDescriptor.RelationByName(load.Name)
		if relation == nil {
			return fmt.Errorf("missing relation %s.%s", query.Entity, load.Name)
		}
		ids := make([]core.Value, 0, len(parents))
		for _, parent := range parents {
			if id, ok := parent[relation.LocKey]; ok {
				ids = append(ids, id)
			}
		}
		childQuery := cloneSelectQuery(load.Query, relation.TargetEntity)
		ensureProjection(childQuery, relation.ForKey)
		bounded := relation.IsMany && childQuery.Slice != nil && childQuery.Slice.Limit != nil
		useProbes := false
		if bounded {
			if childQuery.TopNProbeThreshold != nil {
				useProbes = *childQuery.TopNProbeThreshold > 0 && uint64(len(ids)) <= *childQuery.TopNProbeThreshold
			} else if provider, ok := s.executor.(interface{ TopNRelationPlanPolicy() string }); ok {
				useProbes = provider.TopNRelationPlanPolicy() == "always_probe"
			}
			ensureStableIDOrder(childQuery)
		}
		limit := uint64(0)
		if bounded {
			limit = *childQuery.Slice.Limit
		}
		threshold := "provider-default"
		if childQuery.TopNProbeThreshold != nil {
			threshold = fmt.Sprint(*childQuery.TopNProbeThreshold)
		}
		plan := "window"
		probeCount := 0
		if useProbes {
			plan = "bounded_probes"
			probeCount = len(ids)
		}
		relationContext, relationScope := StartRuntimeOperation(context, telemetry, NewRuntimeOperation("relation_load", query.Entity+"."+load.Name, map[string]RuntimeAttributeValue{
			"teaql.entity.type": query.Entity, "teaql.relation.name": load.Name,
			"teaql.relation.parent_count": len(ids), "teaql.relation.per_parent_limit": limit,
			"teaql.relation.configured_probe_threshold": threshold,
			"teaql.relation.selected_plan":              plan, "teaql.relation.probe_count": probeCount,
		}))
		var children []core.Record
		var err error
		if useProbes {
			children = make([]core.Record, 0)
			for _, id := range ids {
				probe := cloneSelectQuery(childQuery, relation.TargetEntity)
				probe.PartitionBy = nil
				probe.AndFilter(core.ExprEq(relation.ForKey, id))
				var rows []core.Record
				rows, err = s.fetchRows(relationContext, probe)
				if err != nil {
					break
				}
				children = append(children, rows...)
			}
		} else {
			childQuery.AndFilter(core.ExprInList(relation.ForKey, ids))
			if bounded {
				childQuery.PartitionByField(relation.ForKey)
			}
			children, err = s.fetchRows(relationContext, childQuery)
		}
		if err != nil {
			relationScope.Failure(RuntimeErrorType(err))
			return err
		}
		for _, child := range children {
			delete(child, "__teaql_partition_rank")
		}
		if err := s.enhanceRelations(relationContext, children, childQuery); err != nil {
			relationScope.Failure(RuntimeErrorType(err))
			return err
		}
		attachRelationRows(parents, children, load.Name, relation)
		relationScope.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": len(children)})
	}
	return nil
}

func ensureStableIDOrder(query *core.SelectQuery) {
	for _, order := range query.OrderBy {
		if order.Field == "id" {
			return
		}
	}
	query.OrderAsc("id")
}

func cloneSelectQuery(source *core.SelectQuery, entity string) *core.SelectQuery {
	if source == nil {
		return core.NewSelectQuery(entity)
	}
	clone := *source
	clone.Entity = entity
	clone.Projection = append([]string(nil), source.Projection...)
	clone.OrderBy = append([]*core.OrderBy(nil), source.OrderBy...)
	clone.Relations = append([]*core.RelationLoad(nil), source.Relations...)
	clone.RelationAggregates = append([]*core.RelationAggregate(nil), source.RelationAggregates...)
	clone.Aggregates = append([]*core.Aggregate(nil), source.Aggregates...)
	clone.GroupBy = append([]string(nil), source.GroupBy...)
	if source.Slice != nil {
		slice := *source.Slice
		clone.Slice = &slice
	}
	return &clone
}

func (s *RuntimeDataService) enhanceRelationAggregates(context stdcontext.Context, parents []core.Record, query *core.SelectQuery) error {
	if len(parents) == 0 || len(query.RelationAggregates) == 0 {
		return nil
	}
	parentDescriptor := s.metadata.Entity(query.Entity)
	if parentDescriptor == nil {
		return fmt.Errorf("unknown entity %s", query.Entity)
	}
	for _, aggregate := range query.RelationAggregates {
		relation := parentDescriptor.RelationByName(aggregate.RelationName)
		if relation == nil {
			return fmt.Errorf("missing relation %s.%s", query.Entity, aggregate.RelationName)
		}
		ids := make([]core.Value, 0, len(parents))
		for _, parent := range parents {
			if id, ok := parent[relation.LocKey]; ok {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			attachEmptyRelationAggregate(parents, aggregate, aggregate.Query)
			continue
		}
		childQuery := cloneSelectQuery(aggregate.Query, relation.TargetEntity)
		childQuery.Projection = nil
		childQuery.ExprProjection = nil
		childQuery.OrderBy = nil
		childQuery.Slice = nil
		childQuery.Relations = nil
		childQuery.RelationAggregates = nil
		if len(childQuery.Aggregates) == 0 {
			childQuery.Aggregates = []*core.Aggregate{core.AggCountAlias(aggregate.Alias)}
		}
		if !containsField(childQuery.GroupBy, relation.ForKey) {
			childQuery.GroupBy = append(childQuery.GroupBy, relation.ForKey)
		}
		childQuery.AndFilter(core.ExprInList(relation.ForKey, ids))
		rows, err := s.fetchRows(context, childQuery)
		if err != nil {
			return err
		}
		if childDescriptor := s.metadata.Entity(relation.TargetEntity); childDescriptor != nil {
			if property := childDescriptor.PropertyByName(relation.ForKey); property != nil && property.ColName != relation.ForKey {
				for _, row := range rows {
					if value, exists := row[property.ColName]; exists {
						row[relation.ForKey] = value
					}
				}
			}
		}
		attachRelationAggregateRows(parents, rows, relation, aggregate, childQuery)
	}
	return nil
}

func containsField(fields []string, field string) bool {
	for _, candidate := range fields {
		if candidate == field {
			return true
		}
	}
	return false
}

func attachEmptyRelationAggregate(parents []core.Record, aggregate *core.RelationAggregate, query *core.SelectQuery) {
	for _, parent := range parents {
		if aggregate.SingleResult {
			parent[aggregate.Alias] = emptyAggregateValue(query)
		} else {
			parent[aggregate.Alias] = core.Value{V: core.Record{}}
		}
	}
}

func emptyAggregateValue(query *core.SelectQuery) core.Value {
	if query == nil || len(query.Aggregates) == 0 || query.Aggregates[0].Function == core.AggCount {
		return core.ValI64(0)
	}
	return core.ValNull()
}

func attachRelationAggregateRows(parents, rows []core.Record, relation *core.RelationDescriptor, aggregate *core.RelationAggregate, query *core.SelectQuery) {
	buckets := make(map[string]core.Record, len(rows))
	for _, row := range rows {
		if foreignKey, ok := row[relation.ForKey]; ok {
			buckets[relationKey(foreignKey)] = row
		}
	}
	for _, parent := range parents {
		localKey, ok := parent[relation.LocKey]
		if !ok {
			continue
		}
		row, found := buckets[relationKey(localKey)]
		if !found {
			if aggregate.SingleResult {
				parent[aggregate.Alias] = emptyAggregateValue(query)
			} else {
				parent[aggregate.Alias] = core.Value{V: core.Record{}}
			}
			continue
		}
		if aggregate.SingleResult {
			alias := aggregate.Alias
			if len(query.Aggregates) > 0 {
				alias = query.Aggregates[0].Alias
			}
			parent[aggregate.Alias] = row[alias]
		} else {
			values := make(core.Record, len(row))
			for key, value := range row {
				if key != relation.ForKey {
					values[key] = value
				}
			}
			parent[aggregate.Alias] = core.Value{V: values}
		}
	}
}

func ensureProjection(query *core.SelectQuery, field string) {
	// An empty projection means SELECT all entity properties. Appending only the
	// relation foreign key would accidentally turn it into a narrow projection
	// and discard child id and business fields during relation loading.
	if len(query.Projection) == 0 {
		return
	}
	for _, selected := range query.Projection {
		if selected == field {
			return
		}
	}
	query.Projection = append(query.Projection, field)
}

func relationKey(value core.Value) string {
	return fmt.Sprintf("%T:%v", value.V, value.V)
}

func attachRelationRows(parents, children []core.Record, name string, relation *core.RelationDescriptor) {
	buckets := make(map[string][]core.Record)
	for _, child := range children {
		if foreignKey, ok := child[relation.ForKey]; ok {
			key := relationKey(foreignKey)
			buckets[key] = append(buckets[key], child)
		}
	}
	for _, parent := range parents {
		localKey, ok := parent[relation.LocKey]
		if !ok {
			continue
		}
		related := buckets[relationKey(localKey)]
		if relation.IsMany {
			parent[name] = core.Value{V: related}
		} else if len(related) > 0 {
			parent[name] = core.Value{V: related[0]}
		} else {
			parent[name] = core.ValNull()
		}
	}
}

// TODO: fetch entities etc

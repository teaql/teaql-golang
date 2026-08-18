package runtime

import (
	stdcontext "context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func (s *RuntimeDataService) FetchAll(context stdcontext.Context, query *core.SelectQuery) ([]core.Record, error) {
	prepared := cloneSelectQuery(query, query.Entity)
	if err := prepared.PrepareForList(); err != nil {
		return nil, err
	}
	executionQuery, continuous := s.prepareContinuousPage(context, prepared)
	rows, err := s.fetchRows(context, executionQuery)
	if err != nil {
		return nil, err
	}
	if err := s.enhanceRelations(context, rows, executionQuery); err != nil {
		return nil, err
	}
	s.registerContinuousPage(context, continuous, rows)
	return rows, nil
}

func (s *RuntimeDataService) prepareContinuousPage(context stdcontext.Context, query *core.SelectQuery) (*core.SelectQuery, *continuousPageExecution) {
	userCtx, ok := context.(*UserContext)
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
	cursor, err := userCtx.ContinuousPageCursorStore().GetContinuousPageCursor(context, queryKey, query.Slice.Offset)
	if err != nil {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:STORE_UNAVAILABLE", "")
		return query, execution
	}
	if cursor == nil {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:CACHE_MISS", "")
		return query, execution
	}
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
	userCtx, ok := context.(*UserContext)
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
	if err := userCtx.ContinuousPageCursorStore().PutContinuousPageCursor(context, cursor); err != nil {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:STORE_UNAVAILABLE", "")
	} else if execution.optimized {
		userCtx.ObserveContinuousPage("CURSOR_SEEK", execution.seekCursorID)
	} else if execution.originalOffset == 0 {
		userCtx.ObserveContinuousPage("OFFSET_FALLBACK:FIRST_PAGE", "")
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

func (s *RuntimeDataService) fetchRows(context stdcontext.Context, query *core.SelectQuery) ([]core.Record, error) {
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

	return res.Rows, nil
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
		childQuery.AndFilter(core.ExprInList(relation.ForKey, ids))
		if childQuery.Slice != nil {
			childQuery.PartitionByField(relation.ForKey)
		}
		children, err := s.fetchRows(context, childQuery)
		if err != nil {
			return err
		}
		for _, child := range children {
			delete(child, "__teaql_partition_rank")
		}
		if err := s.enhanceRelations(context, children, childQuery); err != nil {
			return err
		}
		attachRelationRows(parents, children, load.Name, relation)
	}
	return nil
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
	if source.Slice != nil {
		slice := *source.Slice
		clone.Slice = &slice
	}
	return &clone
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



package catalog

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"natural-message-service-core-workspace/lib/internal/generatedsearch"
	"natural-message-service-core-workspace/lib/entry"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CatalogRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCatalogRequest struct {
	request *CatalogRequest
}

func NewCatalogRequest() *CatalogRequest {
	r := &CatalogRequest{
		Query: core.NewSelectQuery("catalog"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCatalogMinimalRequest() *CatalogRequest {
	r := NewCatalogRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CatalogRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CatalogRequest) FilterByJson(source []byte) (*CatalogRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CatalogRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CatalogRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CatalogRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCatalog().EntityDescriptor()
}

func (r *CatalogRequest) NewRelationEntity() core.Entity {
	return newLoadedCatalog()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CatalogRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CatalogRequest) teaqlHydrateQueryRecord(rec core.Record) (*Catalog, error) {
	entity := newLoadedCatalog()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["entryList"]; selected {
		entity.markRelationLoaded("entryList")
		factory := r.relationFactories["entryList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation entryList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation entryList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*entry.Entry)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation entryList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.EntryList().Add(childEntity)
			}}
	return entity, nil
}

func (r *CatalogRequest) Comment(comment string) *CatalogRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CatalogRequest) Purpose(purpose string) *ExecutableCatalogRequest {
	r.purposeText = purpose
	return &ExecutableCatalogRequest{request: r}
}

func (r *ExecutableCatalogRequest) Comment(comment string) *ExecutableCatalogRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CatalogRequest) Limit(limit uint64) *CatalogRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CatalogRequest) Offset(offset uint64) *CatalogRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CatalogRequest) OptimizeForContinuousPageFetch() *CatalogRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CatalogRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CatalogRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CatalogRequest) OptimizePaginationWithIDSet() *CatalogRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CatalogRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CatalogRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CatalogRequest) TopNProbeParentThreshold(threshold uint64) *CatalogRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCatalogVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCatalogVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CatalogRequest) WithDeletedRows() *CatalogRequest {
	r.Query.Filter = removeCatalogVersionFilter(r.Query.Filter)
	return r
}

func (r *CatalogRequest) DeletedRowsOnly() *CatalogRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CatalogRequest) SelectId() *CatalogRequest {
	r.Query.Project("id")
	return r
}

func (r *CatalogRequest) WithIdIs(value uint64) *CatalogRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CatalogRequest) WithIdIsNot(value uint64) *CatalogRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CatalogRequest) WithIdIn(values []uint64) *CatalogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CatalogRequest) WithIdNotIn(values []uint64) *CatalogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CatalogRequest) WithIdGreaterThan(value uint64) *CatalogRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CatalogRequest) WithIdGreaterThanOrEqualTo(value uint64) *CatalogRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CatalogRequest) WithIdLessThan(value uint64) *CatalogRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CatalogRequest) WithIdLessThanOrEqualTo(value uint64) *CatalogRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CatalogRequest) WithIdBetween(lower uint64, upper uint64) *CatalogRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CatalogRequest) WithIdIsKnown() *CatalogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CatalogRequest) WithIdIsUnknown() *CatalogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CatalogRequest) OrderByIdAsc() *CatalogRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CatalogRequest) OrderByIdDesc() *CatalogRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CatalogRequest) SelectName() *CatalogRequest {
	r.Query.Project("name")
	return r
}

func (r *CatalogRequest) WithNameIs(value string) *CatalogRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *CatalogRequest) WithNameIsNot(value string) *CatalogRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *CatalogRequest) WithNameIn(values []string) *CatalogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *CatalogRequest) WithNameNotIn(values []string) *CatalogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *CatalogRequest) WithNameGreaterThan(value string) *CatalogRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *CatalogRequest) WithNameGreaterThanOrEqualTo(value string) *CatalogRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *CatalogRequest) WithNameLessThan(value string) *CatalogRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *CatalogRequest) WithNameLessThanOrEqualTo(value string) *CatalogRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *CatalogRequest) WithNameBetween(lower string, upper string) *CatalogRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *CatalogRequest) WithNameIsKnown() *CatalogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *CatalogRequest) WithNameIsUnknown() *CatalogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *CatalogRequest) WithNameContaining(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *CatalogRequest) WithNameNotContaining(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *CatalogRequest) WithNameStartingWith(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *CatalogRequest) WithNameNotStartingWith(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *CatalogRequest) WithNameEndingWith(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *CatalogRequest) WithNameNotEndingWith(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *CatalogRequest) WithNameSoundingLike(term string) *CatalogRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *CatalogRequest) OrderByNameAsc() *CatalogRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *CatalogRequest) OrderByNameDesc() *CatalogRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *CatalogRequest) SelectVersion() *CatalogRequest {
	r.Query.Project("version")
	return r
}

func (r *CatalogRequest) WithVersionIs(value int64) *CatalogRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CatalogRequest) WithVersionIsNot(value int64) *CatalogRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CatalogRequest) WithVersionIn(values []int64) *CatalogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CatalogRequest) WithVersionNotIn(values []int64) *CatalogRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CatalogRequest) WithVersionGreaterThan(value int64) *CatalogRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CatalogRequest) WithVersionGreaterThanOrEqualTo(value int64) *CatalogRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CatalogRequest) WithVersionLessThan(value int64) *CatalogRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CatalogRequest) WithVersionLessThanOrEqualTo(value int64) *CatalogRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CatalogRequest) WithVersionBetween(lower int64, upper int64) *CatalogRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CatalogRequest) WithVersionIsKnown() *CatalogRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CatalogRequest) WithVersionIsUnknown() *CatalogRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CatalogRequest) OrderByVersionAsc() *CatalogRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CatalogRequest) OrderByVersionDesc() *CatalogRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *CatalogRequest) CountEntries() *CatalogRequest {
	return r.CountEntriesAs("countEntries")

}
func (r *CatalogRequest) CountEntriesAs(alias string) *CatalogRequest {
	r.CountEntriesWith(alias, entry.NewEntryRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountEntriesAs", Alias: &alias, Relation: "entryList"}
	return r
}
func (r *CatalogRequest) CountEntriesWith(alias string, child *entry.EntryRequest) *CatalogRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("entryList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountEntriesWith", Alias: &alias, Relation: "entryList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CatalogRequest) MinBornAtOfEntries() *CatalogRequest {
	return r.MinBornAtOfEntriesAs("minBornAtOfEntries", entry.NewEntryRequest())
}
func (r *CatalogRequest) MinBornAtOfEntriesAs(alias string, child *entry.EntryRequest) *CatalogRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("born_at", alias)
	aggregate := core.NewRelationAggregate("entryList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinBornAtOfEntriesAs", Alias: &alias, Relation: "entryList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CatalogRequest) MaxBornAtOfEntries() *CatalogRequest {
	return r.MaxBornAtOfEntriesAs("maxBornAtOfEntries", entry.NewEntryRequest())
}
func (r *CatalogRequest) MaxBornAtOfEntriesAs(alias string, child *entry.EntryRequest) *CatalogRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("born_at", alias)
	aggregate := core.NewRelationAggregate("entryList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxBornAtOfEntriesAs", Alias: &alias, Relation: "entryList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CatalogRequest) MinTouchedAtOfEntries() *CatalogRequest {
	return r.MinTouchedAtOfEntriesAs("minTouchedAtOfEntries", entry.NewEntryRequest())
}
func (r *CatalogRequest) MinTouchedAtOfEntriesAs(alias string, child *entry.EntryRequest) *CatalogRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("touched_at", alias)
	aggregate := core.NewRelationAggregate("entryList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinTouchedAtOfEntriesAs", Alias: &alias, Relation: "entryList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *CatalogRequest) MaxTouchedAtOfEntries() *CatalogRequest {
	return r.MaxTouchedAtOfEntriesAs("maxTouchedAtOfEntries", entry.NewEntryRequest())
}
func (r *CatalogRequest) MaxTouchedAtOfEntriesAs(alias string, child *entry.EntryRequest) *CatalogRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("touched_at", alias)
	aggregate := core.NewRelationAggregate("entryList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxTouchedAtOfEntriesAs", Alias: &alias, Relation: "entryList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *CatalogRequest) SelectEntryList() *CatalogRequest {
	return r.SelectEntryListWith(entry.NewEntryRequest())
}

func (r *CatalogRequest) SelectEntryListWith(child *entry.EntryRequest) *CatalogRequest {
	r.Query.RelationQuerySelection("entryList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectEntryListWith", Relation: "entryList"}
	r.relationFactories["entryList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *CatalogRequest) HaveEntries() *CatalogRequest {
	return r.WithEntryListMatching(entry.NewEntryRequest())
}

func (r *CatalogRequest) HaveNoEntries() *CatalogRequest {
	return r.WithoutEntryListMatching(entry.NewEntryRequest())
}

func (r *CatalogRequest) WithEntryListMatching(child *entry.EntryRequest) *CatalogRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "catalog_id"))
	return r
}

func (r *CatalogRequest) WithoutEntryListMatching(child *entry.EntryRequest) *CatalogRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "catalog_id"))
	return r
}

func (e *ExecutableCatalogRequest) NewEntity(context *runtime.UserContext) *Catalog {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewCatalog()
	initialized := context.InitializeEntity("Catalog", entity)
	typed, ok := initialized.(*Catalog)
	if !ok {
		panic("entity initializer changed Catalog to an incompatible type")
	}
	return typed
}

func (e *ExecutableCatalogRequest) ExecuteForOne(context *runtime.UserContext) (*Catalog, error) {
	request := *e.request
	request.Query = e.request.Query.Clone()
	request.Query.Limit(1)
	executable := *e
	executable.request = &request
	list, err := executable.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableCatalogRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Catalog], error) {
	// Apply trusted policy once. The runtime owns the entire typed list lifecycle,
	// including optional exact count, facets and hydration failures.
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, plan, err := runtime.CaptureQueryPlan(query, r.queryOptions)
	if err != nil { return nil, err }
	query = plan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Catalog, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCatalogRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Catalog) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, err := runtime.CaptureStreamQuery(query, r.queryOptions)
	if err != nil { return err }
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Catalog, error) {
		entity := newLoadedCatalog()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCatalogRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCatalogRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, plan, err := runtime.CaptureQueryPlan(query, r.queryOptions)
	if err != nil { return nil, nil, nil, err }
	query = plan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return nil, nil, nil, err }

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, nil, nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, nil, nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAllWithFacetPlan(context, authorized, plan)
	if err != nil {
		return nil, nil, nil, err
	}
	return rows, authorized, plan, nil
}

// ExecuteForRows preserves aggregate/group projections as records while keeping
// the cross-language SmartList result boundary.
func (e *ExecutableCatalogRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CatalogRequest) Count() *CatalogRequest {
	return r.CountAs("count")
}

func (r *CatalogRequest) CountAs(alias string) *CatalogRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CatalogRequest) GroupById() *CatalogRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CatalogRequest) GroupByName() *CatalogRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *CatalogRequest) GroupByVersion() *CatalogRequest {
	r.Query.WithGroupBy("version")
	return r
}

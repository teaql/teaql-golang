

package supplier

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"supplier-onboarding-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type SupplierRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableSupplierRequest struct {
	request *SupplierRequest
}

func NewSupplierRequest() *SupplierRequest {
	r := &SupplierRequest{
		Query: core.NewSelectQuery("Supplier"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewSupplierMinimalRequest() *SupplierRequest {
	r := NewSupplierRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *SupplierRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *SupplierRequest) FilterByJson(source []byte) (*SupplierRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *SupplierRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *SupplierRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *SupplierRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewSupplier().EntityDescriptor()
}

func (r *SupplierRequest) NewRelationEntity() core.Entity {
	return newLoadedSupplier()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *SupplierRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *SupplierRequest) teaqlHydrateQueryRecord(rec core.Record) (*Supplier, error) {
	entity := newLoadedSupplier()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["warehouseEntity"]; selected {
		entity.markRelationLoaded("warehouseEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["warehouseEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("warehouseEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *SupplierRequest) Comment(comment string) *SupplierRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *SupplierRequest) Purpose(purpose string) *ExecutableSupplierRequest {
	r.purposeText = purpose
	return &ExecutableSupplierRequest{request: r}
}

func (r *ExecutableSupplierRequest) Comment(comment string) *ExecutableSupplierRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *SupplierRequest) Limit(limit uint64) *SupplierRequest {
	r.Query.Limit(limit)
	return r
}

func (r *SupplierRequest) Offset(offset uint64) *SupplierRequest {
	r.Query.Offset(offset)
	return r
}

func (r *SupplierRequest) OptimizeForContinuousPageFetch() *SupplierRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *SupplierRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *SupplierRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *SupplierRequest) OptimizePaginationWithIDSet() *SupplierRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *SupplierRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *SupplierRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *SupplierRequest) TopNProbeParentThreshold(threshold uint64) *SupplierRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeSupplierVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeSupplierVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *SupplierRequest) WithDeletedRows() *SupplierRequest {
	r.Query.Filter = removeSupplierVersionFilter(r.Query.Filter)
	return r
}

func (r *SupplierRequest) DeletedRowsOnly() *SupplierRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *SupplierRequest) SelectId() *SupplierRequest {
	r.Query.Project("id")
	return r
}

func (r *SupplierRequest) WithIdIs(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithIdIsNot(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithIdIn(values []uint64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *SupplierRequest) WithIdNotIn(values []uint64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *SupplierRequest) WithIdGreaterThan(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithIdGreaterThanOrEqualTo(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithIdLessThan(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithIdLessThanOrEqualTo(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithIdBetween(lower uint64, upper uint64) *SupplierRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *SupplierRequest) WithIdIsKnown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *SupplierRequest) WithIdIsUnknown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *SupplierRequest) OrderByIdAsc() *SupplierRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *SupplierRequest) OrderByIdDesc() *SupplierRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *SupplierRequest) SelectWarehouse() *SupplierRequest {
	r.Query.Project("warehouse_id")
	return r
}

func (r *SupplierRequest) WithWarehouseIs(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprEq("warehouse_id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithWarehouseIsNot(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprNe("warehouse_id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithWarehouseIn(values []uint64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("warehouse_id", converted))
	return r
}
func (r *SupplierRequest) WithWarehouseNotIn(values []uint64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("warehouse_id", converted))
	return r
}
func (r *SupplierRequest) WithWarehouseGreaterThan(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGt("warehouse_id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithWarehouseGreaterThanOrEqualTo(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGte("warehouse_id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithWarehouseLessThan(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLt("warehouse_id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithWarehouseLessThanOrEqualTo(value uint64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLte("warehouse_id", core.ValU64(value)))
	return r
}
func (r *SupplierRequest) WithWarehouseBetween(lower uint64, upper uint64) *SupplierRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("warehouse_id", from, to))
	return r
}
func (r *SupplierRequest) WithWarehouseIsKnown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("warehouse_id"))
	return r
}
func (r *SupplierRequest) WithWarehouseIsUnknown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNullNode("warehouse_id"))
	return r
}
func (r *SupplierRequest) FacetByWarehouseAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *SupplierRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "warehouse_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByWarehouseAs", Alias: &name, Relation: "warehouse"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *SupplierRequest) OrderByWarehouseAsc() *SupplierRequest {
	r.Query.OrderAsc("warehouse_id")
	return r
}
func (r *SupplierRequest) OrderByWarehouseDesc() *SupplierRequest {
	r.Query.OrderDesc("warehouse_id")
	return r
}
func (r *SupplierRequest) SelectName() *SupplierRequest {
	r.Query.Project("name")
	return r
}

func (r *SupplierRequest) WithNameIs(value string) *SupplierRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *SupplierRequest) WithNameIsNot(value string) *SupplierRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *SupplierRequest) WithNameIn(values []string) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *SupplierRequest) WithNameNotIn(values []string) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *SupplierRequest) WithNameGreaterThan(value string) *SupplierRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *SupplierRequest) WithNameGreaterThanOrEqualTo(value string) *SupplierRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *SupplierRequest) WithNameLessThan(value string) *SupplierRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *SupplierRequest) WithNameLessThanOrEqualTo(value string) *SupplierRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *SupplierRequest) WithNameBetween(lower string, upper string) *SupplierRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *SupplierRequest) WithNameIsKnown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *SupplierRequest) WithNameIsUnknown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *SupplierRequest) WithNameContaining(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *SupplierRequest) WithNameNotContaining(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *SupplierRequest) WithNameStartingWith(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *SupplierRequest) WithNameNotStartingWith(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *SupplierRequest) WithNameEndingWith(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *SupplierRequest) WithNameNotEndingWith(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *SupplierRequest) WithNameSoundingLike(term string) *SupplierRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *SupplierRequest) OrderByNameAsc() *SupplierRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *SupplierRequest) OrderByNameDesc() *SupplierRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *SupplierRequest) SelectMinimumOrderQuantity() *SupplierRequest {
	r.Query.Project("minimum_order_quantity")
	return r
}

func (r *SupplierRequest) WithMinimumOrderQuantityIs(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprEq("minimum_order_quantity", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityIsNot(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprNe("minimum_order_quantity", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityIn(values []int64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("minimum_order_quantity", converted))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityNotIn(values []int64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("minimum_order_quantity", converted))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityGreaterThan(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGt("minimum_order_quantity", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityGreaterThanOrEqualTo(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGte("minimum_order_quantity", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityLessThan(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLt("minimum_order_quantity", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityLessThanOrEqualTo(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLte("minimum_order_quantity", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityBetween(lower int64, upper int64) *SupplierRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("minimum_order_quantity", from, to))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityIsKnown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("minimum_order_quantity"))
	return r
}
func (r *SupplierRequest) WithMinimumOrderQuantityIsUnknown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNullNode("minimum_order_quantity"))
	return r
}
func (r *SupplierRequest) OrderByMinimumOrderQuantityAsc() *SupplierRequest {
	r.Query.OrderAsc("minimum_order_quantity")
	return r
}
func (r *SupplierRequest) OrderByMinimumOrderQuantityDesc() *SupplierRequest {
	r.Query.OrderDesc("minimum_order_quantity")
	return r
}
func (r *SupplierRequest) SelectVersion() *SupplierRequest {
	r.Query.Project("version")
	return r
}

func (r *SupplierRequest) WithVersionIs(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithVersionIsNot(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithVersionIn(values []int64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *SupplierRequest) WithVersionNotIn(values []int64) *SupplierRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *SupplierRequest) WithVersionGreaterThan(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithVersionGreaterThanOrEqualTo(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithVersionLessThan(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithVersionLessThanOrEqualTo(value int64) *SupplierRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *SupplierRequest) WithVersionBetween(lower int64, upper int64) *SupplierRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *SupplierRequest) WithVersionIsKnown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *SupplierRequest) WithVersionIsUnknown() *SupplierRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *SupplierRequest) OrderByVersionAsc() *SupplierRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *SupplierRequest) OrderByVersionDesc() *SupplierRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *SupplierRequest) SelectWarehouseWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *SupplierRequest {
	runtime.EnsureRelationProjection(r.Query, "warehouse_id")
	r.Query.RelationQuerySelection("warehouseEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectWarehouseWith", Relation: "warehouseEntity"}
	r.relationFactories["warehouseEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *SupplierRequest) WithWarehouseMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *SupplierRequest {
	r.Query.AndFilter(core.ExprInSubQuery("warehouse_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *SupplierRequest) WithoutWarehouseMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *SupplierRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("warehouse_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableSupplierRequest) NewEntity(context *runtime.UserContext) *Supplier {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewSupplier()
	initialized := context.InitializeEntity("Supplier", entity)
	typed, ok := initialized.(*Supplier)
	if !ok {
		panic("entity initializer changed Supplier to an incompatible type")
	}
	return typed
}

func (e *ExecutableSupplierRequest) ExecuteForOne(context *runtime.UserContext) (*Supplier, error) {
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

func (e *ExecutableSupplierRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Supplier], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Supplier, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableSupplierRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Supplier) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Supplier, error) {
		entity := newLoadedSupplier()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableSupplierRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableSupplierRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableSupplierRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *SupplierRequest) Count() *SupplierRequest {
	return r.CountAs("count")
}

func (r *SupplierRequest) CountAs(alias string) *SupplierRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *SupplierRequest) MinMinimumOrderQuantity() *SupplierRequest {
	return r.MinMinimumOrderQuantityAs("minOfMinimumOrderQuantity")
}

func (r *SupplierRequest) MinMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.Min("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) MaxMinimumOrderQuantity() *SupplierRequest {
	return r.MaxMinimumOrderQuantityAs("maxOfMinimumOrderQuantity")
}

func (r *SupplierRequest) MaxMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.Max("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) SumMinimumOrderQuantity() *SupplierRequest {
	return r.SumMinimumOrderQuantityAs("sumOfMinimumOrderQuantity")
}

func (r *SupplierRequest) SumMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.Sum("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) AvgMinimumOrderQuantity() *SupplierRequest {
	return r.AvgMinimumOrderQuantityAs("avgOfMinimumOrderQuantity")
}

func (r *SupplierRequest) AvgMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.Avg("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) StddevMinimumOrderQuantity() *SupplierRequest {
	return r.StddevMinimumOrderQuantityAs("standardDeviationOfMinimumOrderQuantity")
}

func (r *SupplierRequest) StddevMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.Stddev("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) StddevPopMinimumOrderQuantity() *SupplierRequest {
	return r.StddevPopMinimumOrderQuantityAs("squareRootOfPopulationStandardDeviationOfMinimumOrderQuantity")
}

func (r *SupplierRequest) StddevPopMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.StddevPop("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) VarSampMinimumOrderQuantity() *SupplierRequest {
	return r.VarSampMinimumOrderQuantityAs("sampleVarianceOfMinimumOrderQuantity")
}

func (r *SupplierRequest) VarSampMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.VarSamp("minimum_order_quantity", alias)
	return r
}
func (r *SupplierRequest) VarPopMinimumOrderQuantity() *SupplierRequest {
	return r.VarPopMinimumOrderQuantityAs("samplePopulationVarianceOfMinimumOrderQuantity")
}

func (r *SupplierRequest) VarPopMinimumOrderQuantityAs(alias string) *SupplierRequest {
	r.Query.VarPop("minimum_order_quantity", alias)
	return r
}

func (r *SupplierRequest) GroupById() *SupplierRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *SupplierRequest) GroupByWarehouse() *SupplierRequest {
	r.Query.WithGroupBy("warehouse_id")
	return r
}
func (r *SupplierRequest) GroupByName() *SupplierRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *SupplierRequest) GroupByMinimumOrderQuantity() *SupplierRequest {
	r.Query.WithGroupBy("minimum_order_quantity")
	return r
}
func (r *SupplierRequest) GroupByVersion() *SupplierRequest {
	r.Query.WithGroupBy("version")
	return r
}

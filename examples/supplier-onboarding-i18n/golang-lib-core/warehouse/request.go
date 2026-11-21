

package warehouse

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"supplier-onboarding-service-core-workspace/lib/internal/generatedsearch"
	"supplier-onboarding-service-core-workspace/lib/supplier"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type WarehouseRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableWarehouseRequest struct {
	request *WarehouseRequest
}

func NewWarehouseRequest() *WarehouseRequest {
	r := &WarehouseRequest{
		Query: core.NewSelectQuery("Warehouse"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewWarehouseMinimalRequest() *WarehouseRequest {
	r := NewWarehouseRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *WarehouseRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *WarehouseRequest) FilterByJson(source []byte) (*WarehouseRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *WarehouseRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *WarehouseRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *WarehouseRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewWarehouse().EntityDescriptor()
}

func (r *WarehouseRequest) NewRelationEntity() core.Entity {
	return newLoadedWarehouse()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *WarehouseRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *WarehouseRequest) teaqlHydrateQueryRecord(rec core.Record) (*Warehouse, error) {
	entity := newLoadedWarehouse()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["supplierList"]; selected {
		entity.markRelationLoaded("supplierList")
		factory := r.relationFactories["supplierList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation supplierList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation supplierList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*supplier.Supplier)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation supplierList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.SupplierList().Add(childEntity)
			}}
	return entity, nil
}

func (r *WarehouseRequest) Comment(comment string) *WarehouseRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *WarehouseRequest) Purpose(purpose string) *ExecutableWarehouseRequest {
	r.purposeText = purpose
	return &ExecutableWarehouseRequest{request: r}
}

func (r *ExecutableWarehouseRequest) Comment(comment string) *ExecutableWarehouseRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *WarehouseRequest) Limit(limit uint64) *WarehouseRequest {
	r.Query.Limit(limit)
	return r
}

func (r *WarehouseRequest) Offset(offset uint64) *WarehouseRequest {
	r.Query.Offset(offset)
	return r
}

func (r *WarehouseRequest) OptimizeForContinuousPageFetch() *WarehouseRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *WarehouseRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *WarehouseRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *WarehouseRequest) OptimizePaginationWithIDSet() *WarehouseRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *WarehouseRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *WarehouseRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *WarehouseRequest) TopNProbeParentThreshold(threshold uint64) *WarehouseRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeWarehouseVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeWarehouseVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *WarehouseRequest) WithDeletedRows() *WarehouseRequest {
	r.Query.Filter = removeWarehouseVersionFilter(r.Query.Filter)
	return r
}

func (r *WarehouseRequest) DeletedRowsOnly() *WarehouseRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *WarehouseRequest) SelectId() *WarehouseRequest {
	r.Query.Project("id")
	return r
}

func (r *WarehouseRequest) WithIdIs(value uint64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *WarehouseRequest) WithIdIsNot(value uint64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *WarehouseRequest) WithIdIn(values []uint64) *WarehouseRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *WarehouseRequest) WithIdNotIn(values []uint64) *WarehouseRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *WarehouseRequest) WithIdGreaterThan(value uint64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *WarehouseRequest) WithIdGreaterThanOrEqualTo(value uint64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *WarehouseRequest) WithIdLessThan(value uint64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *WarehouseRequest) WithIdLessThanOrEqualTo(value uint64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *WarehouseRequest) WithIdBetween(lower uint64, upper uint64) *WarehouseRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *WarehouseRequest) WithIdIsKnown() *WarehouseRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *WarehouseRequest) WithIdIsUnknown() *WarehouseRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *WarehouseRequest) OrderByIdAsc() *WarehouseRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *WarehouseRequest) OrderByIdDesc() *WarehouseRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *WarehouseRequest) SelectName() *WarehouseRequest {
	r.Query.Project("name")
	return r
}

func (r *WarehouseRequest) WithNameIs(value string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *WarehouseRequest) WithNameIsNot(value string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *WarehouseRequest) WithNameIn(values []string) *WarehouseRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *WarehouseRequest) WithNameNotIn(values []string) *WarehouseRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *WarehouseRequest) WithNameGreaterThan(value string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *WarehouseRequest) WithNameGreaterThanOrEqualTo(value string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *WarehouseRequest) WithNameLessThan(value string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *WarehouseRequest) WithNameLessThanOrEqualTo(value string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *WarehouseRequest) WithNameBetween(lower string, upper string) *WarehouseRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *WarehouseRequest) WithNameIsKnown() *WarehouseRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *WarehouseRequest) WithNameIsUnknown() *WarehouseRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *WarehouseRequest) WithNameContaining(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *WarehouseRequest) WithNameNotContaining(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *WarehouseRequest) WithNameStartingWith(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *WarehouseRequest) WithNameNotStartingWith(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *WarehouseRequest) WithNameEndingWith(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *WarehouseRequest) WithNameNotEndingWith(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *WarehouseRequest) WithNameSoundingLike(term string) *WarehouseRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *WarehouseRequest) OrderByNameAsc() *WarehouseRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *WarehouseRequest) OrderByNameDesc() *WarehouseRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *WarehouseRequest) SelectVersion() *WarehouseRequest {
	r.Query.Project("version")
	return r
}

func (r *WarehouseRequest) WithVersionIs(value int64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *WarehouseRequest) WithVersionIsNot(value int64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *WarehouseRequest) WithVersionIn(values []int64) *WarehouseRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *WarehouseRequest) WithVersionNotIn(values []int64) *WarehouseRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *WarehouseRequest) WithVersionGreaterThan(value int64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *WarehouseRequest) WithVersionGreaterThanOrEqualTo(value int64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *WarehouseRequest) WithVersionLessThan(value int64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *WarehouseRequest) WithVersionLessThanOrEqualTo(value int64) *WarehouseRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *WarehouseRequest) WithVersionBetween(lower int64, upper int64) *WarehouseRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *WarehouseRequest) WithVersionIsKnown() *WarehouseRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *WarehouseRequest) WithVersionIsUnknown() *WarehouseRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *WarehouseRequest) OrderByVersionAsc() *WarehouseRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *WarehouseRequest) OrderByVersionDesc() *WarehouseRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *WarehouseRequest) CountSuppliers() *WarehouseRequest {
	return r.CountSuppliersAs("countSuppliers")

}
func (r *WarehouseRequest) CountSuppliersAs(alias string) *WarehouseRequest {
	r.CountSuppliersWith(alias, supplier.NewSupplierRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountSuppliersAs", Alias: &alias, Relation: "supplierList"}
	return r
}
func (r *WarehouseRequest) CountSuppliersWith(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountSuppliersWith", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *WarehouseRequest) SumMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.SumMinimumOrderQuantityOfSuppliersAs("sumMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) SumMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) MinMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.MinMinimumOrderQuantityOfSuppliersAs("minMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) MinMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) MaxMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.MaxMinimumOrderQuantityOfSuppliersAs("maxMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) MaxMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) AvgMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.AvgMinimumOrderQuantityOfSuppliersAs("avgMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) AvgMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) StandardDeviationMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.StandardDeviationMinimumOrderQuantityOfSuppliersAs("standardDeviationMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) StandardDeviationMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) SquareRootOfPopulationStandardDeviationMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.SquareRootOfPopulationStandardDeviationMinimumOrderQuantityOfSuppliersAs("squareRootOfPopulationStandardDeviationMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) SquareRootOfPopulationStandardDeviationMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) SampleVarianceMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.SampleVarianceMinimumOrderQuantityOfSuppliersAs("sampleVarianceMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) SampleVarianceMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *WarehouseRequest) SamplePopulationVarianceMinimumOrderQuantityOfSuppliers() *WarehouseRequest {
	return r.SamplePopulationVarianceMinimumOrderQuantityOfSuppliersAs("samplePopulationVarianceMinimumOrderQuantityOfSuppliers", supplier.NewSupplierRequest())
}
func (r *WarehouseRequest) SamplePopulationVarianceMinimumOrderQuantityOfSuppliersAs(alias string, child *supplier.SupplierRequest) *WarehouseRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("minimum_order_quantity", alias)
	aggregate := core.NewRelationAggregate("supplierList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceMinimumOrderQuantityOfSuppliersAs", Alias: &alias, Relation: "supplierList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *WarehouseRequest) SelectSupplierList() *WarehouseRequest {
	return r.SelectSupplierListWith(supplier.NewSupplierRequest())
}

func (r *WarehouseRequest) SelectSupplierListWith(child *supplier.SupplierRequest) *WarehouseRequest {
	r.Query.RelationQuerySelection("supplierList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectSupplierListWith", Relation: "supplierList"}
	r.relationFactories["supplierList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *WarehouseRequest) HaveSuppliers() *WarehouseRequest {
	return r.WithSupplierListMatching(supplier.NewSupplierRequest())
}

func (r *WarehouseRequest) HaveNoSuppliers() *WarehouseRequest {
	return r.WithoutSupplierListMatching(supplier.NewSupplierRequest())
}

func (r *WarehouseRequest) WithSupplierListMatching(child *supplier.SupplierRequest) *WarehouseRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "warehouse_id"))
	return r
}

func (r *WarehouseRequest) WithoutSupplierListMatching(child *supplier.SupplierRequest) *WarehouseRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "warehouse_id"))
	return r
}

func (e *ExecutableWarehouseRequest) NewEntity(context *runtime.UserContext) *Warehouse {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewWarehouse()
	initialized := context.InitializeEntity("Warehouse", entity)
	typed, ok := initialized.(*Warehouse)
	if !ok {
		panic("entity initializer changed Warehouse to an incompatible type")
	}
	return typed
}

func (e *ExecutableWarehouseRequest) ExecuteForOne(context *runtime.UserContext) (*Warehouse, error) {
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

func (e *ExecutableWarehouseRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Warehouse], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Warehouse, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableWarehouseRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Warehouse) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Warehouse, error) {
		entity := newLoadedWarehouse()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableWarehouseRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableWarehouseRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableWarehouseRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *WarehouseRequest) Count() *WarehouseRequest {
	return r.CountAs("count")
}

func (r *WarehouseRequest) CountAs(alias string) *WarehouseRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *WarehouseRequest) GroupById() *WarehouseRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *WarehouseRequest) GroupByName() *WarehouseRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *WarehouseRequest) GroupByVersion() *WarehouseRequest {
	r.Query.WithGroupBy("version")
	return r
}

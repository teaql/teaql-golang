

package type_platform

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"scalar-type-conformance-service-core-workspace/lib/internal/generatedsearch"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type TypePlatformRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableTypePlatformRequest struct {
	request *TypePlatformRequest
}

func NewTypePlatformRequest() *TypePlatformRequest {
	r := &TypePlatformRequest{
		Query: core.NewSelectQuery("Type Platform"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewTypePlatformMinimalRequest() *TypePlatformRequest {
	r := NewTypePlatformRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *TypePlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *TypePlatformRequest) FilterByJson(source []byte) (*TypePlatformRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *TypePlatformRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *TypePlatformRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *TypePlatformRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewTypePlatform().EntityDescriptor()
}

func (r *TypePlatformRequest) NewRelationEntity() core.Entity {
	return newLoadedTypePlatform()
}

func (r *TypePlatformRequest) Comment(comment string) *TypePlatformRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *TypePlatformRequest) Purpose(purpose string) *ExecutableTypePlatformRequest {
	r.purposeText = purpose
	return &ExecutableTypePlatformRequest{request: r}
}

func (r *ExecutableTypePlatformRequest) Comment(comment string) *ExecutableTypePlatformRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *TypePlatformRequest) Limit(limit uint64) *TypePlatformRequest {
	r.Query.Limit(limit)
	return r
}

func (r *TypePlatformRequest) Offset(offset uint64) *TypePlatformRequest {
	r.Query.Offset(offset)
	return r
}

func (r *TypePlatformRequest) OptimizeForContinuousPageFetch() *TypePlatformRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *TypePlatformRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *TypePlatformRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *TypePlatformRequest) OptimizePaginationWithIDSet() *TypePlatformRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *TypePlatformRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *TypePlatformRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *TypePlatformRequest) TopNProbeParentThreshold(threshold uint64) *TypePlatformRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeTypePlatformVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeTypePlatformVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *TypePlatformRequest) WithDeletedRows() *TypePlatformRequest {
	r.Query.Filter = removeTypePlatformVersionFilter(r.Query.Filter)
	return r
}

func (r *TypePlatformRequest) DeletedRowsOnly() *TypePlatformRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *TypePlatformRequest) SelectId() *TypePlatformRequest {
	r.Query.Project("id")
	return r
}

func (r *TypePlatformRequest) WithIdIs(value uint64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *TypePlatformRequest) WithIdIsNot(value uint64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *TypePlatformRequest) WithIdIn(values []uint64) *TypePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *TypePlatformRequest) WithIdNotIn(values []uint64) *TypePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *TypePlatformRequest) WithIdGreaterThan(value uint64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *TypePlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *TypePlatformRequest) WithIdLessThan(value uint64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *TypePlatformRequest) WithIdLessThanOrEqualTo(value uint64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *TypePlatformRequest) WithIdBetween(lower uint64, upper uint64) *TypePlatformRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *TypePlatformRequest) WithIdIsKnown() *TypePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *TypePlatformRequest) WithIdIsUnknown() *TypePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *TypePlatformRequest) OrderByIdAsc() *TypePlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *TypePlatformRequest) OrderByIdDesc() *TypePlatformRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *TypePlatformRequest) SelectName() *TypePlatformRequest {
	r.Query.Project("name")
	return r
}

func (r *TypePlatformRequest) WithNameIs(value string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *TypePlatformRequest) WithNameIsNot(value string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *TypePlatformRequest) WithNameIn(values []string) *TypePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *TypePlatformRequest) WithNameNotIn(values []string) *TypePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *TypePlatformRequest) WithNameGreaterThan(value string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *TypePlatformRequest) WithNameGreaterThanOrEqualTo(value string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *TypePlatformRequest) WithNameLessThan(value string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *TypePlatformRequest) WithNameLessThanOrEqualTo(value string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *TypePlatformRequest) WithNameBetween(lower string, upper string) *TypePlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *TypePlatformRequest) WithNameIsKnown() *TypePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *TypePlatformRequest) WithNameIsUnknown() *TypePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *TypePlatformRequest) WithNameContaining(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *TypePlatformRequest) WithNameNotContaining(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *TypePlatformRequest) WithNameStartingWith(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *TypePlatformRequest) WithNameNotStartingWith(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *TypePlatformRequest) WithNameEndingWith(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *TypePlatformRequest) WithNameNotEndingWith(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *TypePlatformRequest) WithNameSoundingLike(term string) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *TypePlatformRequest) OrderByNameAsc() *TypePlatformRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *TypePlatformRequest) OrderByNameDesc() *TypePlatformRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *TypePlatformRequest) SelectVersion() *TypePlatformRequest {
	r.Query.Project("version")
	return r
}

func (r *TypePlatformRequest) WithVersionIs(value int64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *TypePlatformRequest) WithVersionIsNot(value int64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *TypePlatformRequest) WithVersionIn(values []int64) *TypePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *TypePlatformRequest) WithVersionNotIn(values []int64) *TypePlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *TypePlatformRequest) WithVersionGreaterThan(value int64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *TypePlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *TypePlatformRequest) WithVersionLessThan(value int64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *TypePlatformRequest) WithVersionLessThanOrEqualTo(value int64) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *TypePlatformRequest) WithVersionBetween(lower int64, upper int64) *TypePlatformRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *TypePlatformRequest) WithVersionIsKnown() *TypePlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *TypePlatformRequest) WithVersionIsUnknown() *TypePlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *TypePlatformRequest) OrderByVersionAsc() *TypePlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *TypePlatformRequest) OrderByVersionDesc() *TypePlatformRequest {
	r.Query.OrderDesc("version")
	return r
}



func (r *TypePlatformRequest) CountScalarTypeProbes() *TypePlatformRequest {
	return r.CountScalarTypeProbesAs("countScalarTypeProbes")

}
func (r *TypePlatformRequest) CountScalarTypeProbesAs(alias string) *TypePlatformRequest {
	r.CountScalarTypeProbesWith(alias, scalar_type_probe.NewScalarTypeProbeRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	return r
}
func (r *TypePlatformRequest) CountScalarTypeProbesWith(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountScalarTypeProbesWith", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *TypePlatformRequest) SumRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumRequiredIntegerOfScalarTypeProbesAs("sumRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredIntegerOfScalarTypeProbesAs("minRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredIntegerOfScalarTypeProbesAs("maxRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgRequiredIntegerOfScalarTypeProbesAs("avgRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationRequiredIntegerOfScalarTypeProbesAs("standardDeviationRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationRequiredIntegerOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceRequiredIntegerOfScalarTypeProbesAs("sampleVarianceRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceRequiredIntegerOfScalarTypeProbesAs("samplePopulationVarianceRequiredIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("required_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceRequiredIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumOptionalIntegerOfScalarTypeProbesAs("sumOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalIntegerOfScalarTypeProbesAs("minOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalIntegerOfScalarTypeProbesAs("maxOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgOptionalIntegerOfScalarTypeProbesAs("avgOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationOptionalIntegerOfScalarTypeProbesAs("standardDeviationOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationOptionalIntegerOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceOptionalIntegerOfScalarTypeProbesAs("sampleVarianceOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalIntegerOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceOptionalIntegerOfScalarTypeProbesAs("samplePopulationVarianceOptionalIntegerOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalIntegerOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("optional_integer", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceOptionalIntegerOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumRequiredLongOfScalarTypeProbesAs("sumRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredLongOfScalarTypeProbesAs("minRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredLongOfScalarTypeProbesAs("maxRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgRequiredLongOfScalarTypeProbesAs("avgRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationRequiredLongOfScalarTypeProbesAs("standardDeviationRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationRequiredLongOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceRequiredLongOfScalarTypeProbesAs("sampleVarianceRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceRequiredLongOfScalarTypeProbesAs("samplePopulationVarianceRequiredLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("required_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceRequiredLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumOptionalLongOfScalarTypeProbesAs("sumOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalLongOfScalarTypeProbesAs("minOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalLongOfScalarTypeProbesAs("maxOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgOptionalLongOfScalarTypeProbesAs("avgOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationOptionalLongOfScalarTypeProbesAs("standardDeviationOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationOptionalLongOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceOptionalLongOfScalarTypeProbesAs("sampleVarianceOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalLongOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceOptionalLongOfScalarTypeProbesAs("samplePopulationVarianceOptionalLongOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalLongOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("optional_long", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceOptionalLongOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumRequiredDecimalOfScalarTypeProbesAs("sumRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredDecimalOfScalarTypeProbesAs("minRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredDecimalOfScalarTypeProbesAs("maxRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgRequiredDecimalOfScalarTypeProbesAs("avgRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationRequiredDecimalOfScalarTypeProbesAs("standardDeviationRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationRequiredDecimalOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceRequiredDecimalOfScalarTypeProbesAs("sampleVarianceRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceRequiredDecimalOfScalarTypeProbesAs("samplePopulationVarianceRequiredDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("required_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceRequiredDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumOptionalDecimalOfScalarTypeProbesAs("sumOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalDecimalOfScalarTypeProbesAs("minOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalDecimalOfScalarTypeProbesAs("maxOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgOptionalDecimalOfScalarTypeProbesAs("avgOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationOptionalDecimalOfScalarTypeProbesAs("standardDeviationOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationOptionalDecimalOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceOptionalDecimalOfScalarTypeProbesAs("sampleVarianceOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalDecimalOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceOptionalDecimalOfScalarTypeProbesAs("samplePopulationVarianceOptionalDecimalOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalDecimalOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("optional_decimal", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceOptionalDecimalOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumRequiredFloatOfScalarTypeProbesAs("sumRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredFloatOfScalarTypeProbesAs("minRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredFloatOfScalarTypeProbesAs("maxRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgRequiredFloatOfScalarTypeProbesAs("avgRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationRequiredFloatOfScalarTypeProbesAs("standardDeviationRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationRequiredFloatOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceRequiredFloatOfScalarTypeProbesAs("sampleVarianceRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceRequiredFloatOfScalarTypeProbesAs("samplePopulationVarianceRequiredFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("required_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceRequiredFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumOptionalFloatOfScalarTypeProbesAs("sumOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalFloatOfScalarTypeProbesAs("minOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalFloatOfScalarTypeProbesAs("maxOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgOptionalFloatOfScalarTypeProbesAs("avgOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationOptionalFloatOfScalarTypeProbesAs("standardDeviationOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationOptionalFloatOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceOptionalFloatOfScalarTypeProbesAs("sampleVarianceOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalFloatOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceOptionalFloatOfScalarTypeProbesAs("samplePopulationVarianceOptionalFloatOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalFloatOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("optional_float", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceOptionalFloatOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumRequiredDoubleOfScalarTypeProbesAs("sumRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredDoubleOfScalarTypeProbesAs("minRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredDoubleOfScalarTypeProbesAs("maxRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgRequiredDoubleOfScalarTypeProbesAs("avgRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationRequiredDoubleOfScalarTypeProbesAs("standardDeviationRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationRequiredDoubleOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceRequiredDoubleOfScalarTypeProbesAs("sampleVarianceRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceRequiredDoubleOfScalarTypeProbesAs("samplePopulationVarianceRequiredDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceRequiredDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("required_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceRequiredDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SumOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SumOptionalDoubleOfScalarTypeProbesAs("sumOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SumOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalDoubleOfScalarTypeProbesAs("minOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalDoubleOfScalarTypeProbesAs("maxOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) AvgOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.AvgOptionalDoubleOfScalarTypeProbesAs("avgOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) AvgOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) StandardDeviationOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.StandardDeviationOptionalDoubleOfScalarTypeProbesAs("standardDeviationOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) StandardDeviationOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationOptionalDoubleOfScalarTypeProbesAs("squareRootOfPopulationStandardDeviationOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SquareRootOfPopulationStandardDeviationOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SampleVarianceOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SampleVarianceOptionalDoubleOfScalarTypeProbesAs("sampleVarianceOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SampleVarianceOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalDoubleOfScalarTypeProbes() *TypePlatformRequest {
	return r.SamplePopulationVarianceOptionalDoubleOfScalarTypeProbesAs("samplePopulationVarianceOptionalDoubleOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) SamplePopulationVarianceOptionalDoubleOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("optional_double", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceOptionalDoubleOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredDateOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredDateOfScalarTypeProbesAs("minRequiredDateOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredDateOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_date", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredDateOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredDateOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredDateOfScalarTypeProbesAs("maxRequiredDateOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredDateOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_date", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredDateOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalDateOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalDateOfScalarTypeProbesAs("minOptionalDateOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalDateOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_date", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalDateOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalDateOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalDateOfScalarTypeProbesAs("maxOptionalDateOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalDateOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_date", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalDateOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinRequiredDatetimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinRequiredDatetimeOfScalarTypeProbesAs("minRequiredDatetimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinRequiredDatetimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("required_datetime", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinRequiredDatetimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxRequiredDatetimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxRequiredDatetimeOfScalarTypeProbesAs("maxRequiredDatetimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxRequiredDatetimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("required_datetime", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxRequiredDatetimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinOptionalDatetimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinOptionalDatetimeOfScalarTypeProbesAs("minOptionalDatetimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinOptionalDatetimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("optional_datetime", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinOptionalDatetimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxOptionalDatetimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxOptionalDatetimeOfScalarTypeProbesAs("maxOptionalDatetimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxOptionalDatetimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("optional_datetime", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxOptionalDatetimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinCreateTimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinCreateTimeOfScalarTypeProbesAs("minCreateTimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinCreateTimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxCreateTimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxCreateTimeOfScalarTypeProbesAs("maxCreateTimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxCreateTimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MinUpdateTimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MinUpdateTimeOfScalarTypeProbesAs("minUpdateTimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MinUpdateTimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *TypePlatformRequest) MaxUpdateTimeOfScalarTypeProbes() *TypePlatformRequest {
	return r.MaxUpdateTimeOfScalarTypeProbesAs("maxUpdateTimeOfScalarTypeProbes", scalar_type_probe.NewScalarTypeProbeRequest())
}
func (r *TypePlatformRequest) MaxUpdateTimeOfScalarTypeProbesAs(alias string, child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("scalarTypeProbeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfScalarTypeProbesAs", Alias: &alias, Relation: "scalarTypeProbeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *TypePlatformRequest) SelectScalarTypeProbeList() *TypePlatformRequest {
	return r.SelectScalarTypeProbeListWith(scalar_type_probe.NewScalarTypeProbeRequest())
}

func (r *TypePlatformRequest) SelectScalarTypeProbeListWith(child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	r.Query.RelationQuerySelection("scalarTypeProbeList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectScalarTypeProbeListWith", Relation: "scalarTypeProbeList"}
	return r
}

func (r *TypePlatformRequest) HaveScalarTypeProbes() *TypePlatformRequest {
	return r.WithScalarTypeProbeListMatching(scalar_type_probe.NewScalarTypeProbeRequest())
}

func (r *TypePlatformRequest) HaveNoScalarTypeProbes() *TypePlatformRequest {
	return r.WithoutScalarTypeProbeListMatching(scalar_type_probe.NewScalarTypeProbeRequest())
}

func (r *TypePlatformRequest) WithScalarTypeProbeListMatching(child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "type_platform_id"))
	return r
}

func (r *TypePlatformRequest) WithoutScalarTypeProbeListMatching(child *scalar_type_probe.ScalarTypeProbeRequest) *TypePlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "type_platform_id"))
	return r
}

func (e *ExecutableTypePlatformRequest) NewEntity(context *runtime.UserContext) *TypePlatform {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewTypePlatform()
	initialized := context.InitializeEntity("TypePlatform", entity)
	typed, ok := initialized.(*TypePlatform)
	if !ok {
		panic("entity initializer changed TypePlatform to an incompatible type")
	}
	return typed
}

func (e *ExecutableTypePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*TypePlatform, error) {
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

func (e *ExecutableTypePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*TypePlatform], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*TypePlatform, error) {
		entity := newLoadedTypePlatform()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["scalarTypeProbeList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation scalarTypeProbeList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := scalar_type_probe.NewScalarTypeProbe()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.ScalarTypeProbeList().Add(childEntity)
				}}
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableTypePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*TypePlatform) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*TypePlatform, error) {
		entity := newLoadedTypePlatform()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableTypePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableTypePlatformRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableTypePlatformRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *TypePlatformRequest) Count() *TypePlatformRequest {
	return r.CountAs("count")
}

func (r *TypePlatformRequest) CountAs(alias string) *TypePlatformRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *TypePlatformRequest) GroupById() *TypePlatformRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *TypePlatformRequest) GroupByName() *TypePlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *TypePlatformRequest) GroupByVersion() *TypePlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}

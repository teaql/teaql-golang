

package platform

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"freight-quote-settlement-core-workspace/lib/internal/generatedsearch"
	"freight-quote-settlement-core-workspace/lib/freight_quote"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type PlatformRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutablePlatformRequest struct {
	request *PlatformRequest
}

func NewPlatformRequest() *PlatformRequest {
	r := &PlatformRequest{
		Query: core.NewSelectQuery("Dispatch Platform"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewPlatformMinimalRequest() *PlatformRequest {
	r := NewPlatformRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *PlatformRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *PlatformRequest) FilterByJson(source []byte) (*PlatformRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *PlatformRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *PlatformRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *PlatformRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedPlatform().EntityDescriptor()
}

func (r *PlatformRequest) NewRelationEntity() core.Entity {
	return newLoadedPlatform()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *PlatformRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *PlatformRequest) teaqlHydrateQueryRecord(rec core.Record) (*Platform, error) {
	entity := newLoadedPlatform()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["freightQuoteList"]; selected {
		entity.markRelationLoaded("freightQuoteList")
		factory := r.relationFactories["freightQuoteList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation freightQuoteList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation freightQuoteList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*freight_quote.FreightQuote)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation freightQuoteList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.FreightQuoteList().Add(childEntity)
			}}
	return entity, nil
}

func (r *PlatformRequest) Comment(comment string) *PlatformRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *PlatformRequest) Purpose(purpose string) *ExecutablePlatformRequest {
	r.purposeText = purpose
	return &ExecutablePlatformRequest{request: r}
}

func (r *ExecutablePlatformRequest) Comment(comment string) *ExecutablePlatformRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *PlatformRequest) Limit(limit uint64) *PlatformRequest {
	r.Query.Limit(limit)
	return r
}

func (r *PlatformRequest) Offset(offset uint64) *PlatformRequest {
	r.Query.Offset(offset)
	return r
}

func (r *PlatformRequest) OptimizeForContinuousPageFetch() *PlatformRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *PlatformRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *PlatformRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *PlatformRequest) OptimizePaginationWithIDSet() *PlatformRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *PlatformRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *PlatformRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *PlatformRequest) TopNProbeParentThreshold(threshold uint64) *PlatformRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removePlatformVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removePlatformVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *PlatformRequest) WithDeletedRows() *PlatformRequest {
	r.Query.Filter = removePlatformVersionFilter(r.Query.Filter)
	return r
}

func (r *PlatformRequest) DeletedRowsOnly() *PlatformRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *PlatformRequest) SelectId() *PlatformRequest {
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *PlatformRequest) WithIdIs(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIsNot(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithIdNotIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithIdGreaterThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdBetween(lower uint64, upper uint64) *PlatformRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *PlatformRequest) WithIdIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *PlatformRequest) WithIdIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *PlatformRequest) OrderByIdAsc() *PlatformRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *PlatformRequest) OrderByIdDesc() *PlatformRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *PlatformRequest) SelectName() *PlatformRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *PlatformRequest) WithNameIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithNameNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithNameGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *PlatformRequest) WithNameIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *PlatformRequest) WithNameIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *PlatformRequest) WithNameContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *PlatformRequest) WithNameNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *PlatformRequest) WithNameStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *PlatformRequest) WithNameNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *PlatformRequest) WithNameEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *PlatformRequest) WithNameNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *PlatformRequest) WithNameSoundingLike(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *PlatformRequest) OrderByNameAsc() *PlatformRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *PlatformRequest) OrderByNameDesc() *PlatformRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *PlatformRequest) SelectVersion() *PlatformRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *PlatformRequest) WithVersionIs(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIsNot(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithVersionNotIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionBetween(lower int64, upper int64) *PlatformRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *PlatformRequest) WithVersionIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *PlatformRequest) WithVersionIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *PlatformRequest) OrderByVersionAsc() *PlatformRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *PlatformRequest) OrderByVersionDesc() *PlatformRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}



func (r *PlatformRequest) CountFreightQuotes() *PlatformRequest {
	return r.CountFreightQuotesAs("countFreightQuotes")

}
func (r *PlatformRequest) CountFreightQuotesAs(alias string) *PlatformRequest {
	r.CountFreightQuotesWith(alias, freight_quote.NewFreightQuoteRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	return r
}
func (r *PlatformRequest) CountFreightQuotesWith(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountFreightQuotesWith", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PlatformRequest) SumAmountOfFreightQuotes() *PlatformRequest {
	return r.SumAmountOfFreightQuotesAs("sumAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SumAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinAmountOfFreightQuotes() *PlatformRequest {
	return r.MinAmountOfFreightQuotesAs("minAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MinAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxAmountOfFreightQuotes() *PlatformRequest {
	return r.MaxAmountOfFreightQuotesAs("maxAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MaxAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) AvgAmountOfFreightQuotes() *PlatformRequest {
	return r.AvgAmountOfFreightQuotesAs("avgAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) AvgAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) StandardDeviationAmountOfFreightQuotes() *PlatformRequest {
	return r.StandardDeviationAmountOfFreightQuotesAs("standardDeviationAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) StandardDeviationAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationAmountOfFreightQuotes() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationAmountOfFreightQuotesAs("squareRootOfPopulationStandardDeviationAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SampleVarianceAmountOfFreightQuotes() *PlatformRequest {
	return r.SampleVarianceAmountOfFreightQuotesAs("sampleVarianceAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SampleVarianceAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceAmountOfFreightQuotes() *PlatformRequest {
	return r.SamplePopulationVarianceAmountOfFreightQuotesAs("samplePopulationVarianceAmountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceAmountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("amount", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceAmountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SumShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.SumShipmentNumberOfFreightQuotesAs("sumShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SumShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.MinShipmentNumberOfFreightQuotesAs("minShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MinShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.MaxShipmentNumberOfFreightQuotesAs("maxShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MaxShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) AvgShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.AvgShipmentNumberOfFreightQuotesAs("avgShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) AvgShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) StandardDeviationShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.StandardDeviationShipmentNumberOfFreightQuotesAs("standardDeviationShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) StandardDeviationShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationShipmentNumberOfFreightQuotesAs("squareRootOfPopulationStandardDeviationShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SampleVarianceShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.SampleVarianceShipmentNumberOfFreightQuotesAs("sampleVarianceShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SampleVarianceShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceShipmentNumberOfFreightQuotes() *PlatformRequest {
	return r.SamplePopulationVarianceShipmentNumberOfFreightQuotesAs("samplePopulationVarianceShipmentNumberOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceShipmentNumberOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("shipment_number", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceShipmentNumberOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SumPackageCountOfFreightQuotes() *PlatformRequest {
	return r.SumPackageCountOfFreightQuotesAs("sumPackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SumPackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumPackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinPackageCountOfFreightQuotes() *PlatformRequest {
	return r.MinPackageCountOfFreightQuotesAs("minPackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MinPackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinPackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxPackageCountOfFreightQuotes() *PlatformRequest {
	return r.MaxPackageCountOfFreightQuotesAs("maxPackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MaxPackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxPackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) AvgPackageCountOfFreightQuotes() *PlatformRequest {
	return r.AvgPackageCountOfFreightQuotesAs("avgPackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) AvgPackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgPackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) StandardDeviationPackageCountOfFreightQuotes() *PlatformRequest {
	return r.StandardDeviationPackageCountOfFreightQuotesAs("standardDeviationPackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) StandardDeviationPackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationPackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationPackageCountOfFreightQuotes() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationPackageCountOfFreightQuotesAs("squareRootOfPopulationStandardDeviationPackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationPackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationPackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SampleVariancePackageCountOfFreightQuotes() *PlatformRequest {
	return r.SampleVariancePackageCountOfFreightQuotesAs("sampleVariancePackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SampleVariancePackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVariancePackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SamplePopulationVariancePackageCountOfFreightQuotes() *PlatformRequest {
	return r.SamplePopulationVariancePackageCountOfFreightQuotesAs("samplePopulationVariancePackageCountOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) SamplePopulationVariancePackageCountOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("package_count", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVariancePackageCountOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinDeliveryDateOfFreightQuotes() *PlatformRequest {
	return r.MinDeliveryDateOfFreightQuotesAs("minDeliveryDateOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MinDeliveryDateOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("delivery_date", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinDeliveryDateOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxDeliveryDateOfFreightQuotes() *PlatformRequest {
	return r.MaxDeliveryDateOfFreightQuotesAs("maxDeliveryDateOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MaxDeliveryDateOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("delivery_date", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxDeliveryDateOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinQuotedAtOfFreightQuotes() *PlatformRequest {
	return r.MinQuotedAtOfFreightQuotesAs("minQuotedAtOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MinQuotedAtOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("quoted_at", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinQuotedAtOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxQuotedAtOfFreightQuotes() *PlatformRequest {
	return r.MaxQuotedAtOfFreightQuotesAs("maxQuotedAtOfFreightQuotes", freight_quote.NewFreightQuoteRequest())
}
func (r *PlatformRequest) MaxQuotedAtOfFreightQuotesAs(alias string, child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("quoted_at", alias)
	aggregate := core.NewRelationAggregate("freightQuoteList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxQuotedAtOfFreightQuotesAs", Alias: &alias, Relation: "freightQuoteList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PlatformRequest) SelectFreightQuoteList() *PlatformRequest {
	return r.SelectFreightQuoteListWith(freight_quote.NewFreightQuoteRequest())
}

func (r *PlatformRequest) SelectFreightQuoteListWith(child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	r.Query.RelationQuerySelection("freightQuoteList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectFreightQuoteListWith", Relation: "freightQuoteList"}
	r.relationFactories["freightQuoteList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *PlatformRequest) HaveFreightQuotes() *PlatformRequest {
	return r.WithFreightQuoteListMatching(freight_quote.NewFreightQuoteRequest())
}

func (r *PlatformRequest) HaveNoFreightQuotes() *PlatformRequest {
	return r.WithoutFreightQuoteListMatching(freight_quote.NewFreightQuoteRequest())
}

func (r *PlatformRequest) WithFreightQuoteListMatching(child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutFreightQuoteListMatching(child *freight_quote.FreightQuoteRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (e *ExecutablePlatformRequest) NewEntity(context *runtime.UserContext) (*Platform, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewPlatform(context)
}

func (e *ExecutablePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*Platform, error) {
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

func (e *ExecutablePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Platform], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*Platform, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutablePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Platform) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*Platform, error) {
		entity := newLoadedPlatform()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutablePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePlatformRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutablePlatformRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *PlatformRequest) Count() *PlatformRequest {
	return r.CountAs("count")
}

func (r *PlatformRequest) CountAs(alias string) *PlatformRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *PlatformRequest) GroupById() *PlatformRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *PlatformRequest) GroupByIdAs(alias string) *PlatformRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *PlatformRequest) GroupByName() *PlatformRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *PlatformRequest) GroupByNameAs(alias string) *PlatformRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *PlatformRequest) GroupByVersion() *PlatformRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *PlatformRequest) GroupByVersionAs(alias string) *PlatformRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}

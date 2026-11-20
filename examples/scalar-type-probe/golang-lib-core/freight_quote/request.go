

package freight_quote

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"freight-quote-settlement-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type FreightQuoteRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableFreightQuoteRequest struct {
	request *FreightQuoteRequest
}

func NewFreightQuoteRequest() *FreightQuoteRequest {
	r := &FreightQuoteRequest{
		Query: core.NewSelectQuery("Freight Quote"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewFreightQuoteMinimalRequest() *FreightQuoteRequest {
	r := NewFreightQuoteRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *FreightQuoteRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *FreightQuoteRequest) FilterByJson(source []byte) (*FreightQuoteRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *FreightQuoteRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *FreightQuoteRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *FreightQuoteRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedFreightQuote().EntityDescriptor()
}

func (r *FreightQuoteRequest) NewRelationEntity() core.Entity {
	return newLoadedFreightQuote()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *FreightQuoteRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *FreightQuoteRequest) teaqlHydrateQueryRecord(rec core.Record) (*FreightQuote, error) {
	entity := newLoadedFreightQuote()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["platformEntity"]; selected {
		entity.markRelationLoaded("platformEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["platformEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("platformEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *FreightQuoteRequest) Comment(comment string) *FreightQuoteRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *FreightQuoteRequest) Purpose(purpose string) *ExecutableFreightQuoteRequest {
	r.purposeText = purpose
	return &ExecutableFreightQuoteRequest{request: r}
}

func (r *ExecutableFreightQuoteRequest) Comment(comment string) *ExecutableFreightQuoteRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *FreightQuoteRequest) Limit(limit uint64) *FreightQuoteRequest {
	r.Query.Limit(limit)
	return r
}

func (r *FreightQuoteRequest) Offset(offset uint64) *FreightQuoteRequest {
	r.Query.Offset(offset)
	return r
}

func (r *FreightQuoteRequest) OptimizeForContinuousPageFetch() *FreightQuoteRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *FreightQuoteRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *FreightQuoteRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *FreightQuoteRequest) OptimizePaginationWithIDSet() *FreightQuoteRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *FreightQuoteRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *FreightQuoteRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *FreightQuoteRequest) TopNProbeParentThreshold(threshold uint64) *FreightQuoteRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeFreightQuoteVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeFreightQuoteVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *FreightQuoteRequest) WithDeletedRows() *FreightQuoteRequest {
	r.Query.Filter = removeFreightQuoteVersionFilter(r.Query.Filter)
	return r
}

func (r *FreightQuoteRequest) DeletedRowsOnly() *FreightQuoteRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *FreightQuoteRequest) SelectId() *FreightQuoteRequest {
	r.Query.Project("id")
	return r
}

func (r *FreightQuoteRequest) WithIdIs(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdIsNot(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *FreightQuoteRequest) WithIdNotIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *FreightQuoteRequest) WithIdGreaterThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdGreaterThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdLessThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdLessThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithIdBetween(lower uint64, upper uint64) *FreightQuoteRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *FreightQuoteRequest) WithIdIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *FreightQuoteRequest) WithIdIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *FreightQuoteRequest) OrderByIdAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *FreightQuoteRequest) OrderByIdDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *FreightQuoteRequest) SelectPlatform() *FreightQuoteRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *FreightQuoteRequest) WithPlatformIs(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIsNot(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *FreightQuoteRequest) WithPlatformNotIn(values []uint64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *FreightQuoteRequest) WithPlatformGreaterThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformLessThan(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformLessThanOrEqualTo(value uint64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPlatformBetween(lower uint64, upper uint64) *FreightQuoteRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *FreightQuoteRequest) WithPlatformIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *FreightQuoteRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *FreightQuoteRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *FreightQuoteRequest) OrderByPlatformAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *FreightQuoteRequest) OrderByPlatformDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *FreightQuoteRequest) SelectReference() *FreightQuoteRequest {
	r.Query.Project("reference")
	return r
}

func (r *FreightQuoteRequest) WithReferenceIs(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("reference", core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIsNot(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("reference", core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIn(values []string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("reference", converted))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotIn(values []string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("reference", converted))
	return r
}
func (r *FreightQuoteRequest) WithReferenceGreaterThan(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("reference", core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceGreaterThanOrEqualTo(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("reference", core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceLessThan(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("reference", core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceLessThanOrEqualTo(value string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("reference", core.ValText(value)))
	return r
}
func (r *FreightQuoteRequest) WithReferenceBetween(lower string, upper string) *FreightQuoteRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("reference", from, to))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("reference"))
	return r
}
func (r *FreightQuoteRequest) WithReferenceIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("reference"))
	return r
}
func (r *FreightQuoteRequest) WithReferenceContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprContain("reference", term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotContain("reference", term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprBeginWith("reference", term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("reference", term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEndWith("reference", term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceNotEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotEndWith("reference", term))
	return r
}
func (r *FreightQuoteRequest) WithReferenceSoundingLike(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprSoundLike("reference", core.ValText(term)))
	return r
}
func (r *FreightQuoteRequest) OrderByReferenceAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("reference")
	return r
}
func (r *FreightQuoteRequest) OrderByReferenceDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("reference")
	return r
}
func (r *FreightQuoteRequest) SelectAmount() *FreightQuoteRequest {
	r.Query.Project("amount")
	return r
}

func (r *FreightQuoteRequest) WithAmountIs(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("amount", core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountIsNot(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("amount", core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountIn(values []decimal.Decimal) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("amount", converted))
	return r
}
func (r *FreightQuoteRequest) WithAmountNotIn(values []decimal.Decimal) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("amount", converted))
	return r
}
func (r *FreightQuoteRequest) WithAmountGreaterThan(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("amount", core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountGreaterThanOrEqualTo(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("amount", core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountLessThan(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("amount", core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountLessThanOrEqualTo(value decimal.Decimal) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("amount", core.ValDecimal(value)))
	return r
}
func (r *FreightQuoteRequest) WithAmountBetween(lower decimal.Decimal, upper decimal.Decimal) *FreightQuoteRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode("amount", from, to))
	return r
}
func (r *FreightQuoteRequest) WithAmountIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("amount"))
	return r
}
func (r *FreightQuoteRequest) WithAmountIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("amount"))
	return r
}
func (r *FreightQuoteRequest) OrderByAmountAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("amount")
	return r
}
func (r *FreightQuoteRequest) OrderByAmountDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("amount")
	return r
}
func (r *FreightQuoteRequest) SelectShipmentNumber() *FreightQuoteRequest {
	r.Query.Project("shipment_number")
	return r
}

func (r *FreightQuoteRequest) WithShipmentNumberIs(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("shipment_number", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIsNot(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("shipment_number", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("shipment_number", converted))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberNotIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("shipment_number", converted))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberGreaterThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("shipment_number", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberGreaterThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("shipment_number", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberLessThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("shipment_number", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberLessThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("shipment_number", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberBetween(lower int64, upper int64) *FreightQuoteRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("shipment_number", from, to))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("shipment_number"))
	return r
}
func (r *FreightQuoteRequest) WithShipmentNumberIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("shipment_number"))
	return r
}
func (r *FreightQuoteRequest) OrderByShipmentNumberAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("shipment_number")
	return r
}
func (r *FreightQuoteRequest) OrderByShipmentNumberDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("shipment_number")
	return r
}
func (r *FreightQuoteRequest) SelectPackageCount() *FreightQuoteRequest {
	r.Query.Project("package_count")
	return r
}

func (r *FreightQuoteRequest) WithPackageCountIs(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("package_count", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIsNot(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("package_count", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("package_count", converted))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountNotIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("package_count", converted))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountGreaterThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("package_count", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountGreaterThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("package_count", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountLessThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("package_count", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountLessThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("package_count", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountBetween(lower int64, upper int64) *FreightQuoteRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("package_count", from, to))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("package_count"))
	return r
}
func (r *FreightQuoteRequest) WithPackageCountIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("package_count"))
	return r
}
func (r *FreightQuoteRequest) OrderByPackageCountAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("package_count")
	return r
}
func (r *FreightQuoteRequest) OrderByPackageCountDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("package_count")
	return r
}
func (r *FreightQuoteRequest) SelectDeliveryDate() *FreightQuoteRequest {
	r.Query.Project("delivery_date")
	return r
}

func (r *FreightQuoteRequest) WithDeliveryDateIs(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("delivery_date", core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIsNot(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("delivery_date", core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList("delivery_date", converted))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateNotIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList("delivery_date", converted))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateGreaterThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("delivery_date", core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateGreaterThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("delivery_date", core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateLessThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("delivery_date", core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateLessThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("delivery_date", core.ValDate(value)))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateBetween(lower time.Time, upper time.Time) *FreightQuoteRequest {
	value := lower
	from := core.ValDate(value)
	value = upper
	to := core.ValDate(value)
	r.Query.AndFilter(core.ExprBetweenNode("delivery_date", from, to))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("delivery_date"))
	return r
}
func (r *FreightQuoteRequest) WithDeliveryDateIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("delivery_date"))
	return r
}
func (r *FreightQuoteRequest) OrderByDeliveryDateAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("delivery_date")
	return r
}
func (r *FreightQuoteRequest) OrderByDeliveryDateDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("delivery_date")
	return r
}
func (r *FreightQuoteRequest) SelectPickupTime() *FreightQuoteRequest {
	r.Query.Project("pickup_time")
	return r
}

func (r *FreightQuoteRequest) WithPickupTimeIs(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("pickup_time", core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIsNot(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("pickup_time", core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIn(values []core.TimeOfDay) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimeOfDay(value))
	}
	r.Query.AndFilter(core.ExprInList("pickup_time", converted))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeNotIn(values []core.TimeOfDay) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimeOfDay(value))
	}
	r.Query.AndFilter(core.ExprNotInList("pickup_time", converted))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeGreaterThan(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("pickup_time", core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeGreaterThanOrEqualTo(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("pickup_time", core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeLessThan(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("pickup_time", core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeLessThanOrEqualTo(value core.TimeOfDay) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("pickup_time", core.ValTimeOfDay(value)))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeBetween(lower core.TimeOfDay, upper core.TimeOfDay) *FreightQuoteRequest {
	value := lower
	from := core.ValTimeOfDay(value)
	value = upper
	to := core.ValTimeOfDay(value)
	r.Query.AndFilter(core.ExprBetweenNode("pickup_time", from, to))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("pickup_time"))
	return r
}
func (r *FreightQuoteRequest) WithPickupTimeIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("pickup_time"))
	return r
}
func (r *FreightQuoteRequest) OrderByPickupTimeAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("pickup_time")
	return r
}
func (r *FreightQuoteRequest) OrderByPickupTimeDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("pickup_time")
	return r
}
func (r *FreightQuoteRequest) SelectQuotedAt() *FreightQuoteRequest {
	r.Query.Project("quoted_at")
	return r
}

func (r *FreightQuoteRequest) WithQuotedAtIs(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("quoted_at", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIsNot(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("quoted_at", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("quoted_at", converted))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtNotIn(values []time.Time) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("quoted_at", converted))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtGreaterThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("quoted_at", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtGreaterThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("quoted_at", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtLessThan(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("quoted_at", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtLessThanOrEqualTo(value time.Time) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("quoted_at", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtBetween(lower time.Time, upper time.Time) *FreightQuoteRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("quoted_at", from, to))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("quoted_at"))
	return r
}
func (r *FreightQuoteRequest) WithQuotedAtIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("quoted_at"))
	return r
}
func (r *FreightQuoteRequest) OrderByQuotedAtAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("quoted_at")
	return r
}
func (r *FreightQuoteRequest) OrderByQuotedAtDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("quoted_at")
	return r
}
func (r *FreightQuoteRequest) SelectInstructions() *FreightQuoteRequest {
	r.Query.Project("instructions")
	return r
}

func (r *FreightQuoteRequest) WithInstructionsIs(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("instructions", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIsNot(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("instructions", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIn(values []*string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("instructions", converted))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotIn(values []*string) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("instructions", converted))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsGreaterThan(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("instructions", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsGreaterThanOrEqualTo(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("instructions", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsLessThan(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("instructions", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsLessThanOrEqualTo(value *string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("instructions", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsBetween(lower *string, upper *string) *FreightQuoteRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("instructions", from, to))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("instructions"))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("instructions"))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprContain("instructions", term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotContaining(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotContain("instructions", term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprBeginWith("instructions", term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotStartingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("instructions", term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEndWith("instructions", term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsNotEndingWith(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotEndWith("instructions", term))
	return r
}
func (r *FreightQuoteRequest) WithInstructionsSoundingLike(term string) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprSoundLike("instructions", core.ValText(term)))
	return r
}
func (r *FreightQuoteRequest) OrderByInstructionsAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("instructions")
	return r
}
func (r *FreightQuoteRequest) OrderByInstructionsDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("instructions")
	return r
}
func (r *FreightQuoteRequest) SelectVersion() *FreightQuoteRequest {
	r.Query.Project("version")
	return r
}

func (r *FreightQuoteRequest) WithVersionIs(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionIsNot(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *FreightQuoteRequest) WithVersionNotIn(values []int64) *FreightQuoteRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *FreightQuoteRequest) WithVersionGreaterThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionGreaterThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionLessThan(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionLessThanOrEqualTo(value int64) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *FreightQuoteRequest) WithVersionBetween(lower int64, upper int64) *FreightQuoteRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *FreightQuoteRequest) WithVersionIsKnown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *FreightQuoteRequest) WithVersionIsUnknown() *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *FreightQuoteRequest) OrderByVersionAsc() *FreightQuoteRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *FreightQuoteRequest) OrderByVersionDesc() *FreightQuoteRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *FreightQuoteRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *FreightQuoteRequest {
	runtime.EnsureRelationProjection(r.Query, "platform_id")
	r.Query.RelationQuerySelection("platformEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPlatformWith", Relation: "platformEntity"}
	r.relationFactories["platformEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *FreightQuoteRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *FreightQuoteRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *FreightQuoteRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableFreightQuoteRequest) NewEntity(context *runtime.UserContext) (*FreightQuote, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewFreightQuote(context)
}

func (e *ExecutableFreightQuoteRequest) ExecuteForOne(context *runtime.UserContext) (*FreightQuote, error) {
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

func (e *ExecutableFreightQuoteRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*FreightQuote], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*FreightQuote, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableFreightQuoteRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*FreightQuote) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*FreightQuote, error) {
		entity := newLoadedFreightQuote()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableFreightQuoteRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableFreightQuoteRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableFreightQuoteRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *FreightQuoteRequest) Count() *FreightQuoteRequest {
	return r.CountAs("count")
}

func (r *FreightQuoteRequest) CountAs(alias string) *FreightQuoteRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *FreightQuoteRequest) MinAmount() *FreightQuoteRequest {
	return r.MinAmountAs("minOfAmount")
}

func (r *FreightQuoteRequest) MinAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Min("amount", alias)
	return r
}
func (r *FreightQuoteRequest) MaxAmount() *FreightQuoteRequest {
	return r.MaxAmountAs("maxOfAmount")
}

func (r *FreightQuoteRequest) MaxAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Max("amount", alias)
	return r
}
func (r *FreightQuoteRequest) SumAmount() *FreightQuoteRequest {
	return r.SumAmountAs("sumOfAmount")
}

func (r *FreightQuoteRequest) SumAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Sum("amount", alias)
	return r
}
func (r *FreightQuoteRequest) AvgAmount() *FreightQuoteRequest {
	return r.AvgAmountAs("avgOfAmount")
}

func (r *FreightQuoteRequest) AvgAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Avg("amount", alias)
	return r
}
func (r *FreightQuoteRequest) StddevAmount() *FreightQuoteRequest {
	return r.StddevAmountAs("standardDeviationOfAmount")
}

func (r *FreightQuoteRequest) StddevAmountAs(alias string) *FreightQuoteRequest {
	r.Query.Stddev("amount", alias)
	return r
}
func (r *FreightQuoteRequest) StddevPopAmount() *FreightQuoteRequest {
	return r.StddevPopAmountAs("squareRootOfPopulationStandardDeviationOfAmount")
}

func (r *FreightQuoteRequest) StddevPopAmountAs(alias string) *FreightQuoteRequest {
	r.Query.StddevPop("amount", alias)
	return r
}
func (r *FreightQuoteRequest) VarSampAmount() *FreightQuoteRequest {
	return r.VarSampAmountAs("sampleVarianceOfAmount")
}

func (r *FreightQuoteRequest) VarSampAmountAs(alias string) *FreightQuoteRequest {
	r.Query.VarSamp("amount", alias)
	return r
}
func (r *FreightQuoteRequest) VarPopAmount() *FreightQuoteRequest {
	return r.VarPopAmountAs("samplePopulationVarianceOfAmount")
}

func (r *FreightQuoteRequest) VarPopAmountAs(alias string) *FreightQuoteRequest {
	r.Query.VarPop("amount", alias)
	return r
}
func (r *FreightQuoteRequest) MinShipmentNumber() *FreightQuoteRequest {
	return r.MinShipmentNumberAs("minOfShipmentNumber")
}

func (r *FreightQuoteRequest) MinShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Min("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) MaxShipmentNumber() *FreightQuoteRequest {
	return r.MaxShipmentNumberAs("maxOfShipmentNumber")
}

func (r *FreightQuoteRequest) MaxShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Max("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) SumShipmentNumber() *FreightQuoteRequest {
	return r.SumShipmentNumberAs("sumOfShipmentNumber")
}

func (r *FreightQuoteRequest) SumShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Sum("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) AvgShipmentNumber() *FreightQuoteRequest {
	return r.AvgShipmentNumberAs("avgOfShipmentNumber")
}

func (r *FreightQuoteRequest) AvgShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Avg("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) StddevShipmentNumber() *FreightQuoteRequest {
	return r.StddevShipmentNumberAs("standardDeviationOfShipmentNumber")
}

func (r *FreightQuoteRequest) StddevShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.Stddev("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) StddevPopShipmentNumber() *FreightQuoteRequest {
	return r.StddevPopShipmentNumberAs("squareRootOfPopulationStandardDeviationOfShipmentNumber")
}

func (r *FreightQuoteRequest) StddevPopShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.StddevPop("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) VarSampShipmentNumber() *FreightQuoteRequest {
	return r.VarSampShipmentNumberAs("sampleVarianceOfShipmentNumber")
}

func (r *FreightQuoteRequest) VarSampShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.VarSamp("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) VarPopShipmentNumber() *FreightQuoteRequest {
	return r.VarPopShipmentNumberAs("samplePopulationVarianceOfShipmentNumber")
}

func (r *FreightQuoteRequest) VarPopShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.VarPop("shipment_number", alias)
	return r
}
func (r *FreightQuoteRequest) MinPackageCount() *FreightQuoteRequest {
	return r.MinPackageCountAs("minOfPackageCount")
}

func (r *FreightQuoteRequest) MinPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Min("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) MaxPackageCount() *FreightQuoteRequest {
	return r.MaxPackageCountAs("maxOfPackageCount")
}

func (r *FreightQuoteRequest) MaxPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Max("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) SumPackageCount() *FreightQuoteRequest {
	return r.SumPackageCountAs("sumOfPackageCount")
}

func (r *FreightQuoteRequest) SumPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Sum("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) AvgPackageCount() *FreightQuoteRequest {
	return r.AvgPackageCountAs("avgOfPackageCount")
}

func (r *FreightQuoteRequest) AvgPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Avg("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) StddevPackageCount() *FreightQuoteRequest {
	return r.StddevPackageCountAs("standardDeviationOfPackageCount")
}

func (r *FreightQuoteRequest) StddevPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.Stddev("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) StddevPopPackageCount() *FreightQuoteRequest {
	return r.StddevPopPackageCountAs("squareRootOfPopulationStandardDeviationOfPackageCount")
}

func (r *FreightQuoteRequest) StddevPopPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.StddevPop("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) VarSampPackageCount() *FreightQuoteRequest {
	return r.VarSampPackageCountAs("sampleVarianceOfPackageCount")
}

func (r *FreightQuoteRequest) VarSampPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.VarSamp("package_count", alias)
	return r
}
func (r *FreightQuoteRequest) VarPopPackageCount() *FreightQuoteRequest {
	return r.VarPopPackageCountAs("samplePopulationVarianceOfPackageCount")
}

func (r *FreightQuoteRequest) VarPopPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.VarPop("package_count", alias)
	return r
}

func (r *FreightQuoteRequest) GroupById() *FreightQuoteRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *FreightQuoteRequest) GroupByIdAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *FreightQuoteRequest) GroupByPlatform() *FreightQuoteRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}

func (r *FreightQuoteRequest) GroupByPlatformAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("platform_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("platform_id"))
	return r
}
func (r *FreightQuoteRequest) GroupByReference() *FreightQuoteRequest {
	r.Query.WithGroupBy("reference")
	return r
}

func (r *FreightQuoteRequest) GroupByReferenceAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("reference")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("reference"))
	return r
}
func (r *FreightQuoteRequest) GroupByAmount() *FreightQuoteRequest {
	r.Query.WithGroupBy("amount")
	return r
}

func (r *FreightQuoteRequest) GroupByAmountAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("amount")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("amount"))
	return r
}
func (r *FreightQuoteRequest) GroupByShipmentNumber() *FreightQuoteRequest {
	r.Query.WithGroupBy("shipment_number")
	return r
}

func (r *FreightQuoteRequest) GroupByShipmentNumberAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("shipment_number")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("shipment_number"))
	return r
}
func (r *FreightQuoteRequest) GroupByPackageCount() *FreightQuoteRequest {
	r.Query.WithGroupBy("package_count")
	return r
}

func (r *FreightQuoteRequest) GroupByPackageCountAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("package_count")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("package_count"))
	return r
}
func (r *FreightQuoteRequest) GroupByDeliveryDate() *FreightQuoteRequest {
	r.Query.WithGroupBy("delivery_date")
	return r
}

func (r *FreightQuoteRequest) GroupByDeliveryDateAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("delivery_date")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("delivery_date"))
	return r
}
func (r *FreightQuoteRequest) GroupByPickupTime() *FreightQuoteRequest {
	r.Query.WithGroupBy("pickup_time")
	return r
}

func (r *FreightQuoteRequest) GroupByPickupTimeAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("pickup_time")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("pickup_time"))
	return r
}
func (r *FreightQuoteRequest) GroupByQuotedAt() *FreightQuoteRequest {
	r.Query.WithGroupBy("quoted_at")
	return r
}

func (r *FreightQuoteRequest) GroupByQuotedAtAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("quoted_at")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("quoted_at"))
	return r
}
func (r *FreightQuoteRequest) GroupByInstructions() *FreightQuoteRequest {
	r.Query.WithGroupBy("instructions")
	return r
}

func (r *FreightQuoteRequest) GroupByInstructionsAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("instructions")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("instructions"))
	return r
}
func (r *FreightQuoteRequest) GroupByVersion() *FreightQuoteRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *FreightQuoteRequest) GroupByVersionAs(alias string) *FreightQuoteRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}

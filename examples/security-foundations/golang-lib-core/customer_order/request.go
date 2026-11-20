

package customer_order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"merchant-order-privacy-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerOrderRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerOrderRequest struct {
	request *CustomerOrderRequest
}

func NewCustomerOrderRequest() *CustomerOrderRequest {
	r := &CustomerOrderRequest{
		Query: core.NewSelectQuery("Customer Order"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerOrderMinimalRequest() *CustomerOrderRequest {
	r := NewCustomerOrderRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerOrderRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CustomerOrderRequest) FilterByJson(source []byte) (*CustomerOrderRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerOrderRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerOrderRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerOrderRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedCustomerOrder().EntityDescriptor()
}

func (r *CustomerOrderRequest) NewRelationEntity() core.Entity {
	return newLoadedCustomerOrder()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CustomerOrderRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CustomerOrderRequest) teaqlHydrateQueryRecord(rec core.Record) (*CustomerOrder, error) {
	entity := newLoadedCustomerOrder()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["merchantEntity"]; selected {
		entity.markRelationLoaded("merchantEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["merchantEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("merchantEntity", childEntity)
			}
		}
	}
	if relationValue, selected := rec["customerProfileEntity"]; selected {
		entity.markRelationLoaded("customerProfileEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["customerProfileEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("customerProfileEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *CustomerOrderRequest) Comment(comment string) *CustomerOrderRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerOrderRequest) Purpose(purpose string) *ExecutableCustomerOrderRequest {
	r.purposeText = purpose
	return &ExecutableCustomerOrderRequest{request: r}
}

func (r *ExecutableCustomerOrderRequest) Comment(comment string) *ExecutableCustomerOrderRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CustomerOrderRequest) Limit(limit uint64) *CustomerOrderRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerOrderRequest) Offset(offset uint64) *CustomerOrderRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerOrderRequest) OptimizeForContinuousPageFetch() *CustomerOrderRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerOrderRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerOrderRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerOrderRequest) OptimizePaginationWithIDSet() *CustomerOrderRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerOrderRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerOrderRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerOrderRequest) TopNProbeParentThreshold(threshold uint64) *CustomerOrderRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerOrderVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerOrderVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerOrderRequest) WithDeletedRows() *CustomerOrderRequest {
	r.Query.Filter = removeCustomerOrderVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerOrderRequest) DeletedRowsOnly() *CustomerOrderRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerOrderRequest) SelectId() *CustomerOrderRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerOrderRequest) WithIdIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithIdIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerOrderRequest) WithIdIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerOrderRequest) OrderByIdAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerOrderRequest) OrderByIdDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerOrderRequest) SelectMerchant() *CustomerOrderRequest {
	r.Query.Project("merchant_id")
	return r
}

func (r *CustomerOrderRequest) WithMerchantIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithMerchantIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithMerchantIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("merchant_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithMerchantNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("merchant_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithMerchantGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithMerchantGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithMerchantLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithMerchantLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("merchant_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithMerchantBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("merchant_id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithMerchantIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("merchant_id"))
	return r
}
func (r *CustomerOrderRequest) WithMerchantIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("merchant_id"))
	return r
}
func (r *CustomerOrderRequest) FacetByMerchantAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "merchant_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByMerchantAs", Alias: &name, Relation: "merchant"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerOrderRequest) OrderByMerchantAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("merchant_id")
	return r
}
func (r *CustomerOrderRequest) OrderByMerchantDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("merchant_id")
	return r
}
func (r *CustomerOrderRequest) SelectCustomerProfile() *CustomerOrderRequest {
	r.Query.Project("customer_profile_id")
	return r
}

func (r *CustomerOrderRequest) WithCustomerProfileIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("customer_profile_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("customer_profile_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_profile_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_profile_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("customer_profile_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("customer_profile_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("customer_profile_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("customer_profile_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_profile_id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_profile_id"))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_profile_id"))
	return r
}
func (r *CustomerOrderRequest) FacetByCustomerProfileAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "customer_profile_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerProfileAs", Alias: &name, Relation: "customerProfile"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerOrderRequest) OrderByCustomerProfileAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("customer_profile_id")
	return r
}
func (r *CustomerOrderRequest) OrderByCustomerProfileDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("customer_profile_id")
	return r
}
func (r *CustomerOrderRequest) SelectOrderNumber() *CustomerOrderRequest {
	r.Query.Project("order_number")
	return r
}

func (r *CustomerOrderRequest) WithOrderNumberIs(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsNot(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberBetween(lower string, upper string) *CustomerOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("order_number", from, to))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberSoundingLike(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("order_number", core.ValText(term)))
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("order_number")
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("order_number")
	return r
}
func (r *CustomerOrderRequest) SelectFulfillmentStatus() *CustomerOrderRequest {
	r.Query.Project("fulfillment_status")
	return r
}

func (r *CustomerOrderRequest) WithFulfillmentStatusIs(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("fulfillment_status", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusIsNot(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("fulfillment_status", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("fulfillment_status", converted))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusNotIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("fulfillment_status", converted))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusGreaterThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("fulfillment_status", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusGreaterThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("fulfillment_status", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusLessThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("fulfillment_status", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusLessThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("fulfillment_status", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusBetween(lower string, upper string) *CustomerOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("fulfillment_status", from, to))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("fulfillment_status"))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("fulfillment_status"))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprContain("fulfillment_status", term))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusNotContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("fulfillment_status", term))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("fulfillment_status", term))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusNotStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("fulfillment_status", term))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("fulfillment_status", term))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusNotEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("fulfillment_status", term))
	return r
}
func (r *CustomerOrderRequest) WithFulfillmentStatusSoundingLike(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("fulfillment_status", core.ValText(term)))
	return r
}
func (r *CustomerOrderRequest) OrderByFulfillmentStatusAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("fulfillment_status")
	return r
}
func (r *CustomerOrderRequest) OrderByFulfillmentStatusDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("fulfillment_status")
	return r
}
func (r *CustomerOrderRequest) SelectVersion() *CustomerOrderRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerOrderRequest) WithVersionIs(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsNot(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionNotIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionBetween(lower int64, upper int64) *CustomerOrderRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerOrderRequest) OrderByVersionAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerOrderRequest) OrderByVersionDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerOrderRequest) SelectMerchantWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	runtime.EnsureRelationProjection(r.Query, "merchant_id")
	r.Query.RelationQuerySelection("merchantEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectMerchantWith", Relation: "merchantEntity"}
	r.relationFactories["merchantEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}
func (r *CustomerOrderRequest) SelectCustomerProfileWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	runtime.EnsureRelationProjection(r.Query, "customer_profile_id")
	r.Query.RelationQuerySelection("customerProfileEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerProfileWith", Relation: "customerProfileEntity"}
	r.relationFactories["customerProfileEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *CustomerOrderRequest) WithMerchantMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("merchant_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) WithoutMerchantMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("merchant_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *CustomerOrderRequest) WithCustomerProfileMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_profile_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) WithoutCustomerProfileMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_profile_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableCustomerOrderRequest) NewEntity(context *runtime.UserContext) (*CustomerOrder, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewCustomerOrder(context)
}

func (e *ExecutableCustomerOrderRequest) ExecuteForOne(context *runtime.UserContext) (*CustomerOrder, error) {
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

func (e *ExecutableCustomerOrderRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CustomerOrder], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*CustomerOrder, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerOrderRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CustomerOrder) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*CustomerOrder, error) {
		entity := newLoadedCustomerOrder()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerOrderRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerOrderRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCustomerOrderRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerOrderRequest) Count() *CustomerOrderRequest {
	return r.CountAs("count")
}

func (r *CustomerOrderRequest) CountAs(alias string) *CustomerOrderRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CustomerOrderRequest) GroupById() *CustomerOrderRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *CustomerOrderRequest) GroupByIdAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *CustomerOrderRequest) GroupByMerchant() *CustomerOrderRequest {
	r.Query.WithGroupBy("merchant_id")
	return r
}

func (r *CustomerOrderRequest) GroupByMerchantAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("merchant_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("merchant_id"))
	return r
}
func (r *CustomerOrderRequest) GroupByCustomerProfile() *CustomerOrderRequest {
	r.Query.WithGroupBy("customer_profile_id")
	return r
}

func (r *CustomerOrderRequest) GroupByCustomerProfileAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("customer_profile_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("customer_profile_id"))
	return r
}
func (r *CustomerOrderRequest) GroupByOrderNumber() *CustomerOrderRequest {
	r.Query.WithGroupBy("order_number")
	return r
}

func (r *CustomerOrderRequest) GroupByOrderNumberAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("order_number")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) GroupByFulfillmentStatus() *CustomerOrderRequest {
	r.Query.WithGroupBy("fulfillment_status")
	return r
}

func (r *CustomerOrderRequest) GroupByFulfillmentStatusAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("fulfillment_status")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("fulfillment_status"))
	return r
}
func (r *CustomerOrderRequest) GroupByVersion() *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *CustomerOrderRequest) GroupByVersionAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}

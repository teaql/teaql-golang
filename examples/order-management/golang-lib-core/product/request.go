

package product

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/order_line"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type ProductRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableProductRequest struct {
	request *ProductRequest
}

func NewProductRequest() *ProductRequest {
	r := &ProductRequest{
		Query: core.NewSelectQuery("product"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewProductMinimalRequest() *ProductRequest {
	r := NewProductRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *ProductRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *ProductRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewProduct().EntityDescriptor()
}

func (r *ProductRequest) NewRelationEntity() core.Entity {
	return newLoadedProduct()
}

func (r *ProductRequest) Comment(comment string) *ProductRequest {
	r.commentText = comment
	return r
}

func (r *ProductRequest) Purpose(purpose string) *ExecutableProductRequest {
	r.purposeText = purpose
	return &ExecutableProductRequest{request: r}
}

func (r *ExecutableProductRequest) Comment(comment string) *ExecutableProductRequest {
	r.request.commentText = comment
	return r
}

func (r *ProductRequest) Limit(limit uint64) *ProductRequest {
	r.Query.Limit(limit)
	return r
}

func (r *ProductRequest) Offset(offset uint64) *ProductRequest {
	r.Query.Offset(offset)
	return r
}

func (r *ProductRequest) OptimizeForContinuousPageFetch() *ProductRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *ProductRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *ProductRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *ProductRequest) OptimizePaginationWithIDSet() *ProductRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *ProductRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *ProductRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *ProductRequest) TopNProbeParentThreshold(threshold uint64) *ProductRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeProductVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeProductVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *ProductRequest) WithDeletedRows() *ProductRequest {
	r.Query.Filter = removeProductVersionFilter(r.Query.Filter)
	return r
}

func (r *ProductRequest) DeletedRowsOnly() *ProductRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *ProductRequest) SelectId() *ProductRequest {
	r.Query.Project("id")
	return r
}

func (r *ProductRequest) WithIdIs(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdIsNot(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *ProductRequest) WithIdNotIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *ProductRequest) WithIdGreaterThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdGreaterThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdLessThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdLessThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithIdBetween(lower uint64, upper uint64) *ProductRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *ProductRequest) WithIdIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *ProductRequest) WithIdIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *ProductRequest) OrderByIdAsc() *ProductRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *ProductRequest) OrderByIdDesc() *ProductRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *ProductRequest) SelectName() *ProductRequest {
	r.Query.Project("name")
	return r
}

func (r *ProductRequest) WithNameIs(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameIsNot(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *ProductRequest) WithNameNotIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *ProductRequest) WithNameGreaterThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameGreaterThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameLessThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameLessThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithNameBetween(lower string, upper string) *ProductRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *ProductRequest) WithNameIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *ProductRequest) WithNameIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *ProductRequest) WithNameContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *ProductRequest) WithNameNotContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *ProductRequest) WithNameStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *ProductRequest) WithNameNotStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *ProductRequest) WithNameEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *ProductRequest) WithNameNotEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *ProductRequest) WithNameSoundingLike(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *ProductRequest) OrderByNameAsc() *ProductRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *ProductRequest) OrderByNameDesc() *ProductRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *ProductRequest) SelectSku() *ProductRequest {
	r.Query.Project("sku")
	return r
}

func (r *ProductRequest) WithSkuIs(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuIsNot(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("sku", converted))
	return r
}
func (r *ProductRequest) WithSkuNotIn(values []string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("sku", converted))
	return r
}
func (r *ProductRequest) WithSkuGreaterThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuGreaterThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuLessThan(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuLessThanOrEqualTo(value string) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("sku", core.ValText(value)))
	return r
}
func (r *ProductRequest) WithSkuBetween(lower string, upper string) *ProductRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("sku", from, to))
	return r
}
func (r *ProductRequest) WithSkuIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("sku"))
	return r
}
func (r *ProductRequest) WithSkuIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("sku"))
	return r
}
func (r *ProductRequest) WithSkuContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprContain("sku", term))
	return r
}
func (r *ProductRequest) WithSkuNotContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotContain("sku", term))
	return r
}
func (r *ProductRequest) WithSkuStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprBeginWith("sku", term))
	return r
}
func (r *ProductRequest) WithSkuNotStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("sku", term))
	return r
}
func (r *ProductRequest) WithSkuEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprEndWith("sku", term))
	return r
}
func (r *ProductRequest) WithSkuNotEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotEndWith("sku", term))
	return r
}
func (r *ProductRequest) WithSkuSoundingLike(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprSoundLike("sku", core.ValText(term)))
	return r
}
func (r *ProductRequest) OrderBySkuAsc() *ProductRequest {
	r.Query.OrderAsc("sku")
	return r
}
func (r *ProductRequest) OrderBySkuDesc() *ProductRequest {
	r.Query.OrderDesc("sku")
	return r
}
func (r *ProductRequest) SelectImageUrl() *ProductRequest {
	r.Query.Project("image_url")
	return r
}

func (r *ProductRequest) WithImageUrlIs(value *string) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ProductRequest) WithImageUrlIsNot(value *string) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ProductRequest) WithImageUrlIn(values []*string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("image_url", converted))
	return r
}
func (r *ProductRequest) WithImageUrlNotIn(values []*string) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("image_url", converted))
	return r
}
func (r *ProductRequest) WithImageUrlGreaterThan(value *string) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ProductRequest) WithImageUrlGreaterThanOrEqualTo(value *string) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ProductRequest) WithImageUrlLessThan(value *string) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ProductRequest) WithImageUrlLessThanOrEqualTo(value *string) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("image_url", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ProductRequest) WithImageUrlBetween(lower *string, upper *string) *ProductRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("image_url", from, to))
	return r
}
func (r *ProductRequest) WithImageUrlIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("image_url"))
	return r
}
func (r *ProductRequest) WithImageUrlIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("image_url"))
	return r
}
func (r *ProductRequest) WithImageUrlContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprContain("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlNotContaining(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotContain("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprBeginWith("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlNotStartingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprEndWith("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlNotEndingWith(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprNotEndWith("image_url", term))
	return r
}
func (r *ProductRequest) WithImageUrlSoundingLike(term string) *ProductRequest {
	r.Query.AndFilter(core.ExprSoundLike("image_url", core.ValText(term)))
	return r
}
func (r *ProductRequest) OrderByImageUrlAsc() *ProductRequest {
	r.Query.OrderAsc("image_url")
	return r
}
func (r *ProductRequest) OrderByImageUrlDesc() *ProductRequest {
	r.Query.OrderDesc("image_url")
	return r
}
func (r *ProductRequest) SelectCommercePlatform() *ProductRequest {
	r.Query.Project("commerce_platform_id")
	return r
}

func (r *ProductRequest) WithCommercePlatformIs(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformIsNot(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *ProductRequest) WithCommercePlatformNotIn(values []uint64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *ProductRequest) WithCommercePlatformGreaterThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformLessThan(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *ProductRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *ProductRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("commerce_platform_id", from, to))
	return r
}
func (r *ProductRequest) WithCommercePlatformIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("commerce_platform_id"))
	return r
}
func (r *ProductRequest) WithCommercePlatformIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("commerce_platform_id"))
	return r
}
func (r *ProductRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *ProductRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "commerce_platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *ProductRequest) OrderByCommercePlatformAsc() *ProductRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *ProductRequest) OrderByCommercePlatformDesc() *ProductRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}
func (r *ProductRequest) SelectCreateTime() *ProductRequest {
	r.Query.Project("create_time")
	return r
}

func (r *ProductRequest) WithCreateTimeIs(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeIsNot(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *ProductRequest) WithCreateTimeNotIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *ProductRequest) WithCreateTimeGreaterThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeLessThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *ProductRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *ProductRequest) WithCreateTimeIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *ProductRequest) WithCreateTimeIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *ProductRequest) OrderByCreateTimeAsc() *ProductRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *ProductRequest) OrderByCreateTimeDesc() *ProductRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *ProductRequest) SelectUpdateTime() *ProductRequest {
	r.Query.Project("update_time")
	return r
}

func (r *ProductRequest) WithUpdateTimeIs(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeIsNot(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *ProductRequest) WithUpdateTimeNotIn(values []time.Time) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *ProductRequest) WithUpdateTimeGreaterThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeLessThan(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ProductRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *ProductRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *ProductRequest) WithUpdateTimeIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *ProductRequest) WithUpdateTimeIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *ProductRequest) OrderByUpdateTimeAsc() *ProductRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *ProductRequest) OrderByUpdateTimeDesc() *ProductRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *ProductRequest) SelectVersion() *ProductRequest {
	r.Query.Project("version")
	return r
}

func (r *ProductRequest) WithVersionIs(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionIsNot(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionIn(values []int64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *ProductRequest) WithVersionNotIn(values []int64) *ProductRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *ProductRequest) WithVersionGreaterThan(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionGreaterThanOrEqualTo(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionLessThan(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionLessThanOrEqualTo(value int64) *ProductRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *ProductRequest) WithVersionBetween(lower int64, upper int64) *ProductRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *ProductRequest) WithVersionIsKnown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *ProductRequest) WithVersionIsUnknown() *ProductRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *ProductRequest) OrderByVersionAsc() *ProductRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *ProductRequest) OrderByVersionDesc() *ProductRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *ProductRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *ProductRequest {
	runtime.EnsureRelationProjection(r.Query, "commerce_platform_id")
	r.Query.RelationQuery("commercePlatformEntity", child.GetQuery())
	r.relationFactories["commercePlatformEntity"] = child.NewRelationEntity
	return r
}

func (r *ProductRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *ProductRequest {
	r.Query.AndFilter(core.ExprInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *ProductRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *ProductRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *ProductRequest) CountOrderLines() *ProductRequest {
	return r.CountOrderLinesAs("countOrderLines")

}
func (r *ProductRequest) CountOrderLinesAs(alias string) *ProductRequest {
	return r.CountOrderLinesWith(alias, order_line.NewOrderLineRequest())
}
func (r *ProductRequest) CountOrderLinesWith(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}

func (r *ProductRequest) SumQuantityOfOrderLines() *ProductRequest {
	return r.SumQuantityOfOrderLinesAs("sumQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) SumQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Sum("quantity", "sum_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) MinQuantityOfOrderLines() *ProductRequest {
	return r.MinQuantityOfOrderLinesAs("minQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) MinQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Min("quantity", "min_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) MaxQuantityOfOrderLines() *ProductRequest {
	return r.MaxQuantityOfOrderLinesAs("maxQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) MaxQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Max("quantity", "max_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) AvgQuantityOfOrderLines() *ProductRequest {
	return r.AvgQuantityOfOrderLinesAs("avgQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) AvgQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Avg("quantity", "avg_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) StandardDeviationQuantityOfOrderLines() *ProductRequest {
	return r.StandardDeviationQuantityOfOrderLinesAs("standardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) StandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Stddev("quantity", "stdDev_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLines() *ProductRequest {
	return r.SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs("squareRootOfPopulationStandardDeviationQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) SquareRootOfPopulationStandardDeviationQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.StddevPop("quantity", "stdDevPop_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) SampleVarianceQuantityOfOrderLines() *ProductRequest {
	return r.SampleVarianceQuantityOfOrderLinesAs("sampleVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) SampleVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.VarSamp("quantity", "varSamp_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) SamplePopulationVarianceQuantityOfOrderLines() *ProductRequest {
	return r.SamplePopulationVarianceQuantityOfOrderLinesAs("samplePopulationVarianceQuantityOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) SamplePopulationVarianceQuantityOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.VarPop("quantity", "varPop_quantity")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) MinCreateTimeOfOrderLines() *ProductRequest {
	return r.MinCreateTimeOfOrderLinesAs("minCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) MinCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}
func (r *ProductRequest) MaxCreateTimeOfOrderLines() *ProductRequest {
	return r.MaxCreateTimeOfOrderLinesAs("maxCreateTimeOfOrderLines", order_line.NewOrderLineRequest())
}
func (r *ProductRequest) MaxCreateTimeOfOrderLinesAs(alias string, child *order_line.OrderLineRequest) *ProductRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("orderLineList", alias, child.Query, true))
	return r
}

func (r *ProductRequest) SelectOrderLineList() *ProductRequest {
	return r.SelectOrderLineListWith(order_line.NewOrderLineRequest())
}

func (r *ProductRequest) SelectOrderLineListWith(child *order_line.OrderLineRequest) *ProductRequest {
	r.Query.RelationQuery("orderLineList", child.Query)
	return r
}

func (r *ProductRequest) HaveOrderLines() *ProductRequest {
	return r.WithOrderLineListMatching(order_line.NewOrderLineRequest())
}

func (r *ProductRequest) HaveNoOrderLines() *ProductRequest {
	return r.WithoutOrderLineListMatching(order_line.NewOrderLineRequest())
}

func (r *ProductRequest) WithOrderLineListMatching(child *order_line.OrderLineRequest) *ProductRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "product_id"))
	return r
}

func (r *ProductRequest) WithoutOrderLineListMatching(child *order_line.OrderLineRequest) *ProductRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "product_id"))
	return r
}

func (e *ExecutableProductRequest) NewEntity(context *runtime.UserContext) *Product {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewProduct()
	initialized := context.InitializeEntity("Product", entity)
	typed, ok := initialized.(*Product)
	if !ok {
		panic("entity initializer changed Product to an incompatible type")
	}
	return typed
}

func (e *ExecutableProductRequest) ExecuteForOne(context *runtime.UserContext) (*Product, error) {
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

func (e *ExecutableProductRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Product], error) {
	rows, authorized, err := e.executeRecords(context, true)
	if err != nil {
		return nil, err
	}

	var results []*Product
	for _, rec := range rows {
		entity := newLoadedProduct()
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	list := core.NewSmartList(results)
	if len(e.request.queryOptions.Facets) > 0 {
		dsRaw := context.GetResource("dataService")
		ds, ok := dsRaw.(data_service.QueryExecutor)
		if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
		facets, err := runtime.ExecuteFacets(
			context, runtime.NewRuntimeDataService(context.Metadata, ds),
			authorized, e.request.queryOptions)
		if err != nil { return nil, err }
		core.AttachFacets(list, facets)
	}
	return list, nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableProductRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Product], error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	query := r.Query.Clone()
	query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	const countAlias = "__teaql_total"
	var rows []core.Record
	var total uint64
	if authorized.IDSetPagination != nil {
		rows, err = service.FetchAll(context, authorized)
		if err != nil { return nil, err }
		if retainedCount, accuracy := context.IDSetCount(); accuracy == "EXACT" {
			total = retainedCount
		} else {
			countRows, countErr := service.FetchAll(context, authorized.ForExactCount(countAlias))
			if countErr != nil { return nil, countErr }
			if len(countRows) != 1 { return nil, fmt.Errorf("exact count returned %d rows", len(countRows)) }
			var ok bool
			total, ok = countRows[0][countAlias].TryU64()
			if !ok { return nil, fmt.Errorf("exact count did not return an unsigned integer") }
		}
	} else {
		countRows, countErr := service.FetchAll(context, authorized.ForExactCount(countAlias))
		if countErr != nil { return nil, countErr }
		if len(countRows) != 1 { return nil, fmt.Errorf("exact count returned %d rows", len(countRows)) }
		var ok bool
		total, ok = countRows[0][countAlias].TryU64()
		if !ok { return nil, fmt.Errorf("exact count did not return an unsigned integer") }
		rows, err = service.FetchAll(context, authorized)
		if err != nil { return nil, err }
	}
	results := make([]*Product, 0, len(rows))
	for _, rec := range rows {
		entity := newLoadedProduct()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["orderLineList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderLineList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_line.NewOrderLine()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderLineList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableProductRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Product) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareEntityQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req, err := data_service.NewQueryRequest(authorized)
	if err != nil { return err }
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := newLoadedProduct()
			if err := entity.FromRecord(rec); err != nil {
				return err
			}
			if err := yield(entity); err != nil {
				return err
			}
		}
		return nil
	})
}

func (e *ExecutableProductRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context, false)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableProductRequest) executeRecords(context *runtime.UserContext, entityProjection bool) ([]core.Record, *core.SelectQuery, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	prepare := context.PrepareQuery
	if entityProjection { prepare = context.PrepareEntityQuery }
	authorized, err := prepare(query)
	if err != nil { return nil, nil, err }

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAll(context, authorized)
	if err != nil {
		return nil, nil, err
	}
	return rows, authorized, nil
}

// ExecuteForRows preserves aggregate/group projections as records while keeping
// the cross-language SmartList result boundary.
func (e *ExecutableProductRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *ProductRequest) Count() *ProductRequest {
	return r.CountAs("count")
}

func (r *ProductRequest) CountAs(alias string) *ProductRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *ProductRequest) GroupById() *ProductRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *ProductRequest) GroupByName() *ProductRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *ProductRequest) GroupBySku() *ProductRequest {
	r.Query.WithGroupBy("sku")
	return r
}
func (r *ProductRequest) GroupByImageUrl() *ProductRequest {
	r.Query.WithGroupBy("image_url")
	return r
}
func (r *ProductRequest) GroupByCommercePlatform() *ProductRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *ProductRequest) GroupByCreateTime() *ProductRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *ProductRequest) GroupByUpdateTime() *ProductRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *ProductRequest) GroupByVersion() *ProductRequest {
	r.Query.WithGroupBy("version")
	return r
}

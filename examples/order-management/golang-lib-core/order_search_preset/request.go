

package order_search_preset

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderSearchPresetRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableOrderSearchPresetRequest struct {
	request *OrderSearchPresetRequest
}

func NewOrderSearchPresetRequest() *OrderSearchPresetRequest {
	r := &OrderSearchPresetRequest{
		Query: core.NewSelectQuery("order_search_preset"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderSearchPresetMinimalRequest() *OrderSearchPresetRequest {
	r := NewOrderSearchPresetRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderSearchPresetRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *OrderSearchPresetRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewOrderSearchPreset().EntityDescriptor()
}

func (r *OrderSearchPresetRequest) NewRelationEntity() core.Entity {
	return NewOrderSearchPreset()
}

func (r *OrderSearchPresetRequest) Comment(comment string) *OrderSearchPresetRequest {
	r.commentText = comment
	return r
}

func (r *OrderSearchPresetRequest) Purpose(purpose string) *ExecutableOrderSearchPresetRequest {
	r.purposeText = purpose
	return &ExecutableOrderSearchPresetRequest{request: r}
}

func (r *ExecutableOrderSearchPresetRequest) Comment(comment string) *ExecutableOrderSearchPresetRequest {
	r.request.commentText = comment
	return r
}

func (r *OrderSearchPresetRequest) Limit(limit uint64) *OrderSearchPresetRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderSearchPresetRequest) Offset(offset uint64) *OrderSearchPresetRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderSearchPresetRequest) OptimizeForContinuousPageFetch() *OrderSearchPresetRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderSearchPresetRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderSearchPresetRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderSearchPresetRequest) OptimizePaginationWithIDSet() *OrderSearchPresetRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderSearchPresetRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderSearchPresetRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderSearchPresetRequest) TopNProbeParentThreshold(threshold uint64) *OrderSearchPresetRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderSearchPresetVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderSearchPresetVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderSearchPresetRequest) WithDeletedRows() *OrderSearchPresetRequest {
	r.Query.Filter = removeOrderSearchPresetVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderSearchPresetRequest) DeletedRowsOnly() *OrderSearchPresetRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderSearchPresetRequest) SelectId() *OrderSearchPresetRequest {
	r.Query.Project("id")
	return r
}

func (r *OrderSearchPresetRequest) WithIdIs(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsNot(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithIdNotIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithIdGreaterThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdLessThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdLessThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithIdBetween(lower uint64, upper uint64) *OrderSearchPresetRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *OrderSearchPresetRequest) WithIdIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *OrderSearchPresetRequest) OrderByIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *OrderSearchPresetRequest) SelectName() *OrderSearchPresetRequest {
	r.Query.Project("name")
	return r
}

func (r *OrderSearchPresetRequest) WithNameIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithNameGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithNameBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *OrderSearchPresetRequest) WithNameIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *OrderSearchPresetRequest) WithNameContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *OrderSearchPresetRequest) WithNameSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByNameAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *OrderSearchPresetRequest) OrderByNameDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *OrderSearchPresetRequest) SelectFilterJson() *OrderSearchPresetRequest {
	r.Query.Project("filter_json")
	return r
}

func (r *OrderSearchPresetRequest) WithFilterJsonIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("filter_json", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("filter_json", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("filter_json", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("filter_json", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("filter_json"))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("filter_json"))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith("filter_json", term))
	return r
}
func (r *OrderSearchPresetRequest) WithFilterJsonSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike("filter_json", core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByFilterJsonAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("filter_json")
	return r
}
func (r *OrderSearchPresetRequest) OrderByFilterJsonDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("filter_json")
	return r
}
func (r *OrderSearchPresetRequest) SelectRequestId() *OrderSearchPresetRequest {
	r.Query.Project("request_id")
	return r
}

func (r *OrderSearchPresetRequest) WithRequestIdIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("request_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("request_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("request_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("request_id", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("request_id"))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("request_id"))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith("request_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithRequestIdSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike("request_id", core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByRequestIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("request_id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByRequestIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("request_id")
	return r
}
func (r *OrderSearchPresetRequest) SelectOwnerUserId() *OrderSearchPresetRequest {
	r.Query.Project("owner_user_id")
	return r
}

func (r *OrderSearchPresetRequest) WithOwnerUserIdIs(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsNot(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("owner_user_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotIn(values []string) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("owner_user_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdGreaterThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdGreaterThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdLessThan(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdLessThanOrEqualTo(value string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("owner_user_id", core.ValText(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdBetween(lower string, upper string) *OrderSearchPresetRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("owner_user_id", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("owner_user_id"))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("owner_user_id"))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprContain("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotContaining(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotContain("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprBeginWith("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotStartingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEndWith("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdNotEndingWith(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotEndWith("owner_user_id", term))
	return r
}
func (r *OrderSearchPresetRequest) WithOwnerUserIdSoundingLike(term string) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprSoundLike("owner_user_id", core.ValText(term)))
	return r
}
func (r *OrderSearchPresetRequest) OrderByOwnerUserIdAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("owner_user_id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByOwnerUserIdDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("owner_user_id")
	return r
}
func (r *OrderSearchPresetRequest) SelectCommercePlatform() *OrderSearchPresetRequest {
	r.Query.Project("commerce_platform_id")
	return r
}

func (r *OrderSearchPresetRequest) WithCommercePlatformIs(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsNot(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformNotIn(values []uint64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformGreaterThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformLessThan(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *OrderSearchPresetRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("commerce_platform_id", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("commerce_platform_id"))
	return r
}
func (r *OrderSearchPresetRequest) WithCommercePlatformIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("commerce_platform_id"))
	return r
}
func (r *OrderSearchPresetRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderSearchPresetRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "commerce_platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *OrderSearchPresetRequest) OrderByCommercePlatformAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *OrderSearchPresetRequest) OrderByCommercePlatformDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}
func (r *OrderSearchPresetRequest) SelectCreateTime() *OrderSearchPresetRequest {
	r.Query.Project("create_time")
	return r
}

func (r *OrderSearchPresetRequest) WithCreateTimeIs(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsNot(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeNotIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeGreaterThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeLessThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *OrderSearchPresetRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *OrderSearchPresetRequest) WithCreateTimeIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *OrderSearchPresetRequest) OrderByCreateTimeAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *OrderSearchPresetRequest) OrderByCreateTimeDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *OrderSearchPresetRequest) SelectUpdateTime() *OrderSearchPresetRequest {
	r.Query.Project("update_time")
	return r
}

func (r *OrderSearchPresetRequest) WithUpdateTimeIs(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsNot(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeNotIn(values []time.Time) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeGreaterThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeLessThan(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *OrderSearchPresetRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *OrderSearchPresetRequest) WithUpdateTimeIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *OrderSearchPresetRequest) OrderByUpdateTimeAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *OrderSearchPresetRequest) OrderByUpdateTimeDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *OrderSearchPresetRequest) SelectVersion() *OrderSearchPresetRequest {
	r.Query.Project("version")
	return r
}

func (r *OrderSearchPresetRequest) WithVersionIs(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsNot(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIn(values []int64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionNotIn(values []int64) *OrderSearchPresetRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionGreaterThan(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionLessThan(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionLessThanOrEqualTo(value int64) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionBetween(lower int64, upper int64) *OrderSearchPresetRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsKnown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *OrderSearchPresetRequest) WithVersionIsUnknown() *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *OrderSearchPresetRequest) OrderByVersionAsc() *OrderSearchPresetRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderSearchPresetRequest) OrderByVersionDesc() *OrderSearchPresetRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderSearchPresetRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderSearchPresetRequest {
	r.Query.Project("commerce_platform_id")
	r.Query.RelationQuery("commercePlatformEntity", child.GetQuery())
	r.relationFactories["commercePlatformEntity"] = child.NewRelationEntity
	return r
}

func (r *OrderSearchPresetRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderSearchPresetRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderSearchPresetRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableOrderSearchPresetRequest) NewEntity(context *runtime.UserContext) *OrderSearchPreset {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewOrderSearchPreset()
	initialized := context.InitializeEntity("OrderSearchPreset", entity)
	typed, ok := initialized.(*OrderSearchPreset)
	if !ok {
		panic("entity initializer changed OrderSearchPreset to an incompatible type")
	}
	return typed
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteForOne(context *runtime.UserContext) (*OrderSearchPreset, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableOrderSearchPresetRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderSearchPreset], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*OrderSearchPreset
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewOrderSearchPreset()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	list := core.NewSmartList(results)
	if len(e.request.queryOptions.Facets) > 0 {
		dsRaw := context.GetResource("dataService")
		ds, ok := dsRaw.(data_service.QueryExecutor)
		if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
		facets, err := runtime.ExecuteFacets(
			context, runtime.NewRuntimeDataService(context.Metadata, ds),
			e.request.Query, e.request.queryOptions)
		if err != nil { return nil, err }
		core.AttachFacets(list, facets)
	}
	return list, nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableOrderSearchPresetRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*OrderSearchPreset], error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForPage()")
	}
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	r.Query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(r.Query)
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
	results := make([]*OrderSearchPreset, 0, len(rows))
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewOrderSearchPreset()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["commercePlatformEntity"]; selected {
			entity.markRelationLoaded("commercePlatformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["commercePlatformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("commercePlatformEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableOrderSearchPresetRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderSearchPreset) error) error {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForStream()")
	}
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	r.Query.Comment(r.commentText).Purpose(r.purposeText)
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req := &data_service.QueryRequest{
		Query: r.Query, TraceChain: r.Query.TraceChain,
		Comment: r.Query.CommentText, Purpose: r.Query.PurposeText,
	}
	queryRoot := core.NewEntityRoot()
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewOrderSearchPreset()
			entity.AttachEntityRoot(queryRoot)
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

func (e *ExecutableOrderSearchPresetRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	r.Query.Comment(r.commentText).Purpose(r.purposeText)

	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}

	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}

	rows, err := runtime.NewRuntimeDataService(context.Metadata, ds).FetchAll(context, r.Query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ExecuteForRows preserves aggregate/group projections as records while keeping
// the cross-language SmartList result boundary.
func (e *ExecutableOrderSearchPresetRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderSearchPresetRequest) Count() *OrderSearchPresetRequest {
	return r.CountAs("count")
}

func (r *OrderSearchPresetRequest) CountAs(alias string) *OrderSearchPresetRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *OrderSearchPresetRequest) GroupById() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByName() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *OrderSearchPresetRequest) GroupByFilterJson() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("filter_json")
	return r
}
func (r *OrderSearchPresetRequest) GroupByRequestId() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("request_id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByOwnerUserId() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("owner_user_id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByCommercePlatform() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *OrderSearchPresetRequest) GroupByCreateTime() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *OrderSearchPresetRequest) GroupByUpdateTime() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *OrderSearchPresetRequest) GroupByVersion() *OrderSearchPresetRequest {
	r.Query.WithGroupBy("version")
	return r
}

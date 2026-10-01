

package customer

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-management-service-core-workspace/lib/customer_order"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableCustomerRequest struct {
	request *CustomerRequest
}

func NewCustomerRequest() *CustomerRequest {
	r := &CustomerRequest{
		Query: core.NewSelectQuery("customer"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerMinimalRequest() *CustomerRequest {
	r := NewCustomerRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *CustomerRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCustomer().EntityDescriptor()
}

func (r *CustomerRequest) NewRelationEntity() core.Entity {
	return NewCustomer()
}

func (r *CustomerRequest) Comment(comment string) *CustomerRequest {
	r.commentText = comment
	return r
}

func (r *CustomerRequest) Purpose(purpose string) *ExecutableCustomerRequest {
	r.purposeText = purpose
	return &ExecutableCustomerRequest{request: r}
}

func (r *ExecutableCustomerRequest) Comment(comment string) *ExecutableCustomerRequest {
	r.request.commentText = comment
	return r
}

func (r *CustomerRequest) Limit(limit uint64) *CustomerRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerRequest) Offset(offset uint64) *CustomerRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerRequest) OptimizeForContinuousPageFetch() *CustomerRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerRequest) OptimizePaginationWithIDSet() *CustomerRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerRequest) TopNProbeParentThreshold(threshold uint64) *CustomerRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerRequest) WithDeletedRows() *CustomerRequest {
	r.Query.Filter = removeCustomerVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerRequest) DeletedRowsOnly() *CustomerRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerRequest) SelectId() *CustomerRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerRequest) WithIdIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerRequest) WithIdNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerRequest) WithIdGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithIdBetween(lower uint64, upper uint64) *CustomerRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerRequest) WithIdIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerRequest) WithIdIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerRequest) OrderByIdAsc() *CustomerRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerRequest) OrderByIdDesc() *CustomerRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerRequest) SelectName() *CustomerRequest {
	r.Query.Project("name")
	return r
}

func (r *CustomerRequest) WithNameIs(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameIsNot(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *CustomerRequest) WithNameNotIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *CustomerRequest) WithNameGreaterThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameGreaterThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameLessThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameLessThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithNameBetween(lower string, upper string) *CustomerRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *CustomerRequest) WithNameIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *CustomerRequest) WithNameIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *CustomerRequest) WithNameContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *CustomerRequest) WithNameStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameNotEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *CustomerRequest) WithNameSoundingLike(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprSoundLike("name", core.ValText(term)))
	return r
}
func (r *CustomerRequest) OrderByNameAsc() *CustomerRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *CustomerRequest) OrderByNameDesc() *CustomerRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *CustomerRequest) SelectEmail() *CustomerRequest {
	r.Query.Project("email")
	return r
}

func (r *CustomerRequest) WithEmailIs(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailIsNot(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("email", converted))
	return r
}
func (r *CustomerRequest) WithEmailNotIn(values []string) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("email", converted))
	return r
}
func (r *CustomerRequest) WithEmailGreaterThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailGreaterThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailLessThan(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailLessThanOrEqualTo(value string) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("email", core.ValText(value)))
	return r
}
func (r *CustomerRequest) WithEmailBetween(lower string, upper string) *CustomerRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("email", from, to))
	return r
}
func (r *CustomerRequest) WithEmailIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("email"))
	return r
}
func (r *CustomerRequest) WithEmailIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("email"))
	return r
}
func (r *CustomerRequest) WithEmailContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprContain("email", term))
	return r
}
func (r *CustomerRequest) WithEmailNotContaining(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotContain("email", term))
	return r
}
func (r *CustomerRequest) WithEmailStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprBeginWith("email", term))
	return r
}
func (r *CustomerRequest) WithEmailNotStartingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("email", term))
	return r
}
func (r *CustomerRequest) WithEmailEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprEndWith("email", term))
	return r
}
func (r *CustomerRequest) WithEmailNotEndingWith(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotEndWith("email", term))
	return r
}
func (r *CustomerRequest) WithEmailSoundingLike(term string) *CustomerRequest {
	r.Query.AndFilter(core.ExprSoundLike("email", core.ValText(term)))
	return r
}
func (r *CustomerRequest) OrderByEmailAsc() *CustomerRequest {
	r.Query.OrderAsc("email")
	return r
}
func (r *CustomerRequest) OrderByEmailDesc() *CustomerRequest {
	r.Query.OrderDesc("email")
	return r
}
func (r *CustomerRequest) SelectCommercePlatform() *CustomerRequest {
	r.Query.Project("commerce_platform_id")
	return r
}

func (r *CustomerRequest) WithCommercePlatformIs(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformIsNot(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerRequest) WithCommercePlatformNotIn(values []uint64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("commerce_platform_id", converted))
	return r
}
func (r *CustomerRequest) WithCommercePlatformGreaterThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformGreaterThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformLessThan(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformLessThanOrEqualTo(value uint64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("commerce_platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerRequest) WithCommercePlatformBetween(lower uint64, upper uint64) *CustomerRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("commerce_platform_id", from, to))
	return r
}
func (r *CustomerRequest) WithCommercePlatformIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("commerce_platform_id"))
	return r
}
func (r *CustomerRequest) WithCommercePlatformIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("commerce_platform_id"))
	return r
}
func (r *CustomerRequest) FacetByCommercePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "commerce_platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *CustomerRequest) OrderByCommercePlatformAsc() *CustomerRequest {
	r.Query.OrderAsc("commerce_platform_id")
	return r
}
func (r *CustomerRequest) OrderByCommercePlatformDesc() *CustomerRequest {
	r.Query.OrderDesc("commerce_platform_id")
	return r
}
func (r *CustomerRequest) SelectCreateTime() *CustomerRequest {
	r.Query.Project("create_time")
	return r
}

func (r *CustomerRequest) WithCreateTimeIs(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeIsNot(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *CustomerRequest) WithCreateTimeNotIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *CustomerRequest) WithCreateTimeGreaterThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeLessThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *CustomerRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *CustomerRequest) WithCreateTimeIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *CustomerRequest) WithCreateTimeIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *CustomerRequest) OrderByCreateTimeAsc() *CustomerRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *CustomerRequest) OrderByCreateTimeDesc() *CustomerRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *CustomerRequest) SelectUpdateTime() *CustomerRequest {
	r.Query.Project("update_time")
	return r
}

func (r *CustomerRequest) WithUpdateTimeIs(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeIsNot(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *CustomerRequest) WithUpdateTimeNotIn(values []time.Time) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *CustomerRequest) WithUpdateTimeGreaterThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeLessThan(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *CustomerRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *CustomerRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *CustomerRequest) WithUpdateTimeIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *CustomerRequest) WithUpdateTimeIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *CustomerRequest) OrderByUpdateTimeAsc() *CustomerRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *CustomerRequest) OrderByUpdateTimeDesc() *CustomerRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *CustomerRequest) SelectVersion() *CustomerRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerRequest) WithVersionIs(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIsNot(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerRequest) WithVersionNotIn(values []int64) *CustomerRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThan(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerRequest) WithVersionBetween(lower int64, upper int64) *CustomerRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerRequest) WithVersionIsKnown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerRequest) WithVersionIsUnknown() *CustomerRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerRequest) OrderByVersionAsc() *CustomerRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerRequest) OrderByVersionDesc() *CustomerRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerRequest) SelectCommercePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerRequest {
	r.Query.Project("commerce_platform_id")
	r.Query.RelationQuery("commercePlatformEntity", child.GetQuery())
	r.relationFactories["commercePlatformEntity"] = child.NewRelationEntity
	return r
}

func (r *CustomerRequest) WithCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerRequest {
	r.Query.AndFilter(core.ExprInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerRequest) WithoutCommercePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("commerce_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerRequest) CountCustomerOrders() *CustomerRequest {
	return r.CountCustomerOrdersAs("countCustomerOrders")

}
func (r *CustomerRequest) CountCustomerOrdersAs(alias string) *CustomerRequest {
	return r.CountCustomerOrdersWith(alias, customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) CountCustomerOrdersWith(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Count(alias)
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}

func (r *CustomerRequest) MinOrderDateOfCustomerOrders() *CustomerRequest {
	return r.MinOrderDateOfCustomerOrdersAs("minOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MinOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Min("order_date", "min_orderDate")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MaxOrderDateOfCustomerOrders() *CustomerRequest {
	return r.MaxOrderDateOfCustomerOrdersAs("maxOrderDateOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MaxOrderDateOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Max("order_date", "max_orderDate")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) SumTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SumTotalAmountOfCustomerOrdersAs("sumTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SumTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Sum("total_amount", "sum_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MinTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.MinTotalAmountOfCustomerOrdersAs("minTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MinTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Min("total_amount", "min_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MaxTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.MaxTotalAmountOfCustomerOrdersAs("maxTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MaxTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Max("total_amount", "max_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) AvgTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.AvgTotalAmountOfCustomerOrdersAs("avgTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) AvgTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Avg("total_amount", "avg_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) StandardDeviationTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.StandardDeviationTotalAmountOfCustomerOrdersAs("standardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) StandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Stddev("total_amount", "stdDev_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs("squareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SquareRootOfPopulationStandardDeviationTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.StddevPop("total_amount", "stdDevPop_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) SampleVarianceTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SampleVarianceTotalAmountOfCustomerOrdersAs("sampleVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SampleVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.VarSamp("total_amount", "varSamp_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) SamplePopulationVarianceTotalAmountOfCustomerOrders() *CustomerRequest {
	return r.SamplePopulationVarianceTotalAmountOfCustomerOrdersAs("samplePopulationVarianceTotalAmountOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) SamplePopulationVarianceTotalAmountOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.VarPop("total_amount", "varPop_totalAmount")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MinCreateTimeOfCustomerOrders() *CustomerRequest {
	return r.MinCreateTimeOfCustomerOrdersAs("minCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MinCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Min("create_time", "min_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MaxCreateTimeOfCustomerOrders() *CustomerRequest {
	return r.MaxCreateTimeOfCustomerOrdersAs("maxCreateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MaxCreateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Max("create_time", "max_createTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MinUpdateTimeOfCustomerOrders() *CustomerRequest {
	return r.MinUpdateTimeOfCustomerOrdersAs("minUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MinUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Min("update_time", "min_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}
func (r *CustomerRequest) MaxUpdateTimeOfCustomerOrders() *CustomerRequest {
	return r.MaxUpdateTimeOfCustomerOrdersAs("maxUpdateTimeOfCustomerOrders", customer_order.NewCustomerOrderRequest())
}
func (r *CustomerRequest) MaxUpdateTimeOfCustomerOrdersAs(alias string, child *customer_order.CustomerOrderRequest) *CustomerRequest {
	child.Query.Max("update_time", "max_updateTime")
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, core.NewRelationAggregate("customerOrderList", alias, child.Query, true))
	return r
}

func (r *CustomerRequest) SelectCustomerOrderList() *CustomerRequest {
	return r.SelectCustomerOrderListWith(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) SelectCustomerOrderListWith(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.RelationQuery("customerOrderList", child.Query)
	return r
}

func (r *CustomerRequest) HaveCustomerOrders() *CustomerRequest {
	return r.WithCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) HaveNoCustomerOrders() *CustomerRequest {
	return r.WithoutCustomerOrderListMatching(customer_order.NewCustomerOrderRequest())
}

func (r *CustomerRequest) WithCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_id"))
	return r
}

func (r *CustomerRequest) WithoutCustomerOrderListMatching(child *customer_order.CustomerOrderRequest) *CustomerRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_id"))
	return r
}

func (e *ExecutableCustomerRequest) NewEntity(context *runtime.UserContext) *Customer {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewCustomer()
	initialized := context.InitializeEntity("Customer", entity)
	typed, ok := initialized.(*Customer)
	if !ok {
		panic("entity initializer changed Customer to an incompatible type")
	}
	return typed
}

func (e *ExecutableCustomerRequest) ExecuteForOne(context *runtime.UserContext) (*Customer, error) {
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

func (e *ExecutableCustomerRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Customer], error) {
	rows, authorized, err := e.executeRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*Customer
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewCustomer()
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
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
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
func (e *ExecutableCustomerRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Customer], error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, err }
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	query := r.Query.Clone()
	query.Page(offset, size).Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
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
	results := make([]*Customer, 0, len(rows))
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewCustomer()
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
		if relationValue, selected := rec["customerOrderList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation customerOrderList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := customer_order.NewCustomerOrder()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.CustomerOrderList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableCustomerRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Customer) error) error {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return err }
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req, err := data_service.NewQueryRequest(authorized)
	if err != nil { return err }
	queryRoot := core.NewEntityRoot()
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewCustomer()
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

func (e *ExecutableCustomerRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { return nil, nil, err }
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	authorized, err := context.PrepareQuery(query)
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
func (e *ExecutableCustomerRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerRequest) Count() *CustomerRequest {
	return r.CountAs("count")
}

func (r *CustomerRequest) CountAs(alias string) *CustomerRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *CustomerRequest) GroupById() *CustomerRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *CustomerRequest) GroupByName() *CustomerRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *CustomerRequest) GroupByEmail() *CustomerRequest {
	r.Query.WithGroupBy("email")
	return r
}
func (r *CustomerRequest) GroupByCommercePlatform() *CustomerRequest {
	r.Query.WithGroupBy("commerce_platform_id")
	return r
}
func (r *CustomerRequest) GroupByCreateTime() *CustomerRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *CustomerRequest) GroupByUpdateTime() *CustomerRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *CustomerRequest) GroupByVersion() *CustomerRequest {
	r.Query.WithGroupBy("version")
	return r
}

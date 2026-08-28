package school

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

type SchoolRequest struct {
	Query             *core.SelectQuery
	purposeText       string
	commentText       string
	relationFactories map[string]func() core.Entity
}

type ExecutableSchoolRequest struct {
	request *SchoolRequest
}

func NewSchoolRequest() *SchoolRequest {
	r := &SchoolRequest{
		Query:             core.NewSelectQuery("School"),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewSchoolMinimalRequest() *SchoolRequest {
	r := NewSchoolRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *SchoolRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *SchoolRequest) NewRelationEntity() core.Entity {
	return NewSchool()
}

func (r *SchoolRequest) Comment(comment string) *SchoolRequest {
	r.commentText = comment
	return r
}

func (r *SchoolRequest) Purpose(purpose string) *ExecutableSchoolRequest {
	r.purposeText = purpose
	return &ExecutableSchoolRequest{request: r}
}

func (r *ExecutableSchoolRequest) Comment(comment string) *ExecutableSchoolRequest {
	r.request.commentText = comment
	return r
}

func (r *SchoolRequest) Limit(limit uint64) *SchoolRequest {
	r.Query.Limit(limit)
	return r
}

func (r *SchoolRequest) Offset(offset uint64) *SchoolRequest {
	r.Query.Offset(offset)
	return r
}

func (r *SchoolRequest) OptimizeForContinuousPageFetch() *SchoolRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *SchoolRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *SchoolRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func removeSchoolVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil {
		return nil
	}
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd {
		return expr
	}
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeSchoolVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return core.ExprAndNode(parts...)
}

func (r *SchoolRequest) WithDeletedRows() *SchoolRequest {
	r.Query.Filter = removeSchoolVersionFilter(r.Query.Filter)
	return r
}

func (r *SchoolRequest) DeletedRowsOnly() *SchoolRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *SchoolRequest) SelectId() *SchoolRequest {
	r.Query.Project("id")
	return r
}

func (r *SchoolRequest) WithIdIs(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdIsNot(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *SchoolRequest) WithIdNotIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *SchoolRequest) WithIdGreaterThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdGreaterThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdLessThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdLessThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdBetween(lower uint64, upper uint64) *SchoolRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *SchoolRequest) WithIdIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *SchoolRequest) WithIdIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *SchoolRequest) OrderByIdAsc() *SchoolRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *SchoolRequest) OrderByIdDesc() *SchoolRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *SchoolRequest) SelectPlatform() *SchoolRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *SchoolRequest) WithPlatformIs(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformIsNot(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *SchoolRequest) WithPlatformNotIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *SchoolRequest) WithPlatformGreaterThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformLessThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformLessThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformBetween(lower uint64, upper uint64) *SchoolRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *SchoolRequest) WithPlatformIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *SchoolRequest) WithPlatformIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *SchoolRequest) FacetByPlatformAs(name string, nestedReq any) *SchoolRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "platform_id", req.GetQuery())
	}
	return r
}
func (r *SchoolRequest) OrderByPlatformAsc() *SchoolRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *SchoolRequest) OrderByPlatformDesc() *SchoolRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *SchoolRequest) SelectSchoolType() *SchoolRequest {
	r.Query.Project("school_type_id")
	return r
}

func (r *SchoolRequest) WithSchoolTypeIs(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("school_type_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsNot(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("school_type_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("school_type_id", converted))
	return r
}
func (r *SchoolRequest) WithSchoolTypeNotIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("school_type_id", converted))
	return r
}
func (r *SchoolRequest) WithSchoolTypeGreaterThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("school_type_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeGreaterThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("school_type_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeLessThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("school_type_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeLessThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("school_type_id", core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeBetween(lower uint64, upper uint64) *SchoolRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("school_type_id", from, to))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("school_type_id"))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("school_type_id"))
	return r
}
func (r *SchoolRequest) FacetBySchoolTypeAs(name string, nestedReq any) *SchoolRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "school_type_id", req.GetQuery())
	}
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsPrimary() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("school_type_id", core.ValU64(1001)))
	return r
}
func (r *SchoolRequest) OrderBySchoolTypeAsc() *SchoolRequest {
	r.Query.OrderAsc("school_type_id")
	return r
}
func (r *SchoolRequest) OrderBySchoolTypeDesc() *SchoolRequest {
	r.Query.OrderDesc("school_type_id")
	return r
}
func (r *SchoolRequest) SelectName() *SchoolRequest {
	r.Query.Project("name")
	return r
}

func (r *SchoolRequest) WithNameIs(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameIsNot(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *SchoolRequest) WithNameNotIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *SchoolRequest) WithNameGreaterThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameGreaterThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameLessThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameLessThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameBetween(lower string, upper string) *SchoolRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *SchoolRequest) WithNameIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *SchoolRequest) WithNameIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *SchoolRequest) WithNameContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *SchoolRequest) WithNameNotContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *SchoolRequest) WithNameStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *SchoolRequest) WithNameNotStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *SchoolRequest) WithNameEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *SchoolRequest) WithNameNotEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *SchoolRequest) OrderByNameAsc() *SchoolRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *SchoolRequest) OrderByNameDesc() *SchoolRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *SchoolRequest) SelectAddress() *SchoolRequest {
	r.Query.Project("address")
	return r
}

func (r *SchoolRequest) WithAddressIs(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("address", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressIsNot(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("address", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("address", converted))
	return r
}
func (r *SchoolRequest) WithAddressNotIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("address", converted))
	return r
}
func (r *SchoolRequest) WithAddressGreaterThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("address", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressGreaterThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("address", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressLessThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("address", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressLessThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("address", core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressBetween(lower string, upper string) *SchoolRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("address", from, to))
	return r
}
func (r *SchoolRequest) WithAddressIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("address"))
	return r
}
func (r *SchoolRequest) WithAddressIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("address"))
	return r
}
func (r *SchoolRequest) WithAddressContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprContain("address", term))
	return r
}
func (r *SchoolRequest) WithAddressNotContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotContain("address", term))
	return r
}
func (r *SchoolRequest) WithAddressStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprBeginWith("address", term))
	return r
}
func (r *SchoolRequest) WithAddressNotStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("address", term))
	return r
}
func (r *SchoolRequest) WithAddressEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEndWith("address", term))
	return r
}
func (r *SchoolRequest) WithAddressNotEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotEndWith("address", term))
	return r
}
func (r *SchoolRequest) OrderByAddressAsc() *SchoolRequest {
	r.Query.OrderAsc("address")
	return r
}
func (r *SchoolRequest) OrderByAddressDesc() *SchoolRequest {
	r.Query.OrderDesc("address")
	return r
}
func (r *SchoolRequest) SelectEstablishedDate() *SchoolRequest {
	r.Query.Project("established_date")
	return r
}

func (r *SchoolRequest) WithEstablishedDateIs(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("established_date", core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIsNot(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("established_date", core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList("established_date", converted))
	return r
}
func (r *SchoolRequest) WithEstablishedDateNotIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList("established_date", converted))
	return r
}
func (r *SchoolRequest) WithEstablishedDateGreaterThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("established_date", core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateGreaterThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("established_date", core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateLessThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("established_date", core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateLessThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("established_date", core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateBetween(lower time.Time, upper time.Time) *SchoolRequest {
	value := lower
	from := core.ValDate(value)
	value = upper
	to := core.ValDate(value)
	r.Query.AndFilter(core.ExprBetweenNode("established_date", from, to))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("established_date"))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("established_date"))
	return r
}
func (r *SchoolRequest) OrderByEstablishedDateAsc() *SchoolRequest {
	r.Query.OrderAsc("established_date")
	return r
}
func (r *SchoolRequest) OrderByEstablishedDateDesc() *SchoolRequest {
	r.Query.OrderDesc("established_date")
	return r
}
func (r *SchoolRequest) SelectStudentCapacity() *SchoolRequest {
	r.Query.Project("student_capacity")
	return r
}

func (r *SchoolRequest) WithStudentCapacityIs(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("student_capacity", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIsNot(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("student_capacity", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("student_capacity", converted))
	return r
}
func (r *SchoolRequest) WithStudentCapacityNotIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("student_capacity", converted))
	return r
}
func (r *SchoolRequest) WithStudentCapacityGreaterThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("student_capacity", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityGreaterThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("student_capacity", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityLessThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("student_capacity", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityLessThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("student_capacity", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityBetween(lower int64, upper int64) *SchoolRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("student_capacity", from, to))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("student_capacity"))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("student_capacity"))
	return r
}
func (r *SchoolRequest) OrderByStudentCapacityAsc() *SchoolRequest {
	r.Query.OrderAsc("student_capacity")
	return r
}
func (r *SchoolRequest) OrderByStudentCapacityDesc() *SchoolRequest {
	r.Query.OrderDesc("student_capacity")
	return r
}
func (r *SchoolRequest) SelectActive() *SchoolRequest {
	r.Query.Project("active")
	return r
}

func (r *SchoolRequest) WhichAreActive() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("active", core.ValBool(true)))
	return r
}
func (r *SchoolRequest) WhichAreNotActive() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("active", core.ValBool(false)))
	return r
}
func (r *SchoolRequest) OrderByActiveAsc() *SchoolRequest {
	r.Query.OrderAsc("active")
	return r
}
func (r *SchoolRequest) OrderByActiveDesc() *SchoolRequest {
	r.Query.OrderDesc("active")
	return r
}
func (r *SchoolRequest) SelectCreateTime() *SchoolRequest {
	r.Query.Project("create_time")
	return r
}

func (r *SchoolRequest) WithCreateTimeIs(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeIsNot(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *SchoolRequest) WithCreateTimeNotIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *SchoolRequest) WithCreateTimeGreaterThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeLessThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *SchoolRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *SchoolRequest) WithCreateTimeIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *SchoolRequest) WithCreateTimeIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *SchoolRequest) OrderByCreateTimeAsc() *SchoolRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *SchoolRequest) OrderByCreateTimeDesc() *SchoolRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *SchoolRequest) SelectUpdateTime() *SchoolRequest {
	r.Query.Project("update_time")
	return r
}

func (r *SchoolRequest) WithUpdateTimeIs(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIsNot(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *SchoolRequest) WithUpdateTimeNotIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *SchoolRequest) WithUpdateTimeGreaterThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeLessThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *SchoolRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *SchoolRequest) OrderByUpdateTimeAsc() *SchoolRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *SchoolRequest) OrderByUpdateTimeDesc() *SchoolRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *SchoolRequest) SelectVersion() *SchoolRequest {
	r.Query.Project("version")
	return r
}

func (r *SchoolRequest) WithVersionIs(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionIsNot(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *SchoolRequest) WithVersionNotIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *SchoolRequest) WithVersionGreaterThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionGreaterThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionLessThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionLessThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionBetween(lower int64, upper int64) *SchoolRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *SchoolRequest) WithVersionIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *SchoolRequest) WithVersionIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *SchoolRequest) OrderByVersionAsc() *SchoolRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *SchoolRequest) OrderByVersionDesc() *SchoolRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *SchoolRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *SchoolRequest {
	r.Query.Project("platform_id")
	r.Query.RelationQuery("platformEntity", child.GetQuery())
	r.relationFactories["platformEntity"] = child.NewRelationEntity
	return r
}
func (r *SchoolRequest) SelectSchoolTypeWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *SchoolRequest {
	r.Query.Project("school_type_id")
	r.Query.RelationQuery("schoolTypeEntity", child.GetQuery())
	r.relationFactories["schoolTypeEntity"] = child.NewRelationEntity
	return r
}

func (e *ExecutableSchoolRequest) NewEntity(context *runtime.UserContext) *School {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewSchool()
	entity.AttachEntityRoot(context.EntityRoot())
	initialized := context.InitializeEntity("School", entity)
	typed, ok := initialized.(*School)
	if !ok {
		panic("entity initializer changed School to an incompatible type")
	}
	return typed
}

func (e *ExecutableSchoolRequest) ExecuteForOne(context *runtime.UserContext) (*School, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableSchoolRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*School], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*School
	for _, rec := range rows {
		entity := NewSchool()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface{ AttachEntityRoot(*core.EntityRoot) }); ok {
						attachable.AttachEntityRoot(entity.EntityRoot())
					}
					if err := childEntity.FromRecord(childRecord); err != nil {
						return nil, err
					}
					entity.setRelationEntity("platformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["schoolTypeEntity"]; selected {
			entity.markRelationLoaded("schoolTypeEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["schoolTypeEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface{ AttachEntityRoot(*core.EntityRoot) }); ok {
						attachable.AttachEntityRoot(entity.EntityRoot())
					}
					if err := childEntity.FromRecord(childRecord); err != nil {
						return nil, err
					}
					entity.setRelationEntity("schoolTypeEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableSchoolRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*School], error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForPage()")
	}
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	r.Query.Page(offset, size).Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))
	authorized, err := context.PrepareQuery(r.Query)
	if err != nil {
		return nil, err
	}
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor")
	}
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	const countAlias = "__teaql_total"
	countRows, err := service.FetchAll(context, authorized.ForExactCount(countAlias))
	if err != nil {
		return nil, err
	}
	if len(countRows) != 1 {
		return nil, fmt.Errorf("exact count returned %d rows", len(countRows))
	}
	total, ok := countRows[0][countAlias].TryU64()
	if !ok {
		return nil, fmt.Errorf("exact count did not return an unsigned integer")
	}
	rows, err := service.FetchAll(context, authorized)
	if err != nil {
		return nil, err
	}
	results := make([]*School, 0, len(rows))
	for _, rec := range rows {
		entity := NewSchool()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface{ AttachEntityRoot(*core.EntityRoot) }); ok {
						attachable.AttachEntityRoot(entity.EntityRoot())
					}
					if err := childEntity.FromRecord(childRecord); err != nil {
						return nil, err
					}
					entity.setRelationEntity("platformEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["schoolTypeEntity"]; selected {
			entity.markRelationLoaded("schoolTypeEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["schoolTypeEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface{ AttachEntityRoot(*core.EntityRoot) }); ok {
						attachable.AttachEntityRoot(entity.EntityRoot())
					}
					if err := childEntity.FromRecord(childRecord); err != nil {
						return nil, err
					}
					entity.setRelationEntity("schoolTypeEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableSchoolRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*School) error) error {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForStream()")
	}
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	r.Query.Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req := &data_service.QueryRequest{Query: r.Query, TraceChain: r.Query.TraceChain, Comment: r.Query.CommentText}
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewSchool()
			entity.AttachEntityRoot(context.EntityRoot())
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

func (e *ExecutableSchoolRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	r.Query.Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))

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

func (r *SchoolRequest) Count() *SchoolRequest {
	return r.CountAs("count")
}

func (r *SchoolRequest) CountAs(alias string) *SchoolRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *SchoolRequest) MinStudentCapacity() *SchoolRequest {
	return r.MinStudentCapacityAs("minOfStudentCapacity")
}

func (r *SchoolRequest) MinStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Min("student_capacity", alias)
	return r
}
func (r *SchoolRequest) MaxStudentCapacity() *SchoolRequest {
	return r.MaxStudentCapacityAs("maxOfStudentCapacity")
}

func (r *SchoolRequest) MaxStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Max("student_capacity", alias)
	return r
}
func (r *SchoolRequest) SumStudentCapacity() *SchoolRequest {
	return r.SumStudentCapacityAs("sumOfStudentCapacity")
}

func (r *SchoolRequest) SumStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Sum("student_capacity", alias)
	return r
}
func (r *SchoolRequest) AvgStudentCapacity() *SchoolRequest {
	return r.AvgStudentCapacityAs("avgOfStudentCapacity")
}

func (r *SchoolRequest) AvgStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Avg("student_capacity", alias)
	return r
}
func (r *SchoolRequest) StddevStudentCapacity() *SchoolRequest {
	return r.StddevStudentCapacityAs("standardDeviationOfStudentCapacity")
}

func (r *SchoolRequest) StddevStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Stddev("student_capacity", alias)
	return r
}
func (r *SchoolRequest) StddevPopStudentCapacity() *SchoolRequest {
	return r.StddevPopStudentCapacityAs("squareRootOfPopulationStandardDeviationOfStudentCapacity")
}

func (r *SchoolRequest) StddevPopStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.StddevPop("student_capacity", alias)
	return r
}
func (r *SchoolRequest) VarSampStudentCapacity() *SchoolRequest {
	return r.VarSampStudentCapacityAs("sampleVarianceOfStudentCapacity")
}

func (r *SchoolRequest) VarSampStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.VarSamp("student_capacity", alias)
	return r
}
func (r *SchoolRequest) VarPopStudentCapacity() *SchoolRequest {
	return r.VarPopStudentCapacityAs("samplePopulationVarianceOfStudentCapacity")
}

func (r *SchoolRequest) VarPopStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.VarPop("student_capacity", alias)
	return r
}

func (r *SchoolRequest) GroupById() *SchoolRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *SchoolRequest) GroupByPlatform() *SchoolRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}
func (r *SchoolRequest) GroupBySchoolType() *SchoolRequest {
	r.Query.WithGroupBy("school_type_id")
	return r
}
func (r *SchoolRequest) GroupByName() *SchoolRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *SchoolRequest) GroupByAddress() *SchoolRequest {
	r.Query.WithGroupBy("address")
	return r
}
func (r *SchoolRequest) GroupByEstablishedDate() *SchoolRequest {
	r.Query.WithGroupBy("established_date")
	return r
}
func (r *SchoolRequest) GroupByStudentCapacity() *SchoolRequest {
	r.Query.WithGroupBy("student_capacity")
	return r
}
func (r *SchoolRequest) GroupByActive() *SchoolRequest {
	r.Query.WithGroupBy("active")
	return r
}
func (r *SchoolRequest) GroupByCreateTime() *SchoolRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *SchoolRequest) GroupByUpdateTime() *SchoolRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *SchoolRequest) GroupByVersion() *SchoolRequest {
	r.Query.WithGroupBy("version")
	return r
}

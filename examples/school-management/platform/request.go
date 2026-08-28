package platform

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"school-management-service-core-workspace/lib/school"
	"school-management-service-core-workspace/lib/school_type"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type PlatformRequest struct {
	Query             *core.SelectQuery
	purposeText       string
	commentText       string
	relationFactories map[string]func() core.Entity
}

type ExecutablePlatformRequest struct {
	request *PlatformRequest
}

func NewPlatformRequest() *PlatformRequest {
	r := &PlatformRequest{
		Query:             core.NewSelectQuery("Platform"),
		relationFactories: make(map[string]func() core.Entity),
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

func (r *PlatformRequest) NewRelationEntity() core.Entity {
	return NewPlatform()
}

func (r *PlatformRequest) Comment(comment string) *PlatformRequest {
	r.commentText = comment
	return r
}

func (r *PlatformRequest) Purpose(purpose string) *ExecutablePlatformRequest {
	r.purposeText = purpose
	return &ExecutablePlatformRequest{request: r}
}

func (r *ExecutablePlatformRequest) Comment(comment string) *ExecutablePlatformRequest {
	r.request.commentText = comment
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

func removePlatformVersionFilter(expr *core.Expr) *core.Expr {
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
		if kept := removePlatformVersionFilter(part); kept != nil {
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
	r.Query.Project("id")
	return r
}

func (r *PlatformRequest) WithIdIs(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIsNot(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *PlatformRequest) WithIdNotIn(values []uint64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *PlatformRequest) WithIdGreaterThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdGreaterThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThan(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdLessThanOrEqualTo(value uint64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *PlatformRequest) WithIdBetween(lower uint64, upper uint64) *PlatformRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *PlatformRequest) WithIdIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *PlatformRequest) WithIdIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *PlatformRequest) OrderByIdAsc() *PlatformRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PlatformRequest) OrderByIdDesc() *PlatformRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *PlatformRequest) SelectName() *PlatformRequest {
	r.Query.Project("name")
	return r
}

func (r *PlatformRequest) WithNameIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *PlatformRequest) WithNameNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *PlatformRequest) WithNameGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithNameBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("name", from, to))
	return r
}
func (r *PlatformRequest) WithNameIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("name"))
	return r
}
func (r *PlatformRequest) WithNameIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("name"))
	return r
}
func (r *PlatformRequest) WithNameContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *PlatformRequest) WithNameStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *PlatformRequest) WithNameNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("name", term))
	return r
}
func (r *PlatformRequest) OrderByNameAsc() *PlatformRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *PlatformRequest) OrderByNameDesc() *PlatformRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *PlatformRequest) SelectBaseUrl() *PlatformRequest {
	r.Query.Project("base_url")
	return r
}

func (r *PlatformRequest) WithBaseUrlIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("base_url", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("base_url", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("base_url", converted))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("base_url", converted))
	return r
}
func (r *PlatformRequest) WithBaseUrlGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("base_url", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("base_url", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("base_url", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("base_url", core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("base_url", from, to))
	return r
}
func (r *PlatformRequest) WithBaseUrlIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("base_url"))
	return r
}
func (r *PlatformRequest) WithBaseUrlIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("base_url"))
	return r
}
func (r *PlatformRequest) WithBaseUrlContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain("base_url", term))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain("base_url", term))
	return r
}
func (r *PlatformRequest) WithBaseUrlStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith("base_url", term))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("base_url", term))
	return r
}
func (r *PlatformRequest) WithBaseUrlEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith("base_url", term))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith("base_url", term))
	return r
}
func (r *PlatformRequest) OrderByBaseUrlAsc() *PlatformRequest {
	r.Query.OrderAsc("base_url")
	return r
}
func (r *PlatformRequest) OrderByBaseUrlDesc() *PlatformRequest {
	r.Query.OrderDesc("base_url")
	return r
}
func (r *PlatformRequest) SelectCreateTime() *PlatformRequest {
	r.Query.Project("create_time")
	return r
}

func (r *PlatformRequest) WithCreateTimeIs(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeIsNot(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *PlatformRequest) WithCreateTimeNotIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *PlatformRequest) WithCreateTimeGreaterThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeLessThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *PlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *PlatformRequest) WithCreateTimeIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *PlatformRequest) WithCreateTimeIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *PlatformRequest) OrderByCreateTimeAsc() *PlatformRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *PlatformRequest) OrderByCreateTimeDesc() *PlatformRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *PlatformRequest) SelectUpdateTime() *PlatformRequest {
	r.Query.Project("update_time")
	return r
}

func (r *PlatformRequest) WithUpdateTimeIs(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIsNot(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *PlatformRequest) WithUpdateTimeNotIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *PlatformRequest) WithUpdateTimeGreaterThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeLessThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *PlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *PlatformRequest) OrderByUpdateTimeAsc() *PlatformRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *PlatformRequest) OrderByUpdateTimeDesc() *PlatformRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *PlatformRequest) SelectVersion() *PlatformRequest {
	r.Query.Project("version")
	return r
}

func (r *PlatformRequest) WithVersionIs(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIsNot(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *PlatformRequest) WithVersionNotIn(values []int64) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionGreaterThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThan(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionLessThanOrEqualTo(value int64) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *PlatformRequest) WithVersionBetween(lower int64, upper int64) *PlatformRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *PlatformRequest) WithVersionIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *PlatformRequest) WithVersionIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *PlatformRequest) OrderByVersionAsc() *PlatformRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PlatformRequest) OrderByVersionDesc() *PlatformRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *PlatformRequest) CountSchoolTypes() *PlatformRequest {
	r.Query.Count("count_school_types")
	return r
}
func (r *PlatformRequest) CountSchools() *PlatformRequest {
	r.Query.Count("count_schools")
	return r
}

func (r *PlatformRequest) SelectSchoolTypeList() *PlatformRequest {
	return r.SelectSchoolTypeListWith(school_type.NewSchoolTypeRequest())
}

func (r *PlatformRequest) SelectSchoolTypeListWith(child *school_type.SchoolTypeRequest) *PlatformRequest {
	r.Query.RelationQuery("schoolTypeList", child.Query)
	return r
}
func (r *PlatformRequest) SelectSchoolList() *PlatformRequest {
	return r.SelectSchoolListWith(school.NewSchoolRequest())
}

func (r *PlatformRequest) SelectSchoolListWith(child *school.SchoolRequest) *PlatformRequest {
	r.Query.RelationQuery("schoolList", child.Query)
	return r
}

func (e *ExecutablePlatformRequest) NewEntity(context *runtime.UserContext) *Platform {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewPlatform()
	entity.AttachEntityRoot(context.EntityRoot())
	initialized := context.InitializeEntity("Platform", entity)
	typed, ok := initialized.(*Platform)
	if !ok {
		panic("entity initializer changed Platform to an incompatible type")
	}
	return typed
}

func (e *ExecutablePlatformRequest) ExecuteForOne(context *runtime.UserContext) (*Platform, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutablePlatformRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*Platform], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*Platform
	for _, rec := range rows {
		entity := NewPlatform()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["schoolTypeList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation schoolTypeList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := school_type.NewSchoolType()
				childEntity.AttachEntityRoot(entity.EntityRoot())
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.SchoolTypeList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["schoolList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation schoolList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := school.NewSchool()
				childEntity.AttachEntityRoot(entity.EntityRoot())
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.SchoolList().Add(childEntity)
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutablePlatformRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*Platform], error) {
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
	results := make([]*Platform, 0, len(rows))
	for _, rec := range rows {
		entity := NewPlatform()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["schoolTypeList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation schoolTypeList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := school_type.NewSchoolType()
				childEntity.AttachEntityRoot(entity.EntityRoot())
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.SchoolTypeList().Add(childEntity)
			}
		}
		if relationValue, selected := rec["schoolList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
			if !ok {
				return nil, fmt.Errorf("relation schoolList has unexpected runtime type %T", relationValue.V)
			}
			for _, childRecord := range childRecords {
				childEntity := school.NewSchool()
				childEntity.AttachEntityRoot(entity.EntityRoot())
				if err := childEntity.FromRecord(childRecord); err != nil {
					return nil, err
				}
				entity.SchoolList().Add(childEntity)
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutablePlatformRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*Platform) error) error {
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
			entity := NewPlatform()
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

func (e *ExecutablePlatformRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
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

func (r *PlatformRequest) Count() *PlatformRequest {
	return r.CountAs("count")
}

func (r *PlatformRequest) CountAs(alias string) *PlatformRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *PlatformRequest) GroupById() *PlatformRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *PlatformRequest) GroupByName() *PlatformRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *PlatformRequest) GroupByBaseUrl() *PlatformRequest {
	r.Query.WithGroupBy("base_url")
	return r
}
func (r *PlatformRequest) GroupByCreateTime() *PlatformRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *PlatformRequest) GroupByUpdateTime() *PlatformRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *PlatformRequest) GroupByVersion() *PlatformRequest {
	r.Query.WithGroupBy("version")
	return r
}

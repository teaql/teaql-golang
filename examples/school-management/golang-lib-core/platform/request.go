

package platform

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"school-management-service-core-workspace/lib/internal/generatedsearch"
	"school-management-service-core-workspace/lib/school_type"
	"school-management-service-core-workspace/lib/school"
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
		Query: core.NewSelectQuery("Platform"),
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
	if relationValue, selected := rec["schoolTypeList"]; selected {
		entity.markRelationLoaded("schoolTypeList")
		factory := r.relationFactories["schoolTypeList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation schoolTypeList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation schoolTypeList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*school_type.SchoolType)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation schoolTypeList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.SchoolTypeList().Add(childEntity)
			}}
	if relationValue, selected := rec["schoolList"]; selected {
		entity.markRelationLoaded("schoolList")
		factory := r.relationFactories["schoolList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation schoolList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation schoolList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*school.School)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation schoolList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.SchoolList().Add(childEntity)
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
func (r *PlatformRequest) SelectBaseUrl() *PlatformRequest {
	r.Query.Project(teaqlBaseUrlStorageKey)
	return r
}

func (r *PlatformRequest) WithBaseUrlIs(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq(teaqlBaseUrlStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlIsNot(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe(teaqlBaseUrlStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlBaseUrlStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotIn(values []string) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlBaseUrlStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithBaseUrlGreaterThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt(teaqlBaseUrlStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlGreaterThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte(teaqlBaseUrlStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlLessThan(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt(teaqlBaseUrlStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlLessThanOrEqualTo(value string) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte(teaqlBaseUrlStorageKey, core.ValText(value)))
	return r
}
func (r *PlatformRequest) WithBaseUrlBetween(lower string, upper string) *PlatformRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlBaseUrlStorageKey, from, to))
	return r
}
func (r *PlatformRequest) WithBaseUrlIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlBaseUrlStorageKey))
	return r
}
func (r *PlatformRequest) WithBaseUrlIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlBaseUrlStorageKey))
	return r
}
func (r *PlatformRequest) WithBaseUrlContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprContain(teaqlBaseUrlStorageKey, term))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotContaining(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlBaseUrlStorageKey, term))
	return r
}
func (r *PlatformRequest) WithBaseUrlStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlBaseUrlStorageKey, term))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotStartingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlBaseUrlStorageKey, term))
	return r
}
func (r *PlatformRequest) WithBaseUrlEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlBaseUrlStorageKey, term))
	return r
}
func (r *PlatformRequest) WithBaseUrlNotEndingWith(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlBaseUrlStorageKey, term))
	return r
}
func (r *PlatformRequest) WithBaseUrlSoundingLike(term string) *PlatformRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlBaseUrlStorageKey, core.ValText(term)))
	return r
}
func (r *PlatformRequest) OrderByBaseUrlAsc() *PlatformRequest {
	r.Query.OrderAsc(teaqlBaseUrlStorageKey)
	return r
}
func (r *PlatformRequest) OrderByBaseUrlDesc() *PlatformRequest {
	r.Query.OrderDesc(teaqlBaseUrlStorageKey)
	return r
}
func (r *PlatformRequest) SelectCreateTime() *PlatformRequest {
	r.Query.Project(teaqlCreateTimeStorageKey)
	return r
}

func (r *PlatformRequest) WithCreateTimeIs(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeIsNot(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCreateTimeStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithCreateTimeNotIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCreateTimeStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithCreateTimeGreaterThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeLessThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *PlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCreateTimeStorageKey, from, to))
	return r
}
func (r *PlatformRequest) WithCreateTimeIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *PlatformRequest) WithCreateTimeIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *PlatformRequest) OrderByCreateTimeAsc() *PlatformRequest {
	r.Query.OrderAsc(teaqlCreateTimeStorageKey)
	return r
}
func (r *PlatformRequest) OrderByCreateTimeDesc() *PlatformRequest {
	r.Query.OrderDesc(teaqlCreateTimeStorageKey)
	return r
}
func (r *PlatformRequest) SelectUpdateTime() *PlatformRequest {
	r.Query.Project(teaqlUpdateTimeStorageKey)
	return r
}

func (r *PlatformRequest) WithUpdateTimeIs(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprEq(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIsNot(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprNe(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlUpdateTimeStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithUpdateTimeNotIn(values []time.Time) *PlatformRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlUpdateTimeStorageKey, converted))
	return r
}
func (r *PlatformRequest) WithUpdateTimeGreaterThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGt(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprGte(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeLessThan(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLt(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *PlatformRequest {
	r.Query.AndFilter(core.ExprLte(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PlatformRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *PlatformRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlUpdateTimeStorageKey, from, to))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIsKnown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *PlatformRequest) WithUpdateTimeIsUnknown() *PlatformRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *PlatformRequest) OrderByUpdateTimeAsc() *PlatformRequest {
	r.Query.OrderAsc(teaqlUpdateTimeStorageKey)
	return r
}
func (r *PlatformRequest) OrderByUpdateTimeDesc() *PlatformRequest {
	r.Query.OrderDesc(teaqlUpdateTimeStorageKey)
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



func (r *PlatformRequest) CountSchoolTypes() *PlatformRequest {
	return r.CountSchoolTypesAs("countSchoolTypes")

}
func (r *PlatformRequest) CountSchoolTypesAs(alias string) *PlatformRequest {
	r.CountSchoolTypesWith(alias, school_type.NewSchoolTypeRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	return r
}
func (r *PlatformRequest) CountSchoolTypesWith(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountSchoolTypesWith", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PlatformRequest) SumDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.SumDisplayOrderOfSchoolTypesAs("sumDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) SumDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.MinDisplayOrderOfSchoolTypesAs("minDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) MinDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.MaxDisplayOrderOfSchoolTypesAs("maxDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) MaxDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) AvgDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.AvgDisplayOrderOfSchoolTypesAs("avgDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) AvgDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) StandardDeviationDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.StandardDeviationDisplayOrderOfSchoolTypesAs("standardDeviationDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) StandardDeviationDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationDisplayOrderOfSchoolTypesAs("squareRootOfPopulationStandardDeviationDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SampleVarianceDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.SampleVarianceDisplayOrderOfSchoolTypesAs("sampleVarianceDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) SampleVarianceDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceDisplayOrderOfSchoolTypes() *PlatformRequest {
	return r.SamplePopulationVarianceDisplayOrderOfSchoolTypesAs("samplePopulationVarianceDisplayOrderOfSchoolTypes", school_type.NewSchoolTypeRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceDisplayOrderOfSchoolTypesAs(alias string, child *school_type.SchoolTypeRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("display_order", alias)
	aggregate := core.NewRelationAggregate("schoolTypeList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceDisplayOrderOfSchoolTypesAs", Alias: &alias, Relation: "schoolTypeList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) CountSchools() *PlatformRequest {
	return r.CountSchoolsAs("countSchools")

}
func (r *PlatformRequest) CountSchoolsAs(alias string) *PlatformRequest {
	r.CountSchoolsWith(alias, school.NewSchoolRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountSchoolsAs", Alias: &alias, Relation: "schoolList"}
	return r
}
func (r *PlatformRequest) CountSchoolsWith(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountSchoolsWith", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PlatformRequest) MinEstablishedDateOfSchools() *PlatformRequest {
	return r.MinEstablishedDateOfSchoolsAs("minEstablishedDateOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MinEstablishedDateOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("established_date", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinEstablishedDateOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxEstablishedDateOfSchools() *PlatformRequest {
	return r.MaxEstablishedDateOfSchoolsAs("maxEstablishedDateOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MaxEstablishedDateOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("established_date", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxEstablishedDateOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SumStudentCapacityOfSchools() *PlatformRequest {
	return r.SumStudentCapacityOfSchoolsAs("sumStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) SumStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinStudentCapacityOfSchools() *PlatformRequest {
	return r.MinStudentCapacityOfSchoolsAs("minStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MinStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxStudentCapacityOfSchools() *PlatformRequest {
	return r.MaxStudentCapacityOfSchoolsAs("maxStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MaxStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) AvgStudentCapacityOfSchools() *PlatformRequest {
	return r.AvgStudentCapacityOfSchoolsAs("avgStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) AvgStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) StandardDeviationStudentCapacityOfSchools() *PlatformRequest {
	return r.StandardDeviationStudentCapacityOfSchoolsAs("standardDeviationStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) StandardDeviationStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationStudentCapacityOfSchools() *PlatformRequest {
	return r.SquareRootOfPopulationStandardDeviationStudentCapacityOfSchoolsAs("squareRootOfPopulationStandardDeviationStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) SquareRootOfPopulationStandardDeviationStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SampleVarianceStudentCapacityOfSchools() *PlatformRequest {
	return r.SampleVarianceStudentCapacityOfSchoolsAs("sampleVarianceStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) SampleVarianceStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) SamplePopulationVarianceStudentCapacityOfSchools() *PlatformRequest {
	return r.SamplePopulationVarianceStudentCapacityOfSchoolsAs("samplePopulationVarianceStudentCapacityOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) SamplePopulationVarianceStudentCapacityOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("student_capacity", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceStudentCapacityOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinCreateTimeOfSchools() *PlatformRequest {
	return r.MinCreateTimeOfSchoolsAs("minCreateTimeOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MinCreateTimeOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxCreateTimeOfSchools() *PlatformRequest {
	return r.MaxCreateTimeOfSchoolsAs("maxCreateTimeOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MaxCreateTimeOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MinUpdateTimeOfSchools() *PlatformRequest {
	return r.MinUpdateTimeOfSchoolsAs("minUpdateTimeOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MinUpdateTimeOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PlatformRequest) MaxUpdateTimeOfSchools() *PlatformRequest {
	return r.MaxUpdateTimeOfSchoolsAs("maxUpdateTimeOfSchools", school.NewSchoolRequest())
}
func (r *PlatformRequest) MaxUpdateTimeOfSchoolsAs(alias string, child *school.SchoolRequest) *PlatformRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("schoolList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfSchoolsAs", Alias: &alias, Relation: "schoolList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PlatformRequest) SelectSchoolTypeList() *PlatformRequest {
	return r.SelectSchoolTypeListWith(school_type.NewSchoolTypeRequest())
}

func (r *PlatformRequest) SelectSchoolTypeListWith(child *school_type.SchoolTypeRequest) *PlatformRequest {
	r.Query.RelationQuerySelection("schoolTypeList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectSchoolTypeListWith", Relation: "schoolTypeList"}
	r.relationFactories["schoolTypeList"] = child.TeaqlHydrateQueryRecord
	return r
}
func (r *PlatformRequest) SelectSchoolList() *PlatformRequest {
	return r.SelectSchoolListWith(school.NewSchoolRequest())
}

func (r *PlatformRequest) SelectSchoolListWith(child *school.SchoolRequest) *PlatformRequest {
	r.Query.RelationQuerySelection("schoolList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectSchoolListWith", Relation: "schoolList"}
	r.relationFactories["schoolList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *PlatformRequest) HaveSchoolTypes() *PlatformRequest {
	return r.WithSchoolTypeListMatching(school_type.NewSchoolTypeRequest())
}

func (r *PlatformRequest) HaveNoSchoolTypes() *PlatformRequest {
	return r.WithoutSchoolTypeListMatching(school_type.NewSchoolTypeRequest())
}

func (r *PlatformRequest) WithSchoolTypeListMatching(child *school_type.SchoolTypeRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutSchoolTypeListMatching(child *school_type.SchoolTypeRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}
func (r *PlatformRequest) HaveSchools() *PlatformRequest {
	return r.WithSchoolListMatching(school.NewSchoolRequest())
}

func (r *PlatformRequest) HaveNoSchools() *PlatformRequest {
	return r.WithoutSchoolListMatching(school.NewSchoolRequest())
}

func (r *PlatformRequest) WithSchoolListMatching(child *school.SchoolRequest) *PlatformRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "platform_id"))
	return r
}

func (r *PlatformRequest) WithoutSchoolListMatching(child *school.SchoolRequest) *PlatformRequest {
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
func (r *PlatformRequest) GroupByBaseUrl() *PlatformRequest {
	r.Query.WithGroupBy(teaqlBaseUrlStorageKey)
	return r
}

func (r *PlatformRequest) GroupByBaseUrlAs(alias string) *PlatformRequest {
	r.Query.WithGroupBy(teaqlBaseUrlStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlBaseUrlStorageKey))
	return r
}
func (r *PlatformRequest) GroupByCreateTime() *PlatformRequest {
	r.Query.WithGroupBy(teaqlCreateTimeStorageKey)
	return r
}

func (r *PlatformRequest) GroupByCreateTimeAs(alias string) *PlatformRequest {
	r.Query.WithGroupBy(teaqlCreateTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *PlatformRequest) GroupByUpdateTime() *PlatformRequest {
	r.Query.WithGroupBy(teaqlUpdateTimeStorageKey)
	return r
}

func (r *PlatformRequest) GroupByUpdateTimeAs(alias string) *PlatformRequest {
	r.Query.WithGroupBy(teaqlUpdateTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlUpdateTimeStorageKey))
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

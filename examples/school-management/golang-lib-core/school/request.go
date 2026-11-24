

package school

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"school-management-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type SchoolRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableSchoolRequest struct {
	request *SchoolRequest
}

func NewSchoolRequest() *SchoolRequest {
	r := &SchoolRequest{
		Query: core.NewSelectQuery("School"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
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

// FilterByJson validates all clauses before changing this request.
func (r *SchoolRequest) FilterByJson(source []byte) (*SchoolRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *SchoolRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *SchoolRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *SchoolRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedSchool().EntityDescriptor()
}

func (r *SchoolRequest) NewRelationEntity() core.Entity {
	return newLoadedSchool()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *SchoolRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *SchoolRequest) teaqlHydrateQueryRecord(rec core.Record) (*School, error) {
	entity := newLoadedSchool()
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
	if relationValue, selected := rec["schoolTypeEntity"]; selected {
		entity.markRelationLoaded("schoolTypeEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["schoolTypeEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("schoolTypeEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *SchoolRequest) Comment(comment string) *SchoolRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *SchoolRequest) Purpose(purpose string) *ExecutableSchoolRequest {
	r.purposeText = purpose
	return &ExecutableSchoolRequest{request: r}
}

func (r *ExecutableSchoolRequest) Comment(comment string) *ExecutableSchoolRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
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

func (r *SchoolRequest) OptimizePaginationWithIDSet() *SchoolRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *SchoolRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *SchoolRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *SchoolRequest) TopNProbeParentThreshold(threshold uint64) *SchoolRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeSchoolVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeSchoolVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
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
	r.Query.Project(teaqlIdStorageKey)
	return r
}

func (r *SchoolRequest) WithIdIs(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdIsNot(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlIdStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithIdNotIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlIdStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithIdGreaterThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdGreaterThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdLessThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdLessThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlIdStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithIdBetween(lower uint64, upper uint64) *SchoolRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlIdStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithIdIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlIdStorageKey))
	return r
}
func (r *SchoolRequest) WithIdIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlIdStorageKey))
	return r
}
func (r *SchoolRequest) OrderByIdAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlIdStorageKey)
	return r
}
func (r *SchoolRequest) OrderByIdDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlIdStorageKey)
	return r
}
func (r *SchoolRequest) SelectPlatform() *SchoolRequest {
	r.Query.Project(teaqlPlatformStorageKey)
	return r
}

func (r *SchoolRequest) WithPlatformIs(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformIsNot(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithPlatformNotIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlPlatformStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithPlatformGreaterThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformLessThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformLessThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlPlatformStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithPlatformBetween(lower uint64, upper uint64) *SchoolRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlPlatformStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithPlatformIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *SchoolRequest) WithPlatformIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlPlatformStorageKey))
	return r
}
func (r *SchoolRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *SchoolRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlPlatformStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *SchoolRequest) OrderByPlatformAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlPlatformStorageKey)
	return r
}
func (r *SchoolRequest) OrderByPlatformDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlPlatformStorageKey)
	return r
}
func (r *SchoolRequest) SelectSchoolType() *SchoolRequest {
	r.Query.Project(teaqlSchoolTypeStorageKey)
	return r
}

func (r *SchoolRequest) WithSchoolTypeIs(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlSchoolTypeStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsNot(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlSchoolTypeStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlSchoolTypeStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithSchoolTypeNotIn(values []uint64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlSchoolTypeStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithSchoolTypeGreaterThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlSchoolTypeStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeGreaterThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlSchoolTypeStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeLessThan(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlSchoolTypeStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeLessThanOrEqualTo(value uint64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlSchoolTypeStorageKey, core.ValU64(value)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeBetween(lower uint64, upper uint64) *SchoolRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlSchoolTypeStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlSchoolTypeStorageKey))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlSchoolTypeStorageKey))
	return r
}
func (r *SchoolRequest) FacetBySchoolTypeAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *SchoolRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, teaqlSchoolTypeStorageKey, core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetBySchoolTypeAs", Alias: &name, Relation: "schoolType"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsPrimary() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlSchoolTypeStorageKey, core.ValU64(1001)))
	return r
}
func (r *SchoolRequest) WithSchoolTypeIsSecondary() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlSchoolTypeStorageKey, core.ValU64(1002)))
	return r
}
func (r *SchoolRequest) OrderBySchoolTypeAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlSchoolTypeStorageKey)
	return r
}
func (r *SchoolRequest) OrderBySchoolTypeDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlSchoolTypeStorageKey)
	return r
}
func (r *SchoolRequest) SelectName() *SchoolRequest {
	r.Query.Project(teaqlNameStorageKey)
	return r
}

func (r *SchoolRequest) WithNameIs(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameIsNot(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlNameStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithNameNotIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlNameStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithNameGreaterThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameGreaterThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameLessThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameLessThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlNameStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithNameBetween(lower string, upper string) *SchoolRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlNameStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithNameIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlNameStorageKey))
	return r
}
func (r *SchoolRequest) WithNameIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlNameStorageKey))
	return r
}
func (r *SchoolRequest) WithNameContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprContain(teaqlNameStorageKey, term))
	return r
}
func (r *SchoolRequest) WithNameNotContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlNameStorageKey, term))
	return r
}
func (r *SchoolRequest) WithNameStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *SchoolRequest) WithNameNotStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlNameStorageKey, term))
	return r
}
func (r *SchoolRequest) WithNameEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *SchoolRequest) WithNameNotEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlNameStorageKey, term))
	return r
}
func (r *SchoolRequest) WithNameSoundingLike(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlNameStorageKey, core.ValText(term)))
	return r
}
func (r *SchoolRequest) OrderByNameAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlNameStorageKey)
	return r
}
func (r *SchoolRequest) OrderByNameDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlNameStorageKey)
	return r
}
func (r *SchoolRequest) SelectAddress() *SchoolRequest {
	r.Query.Project(teaqlAddressStorageKey)
	return r
}

func (r *SchoolRequest) WithAddressIs(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlAddressStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressIsNot(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlAddressStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlAddressStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithAddressNotIn(values []string) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlAddressStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithAddressGreaterThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlAddressStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressGreaterThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlAddressStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressLessThan(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlAddressStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressLessThanOrEqualTo(value string) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlAddressStorageKey, core.ValText(value)))
	return r
}
func (r *SchoolRequest) WithAddressBetween(lower string, upper string) *SchoolRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlAddressStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithAddressIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlAddressStorageKey))
	return r
}
func (r *SchoolRequest) WithAddressIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlAddressStorageKey))
	return r
}
func (r *SchoolRequest) WithAddressContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprContain(teaqlAddressStorageKey, term))
	return r
}
func (r *SchoolRequest) WithAddressNotContaining(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotContain(teaqlAddressStorageKey, term))
	return r
}
func (r *SchoolRequest) WithAddressStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprBeginWith(teaqlAddressStorageKey, term))
	return r
}
func (r *SchoolRequest) WithAddressNotStartingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotBeginWith(teaqlAddressStorageKey, term))
	return r
}
func (r *SchoolRequest) WithAddressEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprEndWith(teaqlAddressStorageKey, term))
	return r
}
func (r *SchoolRequest) WithAddressNotEndingWith(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotEndWith(teaqlAddressStorageKey, term))
	return r
}
func (r *SchoolRequest) WithAddressSoundingLike(term string) *SchoolRequest {
	r.Query.AndFilter(core.ExprSoundLike(teaqlAddressStorageKey, core.ValText(term)))
	return r
}
func (r *SchoolRequest) OrderByAddressAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlAddressStorageKey)
	return r
}
func (r *SchoolRequest) OrderByAddressDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlAddressStorageKey)
	return r
}
func (r *SchoolRequest) SelectEstablishedDate() *SchoolRequest {
	r.Query.Project(teaqlEstablishedDateStorageKey)
	return r
}

func (r *SchoolRequest) WithEstablishedDateIs(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlEstablishedDateStorageKey, core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIsNot(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlEstablishedDateStorageKey, core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlEstablishedDateStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithEstablishedDateNotIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlEstablishedDateStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithEstablishedDateGreaterThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlEstablishedDateStorageKey, core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateGreaterThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlEstablishedDateStorageKey, core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateLessThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlEstablishedDateStorageKey, core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateLessThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlEstablishedDateStorageKey, core.ValDate(value)))
	return r
}
func (r *SchoolRequest) WithEstablishedDateBetween(lower time.Time, upper time.Time) *SchoolRequest {
	value := lower
	from := core.ValDate(value)
	value = upper
	to := core.ValDate(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlEstablishedDateStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlEstablishedDateStorageKey))
	return r
}
func (r *SchoolRequest) WithEstablishedDateIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlEstablishedDateStorageKey))
	return r
}
func (r *SchoolRequest) OrderByEstablishedDateAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlEstablishedDateStorageKey)
	return r
}
func (r *SchoolRequest) OrderByEstablishedDateDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlEstablishedDateStorageKey)
	return r
}
func (r *SchoolRequest) SelectStudentCapacity() *SchoolRequest {
	r.Query.Project(teaqlStudentCapacityStorageKey)
	return r
}

func (r *SchoolRequest) WithStudentCapacityIs(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlStudentCapacityStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIsNot(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlStudentCapacityStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlStudentCapacityStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithStudentCapacityNotIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlStudentCapacityStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithStudentCapacityGreaterThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlStudentCapacityStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityGreaterThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlStudentCapacityStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityLessThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlStudentCapacityStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityLessThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlStudentCapacityStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithStudentCapacityBetween(lower int64, upper int64) *SchoolRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlStudentCapacityStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlStudentCapacityStorageKey))
	return r
}
func (r *SchoolRequest) WithStudentCapacityIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlStudentCapacityStorageKey))
	return r
}
func (r *SchoolRequest) OrderByStudentCapacityAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlStudentCapacityStorageKey)
	return r
}
func (r *SchoolRequest) OrderByStudentCapacityDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlStudentCapacityStorageKey)
	return r
}
func (r *SchoolRequest) SelectActive() *SchoolRequest {
	r.Query.Project(teaqlActiveStorageKey)
	return r
}

func (r *SchoolRequest) WhichAreActive() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlActiveStorageKey, core.ValBool(true)))
	return r
}
func (r *SchoolRequest) WhichAreNotActive() *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlActiveStorageKey, core.ValBool(false)))
	return r
}
func (r *SchoolRequest) OrderByActiveAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlActiveStorageKey)
	return r
}
func (r *SchoolRequest) OrderByActiveDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlActiveStorageKey)
	return r
}
func (r *SchoolRequest) SelectCreateTime() *SchoolRequest {
	r.Query.Project(teaqlCreateTimeStorageKey)
	return r
}

func (r *SchoolRequest) WithCreateTimeIs(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeIsNot(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlCreateTimeStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithCreateTimeNotIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlCreateTimeStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithCreateTimeGreaterThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeLessThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlCreateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *SchoolRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlCreateTimeStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithCreateTimeIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *SchoolRequest) WithCreateTimeIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *SchoolRequest) OrderByCreateTimeAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlCreateTimeStorageKey)
	return r
}
func (r *SchoolRequest) OrderByCreateTimeDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlCreateTimeStorageKey)
	return r
}
func (r *SchoolRequest) SelectUpdateTime() *SchoolRequest {
	r.Query.Project(teaqlUpdateTimeStorageKey)
	return r
}

func (r *SchoolRequest) WithUpdateTimeIs(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIsNot(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList(teaqlUpdateTimeStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithUpdateTimeNotIn(values []time.Time) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlUpdateTimeStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithUpdateTimeGreaterThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeLessThan(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlUpdateTimeStorageKey, core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *SchoolRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *SchoolRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode(teaqlUpdateTimeStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *SchoolRequest) WithUpdateTimeIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *SchoolRequest) OrderByUpdateTimeAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlUpdateTimeStorageKey)
	return r
}
func (r *SchoolRequest) OrderByUpdateTimeDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlUpdateTimeStorageKey)
	return r
}
func (r *SchoolRequest) SelectVersion() *SchoolRequest {
	r.Query.Project(teaqlVersionStorageKey)
	return r
}

func (r *SchoolRequest) WithVersionIs(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprEq(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionIsNot(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprNe(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithVersionNotIn(values []int64) *SchoolRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList(teaqlVersionStorageKey, converted))
	return r
}
func (r *SchoolRequest) WithVersionGreaterThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionGreaterThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprGte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionLessThan(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLt(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionLessThanOrEqualTo(value int64) *SchoolRequest {
	r.Query.AndFilter(core.ExprLte(teaqlVersionStorageKey, core.ValI64(value)))
	return r
}
func (r *SchoolRequest) WithVersionBetween(lower int64, upper int64) *SchoolRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode(teaqlVersionStorageKey, from, to))
	return r
}
func (r *SchoolRequest) WithVersionIsKnown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode(teaqlVersionStorageKey))
	return r
}
func (r *SchoolRequest) WithVersionIsUnknown() *SchoolRequest {
	r.Query.AndFilter(core.ExprIsNullNode(teaqlVersionStorageKey))
	return r
}
func (r *SchoolRequest) OrderByVersionAsc() *SchoolRequest {
	r.Query.OrderAsc(teaqlVersionStorageKey)
	return r
}
func (r *SchoolRequest) OrderByVersionDesc() *SchoolRequest {
	r.Query.OrderDesc(teaqlVersionStorageKey)
	return r
}

func (r *SchoolRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *SchoolRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlPlatformStorageKey)
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
func (r *SchoolRequest) SelectSchoolTypeWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *SchoolRequest {
	runtime.EnsureRelationProjection(r.Query, teaqlSchoolTypeStorageKey)
	r.Query.RelationQuerySelection("schoolTypeEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectSchoolTypeWith", Relation: "schoolTypeEntity"}
	r.relationFactories["schoolTypeEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *SchoolRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *SchoolRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *SchoolRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlPlatformStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}
func (r *SchoolRequest) WithSchoolTypeMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *SchoolRequest {
	r.Query.AndFilter(core.ExprInSubQuery(teaqlSchoolTypeStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *SchoolRequest) WithoutSchoolTypeMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *SchoolRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery(teaqlSchoolTypeStorageKey, child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableSchoolRequest) NewEntity(context *runtime.UserContext) (*School, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewSchool(context)
}

func (e *ExecutableSchoolRequest) ExecuteForOne(context *runtime.UserContext) (*School, error) {
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

func (e *ExecutableSchoolRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*School], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*School, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableSchoolRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*School) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*School, error) {
		entity := newLoadedSchool()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableSchoolRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableSchoolRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableSchoolRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *SchoolRequest) Count() *SchoolRequest {
	return r.CountAs("count")
}

func (r *SchoolRequest) CountAs(alias string) *SchoolRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *SchoolRequest) MinStudentCapacity() *SchoolRequest {
	return r.MinStudentCapacityAs("minOfStudentCapacity")
}

func (r *SchoolRequest) MinStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Min(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) MaxStudentCapacity() *SchoolRequest {
	return r.MaxStudentCapacityAs("maxOfStudentCapacity")
}

func (r *SchoolRequest) MaxStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Max(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) SumStudentCapacity() *SchoolRequest {
	return r.SumStudentCapacityAs("sumOfStudentCapacity")
}

func (r *SchoolRequest) SumStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Sum(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) AvgStudentCapacity() *SchoolRequest {
	return r.AvgStudentCapacityAs("avgOfStudentCapacity")
}

func (r *SchoolRequest) AvgStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Avg(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) StddevStudentCapacity() *SchoolRequest {
	return r.StddevStudentCapacityAs("standardDeviationOfStudentCapacity")
}

func (r *SchoolRequest) StddevStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.Stddev(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) StddevPopStudentCapacity() *SchoolRequest {
	return r.StddevPopStudentCapacityAs("squareRootOfPopulationStandardDeviationOfStudentCapacity")
}

func (r *SchoolRequest) StddevPopStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.StddevPop(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) VarSampStudentCapacity() *SchoolRequest {
	return r.VarSampStudentCapacityAs("sampleVarianceOfStudentCapacity")
}

func (r *SchoolRequest) VarSampStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.VarSamp(teaqlStudentCapacityStorageKey, alias)
	return r
}
func (r *SchoolRequest) VarPopStudentCapacity() *SchoolRequest {
	return r.VarPopStudentCapacityAs("samplePopulationVarianceOfStudentCapacity")
}

func (r *SchoolRequest) VarPopStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.VarPop(teaqlStudentCapacityStorageKey, alias)
	return r
}

func (r *SchoolRequest) GroupById() *SchoolRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	return r
}

func (r *SchoolRequest) GroupByIdAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlIdStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlIdStorageKey))
	return r
}
func (r *SchoolRequest) GroupByPlatform() *SchoolRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	return r
}

func (r *SchoolRequest) GroupByPlatformAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlPlatformStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlPlatformStorageKey))
	return r
}
func (r *SchoolRequest) GroupBySchoolType() *SchoolRequest {
	r.Query.WithGroupBy(teaqlSchoolTypeStorageKey)
	return r
}

func (r *SchoolRequest) GroupBySchoolTypeAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlSchoolTypeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlSchoolTypeStorageKey))
	return r
}
func (r *SchoolRequest) GroupByName() *SchoolRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	return r
}

func (r *SchoolRequest) GroupByNameAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlNameStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlNameStorageKey))
	return r
}
func (r *SchoolRequest) GroupByAddress() *SchoolRequest {
	r.Query.WithGroupBy(teaqlAddressStorageKey)
	return r
}

func (r *SchoolRequest) GroupByAddressAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlAddressStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlAddressStorageKey))
	return r
}
func (r *SchoolRequest) GroupByEstablishedDate() *SchoolRequest {
	r.Query.WithGroupBy(teaqlEstablishedDateStorageKey)
	return r
}

func (r *SchoolRequest) GroupByEstablishedDateAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlEstablishedDateStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlEstablishedDateStorageKey))
	return r
}
func (r *SchoolRequest) GroupByStudentCapacity() *SchoolRequest {
	r.Query.WithGroupBy(teaqlStudentCapacityStorageKey)
	return r
}

func (r *SchoolRequest) GroupByStudentCapacityAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlStudentCapacityStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlStudentCapacityStorageKey))
	return r
}
func (r *SchoolRequest) GroupByActive() *SchoolRequest {
	r.Query.WithGroupBy(teaqlActiveStorageKey)
	return r
}

func (r *SchoolRequest) GroupByActiveAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlActiveStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlActiveStorageKey))
	return r
}
func (r *SchoolRequest) GroupByCreateTime() *SchoolRequest {
	r.Query.WithGroupBy(teaqlCreateTimeStorageKey)
	return r
}

func (r *SchoolRequest) GroupByCreateTimeAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlCreateTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlCreateTimeStorageKey))
	return r
}
func (r *SchoolRequest) GroupByUpdateTime() *SchoolRequest {
	r.Query.WithGroupBy(teaqlUpdateTimeStorageKey)
	return r
}

func (r *SchoolRequest) GroupByUpdateTimeAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlUpdateTimeStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlUpdateTimeStorageKey))
	return r
}
func (r *SchoolRequest) GroupByVersion() *SchoolRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	return r
}

func (r *SchoolRequest) GroupByVersionAs(alias string) *SchoolRequest {
	r.Query.WithGroupBy(teaqlVersionStorageKey)
	r.Query.ProjectExpr(alias, core.ExprColumnNode(teaqlVersionStorageKey))
	return r
}

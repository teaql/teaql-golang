

package work_item

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

type WorkItemRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
}

type ExecutableWorkItemRequest struct {
	request *WorkItemRequest
}

func NewWorkItemRequest() *WorkItemRequest {
	r := &WorkItemRequest{
		Query: core.NewSelectQuery("Work Item"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewWorkItemMinimalRequest() *WorkItemRequest {
	r := NewWorkItemRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *WorkItemRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *WorkItemRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewWorkItem().EntityDescriptor()
}

func (r *WorkItemRequest) NewRelationEntity() core.Entity {
	return NewWorkItem()
}

func (r *WorkItemRequest) Comment(comment string) *WorkItemRequest {
	r.commentText = comment
	return r
}

func (r *WorkItemRequest) Purpose(purpose string) *ExecutableWorkItemRequest {
	r.purposeText = purpose
	return &ExecutableWorkItemRequest{request: r}
}

func (r *ExecutableWorkItemRequest) Comment(comment string) *ExecutableWorkItemRequest {
	r.request.commentText = comment
	return r
}

func (r *WorkItemRequest) Limit(limit uint64) *WorkItemRequest {
	r.Query.Limit(limit)
	return r
}

func (r *WorkItemRequest) Offset(offset uint64) *WorkItemRequest {
	r.Query.Offset(offset)
	return r
}

func (r *WorkItemRequest) OptimizeForContinuousPageFetch() *WorkItemRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *WorkItemRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *WorkItemRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *WorkItemRequest) OptimizePaginationWithIDSet() *WorkItemRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *WorkItemRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *WorkItemRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *WorkItemRequest) TopNProbeParentThreshold(threshold uint64) *WorkItemRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeWorkItemVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeWorkItemVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *WorkItemRequest) WithDeletedRows() *WorkItemRequest {
	r.Query.Filter = removeWorkItemVersionFilter(r.Query.Filter)
	return r
}

func (r *WorkItemRequest) DeletedRowsOnly() *WorkItemRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *WorkItemRequest) SelectId() *WorkItemRequest {
	r.Query.Project("id")
	return r
}

func (r *WorkItemRequest) WithIdIs(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithIdIsNot(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithIdIn(values []uint64) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *WorkItemRequest) WithIdNotIn(values []uint64) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *WorkItemRequest) WithIdGreaterThan(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithIdGreaterThanOrEqualTo(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithIdLessThan(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithIdLessThanOrEqualTo(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithIdBetween(lower uint64, upper uint64) *WorkItemRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *WorkItemRequest) WithIdIsKnown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *WorkItemRequest) WithIdIsUnknown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *WorkItemRequest) OrderByIdAsc() *WorkItemRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *WorkItemRequest) OrderByIdDesc() *WorkItemRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *WorkItemRequest) SelectTitle() *WorkItemRequest {
	r.Query.Project("title")
	return r
}

func (r *WorkItemRequest) WithTitleIs(value string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEq("title", core.ValText(value)))
	return r
}
func (r *WorkItemRequest) WithTitleIsNot(value string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNe("title", core.ValText(value)))
	return r
}
func (r *WorkItemRequest) WithTitleIn(values []string) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("title", converted))
	return r
}
func (r *WorkItemRequest) WithTitleNotIn(values []string) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("title", converted))
	return r
}
func (r *WorkItemRequest) WithTitleGreaterThan(value string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGt("title", core.ValText(value)))
	return r
}
func (r *WorkItemRequest) WithTitleGreaterThanOrEqualTo(value string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGte("title", core.ValText(value)))
	return r
}
func (r *WorkItemRequest) WithTitleLessThan(value string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLt("title", core.ValText(value)))
	return r
}
func (r *WorkItemRequest) WithTitleLessThanOrEqualTo(value string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLte("title", core.ValText(value)))
	return r
}
func (r *WorkItemRequest) WithTitleBetween(lower string, upper string) *WorkItemRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("title", from, to))
	return r
}
func (r *WorkItemRequest) WithTitleIsKnown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("title"))
	return r
}
func (r *WorkItemRequest) WithTitleIsUnknown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("title"))
	return r
}
func (r *WorkItemRequest) WithTitleContaining(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprContain("title", term))
	return r
}
func (r *WorkItemRequest) WithTitleNotContaining(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotContain("title", term))
	return r
}
func (r *WorkItemRequest) WithTitleStartingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprBeginWith("title", term))
	return r
}
func (r *WorkItemRequest) WithTitleNotStartingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("title", term))
	return r
}
func (r *WorkItemRequest) WithTitleEndingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEndWith("title", term))
	return r
}
func (r *WorkItemRequest) WithTitleNotEndingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotEndWith("title", term))
	return r
}
func (r *WorkItemRequest) WithTitleSoundingLike(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprSoundLike("title", core.ValText(term)))
	return r
}
func (r *WorkItemRequest) OrderByTitleAsc() *WorkItemRequest {
	r.Query.OrderAsc("title")
	return r
}
func (r *WorkItemRequest) OrderByTitleDesc() *WorkItemRequest {
	r.Query.OrderDesc("title")
	return r
}
func (r *WorkItemRequest) SelectDescription() *WorkItemRequest {
	r.Query.Project("description")
	return r
}

func (r *WorkItemRequest) WithDescriptionIs(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEq("description", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *WorkItemRequest) WithDescriptionIsNot(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNe("description", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *WorkItemRequest) WithDescriptionIn(values []*string) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("description", converted))
	return r
}
func (r *WorkItemRequest) WithDescriptionNotIn(values []*string) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("description", converted))
	return r
}
func (r *WorkItemRequest) WithDescriptionGreaterThan(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGt("description", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *WorkItemRequest) WithDescriptionGreaterThanOrEqualTo(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGte("description", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *WorkItemRequest) WithDescriptionLessThan(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLt("description", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *WorkItemRequest) WithDescriptionLessThanOrEqualTo(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLte("description", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *WorkItemRequest) WithDescriptionBetween(lower *string, upper *string) *WorkItemRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("description", from, to))
	return r
}
func (r *WorkItemRequest) WithDescriptionIsKnown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("description"))
	return r
}
func (r *WorkItemRequest) WithDescriptionIsUnknown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("description"))
	return r
}
func (r *WorkItemRequest) WithDescriptionContaining(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprContain("description", term))
	return r
}
func (r *WorkItemRequest) WithDescriptionNotContaining(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotContain("description", term))
	return r
}
func (r *WorkItemRequest) WithDescriptionStartingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprBeginWith("description", term))
	return r
}
func (r *WorkItemRequest) WithDescriptionNotStartingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("description", term))
	return r
}
func (r *WorkItemRequest) WithDescriptionEndingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEndWith("description", term))
	return r
}
func (r *WorkItemRequest) WithDescriptionNotEndingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotEndWith("description", term))
	return r
}
func (r *WorkItemRequest) WithDescriptionSoundingLike(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprSoundLike("description", core.ValText(term)))
	return r
}
func (r *WorkItemRequest) OrderByDescriptionAsc() *WorkItemRequest {
	r.Query.OrderAsc("description")
	return r
}
func (r *WorkItemRequest) OrderByDescriptionDesc() *WorkItemRequest {
	r.Query.OrderDesc("description")
	return r
}
func (r *WorkItemRequest) SelectPlatform() *WorkItemRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *WorkItemRequest) WithPlatformIs(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithPlatformIsNot(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithPlatformIn(values []uint64) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *WorkItemRequest) WithPlatformNotIn(values []uint64) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *WorkItemRequest) WithPlatformGreaterThan(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithPlatformLessThan(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithPlatformLessThanOrEqualTo(value uint64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *WorkItemRequest) WithPlatformBetween(lower uint64, upper uint64) *WorkItemRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *WorkItemRequest) WithPlatformIsKnown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *WorkItemRequest) WithPlatformIsUnknown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *WorkItemRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *WorkItemRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	r.queryOptions.Facets = append(r.queryOptions.Facets, core.NewFacetRequest(
		name, "platform_id", core.NewQuerySelection(nestedReq.GetQuery()), includeAll))
	return r
}
func (r *WorkItemRequest) OrderByPlatformAsc() *WorkItemRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *WorkItemRequest) OrderByPlatformDesc() *WorkItemRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *WorkItemRequest) SelectVersion() *WorkItemRequest {
	r.Query.Project("version")
	return r
}

func (r *WorkItemRequest) WithVersionIs(value int64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *WorkItemRequest) WithVersionIsNot(value int64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *WorkItemRequest) WithVersionIn(values []int64) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *WorkItemRequest) WithVersionNotIn(values []int64) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *WorkItemRequest) WithVersionGreaterThan(value int64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *WorkItemRequest) WithVersionGreaterThanOrEqualTo(value int64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *WorkItemRequest) WithVersionLessThan(value int64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *WorkItemRequest) WithVersionLessThanOrEqualTo(value int64) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *WorkItemRequest) WithVersionBetween(lower int64, upper int64) *WorkItemRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *WorkItemRequest) WithVersionIsKnown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *WorkItemRequest) WithVersionIsUnknown() *WorkItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *WorkItemRequest) OrderByVersionAsc() *WorkItemRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *WorkItemRequest) OrderByVersionDesc() *WorkItemRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *WorkItemRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *WorkItemRequest {
	r.Query.Project("platform_id")
	r.Query.RelationQuery("platformEntity", child.GetQuery())
	r.relationFactories["platformEntity"] = child.NewRelationEntity
	return r
}

func (r *WorkItemRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *WorkItemRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *WorkItemRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableWorkItemRequest) NewEntity(context *runtime.UserContext) *WorkItem {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewWorkItem()
	initialized := context.InitializeEntity("WorkItem", entity)
	typed, ok := initialized.(*WorkItem)
	if !ok {
		panic("entity initializer changed WorkItem to an incompatible type")
	}
	return typed
}

func (e *ExecutableWorkItemRequest) ExecuteForOne(context *runtime.UserContext) (*WorkItem, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableWorkItemRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*WorkItem], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*WorkItem
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewWorkItem()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("platformEntity", childEntity)
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
func (e *ExecutableWorkItemRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*WorkItem], error) {
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
	results := make([]*WorkItem, 0, len(rows))
	queryRoot := core.NewEntityRoot()
	for _, rec := range rows {
		entity := NewWorkItem()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["platformEntity"]; selected {
			entity.markRelationLoaded("platformEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["platformEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("platformEntity", childEntity)
				}
			}
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableWorkItemRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*WorkItem) error) error {
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
			entity := NewWorkItem()
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

func (e *ExecutableWorkItemRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
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
func (e *ExecutableWorkItemRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *WorkItemRequest) Count() *WorkItemRequest {
	return r.CountAs("count")
}

func (r *WorkItemRequest) CountAs(alias string) *WorkItemRequest {
	r.Query.CountField("id", alias)
	return r
}


func (r *WorkItemRequest) GroupById() *WorkItemRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *WorkItemRequest) GroupByTitle() *WorkItemRequest {
	r.Query.WithGroupBy("title")
	return r
}
func (r *WorkItemRequest) GroupByDescription() *WorkItemRequest {
	r.Query.WithGroupBy("description")
	return r
}
func (r *WorkItemRequest) GroupByPlatform() *WorkItemRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}
func (r *WorkItemRequest) GroupByVersion() *WorkItemRequest {
	r.Query.WithGroupBy("version")
	return r
}
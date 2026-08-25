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
	purposeText string
	commentText string
}

type ExecutableWorkItemRequest struct {
	request *WorkItemRequest
}

func NewWorkItemRequest() *WorkItemRequest {
	r := &WorkItemRequest{
		Query: core.NewSelectQuery("Work Item"),
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

func removeWorkItemVersionFilter(expr *core.Expr) *core.Expr {
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
		if kept := removeWorkItemVersionFilter(part); kept != nil {
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
func (r *WorkItemRequest) WithTitleEndingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEndWith("title", term))
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
	r.Query.AndFilter(core.ExprEq("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}()))
	return r
}
func (r *WorkItemRequest) WithDescriptionIsNot(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprNe("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}()))
	return r
}
func (r *WorkItemRequest) WithDescriptionIn(values []*string) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value {
			if value == nil {
				return core.ValNull()
			}
			return core.ValText(*value)
		}())
	}
	r.Query.AndFilter(core.ExprInList("description", converted))
	return r
}
func (r *WorkItemRequest) WithDescriptionNotIn(values []*string) *WorkItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value {
			if value == nil {
				return core.ValNull()
			}
			return core.ValText(*value)
		}())
	}
	r.Query.AndFilter(core.ExprNotInList("description", converted))
	return r
}
func (r *WorkItemRequest) WithDescriptionGreaterThan(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGt("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}()))
	return r
}
func (r *WorkItemRequest) WithDescriptionGreaterThanOrEqualTo(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprGte("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}()))
	return r
}
func (r *WorkItemRequest) WithDescriptionLessThan(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLt("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}()))
	return r
}
func (r *WorkItemRequest) WithDescriptionLessThanOrEqualTo(value *string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprLte("description", func() core.Value {
		if value == nil {
			return core.ValNull()
		}
		return core.ValText(*value)
	}()))
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
func (r *WorkItemRequest) WithDescriptionEndingWith(term string) *WorkItemRequest {
	r.Query.AndFilter(core.ExprEndWith("description", term))
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
func (r *WorkItemRequest) FacetByPlatformAs(name string, nestedReq any) *WorkItemRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "platform_id", req.GetQuery())
	}
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
func (r *WorkItemRequest) OrderByVersionAsc() *WorkItemRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *WorkItemRequest) OrderByVersionDesc() *WorkItemRequest {
	r.Query.OrderDesc("version")
	return r
}

func (e *ExecutableWorkItemRequest) NewEntity(context *runtime.UserContext) *WorkItem {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewWorkItem()
	entity.AttachEntityRoot(context.EntityRoot())
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
	for _, rec := range rows {
		entity := NewWorkItem()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
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
	results := make([]*WorkItem, 0, len(rows))
	for _, rec := range rows {
		entity := NewWorkItem()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
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
	r.Query.Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	req := &data_service.QueryRequest{Query: r.Query, TraceChain: r.Query.TraceChain, Comment: r.Query.CommentText}
	return ds.QueryStream(context, req, chunkSize, func(chunk *data_service.StreamChunk) error {
		for _, rec := range chunk.Rows {
			entity := NewWorkItem()
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

func (e *ExecutableWorkItemRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
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

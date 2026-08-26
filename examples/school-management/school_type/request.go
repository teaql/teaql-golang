

package school_type

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"school-management-service-core-workspace/lib/school"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type SchoolTypeRequest struct {
	Query       *core.SelectQuery
	purposeText string
	commentText string
}

type ExecutableSchoolTypeRequest struct {
	request *SchoolTypeRequest
}

func NewSchoolTypeRequest() *SchoolTypeRequest {
	r := &SchoolTypeRequest{
		Query: core.NewSelectQuery("School Type"),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewSchoolTypeMinimalRequest() *SchoolTypeRequest {
	r := NewSchoolTypeRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *SchoolTypeRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

func (r *SchoolTypeRequest) Comment(comment string) *SchoolTypeRequest {
	r.commentText = comment
	return r
}

func (r *SchoolTypeRequest) Purpose(purpose string) *ExecutableSchoolTypeRequest {
	r.purposeText = purpose
	return &ExecutableSchoolTypeRequest{request: r}
}

func (r *ExecutableSchoolTypeRequest) Comment(comment string) *ExecutableSchoolTypeRequest {
	r.request.commentText = comment
	return r
}

func (r *SchoolTypeRequest) Limit(limit uint64) *SchoolTypeRequest {
	r.Query.Limit(limit)
	return r
}

func (r *SchoolTypeRequest) Offset(offset uint64) *SchoolTypeRequest {
	r.Query.Offset(offset)
	return r
}

func (r *SchoolTypeRequest) OptimizeForContinuousPageFetch() *SchoolTypeRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *SchoolTypeRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *SchoolTypeRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func removeSchoolTypeVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeSchoolTypeVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *SchoolTypeRequest) WithDeletedRows() *SchoolTypeRequest {
	r.Query.Filter = removeSchoolTypeVersionFilter(r.Query.Filter)
	return r
}

func (r *SchoolTypeRequest) DeletedRowsOnly() *SchoolTypeRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *SchoolTypeRequest) SelectPlatform() *SchoolTypeRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *SchoolTypeRequest) WithPlatformIs(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithPlatformIsNot(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithPlatformIn(values []uint64) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *SchoolTypeRequest) WithPlatformNotIn(values []uint64) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *SchoolTypeRequest) WithPlatformGreaterThan(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithPlatformLessThan(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithPlatformLessThanOrEqualTo(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) FacetByPlatformAs(name string, nestedReq any) *SchoolTypeRequest {
	if req, ok := nestedReq.(interface{ GetQuery() *core.SelectQuery }); ok {
		r.Query.WithObjectGroupBy(name, "platform_id", req.GetQuery())
	}
	return r
}
func (r *SchoolTypeRequest) OrderByPlatformAsc() *SchoolTypeRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *SchoolTypeRequest) OrderByPlatformDesc() *SchoolTypeRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *SchoolTypeRequest) SelectId() *SchoolTypeRequest {
	r.Query.Project("id")
	return r
}

func (r *SchoolTypeRequest) WithIdIs(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithIdIsNot(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithIdIn(values []uint64) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *SchoolTypeRequest) WithIdNotIn(values []uint64) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *SchoolTypeRequest) WithIdGreaterThan(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithIdGreaterThanOrEqualTo(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithIdLessThan(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) WithIdLessThanOrEqualTo(value uint64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *SchoolTypeRequest) OrderByIdAsc() *SchoolTypeRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *SchoolTypeRequest) OrderByIdDesc() *SchoolTypeRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *SchoolTypeRequest) SelectName() *SchoolTypeRequest {
	r.Query.Project("name")
	return r
}

func (r *SchoolTypeRequest) WithNameIs(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEq("name", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithNameIsNot(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNe("name", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithNameIn(values []string) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("name", converted))
	return r
}
func (r *SchoolTypeRequest) WithNameNotIn(values []string) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("name", converted))
	return r
}
func (r *SchoolTypeRequest) WithNameGreaterThan(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGt("name", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithNameGreaterThanOrEqualTo(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGte("name", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithNameLessThan(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLt("name", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithNameLessThanOrEqualTo(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLte("name", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithNameContaining(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprContain("name", term))
	return r
}
func (r *SchoolTypeRequest) WithNameNotContaining(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNotContain("name", term))
	return r
}
func (r *SchoolTypeRequest) WithNameStartingWith(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprBeginWith("name", term))
	return r
}
func (r *SchoolTypeRequest) WithNameEndingWith(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEndWith("name", term))
	return r
}
func (r *SchoolTypeRequest) OrderByNameAsc() *SchoolTypeRequest {
	r.Query.OrderAsc("name")
	return r
}
func (r *SchoolTypeRequest) OrderByNameDesc() *SchoolTypeRequest {
	r.Query.OrderDesc("name")
	return r
}
func (r *SchoolTypeRequest) SelectCode() *SchoolTypeRequest {
	r.Query.Project("code")
	return r
}

func (r *SchoolTypeRequest) WithCodeIs(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEq("code", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithCodeIsNot(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNe("code", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithCodeIn(values []string) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("code", converted))
	return r
}
func (r *SchoolTypeRequest) WithCodeNotIn(values []string) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("code", converted))
	return r
}
func (r *SchoolTypeRequest) WithCodeGreaterThan(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGt("code", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithCodeGreaterThanOrEqualTo(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGte("code", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithCodeLessThan(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLt("code", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithCodeLessThanOrEqualTo(value string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLte("code", core.ValText(value)))
	return r
}
func (r *SchoolTypeRequest) WithCodeContaining(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprContain("code", term))
	return r
}
func (r *SchoolTypeRequest) WithCodeNotContaining(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNotContain("code", term))
	return r
}
func (r *SchoolTypeRequest) WithCodeStartingWith(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprBeginWith("code", term))
	return r
}
func (r *SchoolTypeRequest) WithCodeEndingWith(term string) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEndWith("code", term))
	return r
}
func (r *SchoolTypeRequest) OrderByCodeAsc() *SchoolTypeRequest {
	r.Query.OrderAsc("code")
	return r
}
func (r *SchoolTypeRequest) OrderByCodeDesc() *SchoolTypeRequest {
	r.Query.OrderDesc("code")
	return r
}
func (r *SchoolTypeRequest) SelectDisplayOrder() *SchoolTypeRequest {
	r.Query.Project("display_order")
	return r
}

func (r *SchoolTypeRequest) WithDisplayOrderIs(value decimal.Decimal) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEq("display_order", core.ValDecimal(value)))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderIsNot(value decimal.Decimal) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNe("display_order", core.ValDecimal(value)))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderIn(values []decimal.Decimal) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("display_order", converted))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderNotIn(values []decimal.Decimal) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("display_order", converted))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderGreaterThan(value decimal.Decimal) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGt("display_order", core.ValDecimal(value)))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderGreaterThanOrEqualTo(value decimal.Decimal) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGte("display_order", core.ValDecimal(value)))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderLessThan(value decimal.Decimal) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLt("display_order", core.ValDecimal(value)))
	return r
}
func (r *SchoolTypeRequest) WithDisplayOrderLessThanOrEqualTo(value decimal.Decimal) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLte("display_order", core.ValDecimal(value)))
	return r
}
func (r *SchoolTypeRequest) OrderByDisplayOrderAsc() *SchoolTypeRequest {
	r.Query.OrderAsc("display_order")
	return r
}
func (r *SchoolTypeRequest) OrderByDisplayOrderDesc() *SchoolTypeRequest {
	r.Query.OrderDesc("display_order")
	return r
}
func (r *SchoolTypeRequest) SelectVersion() *SchoolTypeRequest {
	r.Query.Project("version")
	return r
}

func (r *SchoolTypeRequest) WithVersionIs(value int64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *SchoolTypeRequest) WithVersionIsNot(value int64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *SchoolTypeRequest) WithVersionIn(values []int64) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *SchoolTypeRequest) WithVersionNotIn(values []int64) *SchoolTypeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *SchoolTypeRequest) WithVersionGreaterThan(value int64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *SchoolTypeRequest) WithVersionGreaterThanOrEqualTo(value int64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *SchoolTypeRequest) WithVersionLessThan(value int64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *SchoolTypeRequest) WithVersionLessThanOrEqualTo(value int64) *SchoolTypeRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *SchoolTypeRequest) OrderByVersionAsc() *SchoolTypeRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *SchoolTypeRequest) OrderByVersionDesc() *SchoolTypeRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *SchoolTypeRequest) CountSchools() *SchoolTypeRequest {
	r.Query.Count("count_schools")
	return r
}

func (r *SchoolTypeRequest) SelectSchoolList() *SchoolTypeRequest {
	return r.SelectSchoolListWith(school.NewSchoolRequest())
}

func (r *SchoolTypeRequest) SelectSchoolListWith(child *school.SchoolRequest) *SchoolTypeRequest {
	r.Query.RelationQuery("schoolList", child.Query)
	return r
}

func (e *ExecutableSchoolTypeRequest) NewEntity(context *runtime.UserContext) *SchoolType {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewSchoolType()
	entity.AttachEntityRoot(context.EntityRoot())
	initialized := context.InitializeEntity("SchoolType", entity)
	typed, ok := initialized.(*SchoolType)
	if !ok {
		panic("entity initializer changed SchoolType to an incompatible type")
	}
	return typed
}

func (e *ExecutableSchoolTypeRequest) ExecuteForOne(context *runtime.UserContext) (*SchoolType, error) {
	list, err := e.ExecuteForList(context)
	if err != nil {
		return nil, err
	}
	if len(list.Data) == 0 {
		return nil, nil // Or a specific Not Found error
	}
	return list.Data[0], nil
}

func (e *ExecutableSchoolTypeRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*SchoolType], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil {
		return nil, err
	}

	var results []*SchoolType
	for _, rec := range rows {
		entity := NewSchoolType()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil {
			return nil, err
		}
		if relationValue, selected := rec["schoolList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation schoolList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := school.NewSchool()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.SchoolList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results), nil
}

// ExecuteForPage applies trusted policy once, then derives exact-count and row
// queries from that same authorized snapshot.
func (e *ExecutableSchoolTypeRequest) ExecuteForPage(context *runtime.UserContext, offset uint64, size uint64) (*core.SmartList[*SchoolType], error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForPage()")
	}
	if size == 0 {
		return nil, fmt.Errorf("QUERY_INVALID_LIMIT: size must be positive")
	}
	r.Query.Page(offset, size).Comment(fmt.Sprintf("comment=%s; purpose=%s", r.commentText, r.purposeText))
	authorized, err := context.PrepareQuery(r.Query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	const countAlias = "__teaql_total"
	countRows, err := service.FetchAll(context, authorized.ForExactCount(countAlias))
	if err != nil { return nil, err }
	if len(countRows) != 1 { return nil, fmt.Errorf("exact count returned %d rows", len(countRows)) }
	total, ok := countRows[0][countAlias].TryU64()
	if !ok { return nil, fmt.Errorf("exact count did not return an unsigned integer") }
	rows, err := service.FetchAll(context, authorized)
	if err != nil { return nil, err }
	results := make([]*SchoolType, 0, len(rows))
	for _, rec := range rows {
		entity := NewSchoolType()
		entity.AttachEntityRoot(context.EntityRoot())
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["schoolList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation schoolList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := school.NewSchool()
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.SchoolList().Add(childEntity)
				}}
		results = append(results, entity)
	}
	return core.NewSmartList(results).WithTotalCount(total), nil
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield cancels iteration and releases the database resources.
func (e *ExecutableSchoolTypeRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*SchoolType) error) error {
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
			entity := NewSchoolType()
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

func (e *ExecutableSchoolTypeRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
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

func (r *SchoolTypeRequest) Count() *SchoolTypeRequest {
	return r.CountAs("count")
}

func (r *SchoolTypeRequest) CountAs(alias string) *SchoolTypeRequest {
	r.Query.CountField("id", alias)
	return r
}

func (r *SchoolTypeRequest) MinDisplayOrder() *SchoolTypeRequest {
	return r.MinDisplayOrderAs("minOfDisplayOrder")
}

func (r *SchoolTypeRequest) MinDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.Min("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) MaxDisplayOrder() *SchoolTypeRequest {
	return r.MaxDisplayOrderAs("maxOfDisplayOrder")
}

func (r *SchoolTypeRequest) MaxDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.Max("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) SumDisplayOrder() *SchoolTypeRequest {
	return r.SumDisplayOrderAs("sumOfDisplayOrder")
}

func (r *SchoolTypeRequest) SumDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.Sum("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) AvgDisplayOrder() *SchoolTypeRequest {
	return r.AvgDisplayOrderAs("avgOfDisplayOrder")
}

func (r *SchoolTypeRequest) AvgDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.Avg("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) StddevDisplayOrder() *SchoolTypeRequest {
	return r.StddevDisplayOrderAs("standardDeviationOfDisplayOrder")
}

func (r *SchoolTypeRequest) StddevDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.Stddev("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) StddevPopDisplayOrder() *SchoolTypeRequest {
	return r.StddevPopDisplayOrderAs("squareRootOfPopulationStandardDeviationOfDisplayOrder")
}

func (r *SchoolTypeRequest) StddevPopDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.StddevPop("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) VarSampDisplayOrder() *SchoolTypeRequest {
	return r.VarSampDisplayOrderAs("sampleVarianceOfDisplayOrder")
}

func (r *SchoolTypeRequest) VarSampDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.VarSamp("display_order", alias)
	return r
}
func (r *SchoolTypeRequest) VarPopDisplayOrder() *SchoolTypeRequest {
	return r.VarPopDisplayOrderAs("samplePopulationVarianceOfDisplayOrder")
}

func (r *SchoolTypeRequest) VarPopDisplayOrderAs(alias string) *SchoolTypeRequest {
	r.Query.VarPop("display_order", alias)
	return r
}

func (r *SchoolTypeRequest) GroupByPlatform() *SchoolTypeRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}
func (r *SchoolTypeRequest) GroupById() *SchoolTypeRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *SchoolTypeRequest) GroupByName() *SchoolTypeRequest {
	r.Query.WithGroupBy("name")
	return r
}
func (r *SchoolTypeRequest) GroupByCode() *SchoolTypeRequest {
	r.Query.WithGroupBy("code")
	return r
}
func (r *SchoolTypeRequest) GroupByDisplayOrder() *SchoolTypeRequest {
	r.Query.WithGroupBy("display_order")
	return r
}
func (r *SchoolTypeRequest) GroupByVersion() *SchoolTypeRequest {
	r.Query.WithGroupBy("version")
	return r
}
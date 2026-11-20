

package order_attachment

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-attachment-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderAttachmentRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableOrderAttachmentRequest struct {
	request *OrderAttachmentRequest
}

func NewOrderAttachmentRequest() *OrderAttachmentRequest {
	r := &OrderAttachmentRequest{
		Query: core.NewSelectQuery("order_attachment"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderAttachmentMinimalRequest() *OrderAttachmentRequest {
	r := NewOrderAttachmentRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderAttachmentRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *OrderAttachmentRequest) FilterByJson(source []byte) (*OrderAttachmentRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *OrderAttachmentRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *OrderAttachmentRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *OrderAttachmentRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return newLoadedOrderAttachment().EntityDescriptor()
}

func (r *OrderAttachmentRequest) NewRelationEntity() core.Entity {
	return newLoadedOrderAttachment()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *OrderAttachmentRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *OrderAttachmentRequest) teaqlHydrateQueryRecord(rec core.Record) (*OrderAttachment, error) {
	entity := newLoadedOrderAttachment()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["customerOrderEntity"]; selected {
		entity.markRelationLoaded("customerOrderEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["customerOrderEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("customerOrderEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *OrderAttachmentRequest) Comment(comment string) *OrderAttachmentRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *OrderAttachmentRequest) Purpose(purpose string) *ExecutableOrderAttachmentRequest {
	r.purposeText = purpose
	return &ExecutableOrderAttachmentRequest{request: r}
}

func (r *ExecutableOrderAttachmentRequest) Comment(comment string) *ExecutableOrderAttachmentRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *OrderAttachmentRequest) Limit(limit uint64) *OrderAttachmentRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderAttachmentRequest) Offset(offset uint64) *OrderAttachmentRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderAttachmentRequest) OptimizeForContinuousPageFetch() *OrderAttachmentRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderAttachmentRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderAttachmentRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderAttachmentRequest) OptimizePaginationWithIDSet() *OrderAttachmentRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderAttachmentRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderAttachmentRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderAttachmentRequest) TopNProbeParentThreshold(threshold uint64) *OrderAttachmentRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderAttachmentVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderAttachmentVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderAttachmentRequest) WithDeletedRows() *OrderAttachmentRequest {
	r.Query.Filter = removeOrderAttachmentVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderAttachmentRequest) DeletedRowsOnly() *OrderAttachmentRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderAttachmentRequest) SelectId() *OrderAttachmentRequest {
	r.Query.Project("id")
	return r
}

func (r *OrderAttachmentRequest) WithIdIs(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithIdIsNot(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithIdIn(values []uint64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderAttachmentRequest) WithIdNotIn(values []uint64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderAttachmentRequest) WithIdGreaterThan(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithIdLessThan(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithIdLessThanOrEqualTo(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithIdBetween(lower uint64, upper uint64) *OrderAttachmentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithIdIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *OrderAttachmentRequest) WithIdIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *OrderAttachmentRequest) OrderByIdAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderAttachmentRequest) OrderByIdDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *OrderAttachmentRequest) SelectCustomerOrder() *OrderAttachmentRequest {
	r.Query.Project("customer_order_id")
	return r
}

func (r *OrderAttachmentRequest) WithCustomerOrderIs(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderIsNot(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderIn(values []uint64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderNotIn(values []uint64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderGreaterThan(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderLessThan(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *OrderAttachmentRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_order_id", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_order_id"))
	return r
}
func (r *OrderAttachmentRequest) WithCustomerOrderIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_order_id"))
	return r
}
func (r *OrderAttachmentRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderAttachmentRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "customer_order_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerOrderAs", Alias: &name, Relation: "customerOrder"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *OrderAttachmentRequest) OrderByCustomerOrderAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *OrderAttachmentRequest) OrderByCustomerOrderDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}
func (r *OrderAttachmentRequest) SelectFileName() *OrderAttachmentRequest {
	r.Query.Project("file_name")
	return r
}

func (r *OrderAttachmentRequest) WithFileNameIs(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("file_name", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameIsNot(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("file_name", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("file_name", converted))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameNotIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("file_name", converted))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameGreaterThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("file_name", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameGreaterThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("file_name", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameLessThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("file_name", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameLessThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("file_name", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameBetween(lower string, upper string) *OrderAttachmentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("file_name", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("file_name"))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("file_name"))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprContain("file_name", term))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameNotContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotContain("file_name", term))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprBeginWith("file_name", term))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameNotStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("file_name", term))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEndWith("file_name", term))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameNotEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotEndWith("file_name", term))
	return r
}
func (r *OrderAttachmentRequest) WithFileNameSoundingLike(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprSoundLike("file_name", core.ValText(term)))
	return r
}
func (r *OrderAttachmentRequest) OrderByFileNameAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("file_name")
	return r
}
func (r *OrderAttachmentRequest) OrderByFileNameDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("file_name")
	return r
}
func (r *OrderAttachmentRequest) SelectContentType() *OrderAttachmentRequest {
	r.Query.Project("content_type")
	return r
}

func (r *OrderAttachmentRequest) WithContentTypeIs(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("content_type", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeIsNot(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("content_type", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("content_type", converted))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeNotIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("content_type", converted))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeGreaterThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("content_type", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeGreaterThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("content_type", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeLessThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("content_type", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeLessThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("content_type", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeBetween(lower string, upper string) *OrderAttachmentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("content_type", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("content_type"))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("content_type"))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprContain("content_type", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeNotContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotContain("content_type", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprBeginWith("content_type", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeNotStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("content_type", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEndWith("content_type", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeNotEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotEndWith("content_type", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentTypeSoundingLike(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprSoundLike("content_type", core.ValText(term)))
	return r
}
func (r *OrderAttachmentRequest) OrderByContentTypeAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("content_type")
	return r
}
func (r *OrderAttachmentRequest) OrderByContentTypeDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("content_type")
	return r
}
func (r *OrderAttachmentRequest) SelectContentLength() *OrderAttachmentRequest {
	r.Query.Project("content_length")
	return r
}

func (r *OrderAttachmentRequest) WithContentLengthIs(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("content_length", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthIsNot(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("content_length", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthIn(values []int64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("content_length", converted))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthNotIn(values []int64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("content_length", converted))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthGreaterThan(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("content_length", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthGreaterThanOrEqualTo(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("content_length", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthLessThan(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("content_length", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthLessThanOrEqualTo(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("content_length", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthBetween(lower int64, upper int64) *OrderAttachmentRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("content_length", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("content_length"))
	return r
}
func (r *OrderAttachmentRequest) WithContentLengthIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("content_length"))
	return r
}
func (r *OrderAttachmentRequest) OrderByContentLengthAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("content_length")
	return r
}
func (r *OrderAttachmentRequest) OrderByContentLengthDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("content_length")
	return r
}
func (r *OrderAttachmentRequest) SelectContentSha256() *OrderAttachmentRequest {
	r.Query.Project("content_sha256")
	return r
}

func (r *OrderAttachmentRequest) WithContentSha256Is(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("content_sha256", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256IsNot(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("content_sha256", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256In(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("content_sha256", converted))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256NotIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("content_sha256", converted))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256GreaterThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("content_sha256", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256GreaterThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("content_sha256", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256LessThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("content_sha256", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256LessThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("content_sha256", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256Between(lower string, upper string) *OrderAttachmentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("content_sha256", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256IsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("content_sha256"))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256IsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("content_sha256"))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256Containing(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprContain("content_sha256", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256NotContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotContain("content_sha256", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256StartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprBeginWith("content_sha256", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256NotStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("content_sha256", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256EndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEndWith("content_sha256", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256NotEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotEndWith("content_sha256", term))
	return r
}
func (r *OrderAttachmentRequest) WithContentSha256SoundingLike(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprSoundLike("content_sha256", core.ValText(term)))
	return r
}
func (r *OrderAttachmentRequest) OrderByContentSha256Asc() *OrderAttachmentRequest {
	r.Query.OrderAsc("content_sha256")
	return r
}
func (r *OrderAttachmentRequest) OrderByContentSha256Desc() *OrderAttachmentRequest {
	r.Query.OrderDesc("content_sha256")
	return r
}
func (r *OrderAttachmentRequest) SelectObjectKey() *OrderAttachmentRequest {
	r.Query.Project("object_key")
	return r
}

func (r *OrderAttachmentRequest) WithObjectKeyIs(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("object_key", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyIsNot(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("object_key", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("object_key", converted))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyNotIn(values []string) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("object_key", converted))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyGreaterThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("object_key", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyGreaterThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("object_key", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyLessThan(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("object_key", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyLessThanOrEqualTo(value string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("object_key", core.ValText(value)))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyBetween(lower string, upper string) *OrderAttachmentRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("object_key", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("object_key"))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("object_key"))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprContain("object_key", term))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyNotContaining(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotContain("object_key", term))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprBeginWith("object_key", term))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyNotStartingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("object_key", term))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEndWith("object_key", term))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeyNotEndingWith(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotEndWith("object_key", term))
	return r
}
func (r *OrderAttachmentRequest) WithObjectKeySoundingLike(term string) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprSoundLike("object_key", core.ValText(term)))
	return r
}
func (r *OrderAttachmentRequest) OrderByObjectKeyAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("object_key")
	return r
}
func (r *OrderAttachmentRequest) OrderByObjectKeyDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("object_key")
	return r
}
func (r *OrderAttachmentRequest) SelectVersion() *OrderAttachmentRequest {
	r.Query.Project("version")
	return r
}

func (r *OrderAttachmentRequest) WithVersionIs(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithVersionIsNot(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithVersionIn(values []int64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderAttachmentRequest) WithVersionNotIn(values []int64) *OrderAttachmentRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderAttachmentRequest) WithVersionGreaterThan(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithVersionLessThan(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithVersionLessThanOrEqualTo(value int64) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderAttachmentRequest) WithVersionBetween(lower int64, upper int64) *OrderAttachmentRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *OrderAttachmentRequest) WithVersionIsKnown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *OrderAttachmentRequest) WithVersionIsUnknown() *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *OrderAttachmentRequest) OrderByVersionAsc() *OrderAttachmentRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderAttachmentRequest) OrderByVersionDesc() *OrderAttachmentRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderAttachmentRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderAttachmentRequest {
	runtime.EnsureRelationProjection(r.Query, "customer_order_id")
	r.Query.RelationQuerySelection("customerOrderEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderWith", Relation: "customerOrderEntity"}
	r.relationFactories["customerOrderEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *OrderAttachmentRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderAttachmentRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderAttachmentRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableOrderAttachmentRequest) NewEntity(context *runtime.UserContext) (*OrderAttachment, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewOrderAttachment(context)
}

func (e *ExecutableOrderAttachmentRequest) ExecuteForOne(context *runtime.UserContext) (*OrderAttachment, error) {
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

func (e *ExecutableOrderAttachmentRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderAttachment], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*OrderAttachment, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableOrderAttachmentRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderAttachment) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*OrderAttachment, error) {
		entity := newLoadedOrderAttachment()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableOrderAttachmentRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderAttachmentRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableOrderAttachmentRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderAttachmentRequest) Count() *OrderAttachmentRequest {
	return r.CountAs("count")
}

func (r *OrderAttachmentRequest) CountAs(alias string) *OrderAttachmentRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *OrderAttachmentRequest) MinContentLength() *OrderAttachmentRequest {
	return r.MinContentLengthAs("minOfContentLength")
}

func (r *OrderAttachmentRequest) MinContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.Min("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) MaxContentLength() *OrderAttachmentRequest {
	return r.MaxContentLengthAs("maxOfContentLength")
}

func (r *OrderAttachmentRequest) MaxContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.Max("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) SumContentLength() *OrderAttachmentRequest {
	return r.SumContentLengthAs("sumOfContentLength")
}

func (r *OrderAttachmentRequest) SumContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.Sum("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) AvgContentLength() *OrderAttachmentRequest {
	return r.AvgContentLengthAs("avgOfContentLength")
}

func (r *OrderAttachmentRequest) AvgContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.Avg("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) StddevContentLength() *OrderAttachmentRequest {
	return r.StddevContentLengthAs("standardDeviationOfContentLength")
}

func (r *OrderAttachmentRequest) StddevContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.Stddev("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) StddevPopContentLength() *OrderAttachmentRequest {
	return r.StddevPopContentLengthAs("squareRootOfPopulationStandardDeviationOfContentLength")
}

func (r *OrderAttachmentRequest) StddevPopContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.StddevPop("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) VarSampContentLength() *OrderAttachmentRequest {
	return r.VarSampContentLengthAs("sampleVarianceOfContentLength")
}

func (r *OrderAttachmentRequest) VarSampContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.VarSamp("content_length", alias)
	return r
}
func (r *OrderAttachmentRequest) VarPopContentLength() *OrderAttachmentRequest {
	return r.VarPopContentLengthAs("samplePopulationVarianceOfContentLength")
}

func (r *OrderAttachmentRequest) VarPopContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.VarPop("content_length", alias)
	return r
}

func (r *OrderAttachmentRequest) GroupById() *OrderAttachmentRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *OrderAttachmentRequest) GroupByIdAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *OrderAttachmentRequest) GroupByCustomerOrder() *OrderAttachmentRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}

func (r *OrderAttachmentRequest) GroupByCustomerOrderAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("customer_order_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("customer_order_id"))
	return r
}
func (r *OrderAttachmentRequest) GroupByFileName() *OrderAttachmentRequest {
	r.Query.WithGroupBy("file_name")
	return r
}

func (r *OrderAttachmentRequest) GroupByFileNameAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("file_name")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("file_name"))
	return r
}
func (r *OrderAttachmentRequest) GroupByContentType() *OrderAttachmentRequest {
	r.Query.WithGroupBy("content_type")
	return r
}

func (r *OrderAttachmentRequest) GroupByContentTypeAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("content_type")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("content_type"))
	return r
}
func (r *OrderAttachmentRequest) GroupByContentLength() *OrderAttachmentRequest {
	r.Query.WithGroupBy("content_length")
	return r
}

func (r *OrderAttachmentRequest) GroupByContentLengthAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("content_length")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("content_length"))
	return r
}
func (r *OrderAttachmentRequest) GroupByContentSha256() *OrderAttachmentRequest {
	r.Query.WithGroupBy("content_sha256")
	return r
}

func (r *OrderAttachmentRequest) GroupByContentSha256As(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("content_sha256")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("content_sha256"))
	return r
}
func (r *OrderAttachmentRequest) GroupByObjectKey() *OrderAttachmentRequest {
	r.Query.WithGroupBy("object_key")
	return r
}

func (r *OrderAttachmentRequest) GroupByObjectKeyAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("object_key")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("object_key"))
	return r
}
func (r *OrderAttachmentRequest) GroupByVersion() *OrderAttachmentRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *OrderAttachmentRequest) GroupByVersionAs(alias string) *OrderAttachmentRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}

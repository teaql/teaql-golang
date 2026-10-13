

package order_item

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"request-tree-order-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type OrderItemRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableOrderItemRequest struct {
	request *OrderItemRequest
}

func NewOrderItemRequest() *OrderItemRequest {
	r := &OrderItemRequest{
		Query: core.NewSelectQuery("Order Item"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewOrderItemMinimalRequest() *OrderItemRequest {
	r := NewOrderItemRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *OrderItemRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *OrderItemRequest) FilterByJson(source []byte) (*OrderItemRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *OrderItemRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *OrderItemRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *OrderItemRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewOrderItem().EntityDescriptor()
}

func (r *OrderItemRequest) NewRelationEntity() core.Entity {
	return NewOrderItem()
}

func (r *OrderItemRequest) Comment(comment string) *OrderItemRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *OrderItemRequest) Purpose(purpose string) *ExecutableOrderItemRequest {
	r.purposeText = purpose
	return &ExecutableOrderItemRequest{request: r}
}

func (r *ExecutableOrderItemRequest) Comment(comment string) *ExecutableOrderItemRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *OrderItemRequest) Limit(limit uint64) *OrderItemRequest {
	r.Query.Limit(limit)
	return r
}

func (r *OrderItemRequest) Offset(offset uint64) *OrderItemRequest {
	r.Query.Offset(offset)
	return r
}

func (r *OrderItemRequest) OptimizeForContinuousPageFetch() *OrderItemRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *OrderItemRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *OrderItemRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *OrderItemRequest) OptimizePaginationWithIDSet() *OrderItemRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *OrderItemRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *OrderItemRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *OrderItemRequest) TopNProbeParentThreshold(threshold uint64) *OrderItemRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeOrderItemVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeOrderItemVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *OrderItemRequest) WithDeletedRows() *OrderItemRequest {
	r.Query.Filter = removeOrderItemVersionFilter(r.Query.Filter)
	return r
}

func (r *OrderItemRequest) DeletedRowsOnly() *OrderItemRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *OrderItemRequest) SelectId() *OrderItemRequest {
	r.Query.Project("id")
	return r
}

func (r *OrderItemRequest) WithIdIs(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdIsNot(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *OrderItemRequest) WithIdNotIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *OrderItemRequest) WithIdGreaterThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdGreaterThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdLessThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdLessThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithIdBetween(lower uint64, upper uint64) *OrderItemRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *OrderItemRequest) WithIdIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *OrderItemRequest) WithIdIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *OrderItemRequest) OrderByIdAsc() *OrderItemRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *OrderItemRequest) OrderByIdDesc() *OrderItemRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *OrderItemRequest) SelectCustomerOrder() *OrderItemRequest {
	r.Query.Project("customer_order_id")
	return r
}

func (r *OrderItemRequest) WithCustomerOrderIs(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIsNot(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("customer_order_id", converted))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderNotIn(values []uint64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("customer_order_id", converted))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderGreaterThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderGreaterThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderLessThan(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderLessThanOrEqualTo(value uint64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("customer_order_id", core.ValU64(value)))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderBetween(lower uint64, upper uint64) *OrderItemRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("customer_order_id", from, to))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("customer_order_id"))
	return r
}
func (r *OrderItemRequest) WithCustomerOrderIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("customer_order_id"))
	return r
}
func (r *OrderItemRequest) FacetByCustomerOrderAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *OrderItemRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "customer_order_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByCustomerOrderAs", Alias: &name, Relation: "customerOrder"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *OrderItemRequest) OrderByCustomerOrderAsc() *OrderItemRequest {
	r.Query.OrderAsc("customer_order_id")
	return r
}
func (r *OrderItemRequest) OrderByCustomerOrderDesc() *OrderItemRequest {
	r.Query.OrderDesc("customer_order_id")
	return r
}
func (r *OrderItemRequest) SelectProductName() *OrderItemRequest {
	r.Query.Project("product_name")
	return r
}

func (r *OrderItemRequest) WithProductNameIs(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("product_name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithProductNameIsNot(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("product_name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithProductNameIn(values []string) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("product_name", converted))
	return r
}
func (r *OrderItemRequest) WithProductNameNotIn(values []string) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("product_name", converted))
	return r
}
func (r *OrderItemRequest) WithProductNameGreaterThan(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("product_name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithProductNameGreaterThanOrEqualTo(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("product_name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithProductNameLessThan(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("product_name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithProductNameLessThanOrEqualTo(value string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("product_name", core.ValText(value)))
	return r
}
func (r *OrderItemRequest) WithProductNameBetween(lower string, upper string) *OrderItemRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("product_name", from, to))
	return r
}
func (r *OrderItemRequest) WithProductNameIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("product_name"))
	return r
}
func (r *OrderItemRequest) WithProductNameIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("product_name"))
	return r
}
func (r *OrderItemRequest) WithProductNameContaining(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprContain("product_name", term))
	return r
}
func (r *OrderItemRequest) WithProductNameNotContaining(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotContain("product_name", term))
	return r
}
func (r *OrderItemRequest) WithProductNameStartingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprBeginWith("product_name", term))
	return r
}
func (r *OrderItemRequest) WithProductNameNotStartingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("product_name", term))
	return r
}
func (r *OrderItemRequest) WithProductNameEndingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEndWith("product_name", term))
	return r
}
func (r *OrderItemRequest) WithProductNameNotEndingWith(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotEndWith("product_name", term))
	return r
}
func (r *OrderItemRequest) WithProductNameSoundingLike(term string) *OrderItemRequest {
	r.Query.AndFilter(core.ExprSoundLike("product_name", core.ValText(term)))
	return r
}
func (r *OrderItemRequest) OrderByProductNameAsc() *OrderItemRequest {
	r.Query.OrderAsc("product_name")
	return r
}
func (r *OrderItemRequest) OrderByProductNameDesc() *OrderItemRequest {
	r.Query.OrderDesc("product_name")
	return r
}
func (r *OrderItemRequest) SelectQuantity() *OrderItemRequest {
	r.Query.Project("quantity")
	return r
}

func (r *OrderItemRequest) WithQuantityIs(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("quantity", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithQuantityIsNot(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("quantity", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithQuantityIn(values []int64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("quantity", converted))
	return r
}
func (r *OrderItemRequest) WithQuantityNotIn(values []int64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("quantity", converted))
	return r
}
func (r *OrderItemRequest) WithQuantityGreaterThan(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("quantity", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithQuantityGreaterThanOrEqualTo(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("quantity", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithQuantityLessThan(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("quantity", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithQuantityLessThanOrEqualTo(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("quantity", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithQuantityBetween(lower int64, upper int64) *OrderItemRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("quantity", from, to))
	return r
}
func (r *OrderItemRequest) WithQuantityIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("quantity"))
	return r
}
func (r *OrderItemRequest) WithQuantityIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("quantity"))
	return r
}
func (r *OrderItemRequest) OrderByQuantityAsc() *OrderItemRequest {
	r.Query.OrderAsc("quantity")
	return r
}
func (r *OrderItemRequest) OrderByQuantityDesc() *OrderItemRequest {
	r.Query.OrderDesc("quantity")
	return r
}
func (r *OrderItemRequest) SelectUnitPrice() *OrderItemRequest {
	r.Query.Project("unit_price")
	return r
}

func (r *OrderItemRequest) WithUnitPriceIs(value decimal.Decimal) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("unit_price", core.ValDecimal(value)))
	return r
}
func (r *OrderItemRequest) WithUnitPriceIsNot(value decimal.Decimal) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("unit_price", core.ValDecimal(value)))
	return r
}
func (r *OrderItemRequest) WithUnitPriceIn(values []decimal.Decimal) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("unit_price", converted))
	return r
}
func (r *OrderItemRequest) WithUnitPriceNotIn(values []decimal.Decimal) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("unit_price", converted))
	return r
}
func (r *OrderItemRequest) WithUnitPriceGreaterThan(value decimal.Decimal) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("unit_price", core.ValDecimal(value)))
	return r
}
func (r *OrderItemRequest) WithUnitPriceGreaterThanOrEqualTo(value decimal.Decimal) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("unit_price", core.ValDecimal(value)))
	return r
}
func (r *OrderItemRequest) WithUnitPriceLessThan(value decimal.Decimal) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("unit_price", core.ValDecimal(value)))
	return r
}
func (r *OrderItemRequest) WithUnitPriceLessThanOrEqualTo(value decimal.Decimal) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("unit_price", core.ValDecimal(value)))
	return r
}
func (r *OrderItemRequest) WithUnitPriceBetween(lower decimal.Decimal, upper decimal.Decimal) *OrderItemRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode("unit_price", from, to))
	return r
}
func (r *OrderItemRequest) WithUnitPriceIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("unit_price"))
	return r
}
func (r *OrderItemRequest) WithUnitPriceIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("unit_price"))
	return r
}
func (r *OrderItemRequest) OrderByUnitPriceAsc() *OrderItemRequest {
	r.Query.OrderAsc("unit_price")
	return r
}
func (r *OrderItemRequest) OrderByUnitPriceDesc() *OrderItemRequest {
	r.Query.OrderDesc("unit_price")
	return r
}
func (r *OrderItemRequest) SelectVersion() *OrderItemRequest {
	r.Query.Project("version")
	return r
}

func (r *OrderItemRequest) WithVersionIs(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionIsNot(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionIn(values []int64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *OrderItemRequest) WithVersionNotIn(values []int64) *OrderItemRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *OrderItemRequest) WithVersionGreaterThan(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionGreaterThanOrEqualTo(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionLessThan(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionLessThanOrEqualTo(value int64) *OrderItemRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *OrderItemRequest) WithVersionBetween(lower int64, upper int64) *OrderItemRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *OrderItemRequest) WithVersionIsKnown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *OrderItemRequest) WithVersionIsUnknown() *OrderItemRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *OrderItemRequest) OrderByVersionAsc() *OrderItemRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *OrderItemRequest) OrderByVersionDesc() *OrderItemRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *OrderItemRequest) SelectCustomerOrderWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *OrderItemRequest {
	r.Query.Project("customer_order_id")
	r.Query.RelationQuerySelection("customerOrderEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectCustomerOrderWith", Relation: "customerOrderEntity"}
	r.relationFactories["customerOrderEntity"] = child.NewRelationEntity
	return r
}

func (r *OrderItemRequest) WithCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderItemRequest {
	r.Query.AndFilter(core.ExprInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *OrderItemRequest) WithoutCustomerOrderMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *OrderItemRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("customer_order_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableOrderItemRequest) NewEntity(context *runtime.UserContext) *OrderItem {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		panic("security audit failure: non-empty Comment() and Purpose() are required before NewEntity()")
	}
	entity := NewOrderItem()
	initialized := context.InitializeEntity("OrderItem", entity)
	typed, ok := initialized.(*OrderItem)
	if !ok {
		panic("entity initializer changed OrderItem to an incompatible type")
	}
	return typed
}

func (e *ExecutableOrderItemRequest) ExecuteForOne(context *runtime.UserContext) (*OrderItem, error) {
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

func (e *ExecutableOrderItemRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*OrderItem], error) {
	// Apply trusted policy once. The runtime owns the entire typed list lifecycle,
	// including optional exact count, facets and hydration failures.
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, plan, err := runtime.CaptureQueryPlan(query, r.queryOptions)
	if err != nil { return nil, err }
	query = plan.WithQueryDiagnostics(query)
	authorized, err := context.PrepareQuery(query)
	if err != nil { return nil, err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.QueryExecutor)
	if !ok { return nil, fmt.Errorf("dataService does not implement data_service.QueryExecutor") }
	service := runtime.NewRuntimeDataService(context.Metadata, ds)
	queryRoot := core.NewEntityRoot()
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*OrderItem, error) {
		entity := NewOrderItem()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["customerOrderEntity"]; selected {
			entity.markRelationLoaded("customerOrderEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["customerOrderEntity"]; factory != nil {
					childEntity := factory()
					if attachable, ok := childEntity.(interface { AttachEntityRoot(*core.EntityRoot) }); ok { attachable.AttachEntityRoot(entity.EntityRoot()) }
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("customerOrderEntity", childEntity)
				}
			}
		}
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableOrderItemRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*OrderItem) error) error {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForStream()")
	}
	if yield == nil {
		return fmt.Errorf("stream consumer must not be nil")
	}
	query := r.Query.Clone()
	query.Comment(r.commentText).Purpose(r.purposeText)
	query, err := runtime.CaptureStreamQuery(query, r.queryOptions)
	if err != nil { return err }
	authorized, err := context.PrepareQuery(query)
	if err != nil { return err }
	dsRaw := context.GetResource("dataService")
	ds, ok := dsRaw.(data_service.StreamQueryExecutor)
	if !ok {
		return fmt.Errorf("dataService does not implement data_service.StreamQueryExecutor")
	}
	queryRoot := core.NewEntityRoot()
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*OrderItem, error) {
		entity := NewOrderItem()
		entity.AttachEntityRoot(queryRoot)
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableOrderItemRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableOrderItemRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
	r := e.request
	if strings.TrimSpace(r.purposeText) == "" || strings.TrimSpace(r.commentText) == "" {
		return nil, nil, nil, fmt.Errorf("security audit failure: Comment() and Purpose() must be called before ExecuteForList()")
	}
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
func (e *ExecutableOrderItemRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *OrderItemRequest) Count() *OrderItemRequest {
	return r.CountAs("count")
}

func (r *OrderItemRequest) CountAs(alias string) *OrderItemRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *OrderItemRequest) MinQuantity() *OrderItemRequest {
	return r.MinQuantityAs("minOfQuantity")
}

func (r *OrderItemRequest) MinQuantityAs(alias string) *OrderItemRequest {
	r.Query.Min("quantity", alias)
	return r
}
func (r *OrderItemRequest) MaxQuantity() *OrderItemRequest {
	return r.MaxQuantityAs("maxOfQuantity")
}

func (r *OrderItemRequest) MaxQuantityAs(alias string) *OrderItemRequest {
	r.Query.Max("quantity", alias)
	return r
}
func (r *OrderItemRequest) SumQuantity() *OrderItemRequest {
	return r.SumQuantityAs("sumOfQuantity")
}

func (r *OrderItemRequest) SumQuantityAs(alias string) *OrderItemRequest {
	r.Query.Sum("quantity", alias)
	return r
}
func (r *OrderItemRequest) AvgQuantity() *OrderItemRequest {
	return r.AvgQuantityAs("avgOfQuantity")
}

func (r *OrderItemRequest) AvgQuantityAs(alias string) *OrderItemRequest {
	r.Query.Avg("quantity", alias)
	return r
}
func (r *OrderItemRequest) StddevQuantity() *OrderItemRequest {
	return r.StddevQuantityAs("standardDeviationOfQuantity")
}

func (r *OrderItemRequest) StddevQuantityAs(alias string) *OrderItemRequest {
	r.Query.Stddev("quantity", alias)
	return r
}
func (r *OrderItemRequest) StddevPopQuantity() *OrderItemRequest {
	return r.StddevPopQuantityAs("squareRootOfPopulationStandardDeviationOfQuantity")
}

func (r *OrderItemRequest) StddevPopQuantityAs(alias string) *OrderItemRequest {
	r.Query.StddevPop("quantity", alias)
	return r
}
func (r *OrderItemRequest) VarSampQuantity() *OrderItemRequest {
	return r.VarSampQuantityAs("sampleVarianceOfQuantity")
}

func (r *OrderItemRequest) VarSampQuantityAs(alias string) *OrderItemRequest {
	r.Query.VarSamp("quantity", alias)
	return r
}
func (r *OrderItemRequest) VarPopQuantity() *OrderItemRequest {
	return r.VarPopQuantityAs("samplePopulationVarianceOfQuantity")
}

func (r *OrderItemRequest) VarPopQuantityAs(alias string) *OrderItemRequest {
	r.Query.VarPop("quantity", alias)
	return r
}
func (r *OrderItemRequest) MinUnitPrice() *OrderItemRequest {
	return r.MinUnitPriceAs("minOfUnitPrice")
}

func (r *OrderItemRequest) MinUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.Min("unit_price", alias)
	return r
}
func (r *OrderItemRequest) MaxUnitPrice() *OrderItemRequest {
	return r.MaxUnitPriceAs("maxOfUnitPrice")
}

func (r *OrderItemRequest) MaxUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.Max("unit_price", alias)
	return r
}
func (r *OrderItemRequest) SumUnitPrice() *OrderItemRequest {
	return r.SumUnitPriceAs("sumOfUnitPrice")
}

func (r *OrderItemRequest) SumUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.Sum("unit_price", alias)
	return r
}
func (r *OrderItemRequest) AvgUnitPrice() *OrderItemRequest {
	return r.AvgUnitPriceAs("avgOfUnitPrice")
}

func (r *OrderItemRequest) AvgUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.Avg("unit_price", alias)
	return r
}
func (r *OrderItemRequest) StddevUnitPrice() *OrderItemRequest {
	return r.StddevUnitPriceAs("standardDeviationOfUnitPrice")
}

func (r *OrderItemRequest) StddevUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.Stddev("unit_price", alias)
	return r
}
func (r *OrderItemRequest) StddevPopUnitPrice() *OrderItemRequest {
	return r.StddevPopUnitPriceAs("squareRootOfPopulationStandardDeviationOfUnitPrice")
}

func (r *OrderItemRequest) StddevPopUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.StddevPop("unit_price", alias)
	return r
}
func (r *OrderItemRequest) VarSampUnitPrice() *OrderItemRequest {
	return r.VarSampUnitPriceAs("sampleVarianceOfUnitPrice")
}

func (r *OrderItemRequest) VarSampUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.VarSamp("unit_price", alias)
	return r
}
func (r *OrderItemRequest) VarPopUnitPrice() *OrderItemRequest {
	return r.VarPopUnitPriceAs("samplePopulationVarianceOfUnitPrice")
}

func (r *OrderItemRequest) VarPopUnitPriceAs(alias string) *OrderItemRequest {
	r.Query.VarPop("unit_price", alias)
	return r
}

func (r *OrderItemRequest) GroupById() *OrderItemRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *OrderItemRequest) GroupByCustomerOrder() *OrderItemRequest {
	r.Query.WithGroupBy("customer_order_id")
	return r
}
func (r *OrderItemRequest) GroupByProductName() *OrderItemRequest {
	r.Query.WithGroupBy("product_name")
	return r
}
func (r *OrderItemRequest) GroupByQuantity() *OrderItemRequest {
	r.Query.WithGroupBy("quantity")
	return r
}
func (r *OrderItemRequest) GroupByUnitPrice() *OrderItemRequest {
	r.Query.WithGroupBy("unit_price")
	return r
}
func (r *OrderItemRequest) GroupByVersion() *OrderItemRequest {
	r.Query.WithGroupBy("version")
	return r
}

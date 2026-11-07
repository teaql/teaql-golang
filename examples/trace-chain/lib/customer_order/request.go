

package customer_order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"trace-chain-service-core-workspace/lib/internal/generatedsearch"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/shipment"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type CustomerOrderRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableCustomerOrderRequest struct {
	request *CustomerOrderRequest
}

func NewCustomerOrderRequest() *CustomerOrderRequest {
	r := &CustomerOrderRequest{
		Query: core.NewSelectQuery("Customer Order"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewCustomerOrderMinimalRequest() *CustomerOrderRequest {
	r := NewCustomerOrderRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *CustomerOrderRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *CustomerOrderRequest) FilterByJson(source []byte) (*CustomerOrderRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *CustomerOrderRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *CustomerOrderRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *CustomerOrderRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewCustomerOrder().EntityDescriptor()
}

func (r *CustomerOrderRequest) NewRelationEntity() core.Entity {
	return newLoadedCustomerOrder()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *CustomerOrderRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *CustomerOrderRequest) teaqlHydrateQueryRecord(rec core.Record) (*CustomerOrder, error) {
	entity := newLoadedCustomerOrder()
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
	if relationValue, selected := rec["orderItemList"]; selected {
		factory := r.relationFactories["orderItemList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation orderItemList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation orderItemList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*order_item.OrderItem)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation orderItemList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.OrderItemList().Add(childEntity)
			}}
	if relationValue, selected := rec["paymentList"]; selected {
		factory := r.relationFactories["paymentList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation paymentList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation paymentList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*payment.Payment)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation paymentList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.PaymentList().Add(childEntity)
			}}
	if relationValue, selected := rec["shipmentList"]; selected {
		factory := r.relationFactories["shipmentList"]
		if factory == nil { return nil, fmt.Errorf("missing generated query hydrator for relation shipmentList") }
		childRecords, ok := relationValue.V.([]core.Record)
			if !ok { return nil, fmt.Errorf("relation shipmentList has unexpected runtime type %T", relationValue.V) }
			for _, childRecord := range childRecords {
				child, err := factory(childRecord)
				if err != nil { return nil, err }
				childEntity, ok := child.(*shipment.Shipment)
				if !ok { return nil, fmt.Errorf("unexpected generated query child type for relation shipmentList") }
				childEntity.AttachEntityRoot(entity.EntityRoot())
				entity.ShipmentList().Add(childEntity)
			}}
	return entity, nil
}

func (r *CustomerOrderRequest) Comment(comment string) *CustomerOrderRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *CustomerOrderRequest) Purpose(purpose string) *ExecutableCustomerOrderRequest {
	r.purposeText = purpose
	return &ExecutableCustomerOrderRequest{request: r}
}

func (r *ExecutableCustomerOrderRequest) Comment(comment string) *ExecutableCustomerOrderRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *CustomerOrderRequest) Limit(limit uint64) *CustomerOrderRequest {
	r.Query.Limit(limit)
	return r
}

func (r *CustomerOrderRequest) Offset(offset uint64) *CustomerOrderRequest {
	r.Query.Offset(offset)
	return r
}

func (r *CustomerOrderRequest) OptimizeForContinuousPageFetch() *CustomerOrderRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *CustomerOrderRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *CustomerOrderRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *CustomerOrderRequest) OptimizePaginationWithIDSet() *CustomerOrderRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *CustomerOrderRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *CustomerOrderRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *CustomerOrderRequest) TopNProbeParentThreshold(threshold uint64) *CustomerOrderRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeCustomerOrderVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeCustomerOrderVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *CustomerOrderRequest) WithDeletedRows() *CustomerOrderRequest {
	r.Query.Filter = removeCustomerOrderVersionFilter(r.Query.Filter)
	return r
}

func (r *CustomerOrderRequest) DeletedRowsOnly() *CustomerOrderRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *CustomerOrderRequest) SelectId() *CustomerOrderRequest {
	r.Query.Project("id")
	return r
}

func (r *CustomerOrderRequest) WithIdIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithIdBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithIdIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *CustomerOrderRequest) WithIdIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *CustomerOrderRequest) OrderByIdAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *CustomerOrderRequest) OrderByIdDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *CustomerOrderRequest) SelectPlatform() *CustomerOrderRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *CustomerOrderRequest) WithPlatformIs(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithPlatformIsNot(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithPlatformIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithPlatformNotIn(values []uint64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *CustomerOrderRequest) WithPlatformGreaterThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithPlatformLessThan(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithPlatformLessThanOrEqualTo(value uint64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *CustomerOrderRequest) WithPlatformBetween(lower uint64, upper uint64) *CustomerOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *CustomerOrderRequest) WithPlatformIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *CustomerOrderRequest) WithPlatformIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *CustomerOrderRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *CustomerOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *CustomerOrderRequest) OrderByPlatformAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *CustomerOrderRequest) OrderByPlatformDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *CustomerOrderRequest) SelectOrderNumber() *CustomerOrderRequest {
	r.Query.Project("order_number")
	return r
}

func (r *CustomerOrderRequest) WithOrderNumberIs(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsNot(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("order_number", converted))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberGreaterThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberLessThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("order_number", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberBetween(lower string, upper string) *CustomerOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("order_number", from, to))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberNotEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("order_number", term))
	return r
}
func (r *CustomerOrderRequest) WithOrderNumberSoundingLike(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("order_number", core.ValText(term)))
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("order_number")
	return r
}
func (r *CustomerOrderRequest) OrderByOrderNumberDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("order_number")
	return r
}
func (r *CustomerOrderRequest) SelectDescription() *CustomerOrderRequest {
	r.Query.Project("description")
	return r
}

func (r *CustomerOrderRequest) WithDescriptionIs(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("description", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionIsNot(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("description", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("description", converted))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionNotIn(values []string) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("description", converted))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionGreaterThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("description", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionGreaterThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("description", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionLessThan(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("description", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionLessThanOrEqualTo(value string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("description", core.ValText(value)))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionBetween(lower string, upper string) *CustomerOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("description", from, to))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("description"))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("description"))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprContain("description", term))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionNotContaining(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("description", term))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("description", term))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionNotStartingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("description", term))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("description", term))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionNotEndingWith(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("description", term))
	return r
}
func (r *CustomerOrderRequest) WithDescriptionSoundingLike(term string) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("description", core.ValText(term)))
	return r
}
func (r *CustomerOrderRequest) OrderByDescriptionAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("description")
	return r
}
func (r *CustomerOrderRequest) OrderByDescriptionDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("description")
	return r
}
func (r *CustomerOrderRequest) SelectVersion() *CustomerOrderRequest {
	r.Query.Project("version")
	return r
}

func (r *CustomerOrderRequest) WithVersionIs(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsNot(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionNotIn(values []int64) *CustomerOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionGreaterThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThan(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionLessThanOrEqualTo(value int64) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *CustomerOrderRequest) WithVersionBetween(lower int64, upper int64) *CustomerOrderRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsKnown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *CustomerOrderRequest) WithVersionIsUnknown() *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *CustomerOrderRequest) OrderByVersionAsc() *CustomerOrderRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *CustomerOrderRequest) OrderByVersionDesc() *CustomerOrderRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *CustomerOrderRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *CustomerOrderRequest {
	runtime.EnsureRelationProjection(r.Query, "platform_id")
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

func (r *CustomerOrderRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *CustomerOrderRequest) CountOrderItems() *CustomerOrderRequest {
	return r.CountOrderItemsAs("countOrderItems")

}
func (r *CustomerOrderRequest) CountOrderItemsAs(alias string) *CustomerOrderRequest {
	r.CountOrderItemsWith(alias, order_item.NewOrderItemRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	return r
}
func (r *CustomerOrderRequest) CountOrderItemsWith(alias string, child *order_item.OrderItemRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsWith", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}


func (r *CustomerOrderRequest) CountPayments() *CustomerOrderRequest {
	return r.CountPaymentsAs("countPayments")

}
func (r *CustomerOrderRequest) CountPaymentsAs(alias string) *CustomerOrderRequest {
	r.CountPaymentsWith(alias, payment.NewPaymentRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountPaymentsAs", Alias: &alias, Relation: "paymentList"}
	return r
}
func (r *CustomerOrderRequest) CountPaymentsWith(alias string, child *payment.PaymentRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("paymentList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountPaymentsWith", Alias: &alias, Relation: "paymentList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}


func (r *CustomerOrderRequest) CountShipments() *CustomerOrderRequest {
	return r.CountShipmentsAs("countShipments")

}
func (r *CustomerOrderRequest) CountShipmentsAs(alias string) *CustomerOrderRequest {
	r.CountShipmentsWith(alias, shipment.NewShipmentRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountShipmentsAs", Alias: &alias, Relation: "shipmentList"}
	return r
}
func (r *CustomerOrderRequest) CountShipmentsWith(alias string, child *shipment.ShipmentRequest) *CustomerOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("shipmentList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountShipmentsWith", Alias: &alias, Relation: "shipmentList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}



func (r *CustomerOrderRequest) SelectOrderItemList() *CustomerOrderRequest {
	return r.SelectOrderItemListWith(order_item.NewOrderItemRequest())
}

func (r *CustomerOrderRequest) SelectOrderItemListWith(child *order_item.OrderItemRequest) *CustomerOrderRequest {
	r.Query.RelationQuerySelection("orderItemList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderItemListWith", Relation: "orderItemList"}
	r.relationFactories["orderItemList"] = child.TeaqlHydrateQueryRecord
	return r
}
func (r *CustomerOrderRequest) SelectPaymentList() *CustomerOrderRequest {
	return r.SelectPaymentListWith(payment.NewPaymentRequest())
}

func (r *CustomerOrderRequest) SelectPaymentListWith(child *payment.PaymentRequest) *CustomerOrderRequest {
	r.Query.RelationQuerySelection("paymentList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectPaymentListWith", Relation: "paymentList"}
	r.relationFactories["paymentList"] = child.TeaqlHydrateQueryRecord
	return r
}
func (r *CustomerOrderRequest) SelectShipmentList() *CustomerOrderRequest {
	return r.SelectShipmentListWith(shipment.NewShipmentRequest())
}

func (r *CustomerOrderRequest) SelectShipmentListWith(child *shipment.ShipmentRequest) *CustomerOrderRequest {
	r.Query.RelationQuerySelection("shipmentList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectShipmentListWith", Relation: "shipmentList"}
	r.relationFactories["shipmentList"] = child.TeaqlHydrateQueryRecord
	return r
}

func (r *CustomerOrderRequest) HaveOrderItems() *CustomerOrderRequest {
	return r.WithOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *CustomerOrderRequest) HaveNoOrderItems() *CustomerOrderRequest {
	return r.WithoutOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *CustomerOrderRequest) WithOrderItemListMatching(child *order_item.OrderItemRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (r *CustomerOrderRequest) WithoutOrderItemListMatching(child *order_item.OrderItemRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}
func (r *CustomerOrderRequest) HavePayments() *CustomerOrderRequest {
	return r.WithPaymentListMatching(payment.NewPaymentRequest())
}

func (r *CustomerOrderRequest) HaveNoPayments() *CustomerOrderRequest {
	return r.WithoutPaymentListMatching(payment.NewPaymentRequest())
}

func (r *CustomerOrderRequest) WithPaymentListMatching(child *payment.PaymentRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (r *CustomerOrderRequest) WithoutPaymentListMatching(child *payment.PaymentRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}
func (r *CustomerOrderRequest) HaveShipments() *CustomerOrderRequest {
	return r.WithShipmentListMatching(shipment.NewShipmentRequest())
}

func (r *CustomerOrderRequest) HaveNoShipments() *CustomerOrderRequest {
	return r.WithoutShipmentListMatching(shipment.NewShipmentRequest())
}

func (r *CustomerOrderRequest) WithShipmentListMatching(child *shipment.ShipmentRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (r *CustomerOrderRequest) WithoutShipmentListMatching(child *shipment.ShipmentRequest) *CustomerOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "customer_order_id"))
	return r
}

func (e *ExecutableCustomerOrderRequest) NewEntity(context *runtime.UserContext) (*CustomerOrder, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	if context == nil { return nil, fmt.Errorf("UserContext is required for entity creation") }
	// Reserve before constructing mutation state or invoking trusted defaults.
	id, err := context.GenerateInternalId("Customer Order")
	if err != nil { return nil, err }
	entity := NewCustomerOrder()
	initialized := context.InitializeEntity("CustomerOrder", entity)
	typed, ok := initialized.(*CustomerOrder)
	if !ok || typed == nil {
		return nil, fmt.Errorf("entity initializer changed CustomerOrder to an incompatible type")
	}
	typed.UpdateId(id)
	if err := typed.base.RecordAllocatedID(id); err != nil { return nil, err }
	return typed, nil
}

func (e *ExecutableCustomerOrderRequest) ExecuteForOne(context *runtime.UserContext) (*CustomerOrder, error) {
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

func (e *ExecutableCustomerOrderRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*CustomerOrder], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*CustomerOrder, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableCustomerOrderRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*CustomerOrder) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*CustomerOrder, error) {
		entity := newLoadedCustomerOrder()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableCustomerOrderRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableCustomerOrderRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableCustomerOrderRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *CustomerOrderRequest) Count() *CustomerOrderRequest {
	return r.CountAs("count")
}

func (r *CustomerOrderRequest) CountAs(alias string) *CustomerOrderRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *CustomerOrderRequest) GroupById() *CustomerOrderRequest {
	r.Query.WithGroupBy("id")
	return r
}

func (r *CustomerOrderRequest) GroupByIdAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *CustomerOrderRequest) GroupByPlatform() *CustomerOrderRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}

func (r *CustomerOrderRequest) GroupByPlatformAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("platform_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("platform_id"))
	return r
}
func (r *CustomerOrderRequest) GroupByOrderNumber() *CustomerOrderRequest {
	r.Query.WithGroupBy("order_number")
	return r
}

func (r *CustomerOrderRequest) GroupByOrderNumberAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("order_number")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("order_number"))
	return r
}
func (r *CustomerOrderRequest) GroupByDescription() *CustomerOrderRequest {
	r.Query.WithGroupBy("description")
	return r
}

func (r *CustomerOrderRequest) GroupByDescriptionAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("description")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("description"))
	return r
}
func (r *CustomerOrderRequest) GroupByVersion() *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *CustomerOrderRequest) GroupByVersionAs(alias string) *CustomerOrderRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}

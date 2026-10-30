

package purchase_order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"checker-fix-advanced-core-workspace/lib/internal/generatedsearch"
	"checker-fix-advanced-core-workspace/lib/order_item"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type PurchaseOrderRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func() core.Entity
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutablePurchaseOrderRequest struct {
	request *PurchaseOrderRequest
}

func NewPurchaseOrderRequest() *PurchaseOrderRequest {
	r := &PurchaseOrderRequest{
		Query: core.NewSelectQuery("purchase_order"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func() core.Entity),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewPurchaseOrderMinimalRequest() *PurchaseOrderRequest {
	r := NewPurchaseOrderRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *PurchaseOrderRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *PurchaseOrderRequest) FilterByJson(source []byte) (*PurchaseOrderRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *PurchaseOrderRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *PurchaseOrderRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *PurchaseOrderRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewPurchaseOrder().EntityDescriptor()
}

func (r *PurchaseOrderRequest) NewRelationEntity() core.Entity {
	return newLoadedPurchaseOrder()
}

func (r *PurchaseOrderRequest) Comment(comment string) *PurchaseOrderRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *PurchaseOrderRequest) Purpose(purpose string) *ExecutablePurchaseOrderRequest {
	r.purposeText = purpose
	return &ExecutablePurchaseOrderRequest{request: r}
}

func (r *ExecutablePurchaseOrderRequest) Comment(comment string) *ExecutablePurchaseOrderRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *PurchaseOrderRequest) Limit(limit uint64) *PurchaseOrderRequest {
	r.Query.Limit(limit)
	return r
}

func (r *PurchaseOrderRequest) Offset(offset uint64) *PurchaseOrderRequest {
	r.Query.Offset(offset)
	return r
}

func (r *PurchaseOrderRequest) OptimizeForContinuousPageFetch() *PurchaseOrderRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *PurchaseOrderRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *PurchaseOrderRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *PurchaseOrderRequest) OptimizePaginationWithIDSet() *PurchaseOrderRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *PurchaseOrderRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *PurchaseOrderRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *PurchaseOrderRequest) TopNProbeParentThreshold(threshold uint64) *PurchaseOrderRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removePurchaseOrderVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removePurchaseOrderVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *PurchaseOrderRequest) WithDeletedRows() *PurchaseOrderRequest {
	r.Query.Filter = removePurchaseOrderVersionFilter(r.Query.Filter)
	return r
}

func (r *PurchaseOrderRequest) DeletedRowsOnly() *PurchaseOrderRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *PurchaseOrderRequest) SelectId() *PurchaseOrderRequest {
	r.Query.Project("id")
	return r
}

func (r *PurchaseOrderRequest) WithIdIs(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithIdIsNot(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithIdIn(values []uint64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *PurchaseOrderRequest) WithIdNotIn(values []uint64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *PurchaseOrderRequest) WithIdGreaterThan(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithIdGreaterThanOrEqualTo(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithIdLessThan(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithIdLessThanOrEqualTo(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithIdBetween(lower uint64, upper uint64) *PurchaseOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithIdIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *PurchaseOrderRequest) WithIdIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *PurchaseOrderRequest) OrderByIdAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *PurchaseOrderRequest) OrderByIdDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *PurchaseOrderRequest) SelectTenant() *PurchaseOrderRequest {
	r.Query.Project("tenant_id")
	return r
}

func (r *PurchaseOrderRequest) WithTenantIs(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("tenant_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithTenantIsNot(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("tenant_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithTenantIn(values []uint64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("tenant_id", converted))
	return r
}
func (r *PurchaseOrderRequest) WithTenantNotIn(values []uint64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("tenant_id", converted))
	return r
}
func (r *PurchaseOrderRequest) WithTenantGreaterThan(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("tenant_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithTenantGreaterThanOrEqualTo(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("tenant_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithTenantLessThan(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("tenant_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithTenantLessThanOrEqualTo(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("tenant_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithTenantBetween(lower uint64, upper uint64) *PurchaseOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("tenant_id", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithTenantIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("tenant_id"))
	return r
}
func (r *PurchaseOrderRequest) WithTenantIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("tenant_id"))
	return r
}
func (r *PurchaseOrderRequest) FacetByTenantAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *PurchaseOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "tenant_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByTenantAs", Alias: &name, Relation: "tenant"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *PurchaseOrderRequest) OrderByTenantAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("tenant_id")
	return r
}
func (r *PurchaseOrderRequest) OrderByTenantDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("tenant_id")
	return r
}
func (r *PurchaseOrderRequest) SelectContactPhone() *PurchaseOrderRequest {
	r.Query.Project("contact_phone")
	return r
}

func (r *PurchaseOrderRequest) WithContactPhoneIs(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("contact_phone", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneIsNot(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("contact_phone", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneIn(values []string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("contact_phone", converted))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneNotIn(values []string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("contact_phone", converted))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneGreaterThan(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("contact_phone", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneGreaterThanOrEqualTo(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("contact_phone", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneLessThan(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("contact_phone", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneLessThanOrEqualTo(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("contact_phone", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneBetween(lower string, upper string) *PurchaseOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("contact_phone", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("contact_phone"))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("contact_phone"))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprContain("contact_phone", term))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneNotContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("contact_phone", term))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("contact_phone", term))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneNotStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("contact_phone", term))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("contact_phone", term))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneNotEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("contact_phone", term))
	return r
}
func (r *PurchaseOrderRequest) WithContactPhoneSoundingLike(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("contact_phone", core.ValText(term)))
	return r
}
func (r *PurchaseOrderRequest) OrderByContactPhoneAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("contact_phone")
	return r
}
func (r *PurchaseOrderRequest) OrderByContactPhoneDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("contact_phone")
	return r
}
func (r *PurchaseOrderRequest) SelectInternalNote() *PurchaseOrderRequest {
	r.Query.Project("internal_note")
	return r
}

func (r *PurchaseOrderRequest) WithInternalNoteIs(value *string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteIsNot(value *string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteIn(values []*string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("internal_note", converted))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteNotIn(values []*string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("internal_note", converted))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteGreaterThan(value *string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteGreaterThanOrEqualTo(value *string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteLessThan(value *string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteLessThanOrEqualTo(value *string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("internal_note", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteBetween(lower *string, upper *string) *PurchaseOrderRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("internal_note", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("internal_note"))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("internal_note"))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprContain("internal_note", term))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteNotContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("internal_note", term))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("internal_note", term))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteNotStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("internal_note", term))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("internal_note", term))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteNotEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("internal_note", term))
	return r
}
func (r *PurchaseOrderRequest) WithInternalNoteSoundingLike(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("internal_note", core.ValText(term)))
	return r
}
func (r *PurchaseOrderRequest) OrderByInternalNoteAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("internal_note")
	return r
}
func (r *PurchaseOrderRequest) OrderByInternalNoteDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("internal_note")
	return r
}
func (r *PurchaseOrderRequest) SelectCreateTime() *PurchaseOrderRequest {
	r.Query.Project("create_time")
	return r
}

func (r *PurchaseOrderRequest) WithCreateTimeIs(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeIsNot(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeIn(values []time.Time) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeNotIn(values []time.Time) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeGreaterThan(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeLessThan(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *PurchaseOrderRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *PurchaseOrderRequest) WithCreateTimeIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *PurchaseOrderRequest) OrderByCreateTimeAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *PurchaseOrderRequest) OrderByCreateTimeDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *PurchaseOrderRequest) SelectUpdateTime() *PurchaseOrderRequest {
	r.Query.Project("update_time")
	return r
}

func (r *PurchaseOrderRequest) WithUpdateTimeIs(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeIsNot(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeIn(values []time.Time) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeNotIn(values []time.Time) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeGreaterThan(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeLessThan(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *PurchaseOrderRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *PurchaseOrderRequest) WithUpdateTimeIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *PurchaseOrderRequest) OrderByUpdateTimeAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *PurchaseOrderRequest) OrderByUpdateTimeDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *PurchaseOrderRequest) SelectVersion() *PurchaseOrderRequest {
	r.Query.Project("version")
	return r
}

func (r *PurchaseOrderRequest) WithVersionIs(value int64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithVersionIsNot(value int64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithVersionIn(values []int64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *PurchaseOrderRequest) WithVersionNotIn(values []int64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *PurchaseOrderRequest) WithVersionGreaterThan(value int64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithVersionGreaterThanOrEqualTo(value int64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithVersionLessThan(value int64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithVersionLessThanOrEqualTo(value int64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithVersionBetween(lower int64, upper int64) *PurchaseOrderRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithVersionIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *PurchaseOrderRequest) WithVersionIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *PurchaseOrderRequest) OrderByVersionAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *PurchaseOrderRequest) OrderByVersionDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *PurchaseOrderRequest) SelectTenantWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *PurchaseOrderRequest {
	runtime.EnsureRelationProjection(r.Query, "tenant_id")
	r.Query.RelationQuerySelection("tenantEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectTenantWith", Relation: "tenantEntity"}
	r.relationFactories["tenantEntity"] = child.NewRelationEntity
	return r
}

func (r *PurchaseOrderRequest) WithTenantMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("tenant_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PurchaseOrderRequest) WithoutTenantMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("tenant_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PurchaseOrderRequest) CountOrderItems() *PurchaseOrderRequest {
	return r.CountOrderItemsAs("countOrderItems")

}
func (r *PurchaseOrderRequest) CountOrderItemsAs(alias string) *PurchaseOrderRequest {
	r.CountOrderItemsWith(alias, order_item.NewOrderItemRequest())
	r.Query.RelationAggregates[len(r.Query.RelationAggregates)-1].Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	return r
}
func (r *PurchaseOrderRequest) CountOrderItemsWith(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Count(alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "count", Method: "CountOrderItemsWith", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PurchaseOrderRequest) SumUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.SumUnitPriceOfOrderItemsAs("sumUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) SumUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Sum("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SumUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) MinUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.MinUnitPriceOfOrderItemsAs("minUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) MinUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) MaxUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.MaxUnitPriceOfOrderItemsAs("maxUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) MaxUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) AvgUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.AvgUnitPriceOfOrderItemsAs("avgUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) AvgUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Avg("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "AvgUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) StandardDeviationUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.StandardDeviationUnitPriceOfOrderItemsAs("standardDeviationUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) StandardDeviationUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Stddev("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "StandardDeviationUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs("squareRootOfPopulationStandardDeviationUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.StddevPop("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SquareRootOfPopulationStandardDeviationUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) SampleVarianceUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.SampleVarianceUnitPriceOfOrderItemsAs("sampleVarianceUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) SampleVarianceUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarSamp("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SampleVarianceUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) SamplePopulationVarianceUnitPriceOfOrderItems() *PurchaseOrderRequest {
	return r.SamplePopulationVarianceUnitPriceOfOrderItemsAs("samplePopulationVarianceUnitPriceOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) SamplePopulationVarianceUnitPriceOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.VarPop("unit_price", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "SamplePopulationVarianceUnitPriceOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) MinCreateTimeOfOrderItems() *PurchaseOrderRequest {
	return r.MinCreateTimeOfOrderItemsAs("minCreateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) MinCreateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("create_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinCreateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) MaxCreateTimeOfOrderItems() *PurchaseOrderRequest {
	return r.MaxCreateTimeOfOrderItemsAs("maxCreateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) MaxCreateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("create_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxCreateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) MinUpdateTimeOfOrderItems() *PurchaseOrderRequest {
	return r.MinUpdateTimeOfOrderItemsAs("minUpdateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) MinUpdateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Min("update_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MinUpdateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}
func (r *PurchaseOrderRequest) MaxUpdateTimeOfOrderItems() *PurchaseOrderRequest {
	return r.MaxUpdateTimeOfOrderItemsAs("maxUpdateTimeOfOrderItems", order_item.NewOrderItemRequest())
}
func (r *PurchaseOrderRequest) MaxUpdateTimeOfOrderItemsAs(alias string, child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	query := child.GetQuery().Clone()
	query.Aggregates = nil
	query.Max("update_time", alias)
	aggregate := core.NewRelationAggregate("orderItemList", alias, query, true)
	aggregate.Operation = &core.QueryOperation{Kind: "aggregate", Method: "MaxUpdateTimeOfOrderItemsAs", Alias: &alias, Relation: "orderItemList"}
	r.Query.RelationAggregates = append(r.Query.RelationAggregates, aggregate)
	return r
}

func (r *PurchaseOrderRequest) SelectOrderItemList() *PurchaseOrderRequest {
	return r.SelectOrderItemListWith(order_item.NewOrderItemRequest())
}

func (r *PurchaseOrderRequest) SelectOrderItemListWith(child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	r.Query.RelationQuerySelection("orderItemList", child.GetQuerySelection())
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectOrderItemListWith", Relation: "orderItemList"}
	return r
}

func (r *PurchaseOrderRequest) HaveOrderItems() *PurchaseOrderRequest {
	return r.WithOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *PurchaseOrderRequest) HaveNoOrderItems() *PurchaseOrderRequest {
	return r.WithoutOrderItemListMatching(order_item.NewOrderItemRequest())
}

func (r *PurchaseOrderRequest) WithOrderItemListMatching(child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "purchase_order_id"))
	return r
}

func (r *PurchaseOrderRequest) WithoutOrderItemListMatching(child *order_item.OrderItemRequest) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("id", child.GetEntityDescriptor(), child.GetQuery(), "purchase_order_id"))
	return r
}

func (e *ExecutablePurchaseOrderRequest) NewEntity(context *runtime.UserContext) *PurchaseOrder {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewPurchaseOrder()
	initialized := context.InitializeEntity("PurchaseOrder", entity)
	typed, ok := initialized.(*PurchaseOrder)
	if !ok {
		panic("entity initializer changed PurchaseOrder to an incompatible type")
	}
	return typed
}

func (e *ExecutablePurchaseOrderRequest) ExecuteForOne(context *runtime.UserContext) (*PurchaseOrder, error) {
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

func (e *ExecutablePurchaseOrderRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*PurchaseOrder], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*PurchaseOrder, error) {
		entity := newLoadedPurchaseOrder()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if relationValue, selected := rec["tenantEntity"]; selected {
			entity.markRelationLoaded("tenantEntity")
			if childRecord, ok := relationValue.V.(core.Record); ok {
				if factory := e.request.relationFactories["tenantEntity"]; factory != nil {
					childEntity := factory()
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.setRelationEntity("tenantEntity", childEntity)
				}
			}
		}
		if relationValue, selected := rec["orderItemList"]; selected {
			childRecords, ok := relationValue.V.([]core.Record)
				if !ok { return nil, fmt.Errorf("relation orderItemList has unexpected runtime type %T", relationValue.V) }
				for _, childRecord := range childRecords {
					childEntity := order_item.NewOrderItem()
					childEntity.EntityRoot().ClearEntity(childEntity.EntityKey())
					childEntity.AttachEntityRoot(entity.EntityRoot())
					if err := childEntity.FromRecord(childRecord); err != nil { return nil, err }
					entity.OrderItemList().Add(childEntity)
				}}
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutablePurchaseOrderRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*PurchaseOrder) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*PurchaseOrder, error) {
		entity := newLoadedPurchaseOrder()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutablePurchaseOrderRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutablePurchaseOrderRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutablePurchaseOrderRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *PurchaseOrderRequest) Count() *PurchaseOrderRequest {
	return r.CountAs("count")
}

func (r *PurchaseOrderRequest) CountAs(alias string) *PurchaseOrderRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}


func (r *PurchaseOrderRequest) GroupById() *PurchaseOrderRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *PurchaseOrderRequest) GroupByTenant() *PurchaseOrderRequest {
	r.Query.WithGroupBy("tenant_id")
	return r
}
func (r *PurchaseOrderRequest) GroupByContactPhone() *PurchaseOrderRequest {
	r.Query.WithGroupBy("contact_phone")
	return r
}
func (r *PurchaseOrderRequest) GroupByInternalNote() *PurchaseOrderRequest {
	r.Query.WithGroupBy("internal_note")
	return r
}
func (r *PurchaseOrderRequest) GroupByCreateTime() *PurchaseOrderRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *PurchaseOrderRequest) GroupByUpdateTime() *PurchaseOrderRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *PurchaseOrderRequest) GroupByVersion() *PurchaseOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}

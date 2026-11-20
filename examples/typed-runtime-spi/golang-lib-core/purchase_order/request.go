

package purchase_order

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"order-service-routing-core-workspace/lib/internal/generatedsearch"
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
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutablePurchaseOrderRequest struct {
	request *PurchaseOrderRequest
}

func NewPurchaseOrderRequest() *PurchaseOrderRequest {
	r := &PurchaseOrderRequest{
		Query: core.NewSelectQuery("purchase_order"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
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
	return newLoadedPurchaseOrder().EntityDescriptor()
}

func (r *PurchaseOrderRequest) NewRelationEntity() core.Entity {
	return newLoadedPurchaseOrder()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *PurchaseOrderRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *PurchaseOrderRequest) teaqlHydrateQueryRecord(rec core.Record) (*PurchaseOrder, error) {
	entity := newLoadedPurchaseOrder()
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
	return entity, nil
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
func (r *PurchaseOrderRequest) SelectPlatform() *PurchaseOrderRequest {
	r.Query.Project("platform_id")
	return r
}

func (r *PurchaseOrderRequest) WithPlatformIs(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("platform_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformIsNot(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("platform_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformIn(values []uint64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("platform_id", converted))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformNotIn(values []uint64) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("platform_id", converted))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformGreaterThan(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("platform_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformGreaterThanOrEqualTo(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("platform_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformLessThan(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("platform_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformLessThanOrEqualTo(value uint64) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("platform_id", core.ValU64(value)))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformBetween(lower uint64, upper uint64) *PurchaseOrderRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("platform_id", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("platform_id"))
	return r
}
func (r *PurchaseOrderRequest) WithPlatformIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("platform_id"))
	return r
}
func (r *PurchaseOrderRequest) FacetByPlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *PurchaseOrderRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByPlatformAs", Alias: &name, Relation: "platform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *PurchaseOrderRequest) OrderByPlatformAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("platform_id")
	return r
}
func (r *PurchaseOrderRequest) OrderByPlatformDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("platform_id")
	return r
}
func (r *PurchaseOrderRequest) SelectReference() *PurchaseOrderRequest {
	r.Query.Project("reference")
	return r
}

func (r *PurchaseOrderRequest) WithReferenceIs(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("reference", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceIsNot(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("reference", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceIn(values []string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("reference", converted))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceNotIn(values []string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("reference", converted))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceGreaterThan(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("reference", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceGreaterThanOrEqualTo(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("reference", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceLessThan(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("reference", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceLessThanOrEqualTo(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("reference", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceBetween(lower string, upper string) *PurchaseOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("reference", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("reference"))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("reference"))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprContain("reference", term))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceNotContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("reference", term))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("reference", term))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceNotStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("reference", term))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("reference", term))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceNotEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("reference", term))
	return r
}
func (r *PurchaseOrderRequest) WithReferenceSoundingLike(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("reference", core.ValText(term)))
	return r
}
func (r *PurchaseOrderRequest) OrderByReferenceAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("reference")
	return r
}
func (r *PurchaseOrderRequest) OrderByReferenceDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("reference")
	return r
}
func (r *PurchaseOrderRequest) SelectDescription() *PurchaseOrderRequest {
	r.Query.Project("description")
	return r
}

func (r *PurchaseOrderRequest) WithDescriptionIs(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEq("description", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionIsNot(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNe("description", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionIn(values []string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("description", converted))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionNotIn(values []string) *PurchaseOrderRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("description", converted))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionGreaterThan(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGt("description", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionGreaterThanOrEqualTo(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprGte("description", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionLessThan(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLt("description", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionLessThanOrEqualTo(value string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprLte("description", core.ValText(value)))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionBetween(lower string, upper string) *PurchaseOrderRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("description", from, to))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionIsKnown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("description"))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionIsUnknown() *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprIsNullNode("description"))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprContain("description", term))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionNotContaining(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotContain("description", term))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprBeginWith("description", term))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionNotStartingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("description", term))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprEndWith("description", term))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionNotEndingWith(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotEndWith("description", term))
	return r
}
func (r *PurchaseOrderRequest) WithDescriptionSoundingLike(term string) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprSoundLike("description", core.ValText(term)))
	return r
}
func (r *PurchaseOrderRequest) OrderByDescriptionAsc() *PurchaseOrderRequest {
	r.Query.OrderAsc("description")
	return r
}
func (r *PurchaseOrderRequest) OrderByDescriptionDesc() *PurchaseOrderRequest {
	r.Query.OrderDesc("description")
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

func (r *PurchaseOrderRequest) SelectPlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *PurchaseOrderRequest {
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

func (r *PurchaseOrderRequest) WithPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *PurchaseOrderRequest) WithoutPlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *PurchaseOrderRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutablePurchaseOrderRequest) NewEntity(context *runtime.UserContext) (*PurchaseOrder, error) {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	return NewPurchaseOrder(context)
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
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
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

func (r *PurchaseOrderRequest) GroupByIdAs(alias string) *PurchaseOrderRequest {
	r.Query.WithGroupBy("id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("id"))
	return r
}
func (r *PurchaseOrderRequest) GroupByPlatform() *PurchaseOrderRequest {
	r.Query.WithGroupBy("platform_id")
	return r
}

func (r *PurchaseOrderRequest) GroupByPlatformAs(alias string) *PurchaseOrderRequest {
	r.Query.WithGroupBy("platform_id")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("platform_id"))
	return r
}
func (r *PurchaseOrderRequest) GroupByReference() *PurchaseOrderRequest {
	r.Query.WithGroupBy("reference")
	return r
}

func (r *PurchaseOrderRequest) GroupByReferenceAs(alias string) *PurchaseOrderRequest {
	r.Query.WithGroupBy("reference")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("reference"))
	return r
}
func (r *PurchaseOrderRequest) GroupByDescription() *PurchaseOrderRequest {
	r.Query.WithGroupBy("description")
	return r
}

func (r *PurchaseOrderRequest) GroupByDescriptionAs(alias string) *PurchaseOrderRequest {
	r.Query.WithGroupBy("description")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("description"))
	return r
}
func (r *PurchaseOrderRequest) GroupByVersion() *PurchaseOrderRequest {
	r.Query.WithGroupBy("version")
	return r
}

func (r *PurchaseOrderRequest) GroupByVersionAs(alias string) *PurchaseOrderRequest {
	r.Query.WithGroupBy("version")
	r.Query.ProjectExpr(alias, core.ExprColumnNode("version"))
	return r
}

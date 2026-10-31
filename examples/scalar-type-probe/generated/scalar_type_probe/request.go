

package scalar_type_probe

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"scalar-type-conformance-service-core-workspace/lib/internal/generatedsearch"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
)

type ScalarTypeProbeRequest struct {
	Query       *core.SelectQuery
	queryOptions *core.QueryOptions
	purposeText string
	commentText string
	relationFactories map[string]func(core.Record) (core.Entity, error)
	dynamicSearchWarnings []core.DynamicSearchWarning
}

type ExecutableScalarTypeProbeRequest struct {
	request *ScalarTypeProbeRequest
}

func NewScalarTypeProbeRequest() *ScalarTypeProbeRequest {
	r := &ScalarTypeProbeRequest{
		Query: core.NewSelectQuery("Scalar Type Probe"),
		queryOptions: core.NewQueryOptions(),
		relationFactories: make(map[string]func(core.Record) (core.Entity, error)),
	}
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(1)))
	return r
}

func NewScalarTypeProbeMinimalRequest() *ScalarTypeProbeRequest {
	r := NewScalarTypeProbeRequest()
	r.Query.Projects("id", "version")
	return r
}

func (r *ScalarTypeProbeRequest) GetQuery() *core.SelectQuery {
	return r.Query
}

// FilterByJson validates all clauses before changing this request.
func (r *ScalarTypeProbeRequest) FilterByJson(source []byte) (*ScalarTypeProbeRequest, error) {
	query, warnings, err := generatedsearch.Apply(r.Query, source, nil)
	if err != nil { return nil, err }
	r.Query = query
	r.dynamicSearchWarnings = append(r.dynamicSearchWarnings, warnings...)
	return r, nil
}

func (r *ScalarTypeProbeRequest) DynamicSearchWarnings() []core.DynamicSearchWarning {
	return append([]core.DynamicSearchWarning(nil), r.dynamicSearchWarnings...)
}

func (r *ScalarTypeProbeRequest) GetQuerySelection() *core.QuerySelection {
	selection := core.NewQuerySelection(r.Query)
	selection.QueryOptions = r.queryOptions
	return selection
}

func (r *ScalarTypeProbeRequest) GetEntityDescriptor() *core.EntityDescriptor {
	return NewScalarTypeProbe().EntityDescriptor()
}

func (r *ScalarTypeProbeRequest) NewRelationEntity() core.Entity {
	return newLoadedScalarTypeProbe()
}

// Generated query infrastructure. Preserve this request's selected nested
// factories without importing parent packages or using a global registry.
func (r *ScalarTypeProbeRequest) TeaqlHydrateQueryRecord(rec core.Record) (core.Entity, error) {
	return r.teaqlHydrateQueryRecord(rec)
}

func (r *ScalarTypeProbeRequest) teaqlHydrateQueryRecord(rec core.Record) (*ScalarTypeProbe, error) {
	entity := newLoadedScalarTypeProbe()
	if err := entity.FromRecord(rec); err != nil { return nil, err }
	if relationValue, selected := rec["typePlatformEntity"]; selected {
		entity.markRelationLoaded("typePlatformEntity")
		if childRecord, ok := relationValue.V.(core.Record); ok {
			if factory := r.relationFactories["typePlatformEntity"]; factory != nil {
				childEntity, err := factory(childRecord)
				if err != nil { return nil, err }
				entity.setRelationEntity("typePlatformEntity", childEntity)
			}
		}
	}
	return entity, nil
}

func (r *ScalarTypeProbeRequest) Comment(comment string) *ScalarTypeProbeRequest {
	r.commentText = comment
	r.Query.Comment(comment)
	return r
}

func (r *ScalarTypeProbeRequest) Purpose(purpose string) *ExecutableScalarTypeProbeRequest {
	r.purposeText = purpose
	return &ExecutableScalarTypeProbeRequest{request: r}
}

func (r *ExecutableScalarTypeProbeRequest) Comment(comment string) *ExecutableScalarTypeProbeRequest {
	r.request.commentText = comment
	r.request.Query.Comment(comment)
	return r
}

func (r *ScalarTypeProbeRequest) Limit(limit uint64) *ScalarTypeProbeRequest {
	r.Query.Limit(limit)
	return r
}

func (r *ScalarTypeProbeRequest) Offset(offset uint64) *ScalarTypeProbeRequest {
	r.Query.Offset(offset)
	return r
}

func (r *ScalarTypeProbeRequest) OptimizeForContinuousPageFetch() *ScalarTypeProbeRequest {
	r.Query.OptimizeForContinuousPageFetch()
	return r
}

func (r *ScalarTypeProbeRequest) OptimizeForContinuousPageFetchWith(namespace string, ttlSeconds uint64) *ScalarTypeProbeRequest {
	r.Query.OptimizeForContinuousPageFetchWith(namespace, ttlSeconds)
	return r
}

func (r *ScalarTypeProbeRequest) OptimizePaginationWithIDSet() *ScalarTypeProbeRequest {
	r.Query.OptimizePaginationWithIDSet()
	return r
}

func (r *ScalarTypeProbeRequest) OptimizePaginationWithIDSetConfig(namespace string, ttlSeconds, maxIDs uint64) *ScalarTypeProbeRequest {
	r.Query.OptimizePaginationWithIDSetConfig(namespace, ttlSeconds, maxIDs)
	return r
}

func (r *ScalarTypeProbeRequest) TopNProbeParentThreshold(threshold uint64) *ScalarTypeProbeRequest {
	r.Query.TopNProbeParentThreshold(threshold)
	return r
}

func removeScalarTypeProbeVersionFilter(expr *core.Expr) *core.Expr {
	if expr == nil { return nil }
	if expr.Type == core.ExprTypeBinary && expr.Left != nil &&
		expr.Left.Type == core.ExprTypeColumn && expr.Left.Column == "version" {
		return nil
	}
	if expr.Type != core.ExprTypeAnd { return expr }
	parts := make([]*core.Expr, 0, len(expr.Parts))
	for _, part := range expr.Parts {
		if kept := removeScalarTypeProbeVersionFilter(part); kept != nil {
			parts = append(parts, kept)
		}
	}
	if len(parts) == 0 { return nil }
	if len(parts) == 1 { return parts[0] }
	return core.ExprAndNode(parts...)
}

func (r *ScalarTypeProbeRequest) WithDeletedRows() *ScalarTypeProbeRequest {
	r.Query.Filter = removeScalarTypeProbeVersionFilter(r.Query.Filter)
	return r
}

func (r *ScalarTypeProbeRequest) DeletedRowsOnly() *ScalarTypeProbeRequest {
	r.WithDeletedRows()
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(-1)))
	return r
}

func (r *ScalarTypeProbeRequest) SelectId() *ScalarTypeProbeRequest {
	r.Query.Project("id")
	return r
}

func (r *ScalarTypeProbeRequest) WithIdIs(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdIsNot(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdIn(values []uint64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("id", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdNotIn(values []uint64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("id", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdGreaterThan(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdGreaterThanOrEqualTo(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdLessThan(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdLessThanOrEqualTo(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdBetween(lower uint64, upper uint64) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("id", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("id"))
	return r
}
func (r *ScalarTypeProbeRequest) WithIdIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("id"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByIdAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("id")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByIdDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("id")
	return r
}
func (r *ScalarTypeProbeRequest) SelectTypePlatform() *ScalarTypeProbeRequest {
	r.Query.Project("type_platform_id")
	return r
}

func (r *ScalarTypeProbeRequest) WithTypePlatformIs(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("type_platform_id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformIsNot(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("type_platform_id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformIn(values []uint64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprInList("type_platform_id", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformNotIn(values []uint64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValU64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("type_platform_id", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformGreaterThan(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("type_platform_id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformGreaterThanOrEqualTo(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("type_platform_id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformLessThan(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("type_platform_id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformLessThanOrEqualTo(value uint64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("type_platform_id", core.ValU64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformBetween(lower uint64, upper uint64) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValU64(value)
	value = upper
	to := core.ValU64(value)
	r.Query.AndFilter(core.ExprBetweenNode("type_platform_id", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("type_platform_id"))
	return r
}
func (r *ScalarTypeProbeRequest) WithTypePlatformIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("type_platform_id"))
	return r
}
func (r *ScalarTypeProbeRequest) FacetByTypePlatformAs(
	name string,
	nestedReq interface{ GetQuery() *core.SelectQuery },
	includeAllFacets ...bool,
) *ScalarTypeProbeRequest {
	includeAll := true
	if len(includeAllFacets) > 0 { includeAll = includeAllFacets[0] }
	facet := core.NewFacetRequest(name, "type_platform_id", core.QuerySelectionFor(nestedReq), includeAll)
	facet.Operation = &core.QueryOperation{Kind: "facet", Method: "FacetByTypePlatformAs", Alias: &name, Relation: "typePlatform"}
	r.queryOptions.Facets = append(r.queryOptions.Facets, facet)
	return r
}
func (r *ScalarTypeProbeRequest) OrderByTypePlatformAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("type_platform_id")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByTypePlatformDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("type_platform_id")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredString() *ScalarTypeProbeRequest {
	r.Query.Project("required_string")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredStringIs(value string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_string", core.ValText(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringIsNot(value string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_string", core.ValText(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringIn(values []string) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprInList("required_string", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringNotIn(values []string) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValText(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_string", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringGreaterThan(value string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_string", core.ValText(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringGreaterThanOrEqualTo(value string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_string", core.ValText(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringLessThan(value string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_string", core.ValText(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringLessThanOrEqualTo(value string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_string", core.ValText(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringBetween(lower string, upper string) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValText(value)
	value = upper
	to := core.ValText(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_string", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_string"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_string"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringContaining(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprContain("required_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringNotContaining(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotContain("required_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringStartingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprBeginWith("required_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringNotStartingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("required_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringEndingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEndWith("required_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringNotEndingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotEndWith("required_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredStringSoundingLike(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprSoundLike("required_string", core.ValText(term)))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredStringAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_string")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredStringDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_string")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalString() *ScalarTypeProbeRequest {
	r.Query.Project("optional_string")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalStringIs(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringIsNot(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringIn(values []*string) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_string", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringNotIn(values []*string) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_string", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringGreaterThan(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringGreaterThanOrEqualTo(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringLessThan(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringLessThanOrEqualTo(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringBetween(lower *string, upper *string) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_string", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_string"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_string"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringContaining(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprContain("optional_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringNotContaining(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotContain("optional_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringStartingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprBeginWith("optional_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringNotStartingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("optional_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringEndingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEndWith("optional_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringNotEndingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotEndWith("optional_string", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalStringSoundingLike(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprSoundLike("optional_string", core.ValText(term)))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalStringAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_string")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalStringDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_string")
	return r
}
func (r *ScalarTypeProbeRequest) SelectLargeText() *ScalarTypeProbeRequest {
	r.Query.Project("large_text")
	return r
}

func (r *ScalarTypeProbeRequest) WithLargeTextIs(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextIsNot(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextIn(values []*string) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("large_text", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextNotIn(values []*string) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("large_text", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextGreaterThan(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextGreaterThanOrEqualTo(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextLessThan(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextLessThanOrEqualTo(value *string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextBetween(lower *string, upper *string) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("large_text", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("large_text"))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("large_text"))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextContaining(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprContain("large_text", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextNotContaining(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotContain("large_text", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextStartingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprBeginWith("large_text", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextNotStartingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotBeginWith("large_text", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextEndingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEndWith("large_text", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextNotEndingWith(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotEndWith("large_text", term))
	return r
}
func (r *ScalarTypeProbeRequest) WithLargeTextSoundingLike(term string) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprSoundLike("large_text", core.ValText(term)))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByLargeTextAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("large_text")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByLargeTextDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("large_text")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredBoolean() *ScalarTypeProbeRequest {
	r.Query.Project("required_boolean")
	return r
}

func (r *ScalarTypeProbeRequest) WhichAreRequiredBoolean() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_boolean", core.ValBool(true)))
	return r
}
func (r *ScalarTypeProbeRequest) WhichAreNotRequiredBoolean() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_boolean", core.ValBool(false)))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredBooleanAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_boolean")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredBooleanDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_boolean")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalBoolean() *ScalarTypeProbeRequest {
	r.Query.Project("optional_boolean")
	return r
}

func (r *ScalarTypeProbeRequest) WhichAreOptionalBoolean() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_boolean", core.ValBool(true)))
	return r
}
func (r *ScalarTypeProbeRequest) WhichAreNotOptionalBoolean() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_boolean", core.ValBool(false)))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalBooleanIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_boolean"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalBooleanIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_boolean"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalBooleanAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_boolean")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalBooleanDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_boolean")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredInteger() *ScalarTypeProbeRequest {
	r.Query.Project("required_integer")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredIntegerIs(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_integer", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerIsNot(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_integer", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerIn(values []int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("required_integer", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerNotIn(values []int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_integer", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerGreaterThan(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_integer", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerGreaterThanOrEqualTo(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_integer", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerLessThan(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_integer", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerLessThanOrEqualTo(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_integer", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerBetween(lower int64, upper int64) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_integer", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_integer"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredIntegerIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_integer"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredIntegerAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_integer")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredIntegerDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_integer")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalInteger() *ScalarTypeProbeRequest {
	r.Query.Project("optional_integer")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalIntegerIs(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerIsNot(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerIn(values []*int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_integer", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerNotIn(values []*int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_integer", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerGreaterThan(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerGreaterThanOrEqualTo(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerLessThan(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerLessThanOrEqualTo(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerBetween(lower *int64, upper *int64) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_integer", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_integer"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalIntegerIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_integer"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalIntegerAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_integer")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalIntegerDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_integer")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredLong() *ScalarTypeProbeRequest {
	r.Query.Project("required_long")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredLongIs(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_long", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongIsNot(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_long", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongIn(values []int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("required_long", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongNotIn(values []int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_long", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongGreaterThan(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_long", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongGreaterThanOrEqualTo(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_long", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongLessThan(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_long", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongLessThanOrEqualTo(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_long", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongBetween(lower int64, upper int64) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_long", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_long"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredLongIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_long"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredLongAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_long")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredLongDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_long")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalLong() *ScalarTypeProbeRequest {
	r.Query.Project("optional_long")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalLongIs(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongIsNot(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongIn(values []*int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_long", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongNotIn(values []*int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_long", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongGreaterThan(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongGreaterThanOrEqualTo(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongLessThan(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongLessThanOrEqualTo(value *int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongBetween(lower *int64, upper *int64) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_long", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_long"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalLongIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_long"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalLongAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_long")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalLongDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_long")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredDecimal() *ScalarTypeProbeRequest {
	r.Query.Project("required_decimal")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredDecimalIs(value decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_decimal", core.ValDecimal(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalIsNot(value decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_decimal", core.ValDecimal(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalIn(values []decimal.Decimal) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprInList("required_decimal", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalNotIn(values []decimal.Decimal) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDecimal(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_decimal", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalGreaterThan(value decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_decimal", core.ValDecimal(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalGreaterThanOrEqualTo(value decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_decimal", core.ValDecimal(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalLessThan(value decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_decimal", core.ValDecimal(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalLessThanOrEqualTo(value decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_decimal", core.ValDecimal(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalBetween(lower decimal.Decimal, upper decimal.Decimal) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValDecimal(value)
	value = upper
	to := core.ValDecimal(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_decimal", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_decimal"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDecimalIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_decimal"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDecimalAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_decimal")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDecimalDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_decimal")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalDecimal() *ScalarTypeProbeRequest {
	r.Query.Project("optional_decimal")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalDecimalIs(value *decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalIsNot(value *decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalIn(values []*decimal.Decimal) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_decimal", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalNotIn(values []*decimal.Decimal) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_decimal", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalGreaterThan(value *decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalGreaterThanOrEqualTo(value *decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalLessThan(value *decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalLessThanOrEqualTo(value *decimal.Decimal) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalBetween(lower *decimal.Decimal, upper *decimal.Decimal) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_decimal", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_decimal"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDecimalIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_decimal"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDecimalAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_decimal")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDecimalDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_decimal")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredFloat() *ScalarTypeProbeRequest {
	r.Query.Project("required_float")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredFloatIs(value float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_float", core.ValF64(float64(value))))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatIsNot(value float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_float", core.ValF64(float64(value))))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatIn(values []float32) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValF64(float64(value)))
	}
	r.Query.AndFilter(core.ExprInList("required_float", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatNotIn(values []float32) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValF64(float64(value)))
	}
	r.Query.AndFilter(core.ExprNotInList("required_float", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatGreaterThan(value float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_float", core.ValF64(float64(value))))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatGreaterThanOrEqualTo(value float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_float", core.ValF64(float64(value))))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatLessThan(value float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_float", core.ValF64(float64(value))))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatLessThanOrEqualTo(value float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_float", core.ValF64(float64(value))))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatBetween(lower float32, upper float32) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValF64(float64(value))
	value = upper
	to := core.ValF64(float64(value))
	r.Query.AndFilter(core.ExprBetweenNode("required_float", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_float"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredFloatIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_float"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredFloatAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_float")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredFloatDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_float")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalFloat() *ScalarTypeProbeRequest {
	r.Query.Project("optional_float")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalFloatIs(value *float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatIsNot(value *float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatIn(values []*float32) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_float", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatNotIn(values []*float32) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_float", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatGreaterThan(value *float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatGreaterThanOrEqualTo(value *float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatLessThan(value *float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatLessThanOrEqualTo(value *float32) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatBetween(lower *float32, upper *float32) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_float", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_float"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalFloatIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_float"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalFloatAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_float")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalFloatDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_float")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredDouble() *ScalarTypeProbeRequest {
	r.Query.Project("required_double")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredDoubleIs(value float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_double", core.ValF64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleIsNot(value float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_double", core.ValF64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleIn(values []float64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValF64(value))
	}
	r.Query.AndFilter(core.ExprInList("required_double", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleNotIn(values []float64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValF64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_double", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleGreaterThan(value float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_double", core.ValF64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleGreaterThanOrEqualTo(value float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_double", core.ValF64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleLessThan(value float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_double", core.ValF64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleLessThanOrEqualTo(value float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_double", core.ValF64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleBetween(lower float64, upper float64) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValF64(value)
	value = upper
	to := core.ValF64(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_double", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_double"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDoubleIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_double"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDoubleAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_double")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDoubleDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_double")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalDouble() *ScalarTypeProbeRequest {
	r.Query.Project("optional_double")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalDoubleIs(value *float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleIsNot(value *float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleIn(values []*float64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_double", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleNotIn(values []*float64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_double", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleGreaterThan(value *float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleGreaterThanOrEqualTo(value *float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleLessThan(value *float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleLessThanOrEqualTo(value *float64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleBetween(lower *float64, upper *float64) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_double", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_double"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDoubleIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_double"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDoubleAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_double")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDoubleDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_double")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredDate() *ScalarTypeProbeRequest {
	r.Query.Project("required_date")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredDateIs(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_date", core.ValDate(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateIsNot(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_date", core.ValDate(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprInList("required_date", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateNotIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValDate(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_date", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateGreaterThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_date", core.ValDate(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateGreaterThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_date", core.ValDate(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateLessThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_date", core.ValDate(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateLessThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_date", core.ValDate(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateBetween(lower time.Time, upper time.Time) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValDate(value)
	value = upper
	to := core.ValDate(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_date", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_date"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDateIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_date"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDateAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_date")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDateDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_date")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalDate() *ScalarTypeProbeRequest {
	r.Query.Project("optional_date")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalDateIs(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateIsNot(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateIn(values []*time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_date", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateNotIn(values []*time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_date", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateGreaterThan(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateGreaterThanOrEqualTo(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateLessThan(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateLessThanOrEqualTo(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateBetween(lower *time.Time, upper *time.Time) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_date", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_date"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDateIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_date"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDateAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_date")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDateDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_date")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredTime() *ScalarTypeProbeRequest {
	r.Query.Project("required_time")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredTimeIs(value core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_time", core.ValTimeOfDay(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeIsNot(value core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_time", core.ValTimeOfDay(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeIn(values []core.TimeOfDay) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimeOfDay(value))
	}
	r.Query.AndFilter(core.ExprInList("required_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeNotIn(values []core.TimeOfDay) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimeOfDay(value))
	}
	r.Query.AndFilter(core.ExprNotInList("required_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeGreaterThan(value core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_time", core.ValTimeOfDay(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeGreaterThanOrEqualTo(value core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_time", core.ValTimeOfDay(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeLessThan(value core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_time", core.ValTimeOfDay(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeLessThanOrEqualTo(value core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_time", core.ValTimeOfDay(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeBetween(lower core.TimeOfDay, upper core.TimeOfDay) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValTimeOfDay(value)
	value = upper
	to := core.ValTimeOfDay(value)
	r.Query.AndFilter(core.ExprBetweenNode("required_time", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_time"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredTimeIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_time"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredTimeAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_time")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredTimeDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_time")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalTime() *ScalarTypeProbeRequest {
	r.Query.Project("optional_time")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalTimeIs(value *core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeIsNot(value *core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeIn(values []*core.TimeOfDay) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeNotIn(values []*core.TimeOfDay) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeGreaterThan(value *core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeGreaterThanOrEqualTo(value *core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeLessThan(value *core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeLessThanOrEqualTo(value *core.TimeOfDay) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeBetween(lower *core.TimeOfDay, upper *core.TimeOfDay) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_time", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_time"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalTimeIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_time"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalTimeAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_time")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalTimeDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_time")
	return r
}
func (r *ScalarTypeProbeRequest) SelectRequiredDatetime() *ScalarTypeProbeRequest {
	r.Query.Project("required_datetime")
	return r
}

func (r *ScalarTypeProbeRequest) WithRequiredDatetimeIs(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("required_datetime", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeIsNot(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("required_datetime", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("required_datetime", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeNotIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("required_datetime", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeGreaterThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("required_datetime", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeGreaterThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("required_datetime", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeLessThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("required_datetime", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeLessThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("required_datetime", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeBetween(lower time.Time, upper time.Time) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("required_datetime", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("required_datetime"))
	return r
}
func (r *ScalarTypeProbeRequest) WithRequiredDatetimeIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("required_datetime"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDatetimeAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("required_datetime")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByRequiredDatetimeDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("required_datetime")
	return r
}
func (r *ScalarTypeProbeRequest) SelectOptionalDatetime() *ScalarTypeProbeRequest {
	r.Query.Project("optional_datetime")
	return r
}

func (r *ScalarTypeProbeRequest) WithOptionalDatetimeIs(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeIsNot(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeIn(values []*time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }())
	}
	r.Query.AndFilter(core.ExprInList("optional_datetime", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeNotIn(values []*time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }())
	}
	r.Query.AndFilter(core.ExprNotInList("optional_datetime", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeGreaterThan(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeGreaterThanOrEqualTo(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeLessThan(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeLessThanOrEqualTo(value *time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeBetween(lower *time.Time, upper *time.Time) *ScalarTypeProbeRequest {
	value := lower
	from := func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()
	value = upper
	to := func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }()
	r.Query.AndFilter(core.ExprBetweenNode("optional_datetime", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("optional_datetime"))
	return r
}
func (r *ScalarTypeProbeRequest) WithOptionalDatetimeIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("optional_datetime"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDatetimeAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("optional_datetime")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByOptionalDatetimeDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("optional_datetime")
	return r
}
func (r *ScalarTypeProbeRequest) SelectCreateTime() *ScalarTypeProbeRequest {
	r.Query.Project("create_time")
	return r
}

func (r *ScalarTypeProbeRequest) WithCreateTimeIs(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeIsNot(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("create_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeNotIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("create_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeGreaterThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeGreaterThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeLessThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeLessThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("create_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeBetween(lower time.Time, upper time.Time) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("create_time", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("create_time"))
	return r
}
func (r *ScalarTypeProbeRequest) WithCreateTimeIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("create_time"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByCreateTimeAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("create_time")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByCreateTimeDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("create_time")
	return r
}
func (r *ScalarTypeProbeRequest) SelectUpdateTime() *ScalarTypeProbeRequest {
	r.Query.Project("update_time")
	return r
}

func (r *ScalarTypeProbeRequest) WithUpdateTimeIs(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeIsNot(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprInList("update_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeNotIn(values []time.Time) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValTimestamp(value.UnixMilli()))
	}
	r.Query.AndFilter(core.ExprNotInList("update_time", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeGreaterThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeGreaterThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeLessThan(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeLessThanOrEqualTo(value time.Time) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("update_time", core.ValTimestamp(value.UnixMilli())))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeBetween(lower time.Time, upper time.Time) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValTimestamp(value.UnixMilli())
	value = upper
	to := core.ValTimestamp(value.UnixMilli())
	r.Query.AndFilter(core.ExprBetweenNode("update_time", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("update_time"))
	return r
}
func (r *ScalarTypeProbeRequest) WithUpdateTimeIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("update_time"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByUpdateTimeAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("update_time")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByUpdateTimeDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("update_time")
	return r
}
func (r *ScalarTypeProbeRequest) SelectVersion() *ScalarTypeProbeRequest {
	r.Query.Project("version")
	return r
}

func (r *ScalarTypeProbeRequest) WithVersionIs(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprEq("version", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionIsNot(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNe("version", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionIn(values []int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprInList("version", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionNotIn(values []int64) *ScalarTypeProbeRequest {
	converted := make([]core.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, core.ValI64(value))
	}
	r.Query.AndFilter(core.ExprNotInList("version", converted))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionGreaterThan(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGt("version", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionGreaterThanOrEqualTo(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprGte("version", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionLessThan(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLt("version", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionLessThanOrEqualTo(value int64) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprLte("version", core.ValI64(value)))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionBetween(lower int64, upper int64) *ScalarTypeProbeRequest {
	value := lower
	from := core.ValI64(value)
	value = upper
	to := core.ValI64(value)
	r.Query.AndFilter(core.ExprBetweenNode("version", from, to))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionIsKnown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNotNullNode("version"))
	return r
}
func (r *ScalarTypeProbeRequest) WithVersionIsUnknown() *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprIsNullNode("version"))
	return r
}
func (r *ScalarTypeProbeRequest) OrderByVersionAsc() *ScalarTypeProbeRequest {
	r.Query.OrderAsc("version")
	return r
}
func (r *ScalarTypeProbeRequest) OrderByVersionDesc() *ScalarTypeProbeRequest {
	r.Query.OrderDesc("version")
	return r
}

func (r *ScalarTypeProbeRequest) SelectTypePlatformWith(child interface {
	GetQuery() *core.SelectQuery
	NewRelationEntity() core.Entity
}) *ScalarTypeProbeRequest {
	runtime.EnsureRelationProjection(r.Query, "type_platform_id")
	r.Query.RelationQuerySelection("typePlatformEntity", core.QuerySelectionFor(child))
	r.Query.Relations[len(r.Query.Relations)-1].Operation = &core.QueryOperation{Kind: "relation", Method: "SelectTypePlatformWith", Relation: "typePlatformEntity"}
	r.relationFactories["typePlatformEntity"] = func(record core.Record) (core.Entity, error) {
		if hydrator, ok := child.(interface { TeaqlHydrateQueryRecord(core.Record) (core.Entity, error) }); ok {
			return hydrator.TeaqlHydrateQueryRecord(record)
		}
		entity := child.NewRelationEntity()
		if err := entity.FromRecord(record); err != nil { return nil, err }
		return entity, nil
	}
	return r
}

func (r *ScalarTypeProbeRequest) WithTypePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprInSubQuery("type_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}

func (r *ScalarTypeProbeRequest) WithoutTypePlatformMatching(child interface {
	GetQuery() *core.SelectQuery
	GetEntityDescriptor() *core.EntityDescriptor
}) *ScalarTypeProbeRequest {
	r.Query.AndFilter(core.ExprNotInSubQuery("type_platform_id", child.GetEntityDescriptor(), child.GetQuery(), "id"))
	return r
}




func (e *ExecutableScalarTypeProbeRequest) NewEntity(context *runtime.UserContext) *ScalarTypeProbe {
	r := e.request
	if _, err := core.NewQueryIntent(&r.commentText, &r.purposeText); err != nil { panic(err) }
	entity := NewScalarTypeProbe()
	initialized := context.InitializeEntity("ScalarTypeProbe", entity)
	typed, ok := initialized.(*ScalarTypeProbe)
	if !ok {
		panic("entity initializer changed ScalarTypeProbe to an incompatible type")
	}
	return typed
}

func (e *ExecutableScalarTypeProbeRequest) ExecuteForOne(context *runtime.UserContext) (*ScalarTypeProbe, error) {
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

func (e *ExecutableScalarTypeProbeRequest) ExecuteForList(context *runtime.UserContext) (*core.SmartList[*ScalarTypeProbe], error) {
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
	return runtime.ExecuteTypedList(context, service, authorized, plan, func(rec core.Record) (*ScalarTypeProbe, error) {
		entity, err := r.teaqlHydrateQueryRecord(rec)
		if err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	})
}

// ExecuteForStream consumes a provider cursor one chunk at a time. Returning
// an error from yield releases the database resources. runtime.ErrStopStream
// records an intentional early stop; other callback errors remain failures.
func (e *ExecutableScalarTypeProbeRequest) ExecuteForStream(context *runtime.UserContext, chunkSize int, yield func(*ScalarTypeProbe) error) error {
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
	return runtime.ExecuteTypedStream(context, ds, authorized, chunkSize, func(rec core.Record) (*ScalarTypeProbe, error) {
		entity := newLoadedScalarTypeProbe()
		if err := entity.FromRecord(rec); err != nil { return nil, err }
		if err := runtime.CaptureEntityQueryProperties(entity, authorized); err != nil { return nil, err }
		return entity, nil
	}, yield)
}

func (e *ExecutableScalarTypeProbeRequest) ExecuteRecords(context *runtime.UserContext) ([]core.Record, error) {
	rows, _, _, err := e.executeRecords(context)
	return rows, err
}

// executeRecords returns the same authorized snapshot used for row execution
// so facets can derive their membership query without reapplying root policy.
func (e *ExecutableScalarTypeProbeRequest) executeRecords(context *runtime.UserContext) ([]core.Record, *core.SelectQuery, *runtime.FacetPlan, error) {
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
func (e *ExecutableScalarTypeProbeRequest) ExecuteForRows(context *runtime.UserContext) (*core.SmartList[core.Record], error) {
	rows, err := e.ExecuteRecords(context)
	if err != nil { return nil, err }
	return core.NewSmartList(rows), nil
}

func (r *ScalarTypeProbeRequest) Count() *ScalarTypeProbeRequest {
	return r.CountAs("count")
}

func (r *ScalarTypeProbeRequest) CountAs(alias string) *ScalarTypeProbeRequest {
	r.Query.CountField("id", alias)
	r.queryOptions.PageCountOperation = &core.QueryOperation{Kind: "count", Method: "CountAs", Alias: &alias}
	return r
}

func (r *ScalarTypeProbeRequest) MinRequiredInteger() *ScalarTypeProbeRequest {
	return r.MinRequiredIntegerAs("minOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) MinRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxRequiredInteger() *ScalarTypeProbeRequest {
	return r.MaxRequiredIntegerAs("maxOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) MaxRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumRequiredInteger() *ScalarTypeProbeRequest {
	return r.SumRequiredIntegerAs("sumOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) SumRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgRequiredInteger() *ScalarTypeProbeRequest {
	return r.AvgRequiredIntegerAs("avgOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) AvgRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevRequiredInteger() *ScalarTypeProbeRequest {
	return r.StddevRequiredIntegerAs("standardDeviationOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) StddevRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopRequiredInteger() *ScalarTypeProbeRequest {
	return r.StddevPopRequiredIntegerAs("squareRootOfPopulationStandardDeviationOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) StddevPopRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampRequiredInteger() *ScalarTypeProbeRequest {
	return r.VarSampRequiredIntegerAs("sampleVarianceOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) VarSampRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopRequiredInteger() *ScalarTypeProbeRequest {
	return r.VarPopRequiredIntegerAs("samplePopulationVarianceOfRequiredInteger")
}

func (r *ScalarTypeProbeRequest) VarPopRequiredIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("required_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinOptionalInteger() *ScalarTypeProbeRequest {
	return r.MinOptionalIntegerAs("minOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) MinOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxOptionalInteger() *ScalarTypeProbeRequest {
	return r.MaxOptionalIntegerAs("maxOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) MaxOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumOptionalInteger() *ScalarTypeProbeRequest {
	return r.SumOptionalIntegerAs("sumOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) SumOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgOptionalInteger() *ScalarTypeProbeRequest {
	return r.AvgOptionalIntegerAs("avgOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) AvgOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevOptionalInteger() *ScalarTypeProbeRequest {
	return r.StddevOptionalIntegerAs("standardDeviationOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) StddevOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopOptionalInteger() *ScalarTypeProbeRequest {
	return r.StddevPopOptionalIntegerAs("squareRootOfPopulationStandardDeviationOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) StddevPopOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampOptionalInteger() *ScalarTypeProbeRequest {
	return r.VarSampOptionalIntegerAs("sampleVarianceOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) VarSampOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopOptionalInteger() *ScalarTypeProbeRequest {
	return r.VarPopOptionalIntegerAs("samplePopulationVarianceOfOptionalInteger")
}

func (r *ScalarTypeProbeRequest) VarPopOptionalIntegerAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("optional_integer", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinRequiredLong() *ScalarTypeProbeRequest {
	return r.MinRequiredLongAs("minOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) MinRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxRequiredLong() *ScalarTypeProbeRequest {
	return r.MaxRequiredLongAs("maxOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) MaxRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumRequiredLong() *ScalarTypeProbeRequest {
	return r.SumRequiredLongAs("sumOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) SumRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgRequiredLong() *ScalarTypeProbeRequest {
	return r.AvgRequiredLongAs("avgOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) AvgRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevRequiredLong() *ScalarTypeProbeRequest {
	return r.StddevRequiredLongAs("standardDeviationOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) StddevRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopRequiredLong() *ScalarTypeProbeRequest {
	return r.StddevPopRequiredLongAs("squareRootOfPopulationStandardDeviationOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) StddevPopRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampRequiredLong() *ScalarTypeProbeRequest {
	return r.VarSampRequiredLongAs("sampleVarianceOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) VarSampRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopRequiredLong() *ScalarTypeProbeRequest {
	return r.VarPopRequiredLongAs("samplePopulationVarianceOfRequiredLong")
}

func (r *ScalarTypeProbeRequest) VarPopRequiredLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("required_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinOptionalLong() *ScalarTypeProbeRequest {
	return r.MinOptionalLongAs("minOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) MinOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxOptionalLong() *ScalarTypeProbeRequest {
	return r.MaxOptionalLongAs("maxOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) MaxOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumOptionalLong() *ScalarTypeProbeRequest {
	return r.SumOptionalLongAs("sumOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) SumOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgOptionalLong() *ScalarTypeProbeRequest {
	return r.AvgOptionalLongAs("avgOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) AvgOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevOptionalLong() *ScalarTypeProbeRequest {
	return r.StddevOptionalLongAs("standardDeviationOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) StddevOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopOptionalLong() *ScalarTypeProbeRequest {
	return r.StddevPopOptionalLongAs("squareRootOfPopulationStandardDeviationOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) StddevPopOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampOptionalLong() *ScalarTypeProbeRequest {
	return r.VarSampOptionalLongAs("sampleVarianceOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) VarSampOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopOptionalLong() *ScalarTypeProbeRequest {
	return r.VarPopOptionalLongAs("samplePopulationVarianceOfOptionalLong")
}

func (r *ScalarTypeProbeRequest) VarPopOptionalLongAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("optional_long", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinRequiredDecimal() *ScalarTypeProbeRequest {
	return r.MinRequiredDecimalAs("minOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) MinRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxRequiredDecimal() *ScalarTypeProbeRequest {
	return r.MaxRequiredDecimalAs("maxOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) MaxRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumRequiredDecimal() *ScalarTypeProbeRequest {
	return r.SumRequiredDecimalAs("sumOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) SumRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgRequiredDecimal() *ScalarTypeProbeRequest {
	return r.AvgRequiredDecimalAs("avgOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) AvgRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevRequiredDecimal() *ScalarTypeProbeRequest {
	return r.StddevRequiredDecimalAs("standardDeviationOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) StddevRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopRequiredDecimal() *ScalarTypeProbeRequest {
	return r.StddevPopRequiredDecimalAs("squareRootOfPopulationStandardDeviationOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) StddevPopRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampRequiredDecimal() *ScalarTypeProbeRequest {
	return r.VarSampRequiredDecimalAs("sampleVarianceOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) VarSampRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopRequiredDecimal() *ScalarTypeProbeRequest {
	return r.VarPopRequiredDecimalAs("samplePopulationVarianceOfRequiredDecimal")
}

func (r *ScalarTypeProbeRequest) VarPopRequiredDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("required_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinOptionalDecimal() *ScalarTypeProbeRequest {
	return r.MinOptionalDecimalAs("minOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) MinOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxOptionalDecimal() *ScalarTypeProbeRequest {
	return r.MaxOptionalDecimalAs("maxOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) MaxOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumOptionalDecimal() *ScalarTypeProbeRequest {
	return r.SumOptionalDecimalAs("sumOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) SumOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgOptionalDecimal() *ScalarTypeProbeRequest {
	return r.AvgOptionalDecimalAs("avgOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) AvgOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevOptionalDecimal() *ScalarTypeProbeRequest {
	return r.StddevOptionalDecimalAs("standardDeviationOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) StddevOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopOptionalDecimal() *ScalarTypeProbeRequest {
	return r.StddevPopOptionalDecimalAs("squareRootOfPopulationStandardDeviationOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) StddevPopOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampOptionalDecimal() *ScalarTypeProbeRequest {
	return r.VarSampOptionalDecimalAs("sampleVarianceOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) VarSampOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopOptionalDecimal() *ScalarTypeProbeRequest {
	return r.VarPopOptionalDecimalAs("samplePopulationVarianceOfOptionalDecimal")
}

func (r *ScalarTypeProbeRequest) VarPopOptionalDecimalAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("optional_decimal", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinRequiredFloat() *ScalarTypeProbeRequest {
	return r.MinRequiredFloatAs("minOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) MinRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxRequiredFloat() *ScalarTypeProbeRequest {
	return r.MaxRequiredFloatAs("maxOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) MaxRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumRequiredFloat() *ScalarTypeProbeRequest {
	return r.SumRequiredFloatAs("sumOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) SumRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgRequiredFloat() *ScalarTypeProbeRequest {
	return r.AvgRequiredFloatAs("avgOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) AvgRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevRequiredFloat() *ScalarTypeProbeRequest {
	return r.StddevRequiredFloatAs("standardDeviationOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) StddevRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopRequiredFloat() *ScalarTypeProbeRequest {
	return r.StddevPopRequiredFloatAs("squareRootOfPopulationStandardDeviationOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) StddevPopRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampRequiredFloat() *ScalarTypeProbeRequest {
	return r.VarSampRequiredFloatAs("sampleVarianceOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) VarSampRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopRequiredFloat() *ScalarTypeProbeRequest {
	return r.VarPopRequiredFloatAs("samplePopulationVarianceOfRequiredFloat")
}

func (r *ScalarTypeProbeRequest) VarPopRequiredFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("required_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinOptionalFloat() *ScalarTypeProbeRequest {
	return r.MinOptionalFloatAs("minOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) MinOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxOptionalFloat() *ScalarTypeProbeRequest {
	return r.MaxOptionalFloatAs("maxOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) MaxOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumOptionalFloat() *ScalarTypeProbeRequest {
	return r.SumOptionalFloatAs("sumOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) SumOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgOptionalFloat() *ScalarTypeProbeRequest {
	return r.AvgOptionalFloatAs("avgOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) AvgOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevOptionalFloat() *ScalarTypeProbeRequest {
	return r.StddevOptionalFloatAs("standardDeviationOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) StddevOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopOptionalFloat() *ScalarTypeProbeRequest {
	return r.StddevPopOptionalFloatAs("squareRootOfPopulationStandardDeviationOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) StddevPopOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampOptionalFloat() *ScalarTypeProbeRequest {
	return r.VarSampOptionalFloatAs("sampleVarianceOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) VarSampOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopOptionalFloat() *ScalarTypeProbeRequest {
	return r.VarPopOptionalFloatAs("samplePopulationVarianceOfOptionalFloat")
}

func (r *ScalarTypeProbeRequest) VarPopOptionalFloatAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("optional_float", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinRequiredDouble() *ScalarTypeProbeRequest {
	return r.MinRequiredDoubleAs("minOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) MinRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxRequiredDouble() *ScalarTypeProbeRequest {
	return r.MaxRequiredDoubleAs("maxOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) MaxRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumRequiredDouble() *ScalarTypeProbeRequest {
	return r.SumRequiredDoubleAs("sumOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) SumRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgRequiredDouble() *ScalarTypeProbeRequest {
	return r.AvgRequiredDoubleAs("avgOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) AvgRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevRequiredDouble() *ScalarTypeProbeRequest {
	return r.StddevRequiredDoubleAs("standardDeviationOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) StddevRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopRequiredDouble() *ScalarTypeProbeRequest {
	return r.StddevPopRequiredDoubleAs("squareRootOfPopulationStandardDeviationOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) StddevPopRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampRequiredDouble() *ScalarTypeProbeRequest {
	return r.VarSampRequiredDoubleAs("sampleVarianceOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) VarSampRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopRequiredDouble() *ScalarTypeProbeRequest {
	return r.VarPopRequiredDoubleAs("samplePopulationVarianceOfRequiredDouble")
}

func (r *ScalarTypeProbeRequest) VarPopRequiredDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("required_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MinOptionalDouble() *ScalarTypeProbeRequest {
	return r.MinOptionalDoubleAs("minOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) MinOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Min("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) MaxOptionalDouble() *ScalarTypeProbeRequest {
	return r.MaxOptionalDoubleAs("maxOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) MaxOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Max("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) SumOptionalDouble() *ScalarTypeProbeRequest {
	return r.SumOptionalDoubleAs("sumOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) SumOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Sum("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) AvgOptionalDouble() *ScalarTypeProbeRequest {
	return r.AvgOptionalDoubleAs("avgOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) AvgOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Avg("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevOptionalDouble() *ScalarTypeProbeRequest {
	return r.StddevOptionalDoubleAs("standardDeviationOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) StddevOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.Stddev("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) StddevPopOptionalDouble() *ScalarTypeProbeRequest {
	return r.StddevPopOptionalDoubleAs("squareRootOfPopulationStandardDeviationOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) StddevPopOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.StddevPop("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarSampOptionalDouble() *ScalarTypeProbeRequest {
	return r.VarSampOptionalDoubleAs("sampleVarianceOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) VarSampOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarSamp("optional_double", alias)
	return r
}
func (r *ScalarTypeProbeRequest) VarPopOptionalDouble() *ScalarTypeProbeRequest {
	return r.VarPopOptionalDoubleAs("samplePopulationVarianceOfOptionalDouble")
}

func (r *ScalarTypeProbeRequest) VarPopOptionalDoubleAs(alias string) *ScalarTypeProbeRequest {
	r.Query.VarPop("optional_double", alias)
	return r
}

func (r *ScalarTypeProbeRequest) GroupById() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("id")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByTypePlatform() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("type_platform_id")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredString() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_string")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalString() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_string")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByLargeText() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("large_text")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredBoolean() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_boolean")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalBoolean() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_boolean")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredInteger() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_integer")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalInteger() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_integer")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredLong() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_long")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalLong() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_long")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredDecimal() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_decimal")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalDecimal() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_decimal")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredFloat() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_float")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalFloat() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_float")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredDouble() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_double")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalDouble() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_double")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredDate() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_date")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalDate() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_date")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredTime() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_time")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalTime() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_time")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByRequiredDatetime() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("required_datetime")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByOptionalDatetime() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("optional_datetime")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByCreateTime() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("create_time")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByUpdateTime() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("update_time")
	return r
}
func (r *ScalarTypeProbeRequest) GroupByVersion() *ScalarTypeProbeRequest {
	r.Query.WithGroupBy("version")
	return r
}

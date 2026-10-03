package runtime

import (
	stdcontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func TestExecuteFacetsCountsAndIncludeAll(t *testing.T) {
	executor := &facetExecutor{}
	service := NewRuntimeDataService(nil, executor)
	outer := core.NewSelectQuery("School").Comment("verify query fixture").Purpose("preserve the query regression contract")
	outer.AndFilter(core.ExprContain("name", "Riverside"))
	nested := core.NewQuerySelection(core.NewSelectQuery("SchoolType").Comment("verify query fixture").Purpose("preserve the query regression contract").Count("schoolCount"))
	options := core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest("types", "schoolType", nested, true))
	result, err := ExecuteFacets(stdcontext.Background(), service, outer, options)
	assert.NoError(t, err)
	assert.True(t, executor.sawOuterFilter)
	assert.Len(t, result["types"].Data, 3)
	count, _ := result["types"].Data[0]["schoolCount"].TryU64()
	assert.Equal(t, uint64(2), count)

	options.Facets[0].IncludeAllFacets = false
	result, err = ExecuteFacets(stdcontext.Background(), service, outer, options)
	assert.NoError(t, err)
	assert.Len(t, result["types"].Data, 1)
}

func TestExecuteFacetsAppliesPolicyToNestedEntityBeforeProvider(t *testing.T) {
	executor := &facetExecutor{}
	service := NewRuntimeDataService(nil, executor)
	outer := core.NewSelectQuery("School").Comment("verify query fixture").Purpose("preserve the query regression contract")
	options := core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest(
		"types", "schoolType", core.NewQuerySelection(core.NewSelectQuery("SchoolType").Comment("verify query fixture").Purpose("preserve the query regression contract")), true))
	policy := &tenantQueryPolicy{rejected: "SchoolType"}
	context := NewUserContext().WithRequestPolicy(policy)

	_, err := ExecuteFacets(context, service, outer, options)
	assert.True(t, errors.Is(err, errQueryDenied))
	assert.Equal(t, 1, executor.calls, "nested facet denial must happen before its provider query")
}

func TestFacetRelationUsesMetadataAndRejectsAmbiguousAliases(t *testing.T) {
	first := core.NewRelationDescriptor("platformEntity", "Platform").LocalKey("platform_id")
	descriptor := core.NewEntityDescriptor("School").Relation(first)
	assert.Same(t, first, facetRelation(descriptor, "platform_id", "Platform"))
	assert.Nil(t, facetRelation(descriptor, "platform_id", "Employee"))
	descriptor.Relation(core.NewRelationDescriptor("otherPlatform", "Platform").LocalKey("platform_id"))
	assert.Nil(t, facetRelation(descriptor, "platform_id", "Platform"))
	assert.Same(t, first, facetRelation(descriptor, "platformEntity", "Platform"))
}

func TestFacetCaptureOwnsNestedSelectionBeforeExecution(t *testing.T) {
	nested := core.NewSelectQuery("SchoolType").WithFilter(core.ExprEq("code", core.ValText("PRIMARY")))
	options := core.NewQueryOptions()
	selection := core.NewQuerySelection(nested)
	options.Facets = append(options.Facets, core.NewFacetRequest("types", "school_type", selection, true))
	selection.QueryOptions.Facets = append(selection.QueryOptions.Facets,
		core.NewFacetRequest("platforms", "platform_id", core.NewQuerySelection(core.NewSelectQuery("Platform")), false))
	captured, err := captureFacets(options, make(map[*core.QueryOptions]bool))
	assert.NoError(t, err)
	nested.Filter = nil
	options.Facets[0].FacetName = "changed"
	selection.QueryOptions.Facets = nil
	assert.Equal(t, "types", captured[0].name)
	assert.NotNil(t, captured[0].query.Filter)
	assert.Len(t, captured[0].children, 1)
}

func TestFacetCycleFailsBeforeProvider(t *testing.T) {
	executor := &facetExecutor{}
	options := core.NewQueryOptions()
	selection := core.NewQuerySelection(core.NewSelectQuery("School"))
	selection.QueryOptions = options
	options.Facets = append(options.Facets, core.NewFacetRequest("cycle", "parent", selection, true))
	_, err := ExecuteFacets(stdcontext.Background(), NewRuntimeDataService(nil, executor),
		core.NewSelectQuery("School").Comment("cyclic facet").Purpose("reject before execution"), options)
	assert.ErrorContains(t, err, "cyclic facet")
	assert.Zero(t, executor.calls)
}

func TestFacetPlanCapturesBeforeCallerChangesAndCanBeReused(t *testing.T) {
	executor := &facetExecutor{}
	service := NewRuntimeDataService(nil, executor)
	options := core.NewQueryOptions()
	nested := core.NewQuerySelection(core.NewSelectQuery("SchoolType").WithFilter(core.ExprEq("code", core.ValText("PRIVATE-PLAN"))))
	options.Facets = append(options.Facets, core.NewFacetRequest("types", "schoolType", nested, false))
	plan, err := CaptureFacetPlan(options)
	assert.NoError(t, err)
	nested.Query.Filter = nil
	options.Facets[0].FacetName = "changed"
	options.Facets = nil
	assert.True(t, plan.HasFacets())
	assert.NotNil(t, plan.facets[0].query.Filter)
	encoded, err := json.Marshal(plan)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(encoded)+fmt.Sprintf("%+v %#v", plan, plan), "PRIVATE-PLAN"))
	for i := 0; i < 2; i++ {
		result, err := plan.Execute(stdcontext.Background(), service,
			core.NewSelectQuery("School").Comment("captured plan").Purpose("immutable execution"))
		assert.NoError(t, err)
		assert.Len(t, result["types"].Data, 1)
		assert.Nil(t, plan.facets[0].query.CommentText, "execution must not mutate captured selection")
	}
}

type facetExecutor struct {
	sawOuterFilter bool
	calls          int
}

func TestRelationFacetCaptureRejectsCyclesAndDetachesBuilders(t *testing.T) {
	root := core.NewSelectQuery("School")
	child := core.NewSelectQuery("SchoolType")
	selection := core.NewQuerySelection(child)
	selection.QueryOptions.Facets = append(selection.QueryOptions.Facets,
		core.NewFacetRequest("platforms", "platform", core.NewQuerySelection(core.NewSelectQuery("Platform")), false))
	root.RelationQuerySelection("type", selection)
	captured, plan, err := CaptureQueryPlan(root, nil)
	assert.NoError(t, err)
	selection.QueryOptions.Facets = nil
	child.AndFilter(core.ExprEq("id", core.ValU64(0)))
	assert.True(t, plan.relations["type"].HasFacets())
	assert.Nil(t, captured.Relations[0].Selection)
	assert.Nil(t, captured.Relations[0].Query.Filter)
	child.RelationQuerySelection("cycle", core.NewQuerySelection(root))
	_, _, err = CaptureQueryPlan(root, nil)
	assert.ErrorContains(t, err, "cyclic relation selection")
}

func (f *facetExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}
func (f *facetExecutor) Query(context stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	f.calls++
	if request.Query.Entity == "School" {
		if len(request.Query.Aggregates) > 0 {
			f.sawOuterFilter = request.Query.Filter != nil
			return &data_service.QueryResult{Rows: []core.Record{
				{"schoolType": core.ValU64(1001), "__teaql_facet_count": core.ValU64(2)},
			}}, nil
		}
		return &data_service.QueryResult{Rows: []core.Record{
			{"id": core.ValU64(1), "schoolType": core.ValU64(1001)},
			{"id": core.ValU64(2), "schoolType": core.ValU64(1001)},
		}}, nil
	}
	return &data_service.QueryResult{Rows: []core.Record{
		{"id": core.ValU64(1001)}, {"id": core.ValU64(1002)}, {"id": core.ValU64(1003)},
	}}, nil
}
func (f *facetExecutor) Mutate(context stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	panic("unused")
}

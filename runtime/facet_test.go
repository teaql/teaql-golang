package runtime

import (
	stdcontext "context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func TestExecuteFacetsCountsAndIncludeAll(t *testing.T) {
	executor := &facetExecutor{}
	service := NewRuntimeDataService(nil, executor)
	outer := core.NewSelectQuery("School")
	outer.AndFilter(core.ExprContain("name", "Riverside"))
	nested := core.NewQuerySelection(core.NewSelectQuery("SchoolType").Count("schoolCount"))
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

type facetExecutor struct{ sawOuterFilter bool }

func (f *facetExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}
func (f *facetExecutor) Query(context stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
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

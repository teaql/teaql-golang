package runtime

import (
	stdcontext "context"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

// ExecuteFacets evaluates relation membership from the filtered outer query,
// then decorates the typed nested-query rows with their matching counts.
func ExecuteFacets(
	context stdcontext.Context,
	service *RuntimeDataService,
	outer *core.SelectQuery,
	options *core.QueryOptions,
) (map[string]*core.SmartList[core.Record], error) {
	request, err := data_service.NewQueryRequest(outer)
	if err != nil {
		return nil, err
	}
	outer = request.Query
	results := make(map[string]*core.SmartList[core.Record])
	for _, facet := range options.Facets {
		membership := outer.ForExactCount("__teaql_facet_count")
		membership.GroupBy = []string{facet.RelationName}
		var intent logprivacy.IntentSource
		rows, err := service.fetchAllWithIntent(context, membership, &intent)
		if err != nil {
			return nil, err
		}
		counts := make(map[string]uint64)
		for _, row := range rows {
			if value, ok := row[facet.RelationName]; ok && value.V != nil {
				count, valid := row["__teaql_facet_count"].TryU64()
				if valid {
					counts[fmt.Sprint(value.V)] = count
				}
			}
		}

		nested := facet.Query.IntoQuery()
		nested = nested.Clone()
		nested.CommentText, nested.PurposeText = request.Comment, request.Purpose
		nested.TraceChain = core.QueryTraceSource(outer.Entity, outer.TraceChain, *request.Comment, *request.Purpose)
		if service.metadata != nil {
			if descriptor := service.metadata.Entity(outer.Entity); descriptor != nil {
				if relation := descriptor.RelationByName(facet.RelationName); relation != nil && relation.TargetEntity == nested.Entity {
					nested.TraceChain = append(nested.TraceChain, core.NewTypedTraceNode("relation", facet.RelationName, outer.Entity+"."+facet.RelationName))
				}
			}
		}
		countAliases := make([]string, 0)
		for _, aggregate := range nested.Aggregates {
			if aggregate.Function == core.AggCount {
				countAliases = append(countAliases, aggregate.Alias)
			}
		}
		nested.Aggregates = nil
		nested.GroupBy = nil
		if userContext, ok := UserContextFrom(context); ok {
			var err error
			nested, err = userContext.PrepareQuery(nested)
			if err != nil {
				return nil, err
			}
		}
		facetRows, err := service.fetchAllWithIntent(context, nested, &intent)
		if err != nil {
			return nil, err
		}
		decorated := make([]core.Record, 0, len(facetRows))
		for _, row := range facetRows {
			id, ok := row["id"]
			if !ok {
				continue
			}
			count := counts[fmt.Sprint(id.V)]
			if !facet.IncludeAllFacets && count == 0 {
				continue
			}
			if len(countAliases) == 0 {
				countAliases = []string{"count"}
			}
			for _, alias := range countAliases {
				row[alias] = core.ValU64(count)
			}
			decorated = append(decorated, row)
		}
		results[facet.FacetName] = core.NewSmartList(decorated)
	}
	return results, nil
}

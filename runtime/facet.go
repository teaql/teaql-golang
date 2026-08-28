package runtime

import (
	stdcontext "context"
	"fmt"

	"github.com/teaql/teaql-golang/core"
)

// ExecuteFacets evaluates relation membership from the filtered outer query,
// then decorates the typed nested-query rows with their matching counts.
func ExecuteFacets(
	context stdcontext.Context,
	service *RuntimeDataService,
	outer *core.SelectQuery,
	options *core.QueryOptions,
) (map[string]*core.SmartList[core.Record], error) {
	results := make(map[string]*core.SmartList[core.Record])
	for _, facet := range options.Facets {
		membership := cloneSelectQuery(outer, outer.Entity)
		membership.Projection = nil
		membership.Relations = nil
		membership.OrderBy = nil
		membership.Slice = nil
		membership.Aggregates = []*core.Aggregate{core.AggCountField("id", "__teaql_facet_count")}
		membership.GroupBy = []string{facet.RelationName}
		membership.ObjectGroupBys = nil
		rows, err := service.FetchAll(context, membership)
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
		countAliases := make([]string, 0)
		for _, aggregate := range nested.Aggregates {
			if aggregate.Function == core.AggCount {
				countAliases = append(countAliases, aggregate.Alias)
			}
		}
		nested.Aggregates = nil
		nested.GroupBy = nil
		facetRows, err := service.FetchAll(context, nested)
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

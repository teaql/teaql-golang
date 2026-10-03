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
	facets, err := captureFacets(options, make(map[*core.QueryOptions]bool))
	if err != nil {
		return nil, err
	}
	return executeCapturedFacets(context, service, outer, facets, logprivacy.IntentSource{})
}

type capturedFacet struct {
	name, relation string
	query          *core.SelectQuery
	includeAll     bool
	children       []capturedFacet
}

func captureFacets(options *core.QueryOptions, active map[*core.QueryOptions]bool) ([]capturedFacet, error) {
	if options == nil {
		return nil, nil
	}
	if active[options] {
		return nil, fmt.Errorf("cyclic facet selection")
	}
	active[options] = true
	defer delete(active, options)
	var result []capturedFacet
	for _, facet := range options.Facets {
		if facet == nil || facet.Query == nil || facet.Query.Query == nil {
			return nil, fmt.Errorf("facet requires a nested query")
		}
		children, err := captureFacets(facet.Query.QueryOptions, active)
		if err != nil {
			return nil, err
		}
		result = append(result, capturedFacet{facet.FacetName, facet.RelationName,
			facet.Query.IntoQuery().Clone(), facet.IncludeAllFacets, children})
	}
	return result, nil
}

func executeCapturedFacets(context stdcontext.Context, service *RuntimeDataService, outer *core.SelectQuery,
	facets []capturedFacet, inherited logprivacy.IntentSource) (map[string]*core.SmartList[core.Record], error) {
	results := make(map[string]*core.SmartList[core.Record])
	for _, facet := range facets {
		membership := outer.ForExactCount("__teaql_facet_count")
		membership.GroupBy = []string{facet.relation}
		intent := inherited
		rows, err := service.fetchAllWithIntent(context, membership, &intent)
		if err != nil {
			return nil, err
		}
		counts := make(map[string]uint64)
		membershipKey := facet.relation
		var descriptor *core.EntityDescriptor
		if service.metadata != nil {
			descriptor = service.metadata.Entity(outer.Entity)
			if descriptor != nil {
				if property := descriptor.PropertyByName(facet.relation); property != nil && property.ColName != "" {
					membershipKey = property.ColName
				}
			}
		}
		var membershipIDs []core.Value
		for _, row := range rows {
			value, ok := row[facet.relation]
			if !ok {
				value, ok = row[membershipKey]
			}
			if ok && value.V != nil {
				count, valid := row["__teaql_facet_count"].TryU64()
				if valid {
					counts[fmt.Sprint(value.V)] = count
					membershipIDs = append(membershipIDs, core.CloneValue(value))
				}
			}
		}

		nested := facet.query.Clone()
		nested.CommentText, nested.PurposeText = outer.CommentText, outer.PurposeText
		nested.TraceChain = core.QueryTraceSource(outer.Entity, outer.TraceChain, *outer.CommentText, *outer.PurposeText)
		if relation := facetRelation(descriptor, facet.relation, nested.Entity); relation != nil {
			nested.TraceChain = append(nested.TraceChain, core.NewTypedTraceNode("relation", relation.Name, outer.Entity+"."+relation.Name))
		}
		countAliases := make([]string, 0)
		for _, aggregate := range nested.Aggregates {
			if aggregate.Function == core.AggCount {
				countAliases = append(countAliases, aggregate.Alias)
			}
		}
		nested.Aggregates = nil
		nested.GroupBy = nil
		if !facet.includeAll {
			nested.AndFilter(core.ExprInList("id", membershipIDs))
		}
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
			if !facet.includeAll && count == 0 {
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
		list := core.NewSmartList(decorated)
		if len(facet.children) > 0 {
			children, err := executeCapturedFacets(context, service, nested, facet.children, intent)
			if err != nil {
				return nil, err
			}
			list.Facets = children
		}
		results[facet.name] = list
	}
	return results, nil
}

// Generated FK storage properties can differ from logical relation names.
// Resolve through metadata, never by stripping a suffix or inventing an edge.
func facetRelation(descriptor *core.EntityDescriptor, field, target string) *core.RelationDescriptor {
	if descriptor == nil {
		return nil
	}
	if relation := descriptor.RelationByName(field); relation != nil && relation.TargetEntity == target {
		return relation
	}
	var matched *core.RelationDescriptor
	for _, relation := range descriptor.Relations {
		if relation.LocKey == field && relation.TargetEntity == target {
			if matched != nil {
				return nil
			} // Ambiguous metadata is not a verified edge.
			matched = relation
		}
	}
	return matched
}

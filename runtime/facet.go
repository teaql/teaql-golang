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
	plan, err := CaptureFacetPlan(options)
	if err != nil {
		return nil, err
	}
	return plan.Execute(context, service, request.Query)
}

// FacetPlan owns an invocation's selections. Capture before policy or provider
// callbacks; never reread the caller's QueryOptions after row execution begins.
type FacetPlan struct {
	facets    []capturedFacet
	relations map[string]*FacetPlan
}

func CaptureFacetPlan(options *core.QueryOptions) (*FacetPlan, error) {
	_, plan, err := CaptureQueryPlan(core.NewSelectQuery(""), options)
	return plan, err
}

// CaptureQueryPlan owns all relation and Facet work before any callback. Keeping
// this in runtime avoids generating a planner for each domain entity.
func CaptureQueryPlan(query *core.SelectQuery, options *core.QueryOptions) (*core.SelectQuery, *FacetPlan, error) {
	return captureQueryPlan(query, options, make(map[*core.SelectQuery]bool), make(map[*core.QueryOptions]bool))
}

func captureQueryPlan(query *core.SelectQuery, options *core.QueryOptions, queries map[*core.SelectQuery]bool,
	optionsStack map[*core.QueryOptions]bool) (*core.SelectQuery, *FacetPlan, error) {
	if query == nil {
		return nil, nil, fmt.Errorf("selection requires a query")
	}
	if queries[query] {
		return nil, nil, fmt.Errorf("cyclic relation selection")
	}
	if options != nil && optionsStack[options] {
		return nil, nil, fmt.Errorf("cyclic facet selection")
	}
	queries[query] = true
	defer delete(queries, query)
	if options != nil {
		optionsStack[options] = true
		defer delete(optionsStack, options)
	}
	// Relations are captured recursively with cycle checks, not through Clone.
	shallow := *query
	shallow.Relations = nil
	captured := shallow.Clone()
	plan := &FacetPlan{relations: make(map[string]*FacetPlan)}
	for _, load := range query.Relations {
		if load == nil {
			return nil, nil, fmt.Errorf("nil relation selection")
		}
		if load.Query == nil && load.Selection == nil {
			captured.Relations = append(captured.Relations, core.NewRelationLoad(load.Name))
			continue
		}
		var childQuery *core.SelectQuery
		var childPlan *FacetPlan
		var err error
		if load.Selection != nil {
			childQuery, childPlan, err = captureSelection(load.Selection, queries, optionsStack)
		} else {
			childQuery, childPlan, err = captureQueryPlan(load.Query, nil, queries, optionsStack)
		}
		if err != nil {
			return nil, nil, err
		}
		captured.RelationQuery(load.Name, childQuery)
		plan.relations[load.Name] = childPlan
	}
	if options != nil {
		for _, facet := range options.Facets {
			if facet == nil || facet.Query == nil || facet.Query.Query == nil {
				return nil, nil, fmt.Errorf("facet requires a nested query")
			}
			child, childPlan, err := captureSelection(facet.Query, queries, optionsStack)
			if err != nil {
				return nil, nil, err
			}
			plan.facets = append(plan.facets, capturedFacet{name: facet.FacetName, relation: facet.RelationName,
				query: child, includeAll: facet.IncludeAllFacets, children: childPlan.facets, plan: childPlan})
		}
	}
	return captured, plan, nil
}

func captureSelection(selection *core.QuerySelection, queries map[*core.SelectQuery]bool,
	optionsStack map[*core.QueryOptions]bool) (*core.SelectQuery, *FacetPlan, error) {
	if selection == nil || selection.Query == nil {
		return nil, nil, fmt.Errorf("selection requires a query")
	}
	if queries[selection.Query] {
		return nil, nil, fmt.Errorf("cyclic relation selection")
	}
	normalized := selectionQuery(selection)
	if normalized != selection.Query {
		queries[selection.Query] = true
		defer delete(queries, selection.Query)
	}
	return captureQueryPlan(normalized, selection.QueryOptions, queries, optionsStack)
}

// Preserve QuerySelection's non-Facet metadata as well as generated selections.
// The shallow shell avoids recursively cloning a cycle before capture rejects it.
func selectionQuery(selection *core.QuerySelection) *core.SelectQuery {
	if selection == nil || selection.Query == nil {
		return nil
	}
	if len(selection.RelationSelections) == 0 && len(selection.ChildEnhancements) == 0 &&
		(selection.QueryOptions == nil || (len(selection.QueryOptions.DynamicProperties) == 0 &&
			len(selection.QueryOptions.RawProjections) == 0 && len(selection.QueryOptions.RelationAggregates) == 0 &&
			len(selection.QueryOptions.ObjectGroupBys) == 0 && len(selection.QueryOptions.RawSqlSearchCriteria) == 0 &&
			selection.QueryOptions.RawSql == nil && selection.QueryOptions.Comment == nil)) {
		return selection.Query
	}
	shell := *selection.Query
	shell.Relations = append([]*core.RelationLoad(nil), selection.Query.Relations...)
	shell.DynamicProperties = append([]*core.RawSqlProjection(nil), selection.Query.DynamicProperties...)
	shell.RawProjections = append([]*core.RawSqlProjection(nil), selection.Query.RawProjections...)
	shell.RawSqlSearchCriteria = append([]string(nil), selection.Query.RawSqlSearchCriteria...)
	shell.ObjectGroupBys = append([]*core.ObjectGroupBy(nil), selection.Query.ObjectGroupBys...)
	shell.ChildEnhancements = append([]*core.SelectQuery(nil), selection.Query.ChildEnhancements...)
	for _, relation := range selection.RelationSelections {
		if relation == nil {
			continue
		}
		shell.RelationQuerySelection(relation.Name, &core.QuerySelection{Query: relation.Query,
			QueryOptions: relation.QueryOptions, RelationSelections: relation.RelationSelections,
			ChildEnhancements: relation.ChildEnhancements})
	}
	if selection.QueryOptions != nil {
		core.ApplyRuntimeMetadata(&shell, selection.QueryOptions, selection.ChildEnhancements)
	}
	return &shell
}

func (p *FacetPlan) HasFacets() bool { return p != nil && len(p.facets) > 0 }
func (*FacetPlan) String() string    { return "<captured facet plan>" }
func (*FacetPlan) GoString() string  { return "<captured facet plan>" }

// WithQueryDiagnostics includes all future Facet bindings in the first SQL's
// privacy scope without scheduling those queries or storing anything in Context.
func (p *FacetPlan) WithQueryDiagnostics(query *core.SelectQuery) *core.SelectQuery {
	var queries []*core.SelectQuery
	var visit func(*FacetPlan)
	visit = func(plan *FacetPlan) {
		if plan == nil {
			return
		}
		for _, facet := range plan.facets {
			queries = append(queries, facet.query)
			visit(facet.plan)
		}
		for _, child := range plan.relations {
			visit(child)
		}
	}
	if p != nil {
		visit(p)
	}
	return query.WithDiagnosticQueries(queries...)
}

func (p *FacetPlan) Execute(context stdcontext.Context, service *RuntimeDataService,
	outer *core.SelectQuery) (map[string]*core.SmartList[core.Record], error) {
	request, err := data_service.NewQueryRequest(outer)
	if err != nil {
		return nil, err
	}
	if !p.HasFacets() {
		return map[string]*core.SmartList[core.Record]{}, nil
	}
	return executeCapturedFacets(context, service, p.WithQueryDiagnostics(request.Query), p.facets, logprivacy.IntentSource{})
}

type capturedFacet struct {
	name, relation string
	query          *core.SelectQuery
	includeAll     bool
	children       []capturedFacet
	plan           *FacetPlan
}

func captureFacets(options *core.QueryOptions, active map[*core.QueryOptions]bool) ([]capturedFacet, error) {
	_, plan, err := captureQueryPlan(core.NewSelectQuery(""), options, make(map[*core.SelectQuery]bool), active)
	if err != nil {
		return nil, err
	}
	return plan.facets, nil
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
		if !facet.includeAll && len(membershipIDs) == 0 {
			results[facet.name] = core.NewSmartList(make([]core.Record, 0))
			continue
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
		facetRows, err := service.fetchAllWithIntent(context, nested, &intent, facet.plan)
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

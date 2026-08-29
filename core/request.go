package core

const (
	CountAlias     = "count"
	TypeField      = "internal_type"
	TypeGroupField = "type_group"
)

type FieldOperator int

const (
	FieldOperatorEqual FieldOperator = iota
	FieldOperatorNotEqual
	FieldOperatorGreaterThan
	FieldOperatorGreaterThanOrEqual
	FieldOperatorLessThan
	FieldOperatorLessThanOrEqual
	FieldOperatorBetween
	FieldOperatorIn
	FieldOperatorNotIn
	FieldOperatorContain
	FieldOperatorNotContain
	FieldOperatorBeginWith
	FieldOperatorNotBeginWith
	FieldOperatorEndWith
	FieldOperatorNotEndWith
	FieldOperatorSoundsLike
	FieldOperatorIsNull
	FieldOperatorIsNotNull
)

type DateRange[T any] struct {
	Start T
	End   T
}

func NewDateRange[T any](start, end T) *DateRange[T] {
	return &DateRange[T]{Start: start, End: end}
}

type QuerySelection struct {
	Query              *SelectQuery
	RelationSelections []*RelationSelection
	RelationFilters    []*RelationFilter
	ChildEnhancements  []*QuerySelection
	QueryOptions       *QueryOptions
}

func NewQuerySelection(query *SelectQuery) *QuerySelection {
	return &QuerySelection{
		Query:              query,
		RelationSelections: make([]*RelationSelection, 0),
		RelationFilters:    make([]*RelationFilter, 0),
		ChildEnhancements:  make([]*QuerySelection, 0),
		QueryOptions:       NewQueryOptions(),
	}
}

func (q *QuerySelection) IntoQuery() *SelectQuery {
	query := ApplyRelationSelections(q.Query, q.RelationSelections)
	return ApplyRuntimeMetadata(query, q.QueryOptions, q.ChildEnhancements)
}

type RelationSelection struct {
	Name               string
	Query              *SelectQuery
	RelationSelections []*RelationSelection
	RelationFilters    []*RelationFilter
	ChildEnhancements  []*QuerySelection
	QueryOptions       *QueryOptions
}

func NewRelationSelection(name string, selection *QuerySelection) *RelationSelection {
	return &RelationSelection{
		Name:               name,
		Query:              selection.Query,
		RelationSelections: selection.RelationSelections,
		RelationFilters:    selection.RelationFilters,
		ChildEnhancements:  selection.ChildEnhancements,
		QueryOptions:       selection.QueryOptions,
	}
}

func (r *RelationSelection) IntoQuery() *SelectQuery {
	query := ApplyRelationSelections(r.Query, r.RelationSelections)
	return ApplyRuntimeMetadata(query, r.QueryOptions, r.ChildEnhancements)
}

type RelationFilter struct {
	Name               string
	Query              *SelectQuery
	RelationSelections []*RelationSelection
	RelationFilters    []*RelationFilter
	ChildEnhancements  []*QuerySelection
	QueryOptions       *QueryOptions
}

func NewRelationFilter(name string, selection *QuerySelection) *RelationFilter {
	return &RelationFilter{
		Name:               name,
		Query:              selection.Query,
		RelationSelections: selection.RelationSelections,
		RelationFilters:    selection.RelationFilters,
		ChildEnhancements:  selection.ChildEnhancements,
		QueryOptions:       selection.QueryOptions,
	}
}

type QueryOptions struct {
	Comment              *string
	RawSql               *string
	RawSqlSearchCriteria []string
	DynamicProperties    []*RawDynamicProperty
	RawProjections       []*RawProjection
	RelationAggregates   []*RelationAggregateBuilder
	ObjectGroupBys       []*ObjectGroupByBuilder
	Facets               []*FacetRequest
}

func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		RawSqlSearchCriteria: make([]string, 0),
		DynamicProperties:    make([]*RawDynamicProperty, 0),
		RawProjections:       make([]*RawProjection, 0),
		RelationAggregates:   make([]*RelationAggregateBuilder, 0),
		ObjectGroupBys:       make([]*ObjectGroupByBuilder, 0),
		Facets:               make([]*FacetRequest, 0),
	}
}

type UnsafeRawSqlSegment struct {
	Sql string
}

func TrustedRawSql(sql string) *UnsafeRawSqlSegment {
	return &UnsafeRawSqlSegment{Sql: sql}
}

type RawDynamicProperty struct {
	PropertyName  string
	RawSqlSegment string
}

func NewRawDynamicProperty(propertyName string, sqlSegment *UnsafeRawSqlSegment) *RawDynamicProperty {
	return &RawDynamicProperty{PropertyName: propertyName, RawSqlSegment: sqlSegment.Sql}
}

type RawProjection struct {
	PropertyName  string
	RawSqlSegment string
}

func NewRawProjection(propertyName string, sqlSegment *UnsafeRawSqlSegment) *RawProjection {
	return &RawProjection{PropertyName: propertyName, RawSqlSegment: sqlSegment.Sql}
}

type RelationAggregateBuilder struct {
	RelationName string
	Alias        string
	Query        *QuerySelection
	SingleResult bool
}

func NewRelationAggregateBuilder(relationName, alias string, query *QuerySelection, singleResult bool) *RelationAggregateBuilder {
	return &RelationAggregateBuilder{
		RelationName: relationName,
		Alias:        alias,
		Query:        query,
		SingleResult: singleResult,
	}
}

type FacetRequest struct {
	FacetName        string
	RelationName     string
	Query            *QuerySelection
	IncludeAllFacets bool
}

func NewFacetRequest(facetName, relationName string, query *QuerySelection, includeAllFacets bool) *FacetRequest {
	return &FacetRequest{
		FacetName:        facetName,
		RelationName:     relationName,
		Query:            query,
		IncludeAllFacets: includeAllFacets,
	}
}

type ObjectGroupByBuilder struct {
	PropertyName string
	StorageField string
	Query        *QuerySelection
}

func NewObjectGroupByBuilder(propertyName, storageField string, query *QuerySelection) *ObjectGroupByBuilder {
	return &ObjectGroupByBuilder{
		PropertyName: propertyName,
		StorageField: storageField,
		Query:        query,
	}
}

func ApplyRelationSelections(query *SelectQuery, selections []*RelationSelection) *SelectQuery {
	for _, selection := range selections {
		query.RelationQuery(selection.Name, selection.IntoQuery())
	}
	return query
}

func ApplyRuntimeMetadata(query *SelectQuery, options *QueryOptions, childEnhancements []*QuerySelection) *SelectQuery {
	if options.Comment != nil {
		query.Comment(*options.Comment)
	}
	query.RawSql = options.RawSql
	query.RawSqlSearchCriteria = append(query.RawSqlSearchCriteria, options.RawSqlSearchCriteria...)

	for _, p := range options.DynamicProperties {
		query.DynamicProperties = append(query.DynamicProperties, NewRawSqlProjection(p.PropertyName, p.RawSqlSegment))
	}

	for _, p := range options.RawProjections {
		query.RawProjections = append(query.RawProjections, NewRawSqlProjection(p.PropertyName, p.RawSqlSegment))
	}

	for _, g := range options.ObjectGroupBys {
		query.ObjectGroupBys = append(query.ObjectGroupBys, NewObjectGroupBy(g.PropertyName, g.StorageField, g.Query.IntoQuery()))
	}
	query.RelationAggregates = RuntimeRelationAggregates(options)

	for _, c := range childEnhancements {
		query.ChildEnhancement(c.IntoQuery())
	}
	return query
}

func RuntimeRelationAggregates(options *QueryOptions) []*RelationAggregate {
	var results []*RelationAggregate
	for _, agg := range options.RelationAggregates {
		results = append(results, NewRelationAggregate(agg.RelationName, agg.Alias, agg.Query.IntoQuery(), agg.SingleResult))
	}
	return results
}

func MergeOuterFilterIntoFacetAggregates(selection *QuerySelection, outerQuery *SelectQuery) {
	if outerQuery.Filter == nil {
		return
	}
	for _, agg := range selection.QueryOptions.RelationAggregates {
		if agg.Query.Query.Entity == outerQuery.Entity {
			agg.Query.Query.AndFilter(outerQuery.Filter)
		}
	}
}

func AttachFacets[T any](rows *SmartList[T], facets map[string]*SmartList[Record]) {
	for name, facet := range facets {
		rows.AddFacet(name, facet)
	}
}

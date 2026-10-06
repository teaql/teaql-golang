package runtime

import "github.com/teaql/teaql-golang/core"

// PrepareEntityQuery authorizes a caller-owned request exactly once and then
// protects the structural fields needed by typed hydration. Record/aggregate
// execution uses PrepareQuery instead; it must not acquire entity projections.
func (c *UserContext) PrepareEntityQuery(query *core.SelectQuery) (*core.SelectQuery, error) {
	prepared, err := c.PrepareQuery(query)
	if err != nil {
		return nil, err
	}
	if c.Metadata == nil || c.Metadata.Entity(prepared.Entity) == nil {
		return nil, &RuntimeError{Type: "MissingEntity", MissingEntityName: prepared.Entity}
	}
	protectEntityProjection(prepared, c.Metadata.Entity(prepared.Entity))
	return prepared, nil
}

func protectEntityProjection(query *core.SelectQuery, descriptor *core.EntityDescriptor) {
	if query.RawSql != nil || len(query.Aggregates) > 0 || len(query.GroupBy) > 0 || selectsAllProperties(query) {
		return
	}
	for _, property := range []*core.PropertyDescriptor{descriptor.IdProperty(), descriptor.VersionProperty()} {
		if property != nil {
			EnsureRelationProjection(query, property.Name)
		}
	}
	for _, load := range query.Relations {
		if relation := descriptor.RelationByName(load.Name); relation != nil {
			EnsureRelationProjection(query, relation.LocKey)
		}
	}
}

// Record queries need only the join keys for requested forward details, not
// typed-entity ID/version protection. Keep raw and aggregate row shapes intact.
func protectForwardRelationProjection(query *core.SelectQuery, descriptor *core.EntityDescriptor) {
	if query.RawSql != nil || len(query.Aggregates) > 0 || len(query.GroupBy) > 0 || selectsAllProperties(query) {
		return
	}
	for _, load := range query.Relations {
		if relation := descriptor.RelationByName(load.Name); relation != nil && !relation.IsMany {
			EnsureRelationProjection(query, relation.LocKey)
		}
	}
}

func selectsAllProperties(query *core.SelectQuery) bool {
	return len(query.Projection) == 0 && len(query.ExprProjection) == 0 &&
		len(query.RawProjections) == 0 && len(query.DynamicProperties) == 0
}

// EnsureRelationProjection retains a relation join key in an explicit narrow
// projection. A default select-all already contains it and must stay select-all.
// This is generated/runtime plumbing, not a way to bypass query authorization.
func EnsureRelationProjection(query *core.SelectQuery, field string) {
	if selectsAllProperties(query) || containsField(query.Projection, field) {
		return
	}
	query.Projection = append(query.Projection, field)
}

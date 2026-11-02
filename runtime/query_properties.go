package runtime

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
)

// QueryPropertyEntity is the shallow generated view already used by JSON. No
// Context, provider, entity ID or mutation ledger is involved in this operation.
type QueryPropertyEntity interface {
	Base() *core.BaseEntityData
	LoadedJSONFields() core.Record
}

// CaptureEntityQueryProperties follows only materialized, explicitly selected
// relations. It cannot issue a query or turn an unselected relation into a value.
// The authorized captured query, not the caller's mutable builder, owns aliases.
func CaptureEntityQueryProperties(entity QueryPropertyEntity, query *core.SelectQuery) error {
	type captured struct {
		base       *core.BaseEntityData
		properties core.Record
	}
	var pending []captured
	var capture func(QueryPropertyEntity, *core.SelectQuery, int) error
	capture = func(current QueryPropertyEntity, selected *core.SelectQuery, depth int) error {
		if selected == nil {
			return nil
		}
		if depth >= 64 {
			return fmt.Errorf("query property capture exceeds the supported graph depth")
		}
		properties := core.Record{}
		captureAlias := func(alias string) {
			if value, present := current.Base().GetDynamic(alias); present {
				properties[alias] = value
			}
		}
		for _, projection := range selected.ExprProjection {
			if projection != nil {
				captureAlias(projection.Alias)
			}
		}
		for _, projection := range selected.RawProjections {
			if projection != nil {
				captureAlias(projection.PropertyName)
			}
		}
		for _, projection := range selected.DynamicProperties {
			if projection != nil {
				captureAlias(projection.PropertyName)
			}
		}
		for _, aggregate := range selected.Aggregates {
			if aggregate != nil {
				captureAlias(aggregate.Alias)
			}
		}
		for _, aggregate := range selected.RelationAggregates {
			if aggregate == nil {
				continue
			}
			captureAlias(aggregate.Alias)
		}
		pending = append(pending, captured{current.Base(), properties})
		if len(selected.Relations) == 0 {
			return nil
		}
		loaded := current.LoadedJSONFields()
		for _, relation := range selected.Relations {
			if relation == nil || relation.Query == nil {
				continue
			}
			value, present := loaded[relation.Name]
			if !present || value.IsNull() {
				continue
			}
			visit := func(value any) error {
				if child, ok := value.(QueryPropertyEntity); ok {
					return capture(child, relation.Query, depth+1)
				}
				return nil
			}
			switch detail := value.V.(type) {
			case []core.Value:
				for _, child := range detail {
					if err := visit(child.V); err != nil {
						return err
					}
				}
			default:
				if err := visit(detail); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := capture(entity, query, 0); err != nil {
		return err
	}
	for _, item := range pending {
		item.base.CaptureQueryProperties(item.properties)
	}
	return nil
}

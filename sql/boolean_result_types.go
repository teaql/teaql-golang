package sql

import "github.com/teaql/teaql-golang/core"

func booleanResultAliases(entity *core.EntityDescriptor, query *core.SelectQuery) map[string]bool {
	if query.RawSql != nil {
		return nil
	}
	aliases := make(map[string]bool)
	add := func(field, alias string) {
		if property := entity.PropertyByName(field); property != nil && property.DataType == core.TypeBool {
			aliases[alias] = true
		}
	}
	if len(query.Aggregates)+len(query.GroupBy)+len(query.Projection)+len(query.ExprProjection)+len(query.RawProjections)+len(query.DynamicProperties) == 0 {
		for _, property := range entity.Properties {
			add(property.Name, property.Name)
		}
	}
	for _, field := range query.Projection {
		add(field, field)
	}
	for _, field := range query.GroupBy {
		if property := entity.PropertyByName(field); property != nil {
			add(field, property.ColName)
		}
	}
	for _, aggregate := range query.Aggregates {
		delete(aliases, aggregate.Alias)
		if aggregate.Function == core.AggMin || aggregate.Function == core.AggMax {
			add(aggregate.Field, aggregate.Alias)
		}
	}
	for _, projection := range query.ExprProjection {
		if projection == nil {
			continue
		}
		delete(aliases, projection.Alias)
		expr := projection.Expr
		if expr == nil || expr.Type != core.ExprTypeFunctionCall || len(expr.Args) != 1 || expr.Args[0] == nil || expr.Args[0].Type != core.ExprTypeColumn {
			continue
		}
		if expr.Function == core.FuncMin || expr.Function == core.FuncMax {
			add(expr.Args[0].Column, projection.Alias)
		}
	}
	for _, projection := range query.RawProjections {
		delete(aliases, projection.PropertyName)
	}
	for _, projection := range query.DynamicProperties {
		delete(aliases, projection.PropertyName)
	}
	if len(aliases) == 0 {
		return nil
	}
	return aliases
}

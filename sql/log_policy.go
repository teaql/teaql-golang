package sql

import (
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

func (d *DefaultSqlDialect) forCompilation() *DefaultSqlDialect {
	return &DefaultSqlDialect{Dialect: d.Dialect, logPolicies: make(map[int]string), generatedSQL: true}
}

func fieldLogPolicy(entity *core.EntityDescriptor, field string) string {
	property := entity.PropertyByName(field)
	if logprivacy.CredentialName(field) || property != nil && logprivacy.CredentialName(property.ColName) {
		return "credential"
	}
	for _, masked := range entity.AuditMaskFlds {
		if masked == field {
			return "masked"
		}
	}
	if property != nil {
		switch property.LogPolicy {
		case "plain", "masked", "credential":
			return property.LogPolicy
		}
	}
	return "unknown"
}

func (d *DefaultSqlDialect) bindField(params *[]core.Value, value core.Value, entity *core.EntityDescriptor, field string) {
	if d.logPolicies != nil {
		d.logPolicies[len(*params)] = fieldLogPolicy(entity, field)
	}
	*params = append(*params, value)
}

func (d *DefaultSqlDialect) parameterPolicies(params []core.Value) []string {
	result := make([]string, len(params))
	for i := range result {
		result[i] = d.logPolicies[i]
		if result[i] == "" {
			result[i] = "unknown"
		}
	}
	return result
}

// Apply a comparison's field policy to its derived binds, preserving more
// specific nested policies (including independently compiled subqueries).
func (d *DefaultSqlDialect) expressionPolicy(entity *core.EntityDescriptor, expr *core.Expr) string {
	if expr == nil {
		return ""
	}
	if expr.Type == core.ExprTypeColumn {
		return fieldLogPolicy(entity, expr.Column)
	}
	var children []*core.Expr
	switch expr.Type {
	case core.ExprTypeBinary:
		children = []*core.Expr{expr.Left, expr.Right}
	case core.ExprTypeBetween:
		children = []*core.Expr{expr.Left, expr.Lower, expr.Upper}
	case core.ExprTypeFunctionCall:
		children = expr.Args
	default:
		return ""
	}
	policies := map[string]bool{}
	for _, child := range children {
		policies[d.expressionPolicy(entity, child)] = true
	}
	for _, policy := range []string{"credential", "unknown", "masked", "plain"} {
		if policies[policy] {
			return policy
		}
	}
	return ""
}

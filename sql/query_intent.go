package sql

import (
	"fmt"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

// Classify future/removed child bindings before the first physical statement.
// No SQL executes and no provenance is installed on the reusable executor or
// Context. Retained bindings affect only safe diagnostic projection.
func (e *SqlDataServiceExecutor) queryIntentSource(request *ds.QueryRequest, compiled *CompiledQuery) (logprivacy.IntentSource, error) {
	sources := []ds.ExecutionMetadata{{InheritedIntent: request.InheritedIntent}}
	stack := []*core.SelectQuery{request.Query}
	seen := make(map[*core.SelectQuery]bool)
	for len(stack) > 0 {
		query := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if query == nil || seen[query] {
			continue
		}
		seen[query] = true
		bindings := compiled
		if query != request.Query {
			// Work on a copy: descriptor resolution must not modify provenance
			// shared by two count requests derived from the same source.
			captured := query.Clone()
			entity := e.SchemaProvider.GetEntity(captured.Entity)
			if entity == nil {
				return logprivacy.IntentSource{}, fmt.Errorf("unknown diagnostic query entity %s", captured.Entity)
			}
			if err := e.resolveSubqueryDescriptors(captured); err != nil {
				return logprivacy.IntentSource{}, err
			}
			var err error
			bindings, err = (&DefaultSqlDialect{Dialect: e.Dialect}).CompileSelect(entity, captured)
			if err != nil {
				return logprivacy.IntentSource{}, err
			}
		}
		source := ds.ExecutionMetadata{GeneratedSQL: bindings.GeneratedSQL,
			ParameterizedSQL: bindings.Sql, ParameterLogPolicies: append([]string(nil), bindings.ParameterLogPolicies...)}
		for _, value := range bindings.Params {
			source.Parameters = append(source.Parameters, core.CloneValue(value))
		}
		for _, operand := range bindings.intentOperands {
			source.Parameters = append(source.Parameters, core.CloneValue(operand.value))
			source.ParameterLogPolicies = append(source.ParameterLogPolicies, operand.policy)
		}
		sources = append(sources, source)
		if origin, ok := logprivacy.ReadIntentSource(query.DiagnosticOrigin()).(*core.SelectQuery); ok {
			stack = append(stack, origin)
		}
		for _, child := range query.Relations {
			if child != nil {
				stack = append(stack, child.Query)
			}
		}
		for _, child := range query.RelationAggregates {
			if child != nil {
				stack = append(stack, child.Query)
			}
		}
		for _, child := range query.ObjectGroupBys {
			if child != nil {
				stack = append(stack, child.Query)
			}
		}
		stack = append(stack, query.ChildEnhancements...)
	}
	return logprivacy.NewIntentSource(sources), nil
}

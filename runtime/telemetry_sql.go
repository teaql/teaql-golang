package runtime

import (
	"context"
	"encoding/json"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"strings"
)

type RuntimeDiagnosticEvent struct {
	Correlation ds.SQLExecutionCorrelation
	Attributes  map[string]RuntimeAttributeValue
}
type SQLRuntimeTelemetry interface {
	RecordsSQLDiagnostics() bool
	RecordDiagnostic(RuntimeDiagnosticEvent)
}
type correlatedTelemetryScope interface {
	Correlation() *ds.SQLExecutionCorrelation
}
type sqlStatementScope struct {
	scope    RuntimeTelemetryScope
	identity *ds.SQLExecutionCorrelation
}

func (s *sqlStatementScope) Correlation() *ds.SQLExecutionCorrelation { return s.identity }
func (s *sqlStatementScope) Success()                                 { s.scope.Success(nil) }
func (s *sqlStatementScope) Failure(kind string)                      { s.scope.Failure(kind) }
func (s *sqlStatementScope) CompleteSQLStream(outcome string, count int, kind string) {
	CompleteRuntimeOperation(s.scope, outcome, map[string]RuntimeAttributeValue{"teaql.result.cardinality": count}, kind)
}
func (c *UserContext) StartMutationOperation(ctx context.Context) (context.Context, ds.SQLStatementObservation) {
	child, scope := StartRuntimeOperation(ctx, c.RuntimeTelemetry(), NewRuntimeOperation("mutation", "data-service.mutate", nil))
	return child, &sqlStatementScope{scope: scope}
}
func (c *UserContext) StartSQLStatement(ctx context.Context, backend string, operation ds.DataServiceOperation) (context.Context, ds.SQLStatementObservation) {
	telemetry, ok := c.RuntimeTelemetry().(SQLRuntimeTelemetry)
	if !ok || !telemetry.RecordsSQLDiagnostics() {
		return ctx, nil
	}
	child, scope := StartRuntimeOperation(ctx, c.RuntimeTelemetry(), NewRuntimeOperation("provider", backend+".statement", map[string]RuntimeAttributeValue{
		"teaql.provider.kind": backend, "teaql.provider.operation": strings.ToLower(string(operation)), "teaql.provider.execution": "statement",
	}))
	var identity *ds.SQLExecutionCorrelation
	if correlated, ok := scope.(correlatedTelemetryScope); ok {
		identity = correlated.Correlation()
	}
	return child, &sqlStatementScope{scope: scope, identity: identity}
}
func sqlDiagnosticEvent(safe ds.ExecutionMetadata, identity ds.SQLExecutionCorrelation) RuntimeDiagnosticEvent {
	text := func(value *string) string {
		if value == nil {
			return ""
		}
		return *value
	}
	nodes := func(source []*core.TraceNode) string {
		path := make([]map[string]string, 0, len(source))
		for _, node := range source {
			if node != nil {
				path = append(path, map[string]string{"kind": node.Kind, "name": node.Name, "detail": node.Comment})
			}
		}
		data, _ := json.Marshal(path)
		return string(data)
	}
	operation := strings.ToLower(string(safe.Operation))
	if safe.Operation == ds.OpQuery {
		operation = "select"
	}
	comment := safe.Comment
	if comment == nil {
		comment = safe.AuditReason
	}
	mode := "masked"
	if safe.OmissionReason != "" {
		mode = "omitted"
	}
	attributes := map[string]RuntimeAttributeValue{
		"teaql.sql.debug": text(safe.DebugQuery), "teaql.sql.log_mode": mode, "teaql.sql.operation": operation, "teaql.sql.outcome": safe.ExecutionOutcome,
		"teaql.sql.elapsed_us": safe.EndedAt.Sub(safe.StartedAt).Microseconds(), "teaql.intent.comment": text(comment), "teaql.intent.purpose": text(safe.Purpose),
		"teaql.trace_chain": nodes(safe.RequestTraceChain), "teaql.trace_path": nodes(safe.TraceChain), "teaql.mutation_lineage": nodes(safe.MutationLineage),
	}
	if safe.ResultCount != nil {
		attributes["teaql.sql.result_count"] = *safe.ResultCount
	}
	if safe.AffectedRows != nil {
		attributes["teaql.sql.affected_rows"] = *safe.AffectedRows
	}
	return RuntimeDiagnosticEvent{Correlation: identity, Attributes: attributes}
}

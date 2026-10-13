package runtime

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

type sqlDiagnosticsProbe struct {
	NoopRuntimeTelemetry
	events          []RuntimeDiagnosticEvent
	panicCapability bool
}

func (p *sqlDiagnosticsProbe) RecordsSQLDiagnostics() bool {
	if p.panicCapability {
		panic("PRIVATE-CAPABILITY")
	}
	return true
}
func (p *sqlDiagnosticsProbe) RecordDiagnostic(event RuntimeDiagnosticEvent) {
	p.events = append(p.events, event)
}
func TestSQLTelemetryRemainsMaskedWithOrdinaryLogsDisabled(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	probe := &sqlDiagnosticsProbe{}
	ordinary := NewSQLExecutionEvidenceStore()
	context := NewUserContext().WithRuntimeTelemetry(probe).WithDiagnosticSQLLogSink(ordinary)
	context.DisableSqlLog()
	comment := "load PRIVATE-VALUE"
	id := uint64(987654321)
	node := core.NewTraceNode("Platform", &id, comment)
	correlation := &ds.SQLExecutionCorrelation{TraceID: "11111111111111111111111111111111", SpanID: "2222222222222222"}
	metadata := ds.ExecutionMetadata{Operation: ds.OpQuery, SQLCorrelation: correlation,
		ParameterizedSQL: "SELECT name FROM platform_data WHERE name = ?", Parameters: []core.Value{core.ValText("PRIVATE-VALUE")},
		ParameterLogPolicies: []string{"masked"}, GeneratedSQL: true, Comment: &comment, Purpose: &comment,
		RequestTraceChain: []*core.TraceNode{node}, TraceChain: []*core.TraceNode{node}, MutationLineage: []*core.TraceNode{node},
		StartedAt: time.Now(), EndedAt: time.Now(), ExecutionOutcome: "success"}
	context.RecordExecutionMetadata(metadata)
	if len(probe.events) != 1 || len(ordinary.Snapshot()) != 0 {
		t.Fatal("telemetry was coupled to ordinary SQL logging")
	}
	encoded, err := json.Marshal(probe.events[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "PRIVATE-VALUE") || strings.Contains(string(encoded), "987654321") || strings.Contains(string(encoded), "EntityId") {
		t.Fatal("telemetry leaked raw data", string(encoded))
	}
	if probe.events[0].Correlation != *correlation || probe.events[0].Attributes["teaql.sql.log_mode"] != "masked" {
		t.Fatal("carried identity or masking mode missing")
	}
	if metadata.Parameters[0].V != "PRIVATE-VALUE" || node.Comment != comment || node.EntityId != &id {
		t.Fatal("diagnostics mutated business inputs")
	}
}
func TestDiagnosticCapabilityPanicCannotEscapeBusinessExecution(t *testing.T) {
	context := NewUserContext().WithRuntimeTelemetry(&sqlDiagnosticsProbe{panicCapability: true})
	context.RecordExecutionMetadata(ds.ExecutionMetadata{})
}

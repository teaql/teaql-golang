package runtime

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlaintextLogAcknowledgementAndFileOutput(t *testing.T) {
	for _, ack := range []string{"", "true", "1", plaintextLogAck + " ", plaintextLogAck} {
		t.Run(ack, func(t *testing.T) {
			t.Setenv(plaintextLogEnv, ack)
			sql := "UPDATE customer SET name = 'PRIVATE-CUSTOMER'"
			metadata := data_service.ExecutionMetadata{ParameterizedSQL: "UPDATE customer SET name = ?", Parameters: []core.Value{core.ValText("PRIVATE-CUSTOMER")}, DebugQuery: &sql}
			path := filepath.Join(t.TempDir(), "runtime.log")
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			NewSensitiveDiagnosticSQLLogSink(file).WriteSQLLog(metadata)
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			output, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(output), "PRIVATE-CUSTOMER") != (ack == plaintextLogAck) {
				t.Fatal("incorrect plaintext gate")
			}
			if metadata.Parameters[0].V != "PRIVATE-CUSTOMER" {
				t.Fatal("execution input mutated")
			}
		})
	}
}

func TestAuditNestedCredentialAndOldValueTraceAreRedacted(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	oldValue := core.ValText("OLD-TOKEN-CANARY")
	newValue := core.Value{V: map[string]any{"access_token": "NEW-TOKEN-CANARY"}}
	event := &RawAuditEvent{Entity: "Customer", Changes: []*EntityPropertyChange{
		{Field: "settings", OldValue: &oldValue, NewValue: &newValue},
	}, TraceChain: []*core.TraceNode{core.NewTraceNode("Customer", nil, "OLD-TOKEN-CANARY NEW-TOKEN-CANARY")}}
	safe := event.BuildSafeEvent(nil, nil)
	if !safe.Fields[0].Masked || *safe.Fields[0].Value != "[REDACTED]" {
		t.Fatal("nested credential exposed")
	}
	if strings.Contains(safe.TraceChain[0].Comment, "TOKEN-CANARY") {
		t.Fatal("audit trace exposed old/new secret")
	}
	if !strings.Contains(event.TraceChain[0].Comment, "TOKEN-CANARY") {
		t.Fatal("business event was mutated")
	}
}

type privacySink struct {
	events []data_service.ExecutionMetadata
}

func (s *privacySink) WriteSQLLog(metadata data_service.ExecutionMetadata) {
	s.events = append(s.events, metadata)
}

func TestSensitiveCustomSinkCannotBypassGate(t *testing.T) {
	for _, ack := range []string{"", plaintextLogAck} {
		t.Setenv(plaintextLogEnv, ack)
		sink := &privacySink{}
		context := NewUserContext().WithDiagnosticSQLLogSink(nil).WithSensitiveDiagnosticSQLLogSink(sink)
		query := "UPDATE customer SET access_token = 'TOKEN-CANARY'"
		comment := "change TOKEN-CANARY"
		context.RecordExecutionMetadata(data_service.ExecutionMetadata{
			ParameterizedSQL: "UPDATE customer SET access_token = ?", Parameters: []core.Value{core.ValText("TOKEN-CANARY")}, DebugQuery: &query, Comment: &comment,
		})
		if len(sink.events) != 1 || strings.Contains(fmt.Sprint(sink.events), "TOKEN-CANARY") {
			t.Fatal("credential exposed to custom sink")
		}
		safe := BuildSafeAuditField("privateKey", &comment, nil, nil)
		if !safe.Masked || *safe.Value != "[REDACTED]" {
			t.Fatal("credential audit value exposed")
		}
	}
}

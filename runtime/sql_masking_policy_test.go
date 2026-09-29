package runtime

import (
	"bytes"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"reflect"
	"strings"
	"testing"
)

func TestSQLMaskPoliciesMixedAndRepeatedBindings(t *testing.T) {
	raw := data_service.ExecutionMetadata{Backend: "postgresql", ParameterizedSQL: "SELECT $1, $2, $1",
		Parameters: []core.Value{core.ValText("O'Reilly"), core.ValBool(true)}, ParameterLogPolicies: []string{"masked", "plain"}}
	safe := projectedSQLMetadata(raw, false)
	if !strings.Contains(*safe.DebugQuery, "'O''****ly' /* masked */, TRUE, 'O''****ly' /* masked */") {
		t.Fatal(*safe.DebugQuery)
	}
	if !reflect.DeepEqual(safe.MaskedParameters, []bool{true, false}) {
		t.Fatal(safe.MaskedParameters)
	}
	if !reflect.DeepEqual(safe, projectedSQLMetadata(safe, false)) {
		t.Fatal("projection is not idempotent")
	}
	if !strings.Contains(*projectedSQLMetadata(safe, true).DebugQuery, "MASKED") {
		t.Fatal("safe projection was upgraded")
	}
	if raw.Parameters[0].V != "O'Reilly" {
		t.Fatal("source modified")
	}
}

func TestSQLMaskPoliciesCredentialOverrideAndInlineOmission(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	for _, raw := range []data_service.ExecutionMetadata{
		{ParameterizedSQL: "UPDATE account SET password = ?", Parameters: []core.Value{core.ValText("CREDENTIAL-CANARY")}, ParameterLogPolicies: []string{"plain"}},
		{ParameterizedSQL: "UPDATE account SET password = 'CREDENTIAL-CANARY'"},
	} {
		var output bytes.Buffer
		NewSensitiveDiagnosticSQLLogSink(&output).WriteSQLLog(raw)
		if strings.Contains(output.String(), "CREDENTIAL-CANARY") {
			t.Fatal("credential escaped debug guard")
		}
		if !strings.Contains(output.String(), "NOT REPLAYABLE") {
			t.Fatal(output.String())
		}
	}
}

func TestGeneratedCredentialColumnPreservesOtherFieldPolicies(t *testing.T) {
	raw := data_service.ExecutionMetadata{Backend: "sqlite", GeneratedSQL: true,
		ParameterizedSQL:     "SELECT password_hash FROM customer_data WHERE display_name = ? AND public_address = ? AND password_hash = ?",
		Parameters:           []core.Value{core.ValText("Riverside"), core.ValText("1 Runtime Road"), core.ValText("PASSWORD-CANARY")},
		ParameterLogPolicies: []string{"masked", "plain", "credential"}}
	for _, debug := range []bool{false, true} {
		safe := projectedSQLMetadata(raw, debug)
		name := "Ri*****de"
		if debug {
			name = "Riverside"
		}
		if !reflect.DeepEqual(safe.ParameterLogPolicies, raw.ParameterLogPolicies) ||
			!strings.Contains(*safe.DebugQuery, name) || !strings.Contains(*safe.DebugQuery, "1 Runtime Road") ||
			strings.Contains(*safe.DebugQuery, "PASSWORD-CANARY") {
			t.Fatal("generated per-field policies overridden", *safe.DebugQuery)
		}
		if !reflect.DeepEqual(safe, projectedSQLMetadata(safe, debug)) {
			t.Fatal("repeated projection changed values")
		}
	}
	if raw.Parameters[2].V != "PASSWORD-CANARY" {
		t.Fatal("driver parameters changed")
	}
}

func TestSQLMaskUnknownBindingsStayPrivateWithDebugOptIn(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	for _, policies := range [][]string{nil, {"unknown"}, {"unsupported-policy"}} {
		comment := "what: locate UNKNOWN-BINDING-CANARY"
		raw := data_service.ExecutionMetadata{ParameterizedSQL: "SELECT ?",
			Parameters:           []core.Value{core.ValText("UNKNOWN-BINDING-CANARY")},
			ParameterLogPolicies: policies, Comment: &comment}
		safe := projectedSQLMetadata(raw, true)
		if strings.Contains(*safe.DebugQuery, "UNKNOWN-BINDING-CANARY") ||
			strings.Contains(*safe.Comment, "UNKNOWN-BINDING-CANARY") || safe.Parameters[0].V != "[REDACTED]" {
			t.Fatal("unknown binding escaped debug projection")
		}
		if !safe.MaskedParameters[0] || !strings.Contains(*safe.DebugQuery, "SELECT") ||
			!strings.Contains(*safe.DebugQuery, "NOT REPLAYABLE") {
			t.Fatal("expected expanded, non-replayable masked SQL")
		}
		if raw.Parameters[0].V != "UNKNOWN-BINDING-CANARY" {
			t.Fatal("driver binding modified")
		}
	}
}

func TestSQLMaskDebugOnlyExposesExplicitBusinessPolicies(t *testing.T) {
	raw := data_service.ExecutionMetadata{ParameterizedSQL: "SELECT ?, ?, ?, ?",
		Parameters: []core.Value{core.ValText("Ordinary"), core.ValText("Riverside"),
			core.ValText("UNKNOWN-BINDING-CANARY"), core.ValText("CREDENTIAL-CANARY")},
		ParameterLogPolicies: []string{"plain", "masked", "unknown", "credential"}}
	safe := projectedSQLMetadata(raw, true)
	if !reflect.DeepEqual(safe.MaskedParameters, []bool{false, false, true, true}) ||
		!reflect.DeepEqual(safe.Parameters, []core.Value{core.ValText("Ordinary"), core.ValText("Riverside"),
			core.ValText("[REDACTED]"), core.ValText("[REDACTED]")}) {
		t.Fatal("debug policy classification mismatch")
	}
}

func TestSQLMaskPoliciesInvalidBindingsAndLiteralTokens(t *testing.T) {
	for _, query := range []string{"SELECT $0", "SELECT $2", "SELECT ?, ?", "SELECT 1", ""} {
		safe := projectedSQLMetadata(data_service.ExecutionMetadata{ParameterizedSQL: query, Parameters: []core.Value{core.ValText("BIND-CANARY")}}, false)
		if safe.OmissionReason == "" || *safe.DebugQuery != "[REDACTED SQL; NOT REPLAYABLE]" {
			t.Fatal(query, safe.OmissionReason, *safe.DebugQuery)
		}
	}
	safe := projectedSQLMetadata(data_service.ExecutionMetadata{ParameterizedSQL: "SELECT `field?`, 'fixed?', ? /* fixed ? */",
		GeneratedSQL: true, Parameters: []core.Value{core.ValText("Riverside")}, ParameterLogPolicies: []string{"masked"}}, false)
	if !strings.Contains(*safe.DebugQuery, "`field?`, 'fixed?', 'Ri*****de' /* masked */") {
		t.Fatal(*safe.DebugQuery)
	}
}

func TestSQLMaskEvidenceStoreCannotBypassProjection(t *testing.T) {
	t.Setenv(plaintextLogEnv, "")
	store := NewSQLExecutionEvidenceStore()
	raw := maskContractEntry("EVIDENCE-CANARY")
	store.RecordExecutionMetadata(raw)
	snapshot := store.Snapshot()
	if strings.Contains(*snapshot[0].DebugQuery, "EVIDENCE-CANARY") || snapshot[0].Parameters[0].V == "EVIDENCE-CANARY" {
		t.Fatal("raw value in evidence store")
	}
	snapshot[0].Parameters[0] = core.ValText("MODIFIED")
	snapshot[0].ParameterLogPolicies[0] = "plain"
	if store.Snapshot()[0].Parameters[0].V == "MODIFIED" || store.Snapshot()[0].ParameterLogPolicies[0] == "plain" {
		t.Fatal("snapshot mutated store")
	}
	if raw.Parameters[0].V != "EVIDENCE-CANARY" {
		t.Fatal("source mutated")
	}
}

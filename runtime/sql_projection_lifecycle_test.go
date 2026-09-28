package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

func inheritedDebugFixture() ds.ExecutionMetadata {
	text := "what: persist Riverside PASSWORD-CANARY ordinary address"
	write := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "UPDATE customer SET name=?, password=?",
		Parameters: []core.Value{core.ValText("Riverside"), core.ValText("PASSWORD-CANARY")}, ParameterLogPolicies: []string{"masked", "credential"}}
	return ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "SELECT name FROM customer WHERE id=? LIMIT 10000",
		Parameters: []core.Value{core.ValI64(1)}, ParameterLogPolicies: []string{"plain"}, Comment: &text, Purpose: &text, AuditReason: &text,
		TraceChain: []*core.TraceNode{core.NewTypedTraceNode("sql", text, text)}, InheritedIntent: logprivacy.NewIntentSource(write)}
}

func TestRetainedDebugIntentDowngradesWithoutRawProvenance(t *testing.T) {
	raw := inheritedDebugFixture()
	debug := projectedSQLMetadata(raw, true)
	if !strings.Contains(*debug.AuditReason, "Riverside") {
		t.Fatal("debug fixture not plaintext")
	}
	safe := projectedSQLMetadata(debug, false)
	if strings.Contains(*safe.AuditReason, "Riverside") {
		t.Fatal("inherited debug intent leaked on downgrade")
	}
	if !strings.Contains(*safe.AuditReason, "ordinary address") {
		t.Fatal("safe intent lost")
	}
	if logprivacy.ReadIntentSource(debug.InheritedIntent) != nil {
		t.Fatal("raw provenance retained")
	}
	if !reflect.DeepEqual(safe, projectedSQLMetadata(debug, false)) {
		t.Fatal("unstable downgrade")
	}
	*safe.AuditReason = "corrupted"
	safe.TraceChain[0].Comment = "corrupted"
	if strings.Contains(*projectedSQLMetadata(debug, false).AuditReason, "corrupted") {
		t.Fatal("safe cache aliased caller")
	}
	if !strings.Contains(*raw.AuditReason, "Riverside") {
		t.Fatal("source mutated")
	}
}

func TestCopiedOrMutatedDebugIntentFailsClosed(t *testing.T) {
	for _, mode := range []string{"json", "modified", "header-only"} {
		t.Run(mode, func(t *testing.T) {
			debug := projectedSQLMetadata(inheritedDebugFixture(), true)
			if mode == "modified" {
				debug.ParameterizedSQL = "SELECT name FROM customer WHERE id=? LIMIT 20000"
			} else {
				encoded, err := json.Marshal(debug)
				if err != nil {
					t.Fatal(err)
				}
				var copied ds.ExecutionMetadata
				if err = json.Unmarshal(encoded, &copied); err != nil {
					t.Fatal(err)
				}
				debug = copied
				if mode == "header-only" {
					debug.LogMode = ""
				}
			}
			safe := projectedSQLMetadata(debug, false)
			encoded, _ := json.Marshal(safe)
			if strings.Contains(string(encoded), "Riverside") || strings.Contains(string(encoded), "PASSWORD-CANARY") {
				t.Fatal("copied/modified intent leaked")
			}
			if *safe.AuditReason != "[REDACTED]" {
				t.Fatal("unproven intent should be hidden")
			}
			if safe.ParameterizedSQL != debug.ParameterizedSQL {
				t.Fatal("SQL structure changed")
			}
		})
	}
}

func TestEvidenceSnapshotAfterDebugRevocationAndFileOutput(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	evidence := NewSQLExecutionEvidenceStore()
	evidence.WriteSQLLog(inheritedDebugFixture())
	if !strings.Contains(*evidence.Snapshot()[0].AuditReason, "Riverside") {
		t.Fatal("debug not active")
	}
	t.Setenv(plaintextLogEnv, "")
	safe := evidence.Snapshot()[0]
	file, err := os.Create(filepath.Join(t.TempDir(), "safe.log"))
	if err != nil {
		t.Fatal(err)
	}
	NewTextDiagnosticSQLLogSink(file).WriteSQLLog(safe)
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(text), "Riverside") || strings.Contains(string(text), "PASSWORD-CANARY") {
		t.Fatal("disk output leaked revoked debug intent")
	}
	if !strings.Contains(string(text), "LIMIT 10000") || !strings.Contains(string(text), "ordinary address") {
		t.Fatal("diagnostic intent/SQL lost")
	}
}

func TestMalformedMaskFlagsAreUnknown(t *testing.T) {
	raw := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "SELECT ?, ?", Parameters: []core.Value{core.ValText("Riverside"), core.ValText("PASSWORD-CANARY")},
		ParameterLogPolicies: []string{"plain", "plain"}, MaskedParameters: []bool{true}}
	for _, debug := range []bool{false, true} {
		safe := projectedSQLMetadata(raw, debug)
		if safe.OmissionReason != "mask-count-mismatch" {
			t.Fatal("malformed flags were accepted")
		}
		if strings.Contains(fmt.Sprint(safe.Parameters), "Riverside") || strings.Contains(fmt.Sprint(safe.Parameters), "CANARY") {
			t.Fatal("malformed mask revealed values")
		}
		if projectedSQLMetadata(safe, false).OmissionReason != "mask-count-mismatch" {
			t.Fatal("repeated projection lost safe omission reason")
		}
	}
}

func TestSQLNullRemainsNullForAllMaskPolicies(t *testing.T) {
	for _, policy := range []string{"plain", "masked", "unknown", "credential"} {
		for _, null := range []core.Value{core.ValNull(), core.ValTypedNull(core.TypeText), core.ValTypedNull(core.TypeDecimal)} {
			raw := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "SELECT ?", Parameters: []core.Value{null}, ParameterLogPolicies: []string{policy}}
			for _, debug := range []bool{false, true} {
				safe := projectedSQLMetadata(raw, debug)
				if !reflect.DeepEqual(safe.Parameters[0], null) || !strings.Contains(*safe.DebugQuery, "NULL") {
					t.Fatalf("null changed for %s/%v", policy, debug)
				}
			}
		}
	}
}

func TestDebugAlternativeDoesNotAliasMutableValues(t *testing.T) {
	raw := inheritedDebugFixture()
	raw.Parameters = []core.Value{{V: []byte{1, 2}}}
	debug := projectedSQLMetadata(raw, true)
	safe := projectedSQLMetadata(debug, false)
	safe.Parameters[0].V.([]byte)[0] = 42
	*safe.Comment = "corrupted"
	again := projectedSQLMetadata(debug, false)
	if again.Parameters[0].V.([]byte)[0] != 1 || strings.Contains(*again.Comment, "corrupted") {
		t.Fatal("safe alternative aliases output")
	}
	debug.Parameters[0].V.([]byte)[0] = 99
	if *projectedSQLMetadata(debug, false).AuditReason != "[REDACTED]" {
		t.Fatal("mutated binary binding did not invalidate provenance")
	}
	if raw.Parameters[0].V.([]byte)[0] != 1 {
		t.Fatal("driver data aliased")
	}
}

func TestDebugAlternativeCanBeReadConcurrently(t *testing.T) {
	debug := projectedSQLMetadata(inheritedDebugFixture(), true)
	var tasks sync.WaitGroup
	for i := 0; i < 16; i++ {
		tasks.Add(1)
		go func() {
			defer tasks.Done()
			safe := projectedSQLMetadata(debug, false)
			if strings.Contains(*safe.AuditReason, "Riverside") {
				t.Error("concurrent downgrade leaked")
			}
			*safe.AuditReason = "caller-owned edit"
		}()
	}
	tasks.Wait()
	if !strings.Contains(*debug.AuditReason, "Riverside") {
		t.Fatal("projection mutated debug record")
	}
}

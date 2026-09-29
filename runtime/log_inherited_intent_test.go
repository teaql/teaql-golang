package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

func TestInheritedIntentProjection(t *testing.T) {
	for _, debug := range []bool{false, true} {
		t.Run(fmt.Sprint(debug), func(t *testing.T) {
			source := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "UPDATE customer SET name = ?, password_hash = ?, other = ?, payload = ?, amount = ?, address = ?",
				Parameters: []core.Value{core.ValText("Riverside"), core.ValText("PASSWORD-CANARY"), core.ValText("UNKNOWN-CANARY"),
					{V: map[string]any{"api_key": "NESTED-CANARY"}}, core.ValF64(12345), core.ValText("ordinary address")},
				ParameterLogPolicies: []string{"masked", "credential", "unknown", "plain", "masked", "plain"}}
			text := "inspect Riverside PASSWORD-CANARY UNKNOWN-CANARY NESTED-CANARY 12345 ordinary address"
			read := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "SELECT id FROM customer WHERE id = ?", Parameters: []core.Value{core.ValI64(1)}, ParameterLogPolicies: []string{"plain"},
				Comment: &text, Purpose: &text, AuditReason: &text, TraceChain: []*core.TraceNode{core.NewTypedTraceNode("sql", text, text)},
				InheritedIntent: logprivacy.NewIntentSource(source)}
			safe := projectedSQLMetadata(read, debug)
			if logprivacy.ReadIntentSource(safe.InheritedIntent) != nil {
				t.Fatal("raw provenance retained")
			}
			for _, value := range []string{*safe.Comment, *safe.Purpose, *safe.AuditReason, safe.TraceChain[0].Name, safe.TraceChain[0].Comment} {
				for _, secret := range []string{"PASSWORD-CANARY", "UNKNOWN-CANARY", "NESTED-CANARY"} {
					if strings.Contains(value, secret) {
						t.Fatalf("leaked %s", secret)
					}
				}
				if strings.Contains(value, "Riverside") != debug || strings.Contains(value, "12345") != debug || !strings.Contains(value, "ordinary address") {
					t.Fatalf("incorrect policy: %s", value)
				}
			}
			encoded, err := json.Marshal(safe)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), "PASSWORD-CANARY") {
				t.Fatal("serialization leaked provenance")
			}
			if !debug {
				again := projectedSQLMetadata(safe, true)
				if strings.Contains(fmt.Sprintf("%+v", again), "Riverside") || strings.Contains(*again.AuditReason, "Riverside") {
					t.Fatal("repeat projection resurrected raw data")
				}
			}
			if source.Parameters[0].V != "Riverside" || *read.AuditReason != text {
				t.Fatal("changed source")
			}
			for _, rendered := range []string{fmt.Sprintf("%v", read.InheritedIntent), fmt.Sprintf("%+v", read.InheritedIntent), fmt.Sprintf("%#v", read.InheritedIntent)} {
				if strings.Contains(rendered, "PASSWORD-CANARY") {
					t.Fatal("opaque provenance formatted raw values")
				}
			}
		})
	}
}

func TestInheritedIntentUnknownInvalidPolicies(t *testing.T) {
	for _, policies := range [][]string{nil, {"plain", "plain"}} {
		source := ds.ExecutionMetadata{GeneratedSQL: true, Parameters: []core.Value{core.ValText("UNKNOWN-CANARY")}, ParameterLogPolicies: policies}
		text := "inspect UNKNOWN-CANARY"
		read := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: "SELECT id FROM customer", Comment: &text, InheritedIntent: logprivacy.NewIntentSource(source)}
		if strings.Contains(*projectedSQLMetadata(read, true).Comment, "UNKNOWN-CANARY") {
			t.Fatal("debug revealed unclassified source")
		}
	}
}

func TestRepeatedMaskProjectionPreservesCompiledSQLStructure(t *testing.T) {
	for _, value := range []core.Value{core.ValI64(1), core.ValText("customer")} {
		query := "SELECT id FROM customer WHERE version = ? LIMIT 10000"
		metadata := ds.ExecutionMetadata{GeneratedSQL: true, ParameterizedSQL: query, Parameters: []core.Value{value}}
		safe := projectedSQLMetadata(metadata, false)
		again := projectedSQLMetadata(safe, true)
		if safe.ParameterizedSQL != query || again.ParameterizedSQL != query {
			t.Fatal("masking a binding corrupted the compiled SQL template")
		}
		if again.DebugQuery == nil || !strings.Contains(*again.DebugQuery, "FROM customer") || !strings.Contains(*again.DebugQuery, "LIMIT 10000") {
			t.Fatal("repeated sink projection corrupted SQL structure")
		}
	}
}

package runtime

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestMutationLineagePrivacyAndOptionalIDAreRetainedAtEveryProjection(t *testing.T) {
	t.Setenv(plaintextLogEnv, "")
	id := uint64(100)
	raw := inheritedDebugFixture()
	frame := core.NewTypedTraceNode("auditReason", "Payment", "approve Riverside PASSWORD-CANARY")
	frame.EntityId = &id
	raw.MutationLineage = []*core.TraceNode{frame}
	safe := projectedSQLMetadata(raw, false)
	encoded, _ := json.Marshal(safe)
	if strings.Contains(string(encoded), "Riverside") || strings.Contains(string(encoded), "PASSWORD-CANARY") {
		t.Fatal("lineage leaks through JSON")
	}
	if len(safe.MutationLineage) != 1 || *safe.MutationLineage[0].EntityId != 100 || safe.MutationLineage[0].Name != "Payment" {
		t.Fatal("typed lineage structure lost")
	}
	var disk bytes.Buffer
	NewTextDiagnosticSQLLogSink(&disk).WriteSQLLog(raw)
	if strings.Contains(disk.String(), "Riverside") || strings.Contains(disk.String(), "PASSWORD-CANARY") {
		t.Fatal("lineage leaks through text output")
	}
	if !strings.Contains(disk.String(), "mutationLineage=") || !strings.Contains(disk.String(), "Payment#100") {
		t.Fatal("text output hides lineage/ID")
	}
	debug := projectedSQLMetadata(raw, true)
	if !strings.Contains(debug.MutationLineage[0].Comment, "Riverside") || strings.Contains(debug.MutationLineage[0].Comment, "PASSWORD-CANARY") {
		t.Fatal("debug/credential contract changed")
	}
	revoked := projectedSQLMetadata(debug, false)
	if strings.Contains(revoked.MutationLineage[0].Comment, "Riverside") {
		t.Fatal("debug downgrade leaks lineage")
	}
	*revoked.MutationLineage[0].EntityId = 200
	if *projectedSQLMetadata(debug, false).MutationLineage[0].EntityId != 100 || id != 100 {
		t.Fatal("projection cache aliases IDs")
	}
	// A modified trace must not reuse a remembered projection for different IDs.
	*debug.MutationLineage[0].EntityId = 300
	changed := projectedSQLMetadata(debug, false)
	if *changed.MutationLineage[0].EntityId != 300 || strings.Contains(changed.MutationLineage[0].Comment, "Riverside") {
		t.Fatal("trace tampering reused stale safe projection")
	}
}

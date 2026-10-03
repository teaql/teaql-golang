package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLoadedScalarSnapshotOwnsOnlyDeclaredScalars(t *testing.T) {
	payload := map[string]any{"secret": "PRIVATE-OLD"}
	record := Record{"name": ValText("PRIVATE-OLD"), "payload": ValJson(payload), "children": ValJson([]any{"not a scalar"})}
	snapshot := NewLoadedScalarSnapshot(record, "name", "payload")
	payload["secret"] = "changed"
	record["name"] = ValText("changed")
	first := snapshot.Values()
	if len(first) != 2 || first["name"].V != "PRIVATE-OLD" || first["payload"].V.(map[string]any)["secret"] != "PRIVATE-OLD" {
		t.Fatal("snapshot aliases input or contains a relation")
	}
	first["payload"].V.(map[string]any)["secret"] = "changed twice"
	if snapshot.Values()["payload"].V.(map[string]any)["secret"] != "PRIVATE-OLD" {
		t.Fatal("snapshot exposes mutable values")
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil || string(encoded) != "{}" || strings.Contains(fmt.Sprintf("%+v %#v", snapshot, snapshot), "PRIVATE-OLD") {
		t.Fatal("snapshot leaked through serialization/formatting")
	}
}

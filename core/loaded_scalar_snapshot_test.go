package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLoadedScalarSnapshotValuesForRequiresExactOwnerVersionAndReturnsACopy(t *testing.T) {
	key := NewEntityKey("Entry", ValU64(7))
	snapshot := NewOwnedLoadedScalarSnapshot(key, 3, Record{"name": ValText("original")}, "name")
	values := snapshot.ValuesFor(key, 3)
	if values["name"] != ValText("original") {
		t.Fatal("matching baseline unavailable")
	}
	values["name"] = ValText("mutated copy")
	if snapshot.ValuesFor(key, 3)["name"] != ValText("original") {
		t.Fatal("baseline was mutable")
	}
	for _, mismatch := range []struct {
		key     EntityKey
		version int64
	}{
		{key, 4}, {NewEntityKey("Other", ValU64(7)), 3}, {NewEntityKey("Entry", ValU64(8)), 3},
	} {
		if snapshot.ValuesFor(mismatch.key, mismatch.version) != nil {
			t.Fatal("unrelated snapshot admitted")
		}
	}
	if NewLoadedScalarSnapshot(Record{"name": ValText("privacy only")}, "name").ValuesFor(key, 3) != nil {
		t.Fatal("unowned snapshot admitted")
	}
}

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

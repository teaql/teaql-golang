package data_service

import (
	"encoding/json"
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"strings"
	"testing"
)

func TestMutationPrivacyCaptureDoesNotAlterPayloadOrSerialize(t *testing.T) {
	record := core.Record{"name": core.ValText("PRIVATE-OLD")}
	entry := NewMutationPrivacyEntry("item", record)
	source := NewMutationPrivacy(entry)
	record["name"] = core.ValText("tampered")
	request, err := NewMutationRequest(&DeleteMutation{Cmd: core.NewDeleteCommand("item", core.ValU64(1))}, "remove item")
	if err != nil {
		t.Fatal(err)
	}
	captured, err := WithMutationPrivacy(request, source)
	if err != nil {
		t.Fatal(err)
	}
	recaptured, err := CaptureMutationRequest(captured)
	if err != nil {
		t.Fatal(err)
	}
	entries := MutationPrivacyEntries(recaptured)
	if len(entries) != 1 || entries[0].Values()["name"].V != "PRIVATE-OLD" {
		t.Fatal("capture lost immutable source")
	}
	entries[0].Values()["name"] = core.ValText("changed copy")
	if MutationPrivacyEntries(captured)[0].Values()["name"].V != "PRIVATE-OLD" || len(MutationPrivacyEntries(request)) != 0 {
		t.Fatal("capture modified input")
	}
	encoded, err := json.Marshal(captured)
	if err != nil || strings.Contains(string(encoded), "PRIVATE-OLD") || strings.Contains(fmt.Sprintf("%+v %#v", source, source), "PRIVATE-OLD") {
		t.Fatal("private provenance leaked")
	}
	var forged DeleteMutation
	if err := json.Unmarshal([]byte(`{"privacy":{"entries":[{"entity":"item","values":{"name":"forged"}}]}}`), &forged); err != nil {
		t.Fatal(err)
	}
	if len(MutationPrivacyEntries(&forged)) != 0 {
		t.Fatal("wire input forged provenance")
	}
}

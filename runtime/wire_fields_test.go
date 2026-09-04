package runtime

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestWireFieldNormalizationAndProvenance(t *testing.T) {
	meta, err := CreateWireEntityMetadata("School", []string{"name", "school_type"}, JsonFieldCamelCase, map[string][]string{"school_type": {"school_type"}})
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := NormalizeWireInput(map[string]any{"school_type": int64(1001)}, meta, "/school")
	if err != nil || normalized.Values["school_type"] != int64(1001) {
		t.Fatalf("normalize: %#v %v", normalized, err)
	}
	results := RetainSubmittedPaths([]CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("school_type")}}, normalized)
	if results[0].SourceInstancePath != "/school/school_type" {
		t.Fatal(results[0])
	}
}
func TestWireFieldUnknownAndCollision(t *testing.T) {
	meta, _ := CreateWireEntityMetadata("School", []string{"school_type"}, JsonFieldCamelCase, map[string][]string{"school_type": {"school_type"}})
	_, err := NormalizeWireInput(map[string]any{"bad/name": 1}, meta, "")
	var wireErr *WireInputError
	if !errors.As(err, &wireErr) || wireErr.Code != "WIRE_UNKNOWN_FIELD" || wireErr.InstancePath != "/bad~1name" {
		t.Fatal(err)
	}
	_, err = NormalizeWireInput(map[string]any{"schoolType": 1, "school_type": 2}, meta, "")
	if !errors.As(err, &wireErr) || wireErr.Code != "WIRE_FIELD_COLLISION" {
		t.Fatal(err)
	}
}
func TestWireCheckResultDTO(t *testing.T) {
	wire := (CheckResult{RuleID: "required", EntityType: "School", CanonicalLocation: Location().Property("school_type"), SourceInstancePath: "/school_type"}).ToWire(JsonFieldCamelCase)
	if wire.InstancePath != "/schoolType" || wire.SourceInstancePath != "/school_type" {
		t.Fatal(wire)
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	segments := decoded["location"].([]any)
	first := segments[0].(map[string]any)
	if first["kind"] != "property" || first["name"] != "school_type" {
		t.Fatalf("unexpected wire location: %s", encoded)
	}
}

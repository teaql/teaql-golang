package entityjson

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
)

type projectionDocument struct{ fields, properties core.Record }

func (d *projectionDocument) EntityName() string               { return "Probe" }
func (d *projectionDocument) LoadedJSONFields() core.Record    { return d.fields }
func (d *projectionDocument) JSONQueryProperties() core.Record { return d.properties }

func TestSelectedQueryPropertiesPreserveNullZeroAndExactValues(t *testing.T) {
	m := model(t, Limits{})
	for _, profile := range profiles {
		t.Run(string(profile), func(t *testing.T) {
			doc := &projectionDocument{fields: core.Record{"reference_code": core.ValText("selected")}, properties: core.Record{
				"total": core.ValNull(), "zero": core.ValI64(0), "count": core.ValU64(math.MaxUint64),
				"amount_sum": core.ValDecimal(decimal.RequireFromString("12345678901234567890.123400")),
			}}
			payload, err := m.Serialize("Probe", doc, profile)
			if err != nil {
				t.Fatal(err)
			}
			var wire map[string]json.RawMessage
			if err := json.Unmarshal(payload, &wire); err != nil {
				t.Fatal(err)
			}
			for key, expected := range map[string]string{"total": "null", "zero": "0", "count": "18446744073709551615", "amount_sum": `"12345678901234567890.123400"`} {
				if string(wire[key]) != expected {
					t.Fatalf("selected query property %s: got %s want %s", key, wire[key], expected)
				}
			}
			if _, present := wire["missing"]; present {
				t.Fatal("absent property fabricated")
			}
			if _, err := m.Parse("Probe", payload, profile); err == nil {
				t.Fatal("output-only query properties became writable input")
			}
			if !doc.properties["total"].IsNull() || doc.fields["reference_code"].V != "selected" {
				t.Fatal("serialization changed the source")
			}
		})
	}
}

func TestQueryPropertiesCannotShadowModeledFieldsOrBypassLimits(t *testing.T) {
	m := model(t, Limits{})
	for _, name := range []string{"reference_code", "referenceCode", "ReferenceCode", "root_record_id", "rootRecordEntity", "childList"} {
		t.Run(name, func(t *testing.T) {
			_, err := m.Serialize("Probe", &projectionDocument{properties: core.Record{name: core.ValNull()}})
			if err == nil || !strings.Contains(err.Error(), "WIRE_QUERY_PROPERTY_COLLISION") {
				t.Fatal("projection shadowed a model field", err)
			}
		})
	}
	for _, value := range []any{math.NaN(), math.Inf(1), float32(math.Inf(-1)), func() {}} {
		_, err := m.Serialize("Probe", &projectionDocument{properties: core.Record{"total": {V: value}}})
		expectError(t, err, "WIRE_INVALID_TYPE", "total", "/total")
	}
	_, err := m.Serialize("Probe", &projectionDocument{properties: core.Record{"total": core.ValDecimal(decimal.New(1, math.MinInt32))}})
	expectError(t, err, "WIRE_DECIMAL_PRECISION", "total", "/total")
	_, err = model(t, Limits{MaxBytes: 32}).Serialize("Probe", &projectionDocument{properties: core.Record{"total": core.ValText(strings.Repeat("secret-value", 20))}})
	expectError(t, err, "WIRE_JSON_LIMIT", "total", "/total")
	_, err = model(t, Limits{MaxDepth: 2}).Serialize("Probe", &projectionDocument{properties: core.Record{"total": core.ValJson([]any{[]any{nil}})}})
	if err == nil || !strings.Contains(err.Error(), "WIRE_JSON_LIMIT") {
		t.Fatal("projection bypassed depth limit", err)
	}
}

func TestUnmarkedDynamicStateIsStillPrivate(t *testing.T) {
	m := model(t, Limits{})
	for _, value := range []any{
		core.Record{"_private": core.ValText("secret-value")},
		&projectionDocument{fields: core.Record{"_private": core.ValText("secret-value")}},
	} {
		payload, err := m.Serialize("Probe", value)
		if err != nil || string(payload) != "{}" {
			t.Fatal("unmarked dynamic state escaped", string(payload), err)
		}
	}
}

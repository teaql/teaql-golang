package entityjson

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

func definitions() []Schema {
	return []Schema{{EntityType: "Probe", Fields: []Field{
		{CanonicalName: "id", Kind: ID}, {CanonicalName: "version", Kind: Version},
		{CanonicalName: "reference_code", Kind: Text}, {CanonicalName: "small_number", Kind: Integer},
		{CanonicalName: "large_number", Kind: Long}, {CanonicalName: "amount", Kind: Decimal},
		{CanonicalName: "active", Kind: Boolean}, {CanonicalName: "rate", Kind: Float}, {CanonicalName: "ratio", Kind: Double},
		{CanonicalName: "event_date", Kind: Date}, {CanonicalName: "clock_time", Kind: Time}, {CanonicalName: "event_time", Kind: DateTime},
		{CanonicalName: "payload", Kind: JSON},
		{CanonicalName: "root_record", RecordName: "root_record_id", Kind: Reference, Target: "Root", RelationName: "rootRecordEntity"},
		{CanonicalName: "child_list", RecordName: "childList", Kind: Many, Target: "Probe"},
	}}, {EntityType: "Root", Fields: []Field{{CanonicalName: "id", Kind: ID}, {CanonicalName: "display_name", Kind: Text}}}}
}

func model(t *testing.T, limits Limits) *Model {
	t.Helper()
	m, err := NewModel(definitions(), runtime.JsonFieldCamelCase, limits)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func expectError(t *testing.T, err error, code, canonical, path string) {
	t.Helper()
	var target *Error
	if !errors.As(err, &target) || target.Code != code || target.Location.ModelPath() != canonical || target.InstancePath != path {
		t.Fatalf("unexpected JSON diagnostic: %#v; want %s %s %s", err, code, canonical, path)
	}
	if strings.Contains(err.Error(), "secret-value") {
		t.Fatal("JSON diagnostic leaked a submitted value")
	}
}

func TestEntityJSONExactValuesAndProfiles(t *testing.T) {
	m := model(t, Limits{})
	clock, _ := core.ParseTimeOfDay("23:59:59.123456789")
	instant := time.Date(2024, 2, 29, 12, 34, 56, 789000000, time.FixedZone("offset", 8*3600))
	record := core.Record{"id": core.ValU64(math.MaxUint64), "version": core.ValI64(7),
		"reference_code": core.ValText("🙂 secret-value"), "small_number": core.ValI64(math.MinInt32),
		"large_number": core.ValI64(9007199254740993), "amount": core.ValDecimal(decimal.RequireFromString("12345678901234567890.123400")),
		"active": core.ValBool(false), "rate": core.ValF64(float64(float32(0.1))), "ratio": core.ValF64(-0.125),
		"event_date": core.ValDate(time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)), "clock_time": core.ValTimeOfDay(clock),
		"event_time": {V: instant}, "payload": core.ValJson(map[string]any{"large": json.Number("9007199254740993"), "null": nil}),
		"root_record_id": core.ValU64(9), "rootRecordEntity": core.ValObject(core.Record{"id": core.ValU64(9), "display_name": core.ValText("Platform")}),
		"childList": core.ValList([]core.Value{}), "_private": core.ValText("must not escape"),
	}
	for _, profile := range profiles {
		t.Run(string(profile), func(t *testing.T) {
			payload, err := m.Serialize("Probe", record, profile)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(payload), "must not escape") {
				t.Fatal("private dynamic state escaped")
			}
			var wire map[string]json.RawMessage
			if err := json.Unmarshal(payload, &wire); err != nil {
				t.Fatal(err)
			}
			if string(wire[profile.Render("id")]) != `"18446744073709551615"` || string(wire[profile.Render("large_number")]) != `"9007199254740993"` ||
				string(wire[profile.Render("amount")]) != `"12345678901234567890.123400"` || string(wire[profile.Render("event_time")]) != `"2024-02-29T04:34:56.789Z"` {
				t.Fatal("wire precision or naming changed", string(payload))
			}
			decoded, err := m.Parse("Probe", payload, profile)
			if err != nil {
				t.Fatal(err)
			}
			if id, ok := decoded["id"].V.(uint64); !ok || id != math.MaxUint64 {
				t.Fatal("unsigned ID lost")
			}
			if got, ok := decoded["large_number"].V.(int64); !ok || got != 9007199254740993 {
				t.Fatal("Long lost precision")
			}
			if got := decoded["amount"].V.(decimal.Decimal); got.Exponent() != -6 || !got.Equal(record["amount"].V.(decimal.Decimal)) {
				t.Fatal("Decimal scale or coefficient changed")
			}
			if decoded["clock_time"].V != clock || decoded["event_time"].V != core.Timestamp(instant.UnixMilli()) || decoded["root_record_id"].V != uint64(9) {
				t.Fatal("typed temporal/FK value changed")
			}
			if _, found := decoded["missing"]; found {
				t.Fatal("absent field was loaded")
			}
			again, err := m.Serialize("Probe", decoded, profile)
			if err != nil || string(again) != string(payload) {
				t.Fatal("non-deterministic exact JSON roundtrip", err, string(again))
			}
		})
	}
}

func TestEntityJSONSparseNullAndKnownReferenceIdentity(t *testing.T) {
	m := model(t, Limits{})
	for _, profile := range profiles {
		wire := `{"` + profile.Render("reference_code") + `":null,"` + profile.Render("root_record") + `":"42","` + profile.Render("child_list") + `":[]}`
		record, err := m.Parse("Probe", []byte(wire), profile)
		if err != nil {
			t.Fatal(err)
		}
		if !record["reference_code"].IsNull() {
			t.Fatal("explicit null disappeared")
		}
		if _, found := record["small_number"]; found {
			t.Fatal("omitted value acquired a default")
		}
		target := record["rootRecordEntity"].V.(core.Record)
		if target["id"].V != uint64(42) || len(target) != 1 {
			t.Fatal("known FK must retain target ID without inventing details")
		}
		payload, err := m.Serialize("Probe", record, profile)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(payload), `"`+profile.Render("root_record")+`":{"`+profile.Render("id")+`":"42"}`) {
			t.Fatal("identity-only reference lost", string(payload))
		}
	}
	// A filtered/missing detail is not a null relationship when the FK is known.
	payload, err := m.Serialize("Probe", core.Record{"root_record_id": core.ValU64(42), "rootRecordEntity": core.ValNull()})
	if err != nil || string(payload) != `{"rootRecord":"42"}` {
		t.Fatal("filtered detail erased a known FK", err, string(payload))
	}
	for _, fk := range []core.Value{core.ValU64(43), core.ValNull()} {
		_, err := m.Serialize("Probe", core.Record{"root_record_id": fk, "rootRecordEntity": core.ValObject(core.Record{"id": core.ValU64(42)})})
		expectError(t, err, "WIRE_REFERENCE_ID_MISMATCH", "root_record", "/rootRecord")
	}
}

func TestEntityJSONRejectsAliasesMalformedValuesAndPrecisePaths(t *testing.T) {
	m := model(t, Limits{})
	cases := []struct{ name, payload, code, canonical, path string }{
		{"alias collision", `{"referenceCode":"x","reference_code":"secret-value"}`, "WIRE_FIELD_COLLISION", "reference_code", "/reference_code"},
		{"repeated name", `{"referenceCode":"x","referenceCode":"secret-value"}`, "WIRE_FIELD_COLLISION", "reference_code", "/referenceCode"},
		{"unknown", `{"unknown/~":"secret-value"}`, "WIRE_UNKNOWN_FIELD", "unknown/~", "/unknown~1~0"},
		{"native name not alias", `{"ReferenceCode":"secret-value"}`, "WIRE_UNKNOWN_FIELD", "ReferenceCode", "/ReferenceCode"},
		{"ID number", `{"id":9007199254740993}`, "WIRE_INVALID_TYPE", "id", "/id"},
		{"negative ID", `{"id":"-1"}`, "WIRE_INVALID_TYPE", "id", "/id"},
		{"ID overflow", `{"id":"18446744073709551616"}`, "WIRE_INVALID_TYPE", "id", "/id"},
		{"I32 overflow", `{"smallNumber":2147483648}`, "WIRE_INVALID_TYPE", "small_number", "/smallNumber"},
		{"I32 fractional", `{"smallNumber":1.5}`, "WIRE_INVALID_TYPE", "small_number", "/smallNumber"},
		{"I32 exponent", `{"smallNumber":1e0}`, "WIRE_INVALID_TYPE", "small_number", "/smallNumber"},
		{"Long number", `{"largeNumber":12}`, "WIRE_INVALID_TYPE", "large_number", "/largeNumber"},
		{"Long overflow", `{"largeNumber":"9223372036854775808"}`, "WIRE_INVALID_TYPE", "large_number", "/largeNumber"},
		{"Decimal number", `{"amount":0.1}`, "WIRE_INVALID_TYPE", "amount", "/amount"},
		{"Decimal exponent", `{"amount":"1e3"}`, "WIRE_INVALID_TYPE", "amount", "/amount"},
		{"Float overflow", `{"rate":1e40}`, "WIRE_INVALID_TYPE", "rate", "/rate"},
		{"Double overflow", `{"ratio":1e309}`, "WIRE_INVALID_TYPE", "ratio", "/ratio"},
		{"invalid date", `{"eventDate":"2024-02-30"}`, "WIRE_INVALID_TYPE", "event_date", "/eventDate"},
		{"invalid clock", `{"clockTime":"24:00:00"}`, "WIRE_INVALID_TYPE", "clock_time", "/clockTime"},
		{"clock zone", `{"clockTime":"12:00:00Z"}`, "WIRE_INVALID_TYPE", "clock_time", "/clockTime"},
		{"naive timestamp", `{"eventTime":"2024-01-01T12:00:00"}`, "WIRE_INVALID_TYPE", "event_time", "/eventTime"},
		{"timestamp precision", `{"eventTime":"2024-01-01T12:00:00.000000001Z"}`, "WIRE_TEMPORAL_PRECISION", "event_time", "/eventTime"},
		{"invalid UTC offset", `{"eventTime":"2024-01-01T12:00:00+24:00"}`, "WIRE_INVALID_TYPE", "event_time", "/eventTime"},
		{"target identity missing", `{"rootRecord":{"displayName":"secret-value"}}`, "WIRE_REFERENCE_ID_REQUIRED", "root_record", "/rootRecord"},
		{"null collection", `{"childList":null}`, "WIRE_INVALID_TYPE", "child_list", "/childList"},
		{"child type", `{"childList":[{"small_number":"secret-value"}]}`, "WIRE_INVALID_TYPE", "child_list[0].small_number", "/childList/0/small_number"},
		{"raw JSON repeated key", `{"payload":{"a":1,"a":2}}`, "WIRE_FIELD_COLLISION", "payload.a", "/payload/a"},
		{"top array", `[]`, "WIRE_ENTITY_OBJECT_REQUIRED", "", ""},
		{"top null", `null`, "WIRE_ENTITY_OBJECT_REQUIRED", "", ""},
		{"trailing data", `{} {}`, "WIRE_INVALID_JSON", "", ""},
		{"truncated data", `{"id":`, "WIRE_INVALID_JSON", "", ""},
		{"surrogate", `{"referenceCode":"\ud800"}`, "WIRE_INVALID_TEXT", "", ""},
		{"low surrogate", `{"referenceCode":"\udc00"}`, "WIRE_INVALID_TEXT", "", ""},
		{"bad UTF8", "{\"referenceCode\":\"\xff\"}", "WIRE_INVALID_TEXT", "", ""},
	}
	for _, vector := range cases {
		t.Run(vector.name, func(t *testing.T) {
			record, err := m.Parse("Probe", []byte(vector.payload))
			if record != nil {
				t.Fatal("failed parse returned a partial record")
			}
			expectError(t, err, vector.code, vector.canonical, vector.path)
		})
	}
	if _, err := m.Parse("Probe", []byte(`{"referenceCode":"\ud83d\ude42"}`)); err != nil {
		t.Fatal("valid surrogate pair was rejected", err)
	}
	if _, err := m.Parse("Probe", []byte(`{"referenceCode":"\\ud800"}`)); err != nil {
		t.Fatal("literal backslash text was rejected", err)
	}
}

type document struct {
	name   string
	fields core.Record
}

func (d *document) EntityName() string            { return d.name }
func (d *document) LoadedJSONFields() core.Record { return d.fields }

func TestEntityJSONBoundsCyclesAndConcurrentImmutableSchema(t *testing.T) {
	definitions := definitions()
	m, err := NewModel(definitions, runtime.JsonFieldCamelCase, Limits{MaxBytes: 200, MaxDepth: 4})
	if err != nil {
		t.Fatal(err)
	}
	definitions[0].Fields[2].CanonicalName = "changed_after_construction"
	if _, err := m.Parse("Probe", []byte(`{"referenceCode":"x"}`)); err != nil {
		t.Fatal("caller mutated owned schema", err)
	}
	_, err = m.Parse("Probe", []byte(`{"referenceCode":"`+strings.Repeat("x", 200)+`"}`))
	expectError(t, err, "WIRE_JSON_LIMIT", "", "")
	if payload, err := m.Serialize("Probe", core.Record{"reference_code": core.ValText(strings.Repeat("x", 201))}); payload != nil || err == nil {
		t.Fatal("oversize output was delivered")
	}
	_, err = m.Parse("Probe", []byte(`{"childList":[{"childList":[{}]}]}`))
	expectError(t, err, "WIRE_JSON_LIMIT", "", "")
	cycle := core.Record{}
	cycle["childList"] = core.ValList([]core.Value{core.ValObject(cycle)})
	_, err = m.Serialize("Probe", cycle)
	expectError(t, err, "WIRE_ENTITY_CYCLE", "child_list[0]", "/childList/0")
	doc := &document{name: "Probe", fields: core.Record{}}
	doc.fields["childList"] = core.ValList([]core.Value{{V: doc}})
	_, err = m.Serialize("Probe", doc)
	expectError(t, err, "WIRE_ENTITY_CYCLE", "child_list[0]", "/childList/0")
	var nilDocument *document
	if _, err := m.Serialize("Probe", nilDocument); err == nil {
		t.Fatal("typed nil document accepted")
	}
	if _, err := m.Serialize("Probe", &document{name: "Wrong"}); err == nil {
		t.Fatal("wrong entity document accepted")
	}
	if _, err := m.Parse("Probe", []byte(`{}`), runtime.JsonFieldNamingProfile("unknown")); err == nil {
		t.Fatal("unknown profile accepted")
	}
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 20; j++ {
				record, err := m.Parse("Probe", []byte(`{"id":"1","referenceCode":null}`))
				if err != nil {
					t.Error(err)
					return
				}
				if _, err := m.Serialize("Probe", record); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	workers.Wait()
}

func TestEntityJSONInvalidNativeValuesNeverCoerceOrLeak(t *testing.T) {
	m := model(t, Limits{})
	for _, vector := range []struct {
		key   string
		value core.Value
	}{
		{"id", core.ValF64(1)}, {"small_number", core.ValI64(math.MaxInt32 + 1)}, {"large_number", core.ValText("secret-value")},
		{"amount", core.ValF64(0.1)}, {"amount", core.ValDecimal(decimal.New(1, math.MaxInt32))},
		{"rate", core.ValF64(0.1)}, {"rate", core.ValF64(math.Inf(1))}, {"ratio", core.ValF64(math.NaN())},
		{"clock_time", core.ValText("12:00:00")}, {"event_date", core.ValDate(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC))},
		{"event_time", core.Value{V: time.Date(2024, 1, 1, 0, 0, 0, 1, time.UTC)}},
		{"payload", core.ValJson(json.Number("null"))}, {"payload", core.ValJson(math.Inf(1))}, {"reference_code", core.ValText("\xff")},
	} {
		t.Run(vector.key, func(t *testing.T) {
			output, err := m.Serialize("Probe", core.Record{vector.key: vector.value})
			if output != nil || err == nil {
				t.Fatal("invalid native value serialized", string(output))
			}
			if strings.Contains(err.Error(), "secret-value") {
				t.Fatal("native value leaked")
			}
		})
	}
}

func TestEntityJSONRejectsInvalidAndAmbiguousSchema(t *testing.T) {
	for _, definitions := range [][]Schema{
		{}, {{EntityType: ""}}, {{EntityType: "Probe"}, {EntityType: "Probe"}},
		{{EntityType: "Probe", Fields: []Field{{CanonicalName: "a", Kind: Kind("unknown")}}}},
		{{EntityType: "Probe", Fields: []Field{{CanonicalName: "a", Kind: Text}, {CanonicalName: "a", Kind: Text}}}},
		{{EntityType: "Probe", Fields: []Field{{CanonicalName: "a_b", Kind: Text}, {CanonicalName: "aB", Kind: Text}}}},
		{{EntityType: "Probe", Fields: []Field{{CanonicalName: "target", Kind: Reference, Target: "Missing", RelationName: "detail"}}}},
		{{EntityType: "Probe", Fields: []Field{{CanonicalName: "target", Kind: Reference, Target: "Probe", RelationName: "detail"}, {CanonicalName: "detail", Kind: Text}}}},
	} {
		if _, err := NewModel(definitions, runtime.JsonFieldCamelCase, Limits{}); err == nil {
			t.Fatal("invalid schema accepted", definitions)
		}
	}
	if _, err := NewModel(definitions(), runtime.JsonFieldCamelCase, Limits{MaxDepth: 65}); err == nil {
		t.Fatal("unbounded depth accepted")
	}
}

func TestEntityJSONIdentityNullAndReferenceBoundaries(t *testing.T) {
	m := model(t, Limits{})
	for _, key := range []string{"id", "version"} {
		t.Run(key, func(t *testing.T) {
			_, err := m.Parse("Probe", []byte(`{"`+key+`":null}`))
			expectError(t, err, "WIRE_INVALID_TYPE", key, "/"+key)
			_, err = m.Serialize("Probe", core.Record{key: core.ValNull()})
			expectError(t, err, "WIRE_INVALID_TYPE", key, "/"+key)
		})
	}
	_, err := m.Serialize("Probe", core.Record{"rootRecordEntity": core.ValObject(core.Record{"display_name": core.ValText("no identity")})})
	expectError(t, err, "WIRE_REFERENCE_ID_REQUIRED", "root_record", "/rootRecord")
	var nilTarget *document
	_, err = m.Serialize("Probe", core.Record{"rootRecordEntity": {V: nilTarget}})
	expectError(t, err, "WIRE_REFERENCE_ID_REQUIRED", "root_record", "/rootRecord")
	_, err = m.Parse("Probe", []byte(`{"childList":[{"small_number":"secret-value"}]}`))
	var diagnostic *Error
	if !errors.As(err, &diagnostic) {
		t.Fatal(err)
	}
	payload, err := json.Marshal(diagnostic)
	if err != nil || strings.Contains(string(payload), "secret-value") || !strings.Contains(string(payload), `"instancePath":"/childList/0/small_number"`) || !strings.Contains(string(payload), `"kind":"index","index":0`) {
		t.Fatal("structured canonical location or submitted JSON path lost", err, string(payload))
	}
}

func TestEntityJSONNativeTypedNullHasNullSemantics(t *testing.T) {
	m := model(t, Limits{})
	for _, dataType := range []core.DataType{core.TypeBool, core.TypeI64, core.TypeText, core.TypeDecimal, core.TypeTime, core.TypeTimestamp} {
		t.Run(strconv.Itoa(int(dataType)), func(t *testing.T) {
			value := core.ValTypedNull(dataType)
			if !value.IsNull() || !value.TeaqlIsEmpty() || value.ToJsonValue() != nil {
				t.Fatal("typed NULL is treated as its type enum")
			}
			if _, ok := value.TryJson(); ok {
				t.Fatal("typed NULL became a JSON numeric value")
			}
			payload, err := m.Serialize("Probe", core.Record{"reference_code": value})
			if err != nil || string(payload) != `{"referenceCode":null}` {
				t.Fatal("typed NULL escaped as data", err, string(payload))
			}
		})
	}
}

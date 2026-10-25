package scalars

// Application-owned probe: discover APIs through the retained entity/action and
// relation Assist; do not inspect or modify the generated implementation.
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/runtime/entityjson"
	lib "scalar-type-conformance-service-core-workspace/lib"
)

func TestGeneratedTypedEntityJSON(t *testing.T) {
	context, observer := configured(t, ".entity-json.sqlite")
	for _, profile := range []runtime.JsonFieldNamingProfile{runtime.JsonFieldCamelCase, runtime.JsonFieldSnakeCase, runtime.JsonFieldPascalCase} {
		t.Run(string(profile), func(t *testing.T) {
			created, err := prepared(context).AuditAs("create typed JSON boundary fixture").Save(context)
			if err != nil {
				t.Fatal(err)
			}
			id, _ := lib.E.ScalarTypeProbe(created).Id().Eval()
			rows, err := lib.Q.ScalarTypeProbes().WithIdIs(id).
				SelectTypePlatformWith(lib.Q.TypePlatforms().Limit(1)).Limit(1).
				Comment("load typed JSON boundary graph").Purpose("retain exact generated scalars and a real forward entity").ExecuteForList(context)
			if err != nil || len(rows.Data) != 1 {
				t.Fatal("load JSON boundary", err)
			}
			loaded := rows.Data[0]
			observer.reset()
			payload, err := loaded.ToJSON(profile)
			if err != nil {
				t.Fatal("serialize loaded typed entity", err)
			}
			decoded, err := lib.ScalarTypeProbeFromJSON(payload, profile)
			if err != nil {
				t.Fatal("decode actual generated entity", err)
			}
			if decoded.IsNew() || len(decoded.EntityRoot().Change(decoded.EntityKey())) != 0 {
				t.Fatal("decoding created insert/update intent")
			}
			e := lib.E.ScalarTypeProbe(decoded)
			expect(t, "json_id", id, e.Id().TryEval)
			expect(t, "json_version", int64(1), e.Version().TryEval)
			expect(t, "json_required_string", "Portable scalar boundaries", e.RequiredString().TryEval)
			expect(t, "json_optional_string", pointer(""), e.OptionalString().TryEval)
			expect(t, "json_large_text", pointer("Unicode scalar fixture: 精确数值"), e.LargeText().TryEval)
			expect(t, "json_boolean", false, e.RequiredBoolean().TryEval)
			expect(t, "json_optional_boolean", pointer(true), e.OptionalBoolean().TryEval)
			expect(t, "json_integer", int64(math.MaxInt32), e.RequiredInteger().TryEval)
			expect(t, "json_optional_integer", pointer(int64(math.MinInt32)), e.OptionalInteger().TryEval)
			expect(t, "json_long", int64(math.MaxInt64), e.RequiredLong().TryEval)
			expect(t, "json_optional_long", pointer(int64(math.MinInt64)), e.OptionalLong().TryEval)
			expect(t, "json_decimal", decimal.RequireFromString("123456789012.3456789"), e.RequiredDecimal().TryEval)
			expect(t, "json_optional_decimal", pointer(decimal.RequireFromString("3.50")), e.OptionalDecimal().TryEval)
			expect(t, "json_float", float32(1.5), e.RequiredFloat().TryEval)
			expect(t, "json_optional_float", pointer(float32(-3.5)), e.OptionalFloat().TryEval)
			expect(t, "json_double", 2.25, e.RequiredDouble().TryEval)
			expect(t, "json_optional_double", pointer(-0.125), e.OptionalDouble().TryEval)
			expect(t, "json_date", time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), e.RequiredDate().TryEval)
			expect(t, "json_optional_date", pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), e.OptionalDate().TryEval)
			expect(t, "json_time", core.TimeOfDay{}, e.RequiredTime().TryEval)
			expect(t, "json_optional_time", pointer(fixtureClock("23:59:59.123456789")), e.OptionalTime().TryEval)
			expect(t, "json_datetime", time.Date(2024, 2, 29, 12, 34, 56, 789000000, time.UTC), e.RequiredDatetime().TryEval)
			expect(t, "json_optional_datetime", pointer(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), e.OptionalDatetime().TryEval)
			expect(t, "json_fk", uint64(1), e.TypePlatformId().TryEval)
			expect(t, "json_target_id", uint64(1), e.TypePlatform().Id().TryEval)
			if _, present, err := e.TypePlatform().Name().TryEval(); err != nil || !present {
				t.Fatal("forward entity detail missing", err)
			}
			var wire map[string]json.RawMessage
			if err := json.Unmarshal(payload, &wire); err != nil {
				t.Fatal(err)
			}
			if string(wire[profile.Render("required_long")]) != `"9223372036854775807"` ||
				string(wire[profile.Render("required_integer")]) != "2147483647" ||
				string(wire[profile.Render("optional_time")]) != `"23:59:59.123456789"` {
				t.Fatal("wire naming/type/precision loss", string(payload))
			}
			standard, err := json.Marshal(loaded)
			if err != nil {
				t.Fatal(err)
			}
			legacy, err := json.Marshal(loaded.IntoJson())
			if err != nil || !bytes.Equal(legacy, standard) {
				t.Fatal("legacy JSON adapter bypassed typed codec", err)
			}
			if observer.snapshot() != (callCounts{}) {
				t.Fatal("JSON or E performed provider I/O")
			}

			// Explicit null is loaded, omitted fields still fail through E. A
			// scalar FK remains a typed target with known ID and missing details.
			partialWire := map[string]any{profile.Render("id"): strconv.FormatUint(id, 10), profile.Render("version"): "1",
				profile.Render("required_string"): "partial", profile.Render("optional_time"): nil, profile.Render("type_platform"): "1"}
			partialPayload, _ := json.Marshal(partialWire)
			partial, err := lib.ScalarTypeProbeFromJSON(partialPayload, profile)
			if err != nil {
				t.Fatal(err)
			}
			pe := lib.E.ScalarTypeProbe(partial)
			if value, present, err := pe.OptionalTime().TryEval(); err != nil || present || value != nil {
				t.Fatal("explicit null lost", err)
			}
			if _, _, err := pe.RequiredDecimal().TryEval(); err == nil {
				t.Fatal("omission was not NotLoaded")
			}
			expect(t, "identity_only_target", uint64(1), pe.TypePlatform().Id().TryEval)
			if _, _, err := pe.TypePlatform().Name().TryEval(); err == nil {
				t.Fatal("identity-only target invented a name")
			}
			partial.UpdateRequiredString("reject incomplete decoded mutation")
			if _, err := partial.AuditAs("reject partial decoded entity").Save(context); err == nil {
				t.Fatal("incomplete decoded mutation succeeded")
			}
			if observer.snapshot() != (callCounts{}) {
				t.Fatal("partial decoded save entered the provider")
			}

			badInputs := []struct {
				name, code, field string
				value             any
			}{
				{"long_number", "WIRE_INVALID_TYPE", "required_long", float64(2)},
				{"integer_overflow", "WIRE_INVALID_TYPE", "required_integer", int64(math.MaxInt32) + 1},
				{"decimal_number", "WIRE_INVALID_TYPE", "required_decimal", 0.1},
				{"clock_zone", "WIRE_INVALID_TYPE", "optional_time", "12:00:00Z"},
				{"datetime_precision", "WIRE_TEMPORAL_PRECISION", "required_datetime", "2024-01-01T00:00:00.000001Z"},
				{"unknown", "WIRE_UNKNOWN_FIELD", "unknown_field", "secret-value"},
			}
			for _, bad := range badInputs {
				copyWire := make(map[string]json.RawMessage)
				for k, v := range wire {
					copyWire[k] = v
				}
				copyWire[profile.Render(bad.field)], _ = json.Marshal(bad.value)
				invalid, _ := json.Marshal(copyWire)
				value, err := lib.ScalarTypeProbeFromJSON(invalid, profile)
				var diagnostic *entityjson.Error
				if value != nil || !errors.As(err, &diagnostic) || diagnostic.Code != bad.code || diagnostic.InstancePath != "/"+profile.Render(bad.field) {
					t.Fatal("invalid JSON boundary", bad.name, err)
				}
			}
			duplicate := append(append([]byte{}, payload[:len(payload)-1]...), []byte(`,"required_string":"duplicate"}`)...)
			if _, err := lib.ScalarTypeProbeFromJSON(duplicate, profile); err == nil {
				t.Fatal("duplicate canonical alias accepted")
			}
			if observer.snapshot() != (callCounts{}) {
				t.Fatal("invalid JSON invoked provider operations")
			}

			// Two decodes are independent roots, including optimistic versions.
			other, err := lib.ScalarTypeProbeFromJSON(payload, profile)
			if err != nil {
				t.Fatal(err)
			}
			if other.EntityRoot() == decoded.EntityRoot() || other.EntityRoot() == loaded.EntityRoot() {
				t.Fatal("decoded roots share mutation ownership")
			}
			decoded.UpdateRequiredString("JSON left " + string(profile))
			other.UpdateRequiredString("JSON stale right " + string(profile))
			if _, err := decoded.AuditAs("persist decoded complete entity").Save(context); err != nil {
				t.Fatal("save decoded complete entity", err)
			}
			if _, err := other.AuditAs("reject stale independently decoded entity").Save(context); err == nil {
				t.Fatal("independent stale version accepted")
			}
			checked := reload(t, context, id)
			expect(t, "saved_decoded_value", "JSON left "+string(profile), lib.E.ScalarTypeProbe(checked).RequiredString().TryEval)
			expect(t, "saved_decoded_version", int64(2), lib.E.ScalarTypeProbe(checked).Version().TryEval)
			if len(other.EntityRoot().Change(other.EntityKey())) == 0 {
				t.Fatal("failed stale save discarded pending mutation")
			}

			// Real generated reverse relation, not a hand-written graph DTO.
			parents, err := lib.Q.TypePlatforms().WithIdIs(1).
				SelectScalarTypeProbeListWith(lib.Q.ScalarTypeProbes().WithIdIs(id).Limit(1)).Limit(1).
				Comment("load typed JSON reverse graph").Purpose("verify decoded lists retain real generated children").ExecuteForList(context)
			if err != nil || len(parents.Data) != 1 {
				t.Fatal("load reverse JSON graph", err)
			}
			observer.reset()
			graphPayload, err := parents.Data[0].ToJSON(profile)
			if err != nil {
				t.Fatal(err)
			}
			graph, err := lib.TypePlatformFromJSON(graphPayload, profile)
			if err != nil {
				t.Fatal(err)
			}
			expect(t, "decoded_list_size", 1, lib.E.TypePlatform(graph).ScalarTypeProbeList().Size().TryEval)
			expect(t, "decoded_list_first_id", id, lib.E.TypePlatform(graph).ScalarTypeProbeList().First().Id().TryEval)
			if observer.snapshot() != (callCounts{}) {
				t.Fatal("reverse graph JSON performed I/O")
			}
			if !t.Failed() {
				fmt.Printf("GO_ENTITY_JSON profile=%s fields=21 exact=true null_not_loaded=true forward=true reverse=true independent_roots=true stale_rejected=true partial_zero_io=true json_zero_io=true id=%d\n", profile, id)
			}
		})
	}
}

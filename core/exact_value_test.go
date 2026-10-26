package core

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestExactValueBaselineEquality(t *testing.T) {
	clock, _ := ParseTimeOfDay("23:59:59.123456789")
	date := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name        string
		left, right any
		equal       bool
	}{
		{"null", nil, nil, true}, {"typed-null", TypeText, nil, true},
		{"text", "same", "same", true}, {"text-not-number", "1", int64(1), false},
		{"bool", true, true, true}, {"bool-not-number", true, int64(1), false},
		{"long-exact", int64(9007199254740993), int64(9007199254740993), true},
		{"long-not-rounded", int64(9007199254740993), float64(9007199254740993), false},
		{"typed-unsigned", uint64(7), int64(7), false},
		{"decimal", decimal.RequireFromString("3.50"), decimal.RequireFromString("3.50"), true},
		{"decimal-scale", decimal.RequireFromString("3.50"), decimal.RequireFromString("3.500"), false},
		{"decimal-not-double", decimal.RequireFromString("3.50"), 3.5, false},
		{"double", 2.25, 2.25, true}, {"signed-zero", 0.0, math.Copysign(0, -1), false},
		{"float32", float32(1.25), float32(1.25), true}, {"float-width", float32(1.25), 1.25, false},
		{"nan", math.NaN(), math.NaN(), false}, {"positive-infinity", math.Inf(1), math.Inf(1), false},
		{"negative-infinity", math.Inf(-1), math.Inf(-1), false},
		{"timestamp", Timestamp(123), Timestamp(123), true}, {"timestamp-not-long", Timestamp(123), int64(123), false},
		{"wall-clock", clock, clock, true}, {"date", date, date, true},
		{"bytes", []byte{1, 2}, []byte{1, 2}, true}, {"nil-not-empty-bytes", []byte(nil), []byte{}, false},
		{"json", map[string]any{"items": []any{json.Number("1.00"), true}}, map[string]any{"items": []any{json.Number("1.00"), true}}, true},
		{"json-number-scale", json.Number("1.00"), json.Number("1.0"), false},
		{"record", Record{"n": ValI64(1)}, Record{"n": ValI64(1)}, true},
		{"value-list", []Value{ValText("a")}, []Value{ValText("a")}, true},
		{"missing-json-key", map[string]any{}, map[string]any{"n": nil}, false},
		{"unknown-pointer", new(int), new(int), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if ExactValueEqual(Value{V: test.left}, Value{V: test.right}) != test.equal || ExactValueEqual(Value{V: test.right}, Value{V: test.left}) != test.equal {
				t.Fatal("incorrect exact equality")
			}
		})
	}
}

func TestOwnedLoadedScalarBaselineIdentityAndUnknowns(t *testing.T) {
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	values := Record{"phone": ValText("5550800"), "note": ValNull()}
	snapshot := NewOwnedLoadedScalarSnapshot(key, 3, values, "phone", "note", "unknown")
	for _, test := range []struct {
		name    string
		key     EntityKey
		version int64
		count   int
	}{
		{"owner", key, 3, 2},
		{"other-entity", NewEntityKey("OrderItem", ValU64(7)), 3, 0},
		{"other-id", NewEntityKey("PurchaseOrder", ValU64(8)), 3, 0},
		{"signed-id", NewEntityKey("PurchaseOrder", ValI64(7)), 3, 0},
		{"version", key, 4, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := Record{"phone": ValText("5550800"), "note": ValNull(), "unknown": ValNull()}
			if len(snapshot.RestoredValues(test.key, test.version, changed)) != test.count {
				t.Fatal("wrong baseline owner or known-field decision")
			}
		})
	}
	if len(NewLoadedScalarSnapshot(values, "phone").RestoredValues(key, 3, values)) != 0 {
		t.Fatal("unowned privacy snapshot authorized restoration")
	}
	if len((LoadedScalarSnapshot{}).RestoredValues(key, 3, values)) != 0 {
		t.Fatal("zero snapshot authorized restoration")
	}
}

func TestOwnedLoadedScalarBaselineOwnsContainersAndRejectsCycles(t *testing.T) {
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	payload := map[string]any{"bytes": []byte{1, 2}, "items": []any{"original"}}
	cycle := map[string]any{}
	cycle["self"] = cycle
	large := make([]any, exactValueBudget+1)
	snapshot := NewOwnedLoadedScalarSnapshot(key, 3, Record{"payload": ValJson(payload), "cycle": ValJson(cycle), "large": ValJson(large)}, "payload", "cycle", "large")
	payload["bytes"].([]byte)[0] = 9
	payload["items"].([]any)[0] = "changed"
	if len(snapshot.RestoredValues(key, 3, Record{"payload": ValJson(payload)})) != 0 {
		t.Fatal("baseline aliases the loaded JSON")
	}
	if len(snapshot.RestoredValues(key, 3, Record{"payload": ValJson(map[string]any{"bytes": []byte{1, 2}, "items": []any{"original"}})})) != 1 {
		t.Fatal("safe JSON baseline lost")
	}
	if ExactValueEqual(ValJson(cycle), ValJson(cycle)) || ExactValueEqual(ValJson(large), ValJson(large)) {
		t.Fatal("unbounded/cyclic equality accepted")
	}
	if _, present := snapshot.Values()["cycle"]; present {
		t.Fatal("cyclic baseline was captured")
	}
}

func TestDiscardRestoredFieldsPreservesOtherLedgerState(t *testing.T) {
	root := NewEntityRoot()
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	child := NewEntityKey("OrderItem", ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", ValText("restored"))
	root.Set(key, "other", ValText("pending"))
	root.Set(child, "phone", ValText("independent"))
	observed := Record{"phone": ValText("restored")}
	root.Set(key, "phone", ValText("replaced-after-comparison"))
	if len(root.DiscardFields(key, observed)) != 0 {
		t.Fatal("concurrent replacement was erased")
	}
	root.Set(key, "phone", ValText("restored"))
	if len(root.DiscardFields(key, observed)) != 1 || len(root.Change(key)) != 1 || len(root.Change(child)) != 1 {
		t.Fatal("cancellation escaped its field/typed owner")
	}
	if version, known := root.OriginalVersion(key); !known || version != 3 {
		t.Fatal("optimistic version was lost")
	}
	root.MarkAsNew(child)
	if len(root.DiscardFields(child, root.Change(child))) != 0 || !root.IsNew(child) {
		t.Fatal("new lifecycle was cancelled")
	}
	root.MarkAsDeleted(key)
	root.Set(key, "phone", ValText("restored"))
	if len(root.DiscardFields(key, observed)) != 0 || !root.IsDeleted(key) {
		t.Fatal("deletion was cancelled")
	}
}

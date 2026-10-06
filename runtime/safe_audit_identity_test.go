package runtime

import (
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

// Reflection keeps this regression executable against the pre-fix API as well:
// a missing identity is a behavior failure, not merely a compile error.
func safeAuditTarget(t *testing.T, event *SafeAuditEvent) *core.Value {
	t.Helper()
	field := reflect.ValueOf(event).Elem().FieldByName("TargetID")
	if !field.IsValid() {
		t.Fatal("safe audit omits independent target identity")
	}
	value, ok := field.Interface().(*core.Value)
	if !ok {
		t.Fatal("safe audit target has the wrong type")
	}
	return value
}

func TestSafeAuditTargetIdentityIndependentOfChangedFields(t *testing.T) {
	id := core.ValU64(1001)
	rootID := uint64(100)
	update := Updated("Order Item", core.Record{"name": core.ValText("new private name")})
	update.TargetID = &id
	for name, event := range map[string]*RawAuditEvent{
		"created":   Created("Order Item", core.Record{"id": id}),
		"updated":   update,
		"deleted":   Deleted("Order Item", id, nil),
		"recovered": Recovered("Order Item", id, 1),
	} {
		t.Run(name, func(t *testing.T) {
			event.TraceChain = []*core.TraceNode{{Kind: "auditReason", Name: "Customer Order", EntityId: &rootID, Comment: "change item 1001"}}
			safe := event.BuildSafeEvent([]string{"name"}, nil)
			target := safeAuditTarget(t, safe)
			if target == nil || !reflect.DeepEqual(*target, id) || safe.Entity != "Order Item" {
				t.Fatalf("wrong typed target identity: entity=%s target=%v", safe.Entity, target)
			}
			if safe.TraceChain[0].Comment != "change item [REDACTED]" || event.TraceChain[0].Comment != "change item 1001" {
				t.Fatal("identity metadata changed the safe-intent policy or caller trace")
			}
			if name != "created" {
				for _, field := range safe.Fields {
					if field.Name == "id" {
						t.Fatal("audit identity was fabricated as a changed field")
					}
				}
			}
			*target = core.ValU64(999)
			original := safeAuditTarget(t, event.BuildSafeEvent([]string{"name"}, nil))
			if original == nil || !reflect.DeepEqual(*original, id) {
				t.Fatal("safe consumer mutation changed the raw event identity")
			}
		})
	}
}

func TestSafeAuditSchemaHasNoInventedTarget(t *testing.T) {
	for _, event := range []*RawAuditEvent{SchemaCreated("Order Item", "order_item_data", 4), SchemaVerified("Order Item", "order_item_data", 4)} {
		if safeAuditTarget(t, event.BuildSafeEvent(nil, nil)) != nil {
			t.Fatal("schema event invented an entity target identity")
		}
	}
}

func TestSafeAuditExplicitTargetWinsWithoutAliasing(t *testing.T) {
	id := core.ValU64(1001)
	event := Created("Order Item", core.Record{"id": core.ValU64(99)})
	event.TargetID = &id
	first := safeAuditTarget(t, event.BuildSafeEvent(nil, nil))
	if first == nil || !reflect.DeepEqual(*first, id) || first == &id {
		t.Fatal("explicit authoritative target was lost or aliased")
	}
	id = core.ValU64(1002)
	if actual, ok := first.TryU64(); !ok || actual != 1001 {
		t.Fatal("captured safe identity changed with the raw caller")
	}
}

package securedocumentreference

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"reflect"
	"testing"
	"time"
)

func documentFixture(t *testing.T) (*ContextBoundReferenceRuntime, TrustedReferencePrincipal, DocumentSnapshot) {
	t.Helper()
	keys, e := NewStaticReferenceKeyProvider("k1", ReferenceKey{"k1", bytes.Repeat([]byte{0x42}, 32)})
	if e != nil {
		t.Fatal(e)
	}
	r, e := NewContextBoundReferenceRuntime("order-service", "test", keys, func(TrustedReferencePrincipal, ReferenceDocumentScope, ReferenceIdentity) error { return nil })
	if e != nil {
		t.Fatal(e)
	}
	r.WithClock(func() time.Time { return time.Unix(1800000000, 0).UTC() }).WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 12)))
	principal := TrustedReferencePrincipal{"oidc", "alice", "Merchant", 7}
	scope := ReferenceDocumentScope{"doc-1", "edit-order", "Order", 10, 4}
	snapshot := DocumentSnapshot{"model-v1", "ORD-X1", scope, ReferenceIdentity{"Order", 10, 4}, []DocumentEntitySnapshot{{"item:0", ReferenceIdentity{"OrderItem", 20, 2}, []string{"product_code", "quantity"}, []string{"quantity"}, []string{"UPDATE"}}}, []DocumentRelationSnapshot{{"items", []string{"item:0"}, "COMPLETE", true, []string{"CREATE_CHILD", "REMOVE_CHILD"}}}}
	return r, principal, snapshot
}

func TestRustTQD1VectorAndScope(t *testing.T) {
	r, principal, snapshot := documentFixture(t)
	token, e := r.IssueDocumentSnapshot(principal, snapshot, 600*time.Second)
	if e != nil {
		t.Fatal(e)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(token))); got != "69fee1ec01cd4a5b69dd3f795aeaffb3a3546e50d48e25cfc9b8f169a527b267" {
		t.Fatalf("Rust tqd1 hash %s", got)
	}
	verified, e := r.ConsumeDocumentSnapshot(principal, token, "Order", "edit-order")
	if e != nil || verified.Document.BusinessID != "ORD-X1" {
		t.Fatalf("verify: %#v %v", verified, e)
	}
	bob := principal
	bob.Subject = "bob"
	if _, e = r.ConsumeDocumentSnapshot(bob, token, "Order", "edit-order"); code(e) != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("actor: %v", e)
	}
	if _, e = r.ConsumeDocumentSnapshot(principal, token, "Order", "read-order"); code(e) != "ROUND_TRIP_REFERENCE_INVALID" {
		t.Fatalf("purpose: %v", e)
	}
	r.WithClock(func() time.Time { return time.Unix(1800000600, 0).UTC() })
	if _, e = r.ConsumeDocumentSnapshot(principal, token, "Order", "edit-order"); code(e) != "ROUND_TRIP_REFERENCE_EXPIRED" {
		t.Fatalf("expiry: %v", e)
	}
}

func TestJavaRustTQD2TQS1Vectors(t *testing.T) {
	r, principal, snapshot := documentFixture(t)
	snapshot.Entities[0].LoadedFields = []string{"attachment", "product_code", "quantity"}
	snapshot.Entities[0].WritableFields = []string{"attachment", "quantity"}
	oldKey := "blob_A"
	document := BoundDocumentSnapshot{snapshot, []DocumentBlobBindingSnapshot{{"item:0", "attachment", "OrderItem", &oldKey, ^uint64(0), true, true}}}
	token, err := r.IssueBoundDocumentSnapshot(principal, document, 300*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(token))); got != "e7c81ca089b0a3e0585faefcc275aeb44b456d666d9a4a17949c92b1891089d7" {
		t.Fatalf("tqd2 SHA %s", got)
	}
	verified, err := r.ConsumeBoundDocumentSnapshot(principal, token, "Order", "edit-order")
	if err != nil || verified.Document.Bindings[0].BindingVersion != ^uint64(0) {
		t.Fatalf("bound verify: %#v %v", verified, err)
	}
	newBlob, err := core.ParseBlobRef("blob_B")
	if err != nil {
		t.Fatal(err)
	}
	target := BlobSelectionTarget{snapshot.Scope, "OrderItem", "attachment", "EXISTING_ROW", "item:0", "command-1"}
	claim := BlobSelectionClaim{target, newBlob}
	r.WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 12)))
	selection, err := r.IssueBlobSelection(principal, token, claim, 120*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(selection))); got != "31303e3e251dd0e8963bf2a4d8897d1eccde63a485aaf287f88f29d40757ba96" {
		t.Fatalf("tqs1 SHA %s", got)
	}
	selected, err := r.ConsumeBlobSelection(principal, token, selection, target)
	if err != nil || selected.Document.Reference.Key() != "blob_B" {
		t.Fatalf("selection verify: %#v %v", selected, err)
	}
	wrong := target
	wrong.CommandID = "other-command"
	if _, err := r.ConsumeBlobSelection(principal, token, selection, wrong); code(err) != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("wrong command: %v", err)
	}
	bob := principal
	bob.Subject = "bob"
	if _, err := r.ConsumeBlobSelection(bob, token, selection, target); code(err) != "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH" {
		t.Fatalf("wrong actor: %v", err)
	}
}

func surfaceDocumentFixture(t *testing.T) (*ContextBoundReferenceRuntime, TrustedReferencePrincipal, BoundDocumentSnapshot) {
	r, actor, snapshot := documentFixture(t)
	snapshot.Entities[0].LoadedFields = []string{"attachment", "product_code", "quantity"}
	snapshot.Entities[0].WritableFields = []string{"attachment", "quantity"}
	key := "blob_A"
	return r, actor, BoundDocumentSnapshot{snapshot, []DocumentBlobBindingSnapshot{{"item:0", "attachment", "OrderItem", &key, 4, true, true}}}
}

func TestIssuedProjectionRejectsInconsistentSnapshot(t *testing.T) {
	for _, scenario := range []string{"writable-unloaded", "duplicate-row", "orphan-relation", "duplicate-relation-row", "readonly-relation-mutation", "binding-unloaded", "binding-readonly", "binding-foreign-entity", "duplicate-binding"} {
		t.Run(scenario, func(t *testing.T) {
			r, actor, d := surfaceDocumentFixture(t)
			row, relation, binding := d.Snapshot.Entities[0], d.Snapshot.Relations[0], d.Bindings[0]
			switch scenario {
			case "writable-unloaded":
				row.WritableFields = []string{"hidden"}
			case "orphan-relation":
				relation.RowKeys = []string{"absent"}
			case "duplicate-relation-row":
				relation.RowKeys = []string{"item:0", "item:0"}
			case "readonly-relation-mutation":
				relation.Writable = false
			case "binding-unloaded":
				row.LoadedFields = []string{"quantity"}
				row.WritableFields = []string{"quantity"}
				binding.Writable = false
			case "binding-readonly":
				row.WritableFields = []string{"quantity"}
			case "binding-foreign-entity":
				binding.EntityType = "OtherItem"
			}
			d.Snapshot.Entities = []DocumentEntitySnapshot{row}
			if scenario == "duplicate-row" {
				d.Snapshot.Entities = append(d.Snapshot.Entities, row)
			}
			d.Snapshot.Relations = []DocumentRelationSnapshot{relation}
			d.Bindings = []DocumentBlobBindingSnapshot{binding}
			if scenario == "duplicate-binding" {
				d.Bindings = append(d.Bindings, binding)
			}
			if _, err := r.IssueBoundDocumentSnapshot(actor, d, time.Minute); code(err) != "DOCUMENT_PROJECTION_VIOLATION" {
				t.Fatalf("inconsistent projection accepted: %v", err)
			}
		})
	}
}

func TestSelectionCannotExpandIssuedSurface(t *testing.T) {
	for _, scenario := range []string{"readonly-existing", "hidden-existing", "partial-child", "readonly-child", "no-create-child"} {
		t.Run(scenario, func(t *testing.T) {
			r, actor, d := surfaceDocumentFixture(t)
			target := BlobSelectionTarget{d.Snapshot.Scope, "OrderItem", "attachment", "EXISTING_ROW", "item:0", "command-1"}
			switch scenario {
			case "readonly-existing":
				d.Snapshot.Entities[0].WritableFields = []string{"quantity"}
				d.Bindings[0].Writable = false
			case "hidden-existing":
				target.CanonicalField = "hidden_attachment"
			default:
				target.Kind = "NEW_CLIENT_KEY"
				target.TargetKey = "client-hidden"
				relation := &d.Snapshot.Relations[0]
				switch scenario {
				case "partial-child":
					relation.Completeness = "PARTIAL"
				case "readonly-child":
					relation.Writable = false
					relation.AllowedMutations = []string{}
				case "no-create-child":
					relation.AllowedMutations = []string{"REMOVE_CHILD"}
				}
			}
			token, err := r.IssueBoundDocumentSnapshot(actor, d, time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			verified, err := r.ConsumeBoundDocumentSnapshot(actor, token, "Order", "edit-order")
			if err != nil || !reflect.DeepEqual(verified.Document, d) {
				t.Fatal("issued metadata was not preserved", err)
			}
			ref, err := core.ParseBlobRef("blob_B")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = r.IssueBlobSelection(actor, token, BlobSelectionClaim{target, ref}, 30*time.Second); code(err) != "DOCUMENT_PROJECTION_VIOLATION" {
				t.Fatalf("selection expanded issued surface: %v", err)
			}
		})
	}
}

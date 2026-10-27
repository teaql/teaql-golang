package securedocumentreference

import (
	"bytes"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type testAssembly struct {
	principal TrustedReferencePrincipal
	document  BoundDocumentSnapshot
}

func (a testAssembly) Principal(*runtime.UserContext) (TrustedReferencePrincipal, error) {
	return a.principal, nil
}
func (a testAssembly) Snapshot(*runtime.UserContext, runtime.AuthorizedResponseGraph) (PreparedBoundDocument, error) {
	return PreparedBoundDocument{a.document, 300 * time.Second}, nil
}

type testGraphAuthorizer struct{ slot runtime.AuthorizedBlobSlot }

func (a testGraphAuthorizer) Authorize(*runtime.UserContext, any, runtime.ViewOptions) (runtime.AuthorizedResponseGraph, error) {
	return runtime.AuthorizedResponseGraph{Presentation: map[string]any{"attachment": a.slot.Reference}, BlobSlots: []runtime.AuthorizedBlobSlot{a.slot}}, nil
}

type testBlobProjector struct{}

func (testBlobProjector) ProjectBatch(*runtime.UserContext, runtime.AuthorizedResponseGraph, time.Time) ([]runtime.ProjectedBlobURL, error) {
	return []runtime.ProjectedBlobURL{{URL: "https://files.example.invalid/document", ExpiresAt: time.Now().Add(time.Minute)}}, nil
}

func TestModulePreparesBoundAuthorizedView(t *testing.T) {
	r, principal, snapshot := documentFixture(t)
	r.WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 12)))
	snapshot.Entities[0].LoadedFields = []string{"attachment", "product_code", "quantity"}
	snapshot.Entities[0].WritableFields = []string{"attachment", "quantity"}
	key := "blob_A"
	document := BoundDocumentSnapshot{snapshot, []DocumentBlobBindingSnapshot{{"item:0", "attachment", "OrderItem", &key, 4, true, true}}}
	ref, e := core.ParseBlobRef(key)
	if e != nil {
		t.Fatal(e)
	}
	slot := runtime.AuthorizedBlobSlot{Path: []any{"attachment"}, Entity: "OrderItem", EntityID: "20", Field: "attachment", BindingVersion: 4, Reference: ref}
	context := runtime.NewUserContext()
	module := &SecureDocumentReferenceModule{Protocol: r}
	if e = module.Install(context, testAssembly{principal, document}); e != nil {
		t.Fatal(e)
	}
	if e = module.Ready(context); e != nil {
		t.Fatal(e)
	}
	context.InsertResource("responseGraphAuthorizer", testGraphAuthorizer{slot})
	context.InsertResource("blobResponseContributor", testBlobProjector{})
	view, e := (runtime.ViewPreparationService{}).Prepare(context, nil, runtime.ViewOptions{Mode: "governed", BusinessOutcome: "committed"})
	if e != nil {
		t.Fatal(e)
	}
	if view.Presentation.(map[string]any)["attachment"] != "https://files.example.invalid/document" {
		t.Fatal("missing projected URL")
	}
	envelope := view.Document.(map[string]any)
	if envelope["businessId"] != "ORD-X1" {
		t.Fatal("missing business ID")
	}
	if _, e = r.ConsumeBoundDocumentSnapshot(principal, envelope["documentToken"].(string), "Order", "edit-order"); e != nil {
		t.Fatal(e)
	}
	if e = module.Install(context, testAssembly{principal, document}); code(e) != "DOCUMENT_CONFIGURATION" {
		t.Fatalf("duplicate module: %v", e)
	}
	slot.BindingVersion = 5
	context.InsertResource("responseGraphAuthorizer", testGraphAuthorizer{slot})
	_, e = (runtime.ViewPreparationService{}).Prepare(context, nil, runtime.ViewOptions{Mode: "governed", BusinessOutcome: "committed"})
	if failure, ok := e.(*runtime.ViewPreparationError); !ok || failure.Code != "DOCUMENT_PROJECTION_VIOLATION" || failure.BusinessState != "committed" {
		t.Fatalf("binding mismatch: %v", e)
	}
}

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

type surfaceAssembly struct{ testAssembly }

func (a surfaceAssembly) Snapshot(*runtime.UserContext, runtime.AuthorizedResponseGraph) (PreparedBoundDocument, error) {
	return PreparedBoundDocument{a.document, time.Minute}, nil
}

type surfaceProjector struct {
	t        *testing.T
	deadline time.Time
	late     bool
	keys     []string
}

func (p *surfaceProjector) ProjectBatch(_ *runtime.UserContext, graph runtime.AuthorizedResponseGraph, ceiling time.Time) ([]runtime.ProjectedBlobURL, error) {
	if !ceiling.Equal(p.deadline) {
		p.t.Fatal("actual signed document deadline was not supplied to Blob projection")
	}
	for _, slot := range graph.BlobSlots {
		p.keys = append(p.keys, slot.Reference.Key())
	}
	expiry := p.deadline.Add(-5 * time.Second)
	if p.late {
		expiry = p.deadline.Add(time.Millisecond)
	}
	return []runtime.ProjectedBlobURL{{URL: "https://files.example.invalid/asset", ExpiresAt: expiry}}, nil
}
func TestGovernedViewHidesBlobAndBoundsURLByDocumentDeadline(t *testing.T) {
	r, actor, d := surfaceDocumentFixture(t)
	now := time.Now().UTC().Truncate(time.Second)
	deadline := now.Add(time.Minute)
	r.WithClock(func() time.Time { return now }).WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 48)))
	visible, _ := core.ParseBlobRef("blob_A")
	hidden, _ := core.ParseBlobRef("blob_hidden")
	source := map[string]any{"attachment": visible, "hidden_attachment": hidden}
	slot := runtime.AuthorizedBlobSlot{Path: []any{"attachment"}, Entity: "OrderItem", EntityID: "20", Field: "attachment", Scope: "tenant-a", BindingVersion: 4, Reference: visible}
	context := runtime.NewUserContext()
	module := &SecureDocumentReferenceModule{Protocol: r}
	if err := module.Install(context, surfaceAssembly{testAssembly{actor, d}}); err != nil {
		t.Fatal(err)
	}
	signer := &surfaceProjector{t: t, deadline: deadline}
	context.InsertResource("responseGraphAuthorizer", testGraphAuthorizer{slot})
	context.InsertResource("blobResponseContributor", signer)
	options := runtime.ViewOptions{Mode: "governed", BusinessOutcome: "committed", ExpiresAt: now.Add(2 * time.Minute)}
	view, err := (runtime.ViewPreparationService{}).Prepare(context, source, options)
	if err != nil {
		t.Fatal(err)
	}
	presentation := view.Presentation.(map[string]any)
	if len(presentation) != 1 || presentation["attachment"] != "https://files.example.invalid/asset" || !view.ExpiresAt.Equal(deadline.Add(-5*time.Second)) {
		t.Fatal("unexpected visible view or deadline")
	}
	if len(signer.keys) != 1 || signer.keys[0] != "blob_A" || source["hidden_attachment"] != hidden {
		t.Fatal("hidden Blob was projected or source mutated")
	}
	signer.late = true
	_, err = (runtime.ViewPreparationService{}).Prepare(context, source, options)
	failure, ok := err.(*runtime.ViewPreparationError)
	if !ok || failure.Code != "BLOB_SIGN_FAILED" || failure.BusinessState != "committed" || failure.Retry() != "read" {
		t.Fatalf("URL beyond document deadline accepted: %v", err)
	}
	if len(signer.keys) != 2 || signer.keys[1] != "blob_A" {
		t.Fatal("hidden Blob reached projection")
	}
}

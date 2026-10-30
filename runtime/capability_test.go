package runtime

import (
	"errors"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

type readyCapability struct {
	name    string
	failure error
}

func (r *readyCapability) Capability() string               { return r.name }
func (r *readyCapability) ValidateReady(*UserContext) error { return r.failure }

func TestReferenceCapabilityManifestAndGraphGate(t *testing.T) {
	left := NewRuntimeModule().RequireCapability("blob.v1")
	right := NewRuntimeModule().RequireCapability("secure_document.v1").RequireCapability("blob.v1")
	both := left.And(right)
	if !reflect.DeepEqual(both.RequiredCapabilities(), []string{"blob.v1", "secure_document.v1"}) {
		t.Fatal("capability union mismatch")
	}
	if !reflect.DeepEqual(left.RequiredCapabilities(), []string{"blob.v1"}) {
		t.Fatal("input module changed")
	}
	context := both.IntoContext()
	reason := "save document"
	intent, err := core.NewMutationIntent(&reason)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	err = context.ExecuteGraphSave(intent, func() error { called = true; return nil })
	if err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" || called {
		t.Fatalf("missing module entered graph work: %v", err)
	}
	if err := context.InstallRuntimeCapability(&readyCapability{name: "blob.v1"}); err != nil {
		t.Fatal(err)
	}
	if err := context.ValidateCapabilities(); err == nil || err.Error() != "DOCUMENT_CONFIGURATION" {
		t.Fatalf("secure module was not required: %v", err)
	}
	if err := context.InstallRuntimeCapability(&readyCapability{name: "secure_document.v1"}); err != nil {
		t.Fatal(err)
	}
	if err := context.ValidateCapabilities(); err != nil {
		t.Fatal(err)
	}
	if err := context.InstallRuntimeCapability(&readyCapability{name: "blob.v1"}); err == nil {
		t.Fatal("duplicate accepted")
	}

	failing := NewRuntimeModule().RequireCapability("blob.v1").IntoContext()
	if err := failing.InstallRuntimeCapability(&readyCapability{name: "blob.v1", failure: errors.New("private provider diagnostic")}); err != nil {
		t.Fatal(err)
	}
	if err := failing.ValidateCapabilities(); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
		t.Fatalf("unsafe readiness failure: %v", err)
	}
}

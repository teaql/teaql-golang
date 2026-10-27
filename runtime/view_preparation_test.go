package runtime

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type viewAuthorizerFunc func(*UserContext, any, ViewOptions) (AuthorizedResponseGraph, error)

func (f viewAuthorizerFunc) Authorize(c *UserContext, value any, options ViewOptions) (AuthorizedResponseGraph, error) {
	return f(c, value, options)
}

type secureViewFunc func(*UserContext, AuthorizedResponseGraph, ViewOptions) (PreparedSecureDocument, error)

func (f secureViewFunc) Prepare(c *UserContext, graph AuthorizedResponseGraph, options ViewOptions) (PreparedSecureDocument, error) {
	return f(c, graph, options)
}

type blobViewFunc func(*UserContext, AuthorizedResponseGraph, time.Time) ([]ProjectedBlobURL, error)

func (f blobViewFunc) ProjectBatch(c *UserContext, graph AuthorizedResponseGraph, expiry time.Time) ([]ProjectedBlobURL, error) {
	return f(c, graph, expiry)
}

func TestViewPreparationOrdersCurrentAuthorizationSecureDocumentAndBlobProjection(t *testing.T) {
	reference, err := core.ParseBlobRef("blob_asset_1")
	if err != nil {
		t.Fatal(err)
	}
	original := map[string]any{"document": reference, "nested": []any{map[string]any{"name": "untouched"}}}
	stages := make([]string, 0, 3)
	context := NewUserContext()
	context.InsertResource("responseGraphAuthorizer", viewAuthorizerFunc(func(_ *UserContext, _ any, _ ViewOptions) (AuthorizedResponseGraph, error) {
		stages = append(stages, "authorize")
		return AuthorizedResponseGraph{Presentation: original, BlobSlots: []AuthorizedBlobSlot{{
			Path: []any{"document"}, Entity: "Asset", EntityID: "1", Field: "document",
			Scope: "tenant-a", BindingVersion: 4, Reference: reference,
		}}}, nil
	}))
	ceiling := time.Now().Add(5 * time.Minute)
	context.InsertResource("secureDocumentContributor", secureViewFunc(func(_ *UserContext, graph AuthorizedResponseGraph, _ ViewOptions) (PreparedSecureDocument, error) {
		stages = append(stages, "secure")
		graph.Presentation.(map[string]any)["governed"] = true
		return PreparedSecureDocument{Graph: graph, Envelope: "sealed-document", ExpiresAt: ceiling}, nil
	}))
	context.InsertResource("blobResponseContributor", blobViewFunc(func(_ *UserContext, graph AuthorizedResponseGraph, expiry time.Time) ([]ProjectedBlobURL, error) {
		stages = append(stages, "blob")
		if !expiry.Equal(ceiling) || len(graph.BlobSlots) != 1 || graph.BlobSlots[0].Reference != reference {
			t.Fatalf("Blob stage lost authorized slot or governed ceiling")
		}
		return []ProjectedBlobURL{{URL: "https://files.example.invalid/read/asset", ExpiresAt: ceiling.Add(-time.Minute)}}, nil
	}))
	view, err := (ViewPreparationService{}).Prepare(context, original, ViewOptions{Mode: "governed", BusinessOutcome: "committed"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stages, []string{"authorize", "secure", "blob"}) {
		t.Fatalf("wrong preparation order: %v", stages)
	}
	if original["document"] != reference || original["governed"] != nil ||
		view.Presentation.(map[string]any)["document"] != "https://files.example.invalid/read/asset" ||
		view.Presentation.(map[string]any)["governed"] != true || view.Document != "sealed-document" ||
		view.CacheControl != "private, no-store" || !view.ExpiresAt.Equal(ceiling.Add(-time.Minute)) {
		t.Fatalf("view mutated source or lost governed result")
	}
}

func TestViewPreparationFailsClosedWithoutLeakingRawKeyOrRepeatingCommittedWork(t *testing.T) {
	reference, err := core.ParseBlobRef("blob_private_canary")
	if err != nil {
		t.Fatal(err)
	}
	context := NewUserContext()
	context.InsertResource("responseGraphAuthorizer", viewAuthorizerFunc(func(_ *UserContext, _ any, _ ViewOptions) (AuthorizedResponseGraph, error) {
		return AuthorizedResponseGraph{Presentation: map[string]any{"document": reference}}, nil
	}))
	service := ViewPreparationService{}
	_, err = service.Prepare(context, nil, ViewOptions{Mode: "ordinary", BusinessOutcome: "committed"})
	var failure *ViewPreparationError
	if !errors.As(err, &failure) || failure.Code != "BLOB_CAPABILITY_MISSING" ||
		failure.BusinessState != "committed" || failure.Retry() != "read" ||
		strings.Contains(failure.Error(), reference.Key()) {
		t.Fatalf("unexpected committed view failure: %v", err)
	}
	context.InsertResource("responseGraphAuthorizer", viewAuthorizerFunc(func(_ *UserContext, _ any, _ ViewOptions) (AuthorizedResponseGraph, error) {
		return AuthorizedResponseGraph{Presentation: map[string]any{"document": reference},
			BlobSlots: []AuthorizedBlobSlot{{Path: []any{"document"}, Reference: reference}}}, nil
	}))
	context.InsertResource("blobResponseContributor", blobViewFunc(func(_ *UserContext, _ AuthorizedResponseGraph, _ time.Time) ([]ProjectedBlobURL, error) {
		return nil, errors.New("provider exposed private URL https://secret.invalid/private?sig=canary")
	}))
	_, err = service.Prepare(context, nil, ViewOptions{Mode: "ordinary", BusinessOutcome: "committed"})
	if !errors.As(err, &failure) || failure.Code != "VIEW_PREPARATION_FAILED" ||
		strings.Contains(failure.Error(), "secret.invalid") {
		t.Fatalf("provider failure was not sanitized: %v", err)
	}
}

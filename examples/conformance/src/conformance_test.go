package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/teaql/teaql-golang/runtime"
	"runtime-example-conformance-service-core-workspace/lib"
)

func TestRetainedMinimumConformance(t *testing.T) {
	t.Setenv("RUNTIME_EXAMPLE_CONFORMANCE_SERVICE_CORE_DATABASE_URL", filepath.Join(t.TempDir(), "conformance.db"))
	context, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer context.GetResource("db").(interface{ Close() error }).Close()

	// 1. Schema changes are explicit and idempotent.
	if err := lib.EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	if err := lib.EnsureSchema(context); err != nil {
		t.Fatal(err)
	}

	platform := lib.Q.Platforms().Comment("create test root").Purpose("initialize conformance data").NewEntity(context)
	platform.UpdateName("Runtime Example")
	platform, err = platform.AuditAs("create test root").Save(context)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Checker rejects invalid input before persistence.
	invalid := lib.Q.WorkItems().Comment("create invalid work item").Purpose("exercise checker").NewEntity(context)
	invalid.UpdatePlatformId(platform.Id())
	if _, err = invalid.AuditAs("reject missing title").Save(context); err == nil {
		t.Fatalf("expected title checker error, got %v", err)
	}
	var checkError *runtime.RuntimeError
	if !errors.As(err, &checkError) || checkError.Type != "Check" || len(checkError.CheckResults) == 0 || checkError.CheckResults[0].Location != "title" {
		t.Fatalf("expected machine-readable title checker error, got %#v", err)
	}
	assertWorkItemCount(t, context, 0)

	// 3. Create returns the persisted, versioned strong type.
	item := lib.Q.WorkItems().Comment("create work item").Purpose("exercise create mutation").NewEntity(context)
	item.UpdateTitle("draft")
	item.UpdatePlatformId(platform.Id())
	item, err = item.AuditAs("create work item").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	if item.Id() == 0 || item.Version() != 1 {
		t.Fatalf("unexpected persisted identity/version: %d/%d", item.Id(), item.Version())
	}

	// 4. Q returns *core.SmartList[*work_item.WorkItem], proven by this generated API.
	items, err := lib.Q.WorkItems().OrderByIdAsc().Comment("load work items").Purpose("verify typed SmartList").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(items.Data) != 1 || items.Data[0].Title() != "draft" {
		t.Fatalf("unexpected Q result: %#v", items.Data)
	}

	// 5. E preserves loaded, null, and not-loaded states.
	title, present, evalErr := lib.E.WorkItem(items.Data[0]).Title().TryEval()
	if evalErr != nil || !present || title != "draft" {
		t.Fatalf("loaded title: %q %v %v", title, present, evalErr)
	}
	description, present, evalErr := lib.E.WorkItem(items.Data[0]).Description().TryEval()
	if evalErr != nil || present || description != nil {
		t.Fatalf("loaded null description: %#v %v %v", description, present, evalErr)
	}
	minimal, err := lib.Q.WorkItemsMinimal().WithIdIs(item.Id()).Comment("load minimal projection").Purpose("verify not-loaded state").ExecuteForOne(context)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, evalErr = lib.E.WorkItem(minimal).Title().TryEval(); evalErr == nil {
		t.Fatal("expected not-loaded title error")
	}

	// 6. Update returns the next persisted version.
	item.UpdateTitle("done")
	item, err = item.AuditAs("complete work item").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	if item.Version() != 2 || item.Title() != "done" {
		t.Fatalf("unexpected update result: version=%d title=%q", item.Version(), item.Title())
	}

	// 7. Delete is hidden from normal Q and visible through DeletedRowsOnly.
	item.MarkForDeletion()
	if _, err = item.AuditAs("delete work item").Save(context); err != nil {
		t.Fatal(err)
	}
	assertWorkItemCount(t, context, 0)
	deleted, err := lib.Q.WorkItems().DeletedRowsOnly().Comment("load deleted work items").Purpose("verify soft delete visibility").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted.Data) != 1 {
		t.Fatalf("expected one deleted item, got %d", len(deleted.Data))
	}
}

func assertWorkItemCount(t *testing.T, context *runtime.UserContext, expected int) {
	t.Helper()
	items, err := lib.Q.WorkItems().Comment("load visible work items").Purpose("verify persistence boundary").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(items.Data) != expected {
		t.Fatalf("expected %d visible items, got %d", expected, len(items.Data))
	}
}

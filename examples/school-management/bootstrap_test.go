package lib

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/runtime"
	"school-management-service-core-workspace/lib/school"
)

func TestSchoolBootstrapWithLocalRuntime(t *testing.T) {
	database := filepath.Join(t.TempDir(), "school.sqlite")
	t.Setenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL", database)
	context, err := ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSchema(context); err != nil {
		t.Fatal(err)
	}

	platforms, err := Q.Platforms().Comment("verify seeded root").Purpose("local runtime verification").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	constants, err := Q.SchoolTypes().OrderByIdAsc().Comment("verify seeded constants").Purpose("local runtime verification").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(platforms.Data) != 1 || platforms.Data[0].Id() != 1 {
		t.Fatalf("platforms=%v", platforms.Data)
	}
	if len(constants.Data) != 2 || constants.Data[0].Id() != 1001 || constants.Data[1].Id() != 1002 {
		t.Fatalf("constants=%v", constants.Data)
	}
	if constants.Data[0].Version() != 1 || constants.Data[1].Version() != 1 {
		t.Fatalf("idempotent versions=%d,%d", constants.Data[0].Version(), constants.Data[1].Version())
	}

	db := context.GetResource("db").(*sql.DB)
	var idFloor uint64
	if err := db.QueryRow("SELECT current_level FROM teaql_id_space WHERE type_name = ?", "School Type").Scan(&idFloor); err != nil {
		t.Fatal(err)
	}
	if idFloor < 1002 {
		t.Fatalf("constant ID floor=%d", idFloor)
	}
	if _, err := db.Exec("UPDATE platform_data SET name = 'Deployment Campus' WHERE id = 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE school_type_data SET name = 'Drifted Primary' WHERE id = 1001"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	preservedRoot, err := Q.Platforms().WithIdIs(1).Comment("verify preserved root").Purpose("local runtime verification").ExecuteForOne(context)
	if err != nil {
		t.Fatal(err)
	}
	changed, err := Q.SchoolTypes().WithIdIs(1001).Comment("verify constant reconciliation").Purpose("local runtime verification").ExecuteForOne(context)
	if err != nil {
		t.Fatal(err)
	}
	if preservedRoot.Name() != "Deployment Campus" {
		t.Fatalf("root was overwritten: %q", preservedRoot.Name())
	}
	if changed.Name() != "Primary" || changed.Version() != 2 {
		t.Fatalf("name=%q version=%d", changed.Name(), changed.Version())
	}

	fixture := Q.Schools().Comment("create School Query conformance fixture").Purpose("execute the shared School example").NewEntity(context)
	fixture.UpdatePlatformId(1).UpdateSchoolTypeToPrimary().UpdateName("Riverside Primary School")
	fixture.UpdateAddress("12 River Road, Springfield").UpdateEstablishedDate(time.Date(1995, 9, 1, 0, 0, 0, 0, time.UTC))
	fixture.UpdateStudentCapacity(800).UpdateActive(true)
	if _, err := fixture.AuditAs("create School Query conformance fixture").Save(context); err != nil {
		t.Fatal(err)
	}

	assertCount := func(label string, request *school.SchoolRequest, expected int) {
		t.Helper()
		result, err := request.Comment("Query parity: " + label).Purpose("Execute the shared School Query conformance case").ExecuteForList(context)
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		if len(result.Data) != expected {
			t.Fatalf("%s: expected %d, got %d", label, expected, len(result.Data))
		}
	}
	assertCount("string equality", Q.Schools().WithNameIs("Riverside Primary School"), 1)
	assertCount("string inequality", Q.Schools().WithNameIsNot("Another School"), 1)
	assertCount("string membership", Q.Schools().WithNameIn([]string{"Riverside Primary School", "Another School"}), 1)
	assertCount("negative membership", Q.Schools().WithNameNotIn([]string{"Another School"}), 1)
	assertCount("contains", Q.Schools().WithNameContaining("Primary"), 1)
	assertCount("negative contains", Q.Schools().WithNameNotContaining("Secondary"), 1)
	assertCount("starts with", Q.Schools().WithNameStartingWith("Riverside"), 1)
	assertCount("negative starts with", Q.Schools().WithNameNotStartingWith("Lakeside"), 1)
	assertCount("ends with", Q.Schools().WithNameEndingWith("School"), 1)
	assertCount("negative ends with", Q.Schools().WithNameNotEndingWith("Academy"), 1)
	assertCount("number range", Q.Schools().WithStudentCapacityBetween(700, 900), 1)
	assertCount("strict comparison", Q.Schools().WithStudentCapacityGreaterThan(799).WithStudentCapacityLessThan(801), 1)
	assertCount("date range", Q.Schools().WithEstablishedDateBetween(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC)), 1)
	assertCount("known", Q.Schools().WithAddressIsKnown(), 1)
	assertCount("unknown", Q.Schools().WithAddressIsUnknown(), 0)
	assertCount("boolean true", Q.Schools().WhichAreActive(), 1)
	assertCount("boolean false", Q.Schools().WhichAreNotActive(), 0)
	assertCount("constant relation", Q.Schools().WithSchoolTypeIsPrimary(), 1)
	related, err := Q.Schools().WithNameIs("Riverside Primary School").
		SelectPlatformWith(Q.PlatformsMinimal().SelectName()).
		SelectSchoolTypeWith(Q.SchoolTypesMinimal().SelectCode()).
		Comment("Query parity: typed forward relations").
		Purpose("Execute the shared School Query conformance case").ExecuteForOne(context)
	if err != nil {
		t.Fatal(err)
	}
	platformName, platformOK := E.School(related).Platform().Name().Eval()
	typeCode, typeOK := E.School(related).SchoolType().Code().Eval()
	if !platformOK || platformName != "Deployment Campus" || !typeOK || typeCode != "PRIMARY" {
		t.Fatalf("forward relations platform=%q/%t schoolType=%q/%t", platformName, platformOK, typeCode, typeOK)
	}
	projected, err := Q.Schools().SelectName().OrderByIdDesc().Comment("Query parity: projection and ordering").Purpose("Execute the shared School Query conformance case").ExecuteForList(context)
	if err != nil {
		t.Fatal(err)
	}
	if len(projected.Data) != 1 || projected.Data[0].Name() != "Riverside Primary School" {
		t.Fatal("projection/order query did not preserve typed School result")
	}

	if _, err := os.Stat(database); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratedBootstrapConvergesAcrossContexts(t *testing.T) {
	database := filepath.Join(t.TempDir(), "concurrent-school.sqlite")
	t.Setenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL", database)
	initial, err := ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureSchema(initial); err != nil {
		t.Fatal(err)
	}
	db := initial.GetResource("db").(*sql.DB)
	if _, err := db.Exec("DELETE FROM school_type_data"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("DELETE FROM platform_data"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	left, err := ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	right, err := ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	runTogether := func() {
		t.Helper()
		failures := make(chan error, 2)
		var wait sync.WaitGroup
		for _, context := range []*runtime.UserContext{left, right} {
			wait.Add(1)
			go func(current *runtime.UserContext) { defer wait.Done(); failures <- EnsureSchema(current) }(context)
		}
		wait.Wait()
		close(failures)
		for failure := range failures {
			if failure != nil {
				t.Fatal(failure)
			}
		}
	}
	runTogether()
	roots, err := Q.Platforms().Comment("read root").Purpose("verify concurrent bootstrap").ExecuteForList(left)
	if err != nil {
		t.Fatal(err)
	}
	constants, err := Q.SchoolTypes().Comment("read constants").Purpose("verify concurrent bootstrap").ExecuteForList(left)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots.Data) != 1 || len(constants.Data) != 2 {
		t.Fatal("concurrent bootstrap did not converge")
	}

	db = left.GetResource("db").(*sql.DB)
	if _, err := db.Exec("UPDATE school_type_data SET name = 'DRIFT' WHERE id = 1001"); err != nil {
		t.Fatal(err)
	}
	runTogether()
	reconciled, err := Q.SchoolTypes().WithIdIs(1001).Comment("read constant").Purpose("verify concurrent reconcile").ExecuteForOne(left)
	if err != nil {
		t.Fatal(err)
	}
	if reconciled.Name() != "Primary" || reconciled.Version() != 2 {
		t.Fatalf("name=%q version=%d", reconciled.Name(), reconciled.Version())
	}
}

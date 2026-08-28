package lib

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
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

	generatedInitialGraphs[0].Values["name"] = core.ValText("Primary School")
	if err := EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	changed, err := Q.SchoolTypes().WithIdIs(1001).Comment("verify constant reconciliation").Purpose("local runtime verification").ExecuteForOne(context)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Name() != "Primary School" || changed.Version() != 2 {
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
	if !platformOK || platformName != "Campus Learning Platform" || !typeOK || typeCode != "PRIMARY" {
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

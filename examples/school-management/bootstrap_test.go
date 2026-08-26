package lib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestSchoolBootstrapWithLocalRuntime(t *testing.T) {
	database := filepath.Join(t.TempDir(), "school.sqlite")
	t.Setenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL", database)
	context, err := ServiceRuntimeFromEnv()
	if err != nil { t.Fatal(err) }
	if err := EnsureSchema(context); err != nil { t.Fatal(err) }
	if err := EnsureSchema(context); err != nil { t.Fatal(err) }

	platforms, err := Q.Platforms().Comment("verify seeded root").Purpose("local runtime verification").ExecuteForList(context)
	if err != nil { t.Fatal(err) }
	constants, err := Q.SchoolTypes().OrderByIdAsc().Comment("verify seeded constants").Purpose("local runtime verification").ExecuteForList(context)
	if err != nil { t.Fatal(err) }
	if len(platforms.Data) != 1 || platforms.Data[0].Id() != 1 { t.Fatalf("platforms=%v", platforms.Data) }
	if len(constants.Data) != 2 || constants.Data[0].Id() != 1001 || constants.Data[1].Id() != 1002 { t.Fatalf("constants=%v", constants.Data) }
	if constants.Data[0].Version() != 1 || constants.Data[1].Version() != 1 { t.Fatalf("idempotent versions=%d,%d", constants.Data[0].Version(), constants.Data[1].Version()) }

	generatedInitialGraphs[0].Values["name"] = core.ValText("Primary School")
	if err := EnsureSchema(context); err != nil { t.Fatal(err) }
	changed, err := Q.SchoolTypes().WithIdIs(1001).Comment("verify constant reconciliation").Purpose("local runtime verification").ExecuteForOne(context)
	if err != nil { t.Fatal(err) }
	if changed.Name() != "Primary School" || changed.Version() != 2 { t.Fatalf("name=%q version=%d", changed.Name(), changed.Version()) }

	if _, err := os.Stat(database); err != nil { t.Fatal(err) }
}

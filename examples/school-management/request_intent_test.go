package lib

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"school-management-service-core-workspace/lib/school"
)

// App-owned regression: generated source is replaced only through generation.
// Disabling both SQL logs must not turn intent into optional logging metadata.
func TestGeneratedSchoolRequestIntentWithLoggingDisabled(t *testing.T) {
	t.Setenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL", filepath.Join(t.TempDir(), "intent.sqlite"))
	context, err := ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	context.DisableSqlLog()
	if context.QuerySqlLogEnabled() || context.MutationSqlLogEnabled() {
		t.Fatal("test did not disable logs")
	}
	if err = EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	if err = EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	assertRequired := func(err error, code string) {
		t.Helper()
		var required *core.RequestIntentError
		if !errors.As(err, &required) || required.Code != code {
			t.Fatalf("expected %s, got %v", code, err)
		}
	}
	_, err = Q.Schools().Comment("\u0085").Purpose("render schools").ExecuteForList(context)
	assertRequired(err, "REQUEST_COMMENT_REQUIRED")
	_, err = Q.Schools().Comment("load schools").Purpose("\u00a0").ExecuteForList(context)
	assertRequired(err, "QUERY_PURPOSE_REQUIRED")
	_, err = Q.Schools().Comment(" ").Purpose("render schools").ExecuteForPage(context, 0, 5)
	assertRequired(err, "REQUEST_COMMENT_REQUIRED")
	err = Q.Schools().Comment(" ").Purpose("render schools").ExecuteForStream(context, 1, func(*school.School) error { t.Fatal("invalid request yielded an entity"); return nil })
	assertRequired(err, "REQUEST_COMMENT_REQUIRED")
	_, err = school.NewSchool().Save(context)
	assertRequired(err, "REQUEST_COMMENT_REQUIRED")
	rows, err := Q.Schools().Comment("verify rejected write absence").Purpose("verify request intent gates").ExecuteForList(context)
	if err != nil || len(rows.Data) != 0 {
		t.Fatalf("invalid mutation reached the database: %v", err)
	}
	constants, err := Q.SchoolTypes().Comment("verify runtime-owned bootstrap intent").Purpose("verify constants remain available").ExecuteForList(context)
	if err != nil || len(constants.Data) != 2 {
		t.Fatalf("bootstrap did not use valid runtime intent: %v", err)
	}
	t.Log("PASS generated School request intent gates with both logs disabled")
}

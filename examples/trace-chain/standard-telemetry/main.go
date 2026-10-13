package main

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teaql/teaql-golang/runtime"
	_ "github.com/teaql/teaql-golang/telemetry/standard"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/platform"
)

type audits struct{ count int }

func (a *audits) OnSafeEvent(_ *runtime.UserContext, _ *runtime.SafeAuditEvent) error {
	a.count++
	return nil
}

func execute() error {
	// Fresh isolated database per run; retained for evidence, never deleted to make a rerun pass.
	directory, err := os.MkdirTemp("", "teaql-go-standard-telemetry-")
	if err != nil {
		return err
	}
	database := filepath.Join(directory, "example.sqlite")
	if err := os.Setenv("TRACE_CHAIN_SERVICE_CORE_DATABASE_URL", database); err != nil {
		return err
	}
	context, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	defer context.GetResource("db").(*sql.DB).Close()
	defer context.RuntimeTelemetry().Shutdown(stdcontext.Background())
	// Ordinary SQL logs are off; enabling the standard module still exports safe SQL diagnostics.
	ordinary := runtime.NewSQLExecutionEvidenceStore()
	audit := &audits{}
	context.WithDiagnosticSQLLogSink(ordinary).WithAppAuditEventSink(audit).DisableSqlLog()
	if err := lib.EnsureSchema(context); err != nil {
		return err
	}
	first := audit.count
	if err := lib.EnsureSchema(context); err != nil {
		return err
	}
	if first != 1 || audit.count != first {
		return fmt.Errorf("expected one audited seed and no repeated seed, first=%d total=%d", first, audit.count)
	}
	for round := 0; round < 2; round++ {
		rows, err := lib.Q.PlatformsMinimal().WithIdIs(1).SelectName().Limit(1).
			Comment("load seeded platform").Purpose("verify standard telemetry through generated Q and E").ExecuteForList(context)
		if err != nil {
			return err
		}
		if len(rows.Data) != 1 {
			return fmt.Errorf("seeded platform missing")
		}
		id, loaded := platform.NewPlatformExpression(rows.Data[0]).Id().Eval()
		name, named := platform.NewPlatformExpression(rows.Data[0]).Name().Eval()
		if !loaded || id != 1 || !named || name == "" {
			return fmt.Errorf("generated E failed to verify seed")
		}
	}
	if len(ordinary.Snapshot()) != 0 {
		return fmt.Errorf("ordinary SQL logging switch ignored")
	}
	if err := context.RuntimeTelemetry().Flush(stdcontext.Background()); err != nil {
		return err
	}
	fmt.Printf("GO_STANDARD_TELEMETRY_EXAMPLE PASSED first_seed_audits=%d repeat_seed_audits=0 queries=2 ordinary_sql_logs=0 database=%s\n", first, database)
	return nil
}
func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package lib

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type bootstrapTraceEvidence struct {
	db       *sql.DB
	versions map[uint64]int64
	events   []*runtime.SafeAuditEvent
}

func (s *bootstrapTraceEvidence) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if len(event.TraceChain) != 1 || event.TraceChain[0].EntityId == nil {
		return fmt.Errorf("bootstrap audit missing assigned lineage: %+v", event)
	}
	id := *event.TraceChain[0].EntityId
	table := map[string]string{"Platform": "platform_data", "School Type": "school_type_data"}[event.Entity]
	if table == "" {
		return fmt.Errorf("unexpected entity: %s", event.Entity)
	}
	var version int64
	// Separate read-only connection: an audit callback must not precede commit.
	if err := s.db.QueryRow("SELECT version FROM "+table+" WHERE id = ?", id).Scan(&version); err != nil {
		return err
	}
	if version != s.versions[id] {
		return fmt.Errorf("audit before committed version: id=%d got=%d want=%d", id, version, s.versions[id])
	}
	s.events = append(s.events, event)
	return nil
}

func TestGeneratedBootstrapTrace(t *testing.T) {
	database := os.Getenv("TEAQL_SCHOOL_BOOTSTRAP_DB")
	if database == "" {
		database = filepath.Join(t.TempDir(), "bootstrap.sqlite")
	}
	t.Setenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL", database)
	ctx, err := ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ctx.GetResource("db").(*sql.DB).Close() })
	observer, err := sql.Open("sqlite3", "file:"+database+"?mode=ro&_busy_timeout=1000")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { observer.Close() })
	evidence := &bootstrapTraceEvidence{db: observer, versions: map[uint64]int64{1: 1, 1001: 1, 1002: 1}}
	logs := runtime.NewSQLExecutionEvidenceStore()
	ctx.WithAppAuditEventSink(evidence).WithDiagnosticSQLLogSink(logs).WithUserIdentifier("school-example-user")
	logging := os.Getenv("TEAQL_SCHOOL_BOOTSTRAP_LOGGING") != "off"
	if !logging {
		ctx.DisableSqlLog()
	}
	clear := func() { evidence.events = nil; logs.EnableAll() }
	verify := func(writes, reads int) {
		t.Helper()
		if len(evidence.events) != writes {
			t.Fatalf("audits=%d want=%d", len(evidence.events), writes)
		}
		audits := map[uint64]*runtime.SafeAuditEvent{}
		for _, event := range evidence.events {
			if event.Actor != "teaql-generated-bootstrap" || event.Category != "runtime-bootstrap" {
				t.Fatalf("bootstrap identity: %+v", event)
			}
			node := event.TraceChain[0]
			if node.Kind != "auditReason" || node.Comment == "" || node.Name != event.Entity {
				t.Fatalf("lineage=%+v entity=%s", node, event.Entity)
			}
			audits[*node.EntityId] = event
		}
		sql := logs.Snapshot()
		if !logging {
			if len(sql) != 0 {
				t.Fatalf("SQL logs disabled: %d", len(sql))
			}
			return
		}
		if len(sql) != writes+reads {
			t.Fatalf("SQL facts=%d want=%d", len(sql), writes+reads)
		}
		actualWrites := 0
		for _, row := range sql {
			kinds := []string{}
			for _, node := range row.TraceChain {
				kinds = append(kinds, node.Kind)
			}
			middle := "entity"
			if row.Operation == data_service.OpQuery {
				middle = "request"
			}
			if !reflect.DeepEqual(kinds, []string{"operation", middle, "provider", "sql"}) {
				t.Fatalf("SQL path=%v", row.TraceChain)
			}
			if row.Operation == data_service.OpQuery {
				if row.Comment == nil || *row.Comment == "" || row.Purpose == nil || *row.Purpose == "" {
					t.Fatal("lookup/readback intent absent")
				}
			} else {
				actualWrites++
				if len(row.MutationLineage) != 1 || row.MutationLineage[0].EntityId == nil {
					t.Fatal("SQL lineage absent")
				}
				node := row.MutationLineage[0]
				audit := audits[*node.EntityId]
				if audit == nil || !reflect.DeepEqual(row.MutationLineage, core.CloneTraceNodes(audit.TraceChain)) {
					t.Fatalf("SQL/audit lineage mismatch: %v / %+v", row.MutationLineage, audit)
				}
				if row.AuditReason == nil || audit.AuditReason == nil || *row.AuditReason != *audit.AuditReason {
					t.Fatal("SQL/audit intent mismatch")
				}
			}
		}
		if actualWrites != writes {
			t.Fatalf("writes=%d want=%d", actualWrites, writes)
		}
	}
	if err := EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	fresh := len(evidence.events) == 3
	if fresh {
		verify(3, 6)
	} else {
		verify(0, 3)
	}
	clear()
	if err := EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	verify(0, 3)
	primary, err := Q.SchoolTypes().WithIdIs(1001).Comment("load Primary for audited drift").Purpose("verify bootstrap reconciliation").ExecuteForOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	version := primary.Version()
	evidence.versions[1001] = int64(version + 1)
	primary.UpdateName("Drifted Primary")
	if _, err := primary.AuditAs("simulate constant drift").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if len(evidence.events) != 1 || evidence.events[0].Actor != "school-example-user" || evidence.events[0].Category == "runtime-bootstrap" {
		t.Fatal("application save inherited bootstrap identity")
	}
	clear()
	evidence.versions[1001] = int64(version + 2)
	if err := EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	verify(1, 4)
	changed, err := Q.SchoolTypes().WithIdIs(1001).Comment("verify corrected constant").Purpose("verify persisted bootstrap update").ExecuteForOne(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Name() != "Primary" || changed.Version() != version+2 {
		t.Fatal("constant not corrected exactly once")
	}
	clear()
	if err := EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	verify(0, 3)
	if ctx.UserIdentifier() != "school-example-user" {
		t.Fatal("bootstrap leaked actor")
	}
	t.Logf("PASS Go generated bootstrap trace logging=%t fresh=%t originalVersion=%d", logging, fresh, version)
}

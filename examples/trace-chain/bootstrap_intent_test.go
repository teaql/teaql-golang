package tracechain_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	lib "trace-chain-service-core-workspace/lib"
)

// Observe real requests/results only. No expected trace is injected into them.
type bootstrapQueryFact struct {
	comment, purpose string
	metadata         data_service.ExecutionMetadata
}

func (p *graphObserver) observeBootstrapQuery(request *data_service.QueryRequest, result *data_service.QueryResult) {
	if request.Comment == nil || request.Purpose == nil {
		panic("bootstrap query omitted request-owned intent")
	}
	p.mu.Lock()
	p.queryFacts = append(p.queryFacts, bootstrapQueryFact{*request.Comment, *request.Purpose, result.Metadata})
	p.mu.Unlock()
}

type bootstrapCommittedAudit struct {
	*safeEvents
	db *sql.DB
}

func (s *bootstrapCommittedAudit) OnSafeEvent(ctx *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if event.Entity != "Platform" || len(event.TraceChain) != 1 || event.TraceChain[0].EntityId == nil {
		return fmt.Errorf("bootstrap audit omitted assigned Platform identity")
	}
	var version int64
	// Test oracle only: independent read-only connection proves commit timing.
	if err := s.db.QueryRow("SELECT version FROM platform_data WHERE id = ?", *event.TraceChain[0].EntityId).Scan(&version); err != nil {
		return err
	}
	if version != 1 {
		return fmt.Errorf("bootstrap audit before committed Platform version: %d", version)
	}
	return s.safeEvents.OnSafeEvent(ctx, event)
}

func TestGeneratedBootstrapOwnedIntent(t *testing.T) {
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging-%t", logging), func(t *testing.T) {
			directory := os.Getenv("TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY")
			if directory == "" {
				directory = t.TempDir()
			}
			database := filepath.Join(directory, t.Name()+".sqlite")
			if err := os.MkdirAll(filepath.Dir(database), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("TRACE_CHAIN_SERVICE_CORE_DATABASE_URL", database)
			ctx, err := lib.ServiceRuntimeFromEnv()
			if err != nil {
				t.Fatal(err)
			}
			db := ctx.GetResource("db").(*sql.DB)
			t.Cleanup(func() { db.Close() })
			readonly, err := sql.Open("sqlite3", "file:"+database+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { readonly.Close() })
			observer := &graphObserver{TransactionExecutor: ctx.GetResource("dataService").(data_service.TransactionExecutor)}
			ctx.InsertResource("dataService", observer)
			sink := &bootstrapCommittedAudit{safeEvents: &safeEvents{}, db: readonly}
			logs := runtime.NewSQLExecutionEvidenceStore()
			ctx.WithAppAuditEventSink(sink).WithDiagnosticSQLLogSink(logs).WithUserIdentifier("bootstrap-caller")
			if !logging {
				ctx.DisableSqlLog()
			}
			if err := lib.EnsureSchema(ctx); err != nil {
				t.Fatal(err)
			}
			commands := observer.snapshot()
			firstWrites := len(commands)
			if firstWrites > 1 || len(sink.snapshot()) != firstWrites || len(observer.mutationFacts) != firstWrites {
				t.Fatal("bootstrap write/result/committed audit cardinality")
			}
			// The provider performs its readback internally, not through QueryExecutor.
			// Raw query observation proves the lookup; physical readback is verified
			// below when diagnostics are enabled, with committed state checked in both modes.
			if len(observer.queryFacts) != 1 {
				t.Fatal("observe actual bootstrap lookup")
			}
			lookup := observer.queryFacts[0]
			if strings.TrimSpace(lookup.comment) == "" || strings.TrimSpace(lookup.purpose) == "" {
				t.Fatal("bootstrap lookup owns nonblank intent")
			}
			for _, query := range observer.queryFacts {
				assertBootstrapQuery(t, query)
			}
			for i, command := range commands {
				if command.Comment() == nil || strings.TrimSpace(*command.Comment()) == "" {
					t.Fatal("bootstrap mutation request owns nonblank comment")
				}
				assertLineage(t, command.TraceChain(), []*core.TraceNode{reasonNode("Platform", 1, *command.Comment())})
				fact := observer.mutationFacts[i]
				if fact.AuditReason == nil || *fact.AuditReason != *command.Comment() {
					t.Fatal("bootstrap raw SQL intent matches actual mutation request")
				}
				assertBootstrapPhysical(t, fact, "entity", "insert")
				audit := sink.snapshot()[i]
				if audit.Actor != "teaql-generated-bootstrap" || audit.Category != "runtime-bootstrap" {
					t.Fatal("bootstrap responsibility")
				}
				safeReason := strings.ReplaceAll(*command.Comment(), "1", "[REDACTED]")
				if audit.AuditReason == nil || *audit.AuditReason != safeReason {
					t.Fatal("bootstrap safe audit preserves governed request intent")
				}
				assertLineage(t, audit.TraceChain, []*core.TraceNode{reasonNode("Platform", 1, safeReason)})
			}
			if logging && len(logs.Snapshot()) != 1+2*firstWrites || !logging && len(logs.Snapshot()) != 0 {
				t.Fatal("bootstrap SQL logging switch does not alter audit")
			}
			if logging {
				for _, fact := range logs.Snapshot() {
					middle, kind := "request", "select"
					if fact.Operation == data_service.OpInsert {
						middle, kind = "entity", "insert"
					}
					assertBootstrapPhysical(t, fact, middle, kind)
					if fact.AuditReason != nil {
						if len(commands) != 1 || *fact.AuditReason != *sink.snapshot()[0].AuditReason {
							t.Fatal("bootstrap physical readback retains request audit intent")
						}
						assertLineage(t, fact.MutationLineage, sink.snapshot()[0].TraceChain)
					}
				}
			}
			var persistedVersion int64
			if err := readonly.QueryRow("SELECT version FROM platform_data WHERE id = 1").Scan(&persistedVersion); err != nil || persistedVersion != 1 {
				t.Fatalf("bootstrap committed root: %d %v", persistedVersion, err)
			}
			observer.reset()
			sink.reset()
			logs.EnableAll()
			if err := lib.EnsureSchema(ctx); err != nil {
				t.Fatal(err)
			}
			if len(observer.snapshot()) != 0 || len(observer.mutationFacts) != 0 || len(sink.snapshot()) != 0 {
				t.Fatal("repeated bootstrap must not write or audit")
			}
			if len(observer.queryFacts) != 1 {
				t.Fatal("repeated bootstrap retains lookup")
			}
			repeat := observer.queryFacts[0]
			if repeat.comment != lookup.comment || repeat.purpose != lookup.purpose {
				t.Fatal("generated bootstrap owns stable lookup comment and purpose")
			}
			assertBootstrapQuery(t, repeat)
			if logging && len(logs.Snapshot()) != 1 || !logging && len(logs.Snapshot()) != 0 {
				t.Fatal("repeated bootstrap logging policy")
			}
			if ctx.UserIdentifier() != "bootstrap-caller" || ctx.GetResource("dataService") != observer {
				t.Fatal("bootstrap restores caller context")
			}
			// Ordinary application inspection remains through generated Q and E.
			root, err := lib.Q.Platforms().WithIdIs(1).Limit(1).Comment("load bootstrapped root").Purpose("verify generated identity").ExecuteForOne(ctx)
			if err != nil || root == nil {
				t.Fatalf("generated root reload failed: %v", err)
			}
			// The default root ID is exercised by existing generated Q/E tests; no
			// generated source discovery or manual seed is used here.
			t.Logf("TC-REQ-09 GO GENERATED BOOTSTRAP PASSED logging=%t first_writes=%d repeat_writes=0 comment=%s purpose=%s", logging, firstWrites, lookup.comment, lookup.purpose)
		})
	}
}

func assertBootstrapQuery(t *testing.T, query bootstrapQueryFact) {
	t.Helper()
	if query.metadata.Comment == nil || *query.metadata.Comment != query.comment || query.metadata.Purpose == nil || *query.metadata.Purpose != query.purpose {
		t.Fatal("bootstrap query physical metadata retains owned intent")
	}
	assertBootstrapPhysical(t, query.metadata, "request", "select")
}

func assertBootstrapPhysical(t *testing.T, fact data_service.ExecutionMetadata, middle, sqlKind string) {
	t.Helper()
	kinds := []string{}
	for _, node := range fact.TraceChain {
		kinds = append(kinds, node.Kind)
	}
	if !reflect.DeepEqual(kinds, []string{"operation", middle, "provider", "sql"}) {
		t.Fatalf("bootstrap canonical route: %v", kinds)
	}
	if fact.TraceChain[2].Name != "sqlite" || fact.TraceChain[3].Name != sqlKind || fact.ExecutionOutcome != "success" {
		t.Fatal("bootstrap actual SQL provider and outcome")
	}
}

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

type blankTailAudit struct {
	t           *testing.T
	independent *sql.DB
	probe       *blankReasonTransport
	capture     *batchAuditCapture
}

func (s *blankTailAudit) OnSafeEvent(ctx *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if s.probe.commits != 1 || event.Entity != "Customer" || event.TargetID == nil {
		s.t.Fatal("audit emitted without committed typed target")
	}
	var version int64
	if err := s.independent.QueryRow("SELECT version FROM customer_data WHERE id=?", event.TargetID.V).
		Scan(&version); err != nil || version != 1 {
		s.t.Fatal("independent connection cannot see committed audit target", err)
	}
	return s.capture.OnSafeEvent(ctx, event)
}

// TC-REQ-13 deliberately supplies caller route-tail stimuli. This proves actual
// native execution, not generated relation traversal or generated graph paths.
func TestNativeExplicitCommentSurvivesBlankRouteTailsAtRealSinks(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) {
			for _, kind := range []string{"entity", "provider", "sql"} {
				t.Run(kind, func(t *testing.T) {
					e := nativeBatchEnvironment(t)
					if !logging {
						e.ctx.DisableSqlLog()
					}
					var seq int
					var name, databasePath string
					if err := e.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &databasePath); err != nil {
						t.Fatal(err)
					}
					independent, err := sql.Open("sqlite3", databasePath)
					if err != nil {
						t.Fatal(err)
					}
					defer independent.Close()
					probe := &blankReasonTransport{SqlTransactionTransport: e.executor.Transport.(tsql.SqlTransactionTransport)}
					e.executor.Transport = probe
					e.ctx.SetAppAuditEventSink(&blankTailAudit{t, independent, probe, e.audit})
					owner := &groupedLineageOwner{UserContext: e.ctx}
					invocation := context.WithValue(e.ctx, mutationaudit.ContextKey{}, owner)
					comment := "explicit root comment"
					id := int64(601)
					leaf := nativeBatchLeaf(id, "native tail fixture", "").(*ds.InsertMutation)
					lineage := []*core.TraceNode{
						core.NewTypedTraceNode("auditReason", "Customer", comment),
						core.NewTypedTraceNode(kind, "tail", ""),
					}
					leaf.Cmd.TraceChain = lineage
					request, err := ds.NewMutationRequest(leaf, comment)
					if err != nil {
						t.Fatal(err)
					}
					if request.TraceChain()[1].Kind != kind || request.TraceChain()[1].Comment != "" {
						t.Fatal("request did not retain blank typed tail")
					}
					intent, err := core.NewMutationIntent(&comment)
					if err != nil {
						t.Fatal(err)
					}
					var commands *blankReasonExecutor
					err = e.ctx.ExecuteGraphSave(intent, func() error {
						commands = &blankReasonExecutor{MutationExecutor: e.ctx.GetResource("dataService").(ds.MutationExecutor)}
						result, err := commands.Mutate(invocation, request)
						if err != nil {
							return err
						}
						if result.AffectedRows != 1 || result.PersistedRecord == nil || len(e.audit.events) != 0 {
							t.Fatal("real write/readback failed or audit escaped before commit")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					if probe.begins != 1 || probe.commits != 1 || probe.rollbacks != 0 ||
						len(commands.commands) != 1 || len(probe.writes) != 1 || len(probe.reads) != 1 ||
						len(owner.statements) != 2 || len(e.audit.events) != 1 {
						t.Fatal("wrong physical SQL, request, commit or audit counts")
					}
					captured, event := commands.commands[0], e.audit.events[0]
					if captured.Comment() == nil || *captured.Comment() != comment || !reflect.DeepEqual(captured.TraceChain(), lineage) ||
						!reflect.DeepEqual(event.TraceChain, lineage) || event.TargetID == nil || *event.TargetID != core.ValI64(id) ||
						event.AuditReason == nil || *event.AuditReason != comment {
						t.Fatal("blank route tail changed command or committed audit lineage/identity")
					}
					for index, compiled := range []*tsql.CompiledQuery{probe.writes[0], probe.reads[0]} {
						fact := owner.statements[index]
						op, tail := ds.OpInsert, "insert"
						if index == 1 {
							op, tail = ds.OpQuery, "select"
						}
						if fact.Operation != op || fact.ExecutionOutcome != "success" || fact.ParameterizedSQL != compiled.Sql ||
							!reflect.DeepEqual(fact.Parameters, compiled.Params) || !reflect.DeepEqual(fact.MutationLineage, lineage) ||
							fact.Comment == nil || *fact.Comment != comment || fact.AuditReason == nil || *fact.AuditReason != comment ||
							len(fact.TraceChain) != 4 || fact.TraceChain[2].Name != "sqlite" || fact.TraceChain[3].Name != tail {
							t.Fatal("blank route tail replaced request-owned SQL reason or corrupted canonical route")
						}
						for _, node := range fact.TraceChain {
							if node.Kind == "auditReason" {
								t.Fatal("audit lineage leaked into canonical SQL route")
							}
						}
					}
					wantLogs := 0
					if logging {
						wantLogs = 2
					}
					if len(e.logs.entries) != wantLogs || (!logging && e.output.Len() != 0) {
						t.Fatal("logging mode changed SQL diagnostics")
					}
					t.Logf("TC-REQ-13 GO REAL SINKS PASSED logging=%t tail=%s", logging, kind)
				})
			}
		})
	}
}

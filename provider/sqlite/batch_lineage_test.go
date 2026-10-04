package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
	"github.com/teaql/teaql-golang/runtime"
)

// Observe provider-produced raw metadata even when ordinary SQL logging is off.
// Embedding preserves the real trusted audit owner; no event/frame is supplied.
type groupedLineageOwner struct {
	*runtime.UserContext
	statements []ds.ExecutionMetadata
}

func (o *groupedLineageOwner) RecordExecutionMetadata(metadata ds.ExecutionMetadata) {
	o.statements = append(o.statements, metadata)
	o.UserContext.RecordExecutionMetadata(metadata)
}

func groupedLineageRequest(t *testing.T, update bool) ds.MutationRequest {
	t.Helper()
	var items []ds.MutationRequest
	// Deliberately reverse identity order: index is execution order, not sorted ID.
	for index, id := range []int64{2, 1} {
		reason := []string{"second responsibility", "first responsibility"}[index]
		localIntent, err := core.NewMutationIntent(&reason)
		if err != nil {
			t.Fatal(err)
		}
		key := core.NewEntityKey("Customer", core.ValI64(id))
		scope, err := core.MutationScopeForEntity(nil, key, localIntent, nil)
		if err != nil {
			t.Fatal(err)
		}
		// This is the exact runtime scope recovery used by generated mutations,
		// not a manually constructed expected trace or a generated-entry claim.
		lineage := core.MutationTraceForEntity(core.NewEntityRoot(), key, scope)
		if update {
			command := core.NewUpdateCommand("Customer", core.ValI64(id)).WithExpectedVersion(1).
				Value("display_name", core.ValText(batchPrivate+" updated"))
			command.OldValues = core.Record{"display_name": core.ValText(batchPrivate), "password_hash": core.ValText(batchCredential)}
			command.TraceChain = lineage
			items = append(items, &ds.UpdateMutation{Cmd: command})
		} else {
			item := nativeBatchLeaf(id, batchPrivate, batchCredential).(*ds.InsertMutation)
			item.Cmd.TraceChain = lineage
			items = append(items, item)
		}
	}
	request, err := ds.NewMutationRequest(&ds.BatchMutation{Mutations: items},
		"group "+batchPrivate+" "+batchCredential+" "+batchVisible)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func executeGroupedLineage(t *testing.T, e *batchEnvironment, request ds.MutationRequest) (*ds.MutationResult, []ds.ExecutionMetadata, error) {
	t.Helper()
	intent, err := core.NewMutationIntent(request.Comment())
	if err != nil {
		t.Fatal(err)
	}
	owner := &groupedLineageOwner{UserContext: e.ctx}
	var result *ds.MutationResult
	err = e.ctx.ExecuteGraphSave(intent, func() error {
		ctx := context.WithValue(e.ctx, mutationaudit.ContextKey{}, owner)
		var mutateErr error
		result, mutateErr = e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(ctx, request)
		if len(e.audit.events) != 0 {
			t.Error("batch emitted committed audit before transaction COMMIT")
		}
		return mutateErr
	})
	return result, owner.statements, err
}

func assertGroupedLineage(t *testing.T, trace []*core.TraceNode, id uint64, local, root string) {
	t.Helper()
	if len(trace) != 2 {
		t.Errorf("batch root responsibility missing or item chain collapsed: id=%d lineage=%+v", id, trace)
		return
	}
	if trace[0].Kind != "auditReason" || trace[0].EntityType != "Customer" || trace[0].EntityId != nil || trace[0].Comment != root {
		t.Errorf("batch prefix invented identity or lost owned intent: %+v", trace[0])
	}
	if trace[1].Kind != "auditReason" || trace[1].EntityType != "Customer" || trace[1].EntityId == nil || *trace[1].EntityId != id || trace[1].Comment != local {
		t.Errorf("wrong item responsibility for id=%d: %+v", id, trace[1])
	}
}

func TestNativeBatchSameTypeItemsKeepOrderedCommandSQLAndAuditLineage(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			if !logging {
				e.ctx.DisableSqlLog()
			}
			for _, update := range []bool{false, true} {
				e.audit.events, e.logs.entries = nil, nil
				e.output.Reset()
				request := groupedLineageRequest(t, update)
				input, _ := json.Marshal(request)
				result, statements, err := executeGroupedLineage(t, e, request)
				if err != nil || result == nil || result.AffectedRows != 2 || result.Metadata.Operation != ds.OpBatch {
					t.Fatalf("real same-type batch failed: result=%+v error=%v", result, err)
				}
				if len(statements) != 4 || len(e.audit.events) != 2 {
					t.Fatalf("want two physical writes/readbacks and two committed events; statements=%d audits=%d", len(statements), len(e.audit.events))
				}
				batch := request.(*ds.BatchMutation)
				for index, id := range []uint64{2, 1} {
					local := []string{"second responsibility", "first responsibility"}[index]
					assertGroupedLineage(t, batch.Mutations[index].TraceChain(), id, local, *request.Comment())
					for _, statement := range statements[index*2 : index*2+2] {
						assertGroupedLineage(t, statement.MutationLineage, id, local, *request.Comment())
						if statement.Comment == nil || *statement.Comment != *request.Comment() || statement.AuditReason == nil || *statement.AuditReason != *request.Comment() ||
							len(statement.TraceChain) != 4 || statement.TraceChain[0].Name != "Customer" || statement.TraceChain[2].Name != "sqlite" || statement.TraceChain[3].Kind != "sql" {
							t.Error("physical statement lost original batch intent or canonical route")
						}
					}
					write := statements[index*2]
					wantOperation := ds.OpInsert
					if update {
						wantOperation = ds.OpUpdate
					}
					if write.Operation != wantOperation || write.AffectedRows == nil || *write.AffectedRows != 1 || statements[index*2+1].Operation != ds.OpQuery ||
						statements[index*2+1].ResultCount == nil || *statements[index*2+1].ResultCount != 1 {
						t.Fatal("metadata did not describe actual individual writes/readbacks")
					}
					safeRoot := "group [REDACTED] [REDACTED] " + batchVisible
					assertGroupedLineage(t, e.audit.events[index].TraceChain, id, local, safeRoot)
					if e.audit.events[index].AuditReason == nil || *e.audit.events[index].AuditReason != safeRoot {
						t.Fatal("safe audit lost shared batch privacy")
					}
				}
				if statements[0].ParameterizedSQL != statements[2].ParameterizedSQL || reflect.DeepEqual(statements[0].Parameters, statements[2].Parameters) {
					t.Fatal("same-column items did not retain equal SQL shape and separate bindings")
				}
				after, _ := json.Marshal(request)
				if string(input) != string(after) || !strings.Contains(*request.Comment(), batchPrivate) {
					t.Fatal("execution mutated the captured command/lineage/intent")
				}
				wantLogs := 0
				if logging {
					wantLogs = 4
				}
				if len(e.logs.entries) != wantLogs {
					t.Fatal("logging changed physical work or emitted unexpected diagnostics")
				}
				for _, payload := range []any{e.logs.entries, e.audit.events} {
					encoded, _ := json.Marshal(payload)
					if strings.Contains(string(encoded), batchPrivate) || strings.Contains(string(encoded), batchCredential) {
						t.Fatal("safe batch boundary leaked a private binding")
					}
				}
				query := core.NewSelectQuery("Customer").Limit(10).Comment("independent " + batchPrivate).
					Purpose("verify persisted batch values and invocation privacy isolation")
				rows, err := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
				if err != nil || len(rows.Rows) != 2 {
					t.Fatalf("real batch rows missing: %v", err)
				}
				for _, row := range rows.Rows {
					name, version := batchPrivate, int64(1)
					if update {
						name, version = batchPrivate+" updated", 2
					}
					if row["display_name"].V != name || row["password_hash"].V != batchCredential || row["version"].V != version {
						t.Fatal("privacy changed database values or grouped version association")
					}
				}
				if logging && *e.logs.entries[len(e.logs.entries)-1].Comment != *query.CommentText {
					t.Fatal("batch privacy contaminated the next independent query")
				}
				t.Logf("one native same-type BatchMutation update=%t; two ordered physical writes, two readbacks, two committed per-item audits", update)
			}
		})
	}
}

func TestNativeBatchSameTypeFailureRetainsItemLineageWithoutCommittedAudit(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			if !logging {
				e.ctx.DisableSqlLog()
			}
			if _, err := e.db.Exec("CREATE UNIQUE INDEX grouped_public_marker ON customer_data(public_marker)"); err != nil {
				t.Fatal(err)
			}
			request := groupedLineageRequest(t, false)
			_, statements, err := executeGroupedLineage(t, e, request)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unique") || len(statements) != 3 || len(e.audit.events) != 0 {
				t.Fatalf("second real write must fail after first write/readback with no committed audit: err=%v SQL=%d audits=%d", err, len(statements), len(e.audit.events))
			}
			assertGroupedLineage(t, statements[2].MutationLineage, 1, "first responsibility", *request.Comment())
			if statements[2].ExecutionOutcome != "failure" || statements[2].AffectedRows != nil {
				t.Fatal("failed SQL fabricated affected rows")
			}
			wantLogs := 0
			if logging {
				wantLogs = 3
			}
			encoded, _ := json.Marshal(e.logs.entries)
			if len(e.logs.entries) != wantLogs || strings.Contains(string(encoded), batchPrivate) || strings.Contains(string(encoded), batchCredential) {
				t.Fatal("failed batch lost log mode or shared privacy")
			}
			query := core.NewSelectQuery("Customer").Limit(10).Comment("independent " + batchPrivate).Purpose("no partial commit or retained privacy")
			rows, queryErr := e.executor.Query(e.ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
			if queryErr != nil || len(rows.Rows) != 0 {
				t.Fatalf("failed batch retained earlier rows: result=%+v error=%v", rows, queryErr)
			}
			if logging && *e.logs.entries[len(e.logs.entries)-1].Comment != *query.CommentText {
				t.Fatal("failed batch privacy escaped into following request")
			}
		})
	}
}

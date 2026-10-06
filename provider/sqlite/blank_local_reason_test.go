package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

// Delegate every SQL operation to real SQLite. Captures are observations, not
// supplied metadata or expected trace nodes. The native fixture models Customer
// only; this is not evidence for generated relation traversal.
type blankReasonTransport struct {
	tsql.SqlTransactionTransport
	begins, commits, rollbacks int
	writes, reads              []*tsql.CompiledQuery
}

func (p *blankReasonTransport) FetchAllSql(ctx context.Context, query *tsql.CompiledQuery) ([]core.Record, error) {
	p.reads = append(p.reads, query)
	return p.SqlTransactionTransport.FetchAllSql(ctx, query)
}

func (p *blankReasonTransport) ExecuteSql(ctx context.Context, query *tsql.CompiledQuery) (uint64, error) {
	p.writes = append(p.writes, query)
	return p.SqlTransactionTransport.ExecuteSql(ctx, query)
}

func (p *blankReasonTransport) BeginSql(ctx context.Context) (tsql.SqlTransactionTransportTx, error) {
	p.begins++
	tx, err := p.SqlTransactionTransport.BeginSql(ctx)
	if err != nil {
		return nil, err
	}
	return &blankReasonTx{SqlTransactionTransportTx: tx, owner: p}, nil
}

type blankReasonTx struct {
	tsql.SqlTransactionTransportTx
	owner *blankReasonTransport
}

func (tx *blankReasonTx) FetchAllSql(ctx context.Context, query *tsql.CompiledQuery) ([]core.Record, error) {
	tx.owner.reads = append(tx.owner.reads, query)
	return tx.SqlTransactionTransportTx.FetchAllSql(ctx, query)
}

func (tx *blankReasonTx) ExecuteSql(ctx context.Context, query *tsql.CompiledQuery) (uint64, error) {
	tx.owner.writes = append(tx.owner.writes, query)
	return tx.SqlTransactionTransportTx.ExecuteSql(ctx, query)
}

func (tx *blankReasonTx) CommitSql(ctx context.Context) error {
	err := tx.SqlTransactionTransportTx.CommitSql(ctx)
	if err == nil {
		tx.owner.commits++
	}
	return err
}

func (tx *blankReasonTx) RollbackSql(ctx context.Context) error {
	tx.owner.rollbacks++
	return tx.SqlTransactionTransportTx.RollbackSql(ctx)
}

type blankReasonAuditSink struct {
	t       *testing.T
	probe   *blankReasonTransport
	capture *batchAuditCapture
}

func (s *blankReasonAuditSink) OnSafeEvent(ctx *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if s.probe.commits != 1 {
		s.t.Fatal("safe audit escaped before actual SQLite commit")
	}
	return s.capture.OnSafeEvent(ctx, event)
}

type blankReasonExecutor struct {
	ds.MutationExecutor
	commands []ds.MutationRequest
}

func (e *blankReasonExecutor) Mutate(ctx context.Context, request ds.MutationRequest) (*ds.MutationResult, error) {
	// Snapshot the request at the real executor boundary; pass the original on.
	captured, err := ds.CaptureMutationRequest(request)
	if err != nil {
		return nil, err
	}
	e.commands = append(e.commands, captured)
	return e.MutationExecutor.Mutate(ctx, request)
}

func testNativeLocalReasons(t *testing.T, reasons []*string, inherit, logging bool) {
	t.Helper()
	e := nativeBatchEnvironment(t)
	if !logging {
		e.ctx.DisableSqlLog()
	}
	probe := &blankReasonTransport{SqlTransactionTransport: e.executor.Transport.(tsql.SqlTransactionTransport)}
	e.executor.Transport = probe
	e.ctx.SetAppAuditEventSink(&blankReasonAuditSink{t, probe, e.audit})
	owner := &groupedLineageOwner{UserContext: e.ctx}
	invocation := context.WithValue(e.ctx, mutationaudit.ContextKey{}, owner)
	if inherit {
		for _, blank := range reasons {
			if blank != nil {
				_, err := ds.NewMutationRequest(nativeBatchLeaf(400, "rejected customer", ""), *blank)
				assertBlankReasonError(t, err)
			}
			request := nativeBatchLeaf(400, "rejected customer", "").(*ds.InsertMutation)
			request.RootComment = blank
			_, err := e.executor.Mutate(invocation, request)
			assertBlankReasonError(t, err)
		}
		if probe.begins != 0 || len(probe.reads)+len(probe.writes)+len(owner.statements)+len(e.audit.events)+len(e.logs.entries) != 0 {
			t.Fatal("public invalid comment reached transaction, physical SQL or audit")
		}
	}
	rootReason, parentReason, siblingReason := "save customer graph", "approve related customer", "review sibling customer"
	intent, err := core.NewMutationIntent(&rootReason)
	if err != nil {
		t.Fatal(err)
	}
	var commands *blankReasonExecutor
	var expected [][]*core.TraceNode
	err = e.ctx.ExecuteGraphSave(intent, func() error {
		root, err := core.MutationScopeForEntity(nil, core.NewEntityKey("Customer", core.ValI64(100)), intent, nil)
		if err != nil {
			return err
		}
		parent, err := core.MutationScopeForEntity(root, core.NewEntityKey("Customer", core.ValI64(201)), intent, &parentReason)
		if err != nil {
			return err
		}
		sibling, err := core.MutationScopeForEntity(root, core.NewEntityKey("Customer", core.ValI64(301)), intent, &siblingReason)
		if err != nil {
			return err
		}
		commands = &blankReasonExecutor{MutationExecutor: e.ctx.GetResource("dataService").(ds.MutationExecutor)}
		for index := 0; index <= len(reasons); index++ {
			id, scope := int64(301), sibling
			if index < len(reasons) {
				id = int64(400 + index)
				scope, err = core.MutationScopeForEntity(parent, core.NewEntityKey("Customer", core.ValI64(id)), intent, reasons[index])
				if err != nil {
					return err
				}
				if (scope == parent) != inherit {
					t.Fatal("local reason classification disagrees with the public Unicode contract")
				}
			}
			lineage := core.MutationTraceForEntity(core.NewEntityRoot(), core.NewEntityKey("Customer", core.ValI64(id)), scope)
			expected = append(expected, lineage)
			leaf := nativeBatchLeaf(id, "native local customer", "").(*ds.InsertMutation)
			leaf.Cmd.TraceChain = lineage
			request, err := ds.NewMutationRequest(leaf, rootReason)
			if err != nil {
				return err
			}
			result, err := commands.Mutate(invocation, request)
			if err != nil {
				return err
			}
			if result == nil || result.AffectedRows != 1 || result.PersistedRecord == nil || len(e.audit.events) != 0 {
				t.Fatal("real write/readback failed or audit escaped before graph commit")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	count := len(reasons) + 1
	if probe.begins != 1 || probe.commits != 1 || probe.rollbacks != 0 || len(commands.commands) != count ||
		len(probe.writes) != count || len(probe.reads) != count || len(owner.statements) != 2*count || len(e.audit.events) != count {
		t.Fatal("wrong command, driver write/readback, commit or audit counts")
	}
	// Recovering a scope supplies the actual lineage. Independent expected node
	// assertions below prevent a broken Recover implementation from self-passing.
	for index, request := range commands.commands {
		id, local := int64(301), siblingReason
		chain := expected[index]
		if index < len(reasons) {
			id, local = int64(400+index), parentReason
		}
		wantNodes := 2
		if !inherit && index < len(reasons) {
			wantNodes = 3
		}
		if len(chain) != wantNodes || chain[0].Comment != rootReason || chain[0].EntityId == nil || *chain[0].EntityId != 100 ||
			chain[1].Comment != local || chain[1].EntityId == nil || *chain[1].EntityId != uint64(map[bool]int64{true: 201, false: 301}[index < len(reasons)]) {
			t.Fatal("root/parent/sibling reasons or IDs drifted")
		}
		if wantNodes == 3 && (chain[2].Comment != *reasons[index] || chain[2].EntityId == nil || *chain[2].EntityId != uint64(id)) {
			t.Fatal("valid non-White_Space local reason was discarded")
		}
		for _, node := range chain {
			if node.Kind != "auditReason" || node.Name != "Customer" || node.EntityType != "Customer" {
				t.Fatal("responsibility node lost its typed entity")
			}
		}
		if !reflect.DeepEqual(request.TraceChain(), chain) || !reflect.DeepEqual(e.audit.events[index].TraceChain, chain) ||
			e.audit.events[index].Entity != "Customer" || e.audit.events[index].TargetID == nil || *e.audit.events[index].TargetID != core.ValI64(id) {
			t.Fatal("actual command or independent committed audit identity/lineage changed")
		}
		for offset, compiled := range []*tsql.CompiledQuery{probe.writes[index], probe.reads[index]} {
			fact := owner.statements[index*2+offset]
			operation, tail := ds.OpInsert, "insert"
			if offset == 1 {
				operation, tail = ds.OpQuery, "select"
			}
			if fact.Operation != operation || fact.ExecutionOutcome != "success" || fact.ParameterizedSQL != compiled.Sql ||
				!reflect.DeepEqual(fact.Parameters, compiled.Params) || !reflect.DeepEqual(fact.MutationLineage, chain) ||
				fact.Comment == nil || *fact.Comment != rootReason || fact.AuditReason == nil || *fact.AuditReason != rootReason ||
				len(fact.TraceChain) != 4 || fact.TraceChain[2].Name != "sqlite" || fact.TraceChain[3].Name != tail {
				t.Fatal("actual SQLite bindings, lineage, intent or write/readback route changed")
			}
		}
		var version int64
		if err := e.db.QueryRow("SELECT version FROM customer_data WHERE id=?", id).Scan(&version); err != nil || version != 1 {
			t.Fatal("committed customer row missing or version changed", err)
		}
	}
	wantLogs := 0
	if logging {
		wantLogs = 2 * count
	}
	if len(e.logs.entries) != wantLogs {
		t.Fatal("logging mode changed diagnostic count")
	}
	for index, entry := range e.logs.entries {
		if !reflect.DeepEqual(entry.MutationLineage, expected[index/2]) {
			t.Fatal("diagnostic SQL lost inherited/sibling lineage")
		}
	}
	if !logging && e.output.Len() != 0 {
		t.Fatal("logging-off emitted diagnostic text")
	}
	observed, err := json.Marshal(map[string]any{"logging": logging, "inherit": inherit, "commands": commands.commands,
		"physical": owner.statements, "committedAudit": e.audit.events})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("NATIVE_LOCAL_REASON_OBSERVED " + string(observed))
}

func assertBlankReasonError(t *testing.T, err error) {
	t.Helper()
	var intent *core.RequestIntentError
	if !errors.As(err, &intent) || intent.Code != "REQUEST_COMMENT_REQUIRED" || intent.Field != "comment" || intent.RequestKind != "mutation" {
		t.Fatalf("wrong public blank-comment error: %v", err)
	}
}

func TestNativeBlankLocalReasonsInheritAtCommandSQLAndCommittedAudit(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	values := []string{"", " \t\r\n", "\u0085", "\u00a0", "\u2003"}
	reasons := []*string{nil}
	for index := range values {
		reasons = append(reasons, &values[index])
	}
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) { testNativeLocalReasons(t, reasons, true, logging) })
	}
}

func TestNativeNonContractWhitespaceLocalReasonsSurviveAtRealSinks(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	values := []string{"\u001c", "\u001d", "\u001e", "\u001f"}
	var reasons []*string
	for index := range values {
		reasons = append(reasons, &values[index])
	}
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) { testNativeLocalReasons(t, reasons, false, logging) })
	}
}

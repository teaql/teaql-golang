package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	tsql "github.com/teaql/teaql-golang/sql"
)

// Only the first real transaction's readback transport is faulted. Statements,
// requests, results and metadata otherwise use the ordinary provider pipeline.
type throwingReadbackTransport struct {
	tsql.SqlTransactionTransport
	failure            error
	failNext           bool
	writes, reads      []*tsql.CompiledQuery
	commits, rollbacks int
	lastWriteRows      uint64
}

func (p *throwingReadbackTransport) BeginSql(ctx context.Context) (tsql.SqlTransactionTransportTx, error) {
	tx, err := p.SqlTransactionTransport.BeginSql(ctx)
	if err != nil {
		return nil, err
	}
	return &throwingReadbackTx{SqlTransactionTransportTx: tx, owner: p}, nil
}

type throwingReadbackTx struct {
	tsql.SqlTransactionTransportTx
	owner *throwingReadbackTransport
}

func (tx *throwingReadbackTx) ExecuteSql(ctx context.Context, query *tsql.CompiledQuery) (uint64, error) {
	rows, err := tx.SqlTransactionTransportTx.ExecuteSql(ctx, query)
	tx.owner.writes = append(tx.owner.writes, query)
	tx.owner.lastWriteRows = rows
	return rows, err
}

func (tx *throwingReadbackTx) FetchAllSql(ctx context.Context, query *tsql.CompiledQuery) ([]core.Record, error) {
	tx.owner.reads = append(tx.owner.reads, query)
	if tx.owner.failNext {
		tx.owner.failNext = false
		return nil, tx.owner.failure
	}
	return tx.SqlTransactionTransportTx.FetchAllSql(ctx, query)
}

func (tx *throwingReadbackTx) CommitSql(ctx context.Context) error {
	tx.owner.commits++
	return tx.SqlTransactionTransportTx.CommitSql(ctx)
}

func (tx *throwingReadbackTx) RollbackSql(ctx context.Context) error {
	tx.owner.rollbacks++
	return tx.SqlTransactionTransportTx.RollbackSql(ctx)
}

func TestNativeReadbackTransportFailureRetainsSuccessfulWrite(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			if !logging {
				e.ctx.DisableSqlLog()
			}
			failure := errors.New("deliberate authoritative readback transport failure")
			probe := &throwingReadbackTransport{
				SqlTransactionTransport: e.executor.Transport.(tsql.SqlTransactionTransport),
				failure:                 failure, failNext: true,
			}
			e.executor.Transport = probe
			reason := "save " + batchPrivate
			intent, err := core.NewMutationIntent(&reason)
			if err != nil {
				t.Fatal(err)
			}
			key := core.NewEntityKey("Customer", core.ValI64(7))
			scope, err := core.MutationScopeForEntity(nil, key, intent, nil)
			if err != nil {
				t.Fatal(err)
			}
			leaf := nativeBatchLeaf(7, batchPrivate, batchCredential).(*ds.InsertMutation)
			leaf.Cmd.TraceChain = scope.Recover()
			request, err := ds.NewMutationRequest(leaf, reason)
			if err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(request)
			result, facts, err := executeGroupedLineage(t, e, request)
			if !errors.Is(err, failure) || result != nil || len(facts) != 2 || len(probe.writes) != 1 || len(probe.reads) != 1 {
				t.Fatalf("normal mutation/readback did not propagate the transport fault: result=%v error=%v facts=%d writes=%d reads=%d", result, err, len(facts), len(probe.writes), len(probe.reads))
			}
			write, read := facts[0], facts[1]
			if probe.lastWriteRows != 1 || write.Operation != ds.OpInsert || write.ExecutionOutcome != "success" || write.AffectedRows == nil || *write.AffectedRows != 1 {
				t.Fatal("readback failure erased the actual successful SQLite INSERT")
			}
			if read.Operation != ds.OpQuery || read.ExecutionOutcome != "failure" || read.ResultCount != nil || read.AffectedRows != nil {
				t.Fatal("failed SELECT must have its own failure and unknown row count, not successful zero rows")
			}
			for index, fact := range facts {
				compiled := []*tsql.CompiledQuery{probe.writes[0], probe.reads[0]}[index]
				if fact.ParameterizedSQL != compiled.Sql || !reflect.DeepEqual(fact.Parameters, compiled.Params) ||
					!reflect.DeepEqual(fact.MutationLineage, request.TraceChain()) || fact.Comment == nil || *fact.Comment != reason ||
					fact.AuditReason == nil || *fact.AuditReason != reason {
					t.Fatal("actual physical SQL, bindings, request-owned intent or lineage changed")
				}
				middle, tail := "entity", "insert"
				if index == 1 {
					middle, tail = "request", "select"
				}
				if len(fact.TraceChain) != 4 || fact.TraceChain[0].Kind != "operation" || fact.TraceChain[0].Name != "Customer" ||
					fact.TraceChain[1].Kind != middle || fact.TraceChain[1].Name != "Customer" ||
					fact.TraceChain[2].Kind != "provider" || fact.TraceChain[2].Name != "sqlite" ||
					fact.TraceChain[3].Kind != "sql" || fact.TraceChain[3].Name != tail {
					t.Fatal("write/readback routes were conflated")
				}
			}
			if probe.commits != 0 || probe.rollbacks != 1 || len(e.audit.events) != 0 {
				t.Fatal("failed readback committed or published an audit")
			}
			after, _ := json.Marshal(request)
			if string(before) != string(after) {
				t.Fatal("fault handling changed caller input")
			}
			wantLogs := 0
			if logging {
				wantLogs = 2
			}
			if len(e.logs.entries) != wantLogs {
				t.Fatal("logging mode changed execution or emitted an unexpected diagnostic")
			}
			for _, entry := range e.logs.entries {
				if entry.Comment == nil || *entry.Comment != "save [REDACTED]" || len(entry.MutationLineage) != 1 ||
					entry.MutationLineage[0].Comment != "save [REDACTED]" {
					t.Fatal("safe failed readback lost lineage or private intent projection")
				}
			}
			encoded, _ := json.Marshal(e.logs.entries)
			if strings.Contains(string(encoded)+e.output.String(), batchPrivate) || strings.Contains(string(encoded)+e.output.String(), batchCredential) {
				t.Fatal("safe failed readback leaked a private value")
			}
			if rows := queryNativeCustomers(t, e); len(rows) != 0 {
				t.Fatal("successful INSERT survived the graph rollback")
			}
			nextReason := "independent " + batchPrivate
			next, err := ds.NewMutationRequest(nativeBatchLeaf(8, "different private value", ""), nextReason)
			if err != nil {
				t.Fatal(err)
			}
			if err := e.save(t, next); err != nil {
				t.Fatal(err)
			}
			if probe.commits != 1 || probe.rollbacks != 1 || len(e.audit.events) != 1 || *e.audit.events[0].AuditReason != nextReason {
				t.Fatal("failed graph contaminated the following committed save")
			}
			if logging && *e.logs.entries[len(e.logs.entries)-1].Comment != nextReason {
				t.Fatal("failed graph privacy contaminated the following readback")
			}
			rows := queryNativeCustomers(t, e)
			if len(rows) != 1 || rows[0]["id"].V != int64(8) || rows[0]["display_name"].V != "different private value" {
				t.Fatal("following independent save did not persist its own row")
			}
			t.Log("actual INSERT success -> transport readback error -> rollback/no audit -> independent committed save")
		})
	}
}

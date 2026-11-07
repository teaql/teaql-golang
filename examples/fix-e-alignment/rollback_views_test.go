package fixe

// Fault injection wraps real SQLite transactions. Only direct SQL reads are used
// as storage oracles; creation, loading, changes and retries use generated APIs.
import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

var queryViewInjectedFailure = errors.New("controlled query-view transaction failure")

type queryViewFaultProvider struct {
	*guardedExampleProvider
	mode string
}

type queryViewFaultTransaction struct {
	*guardedExampleTransaction
	owner *queryViewFaultProvider
}

func (p *queryViewFaultProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	tx, err := p.guardedExampleProvider.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &queryViewFaultTransaction{tx.(*guardedExampleTransaction), p}, nil
}

func (tx *queryViewFaultTransaction) Mutate(ctx stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	result, err := tx.guardedExampleTransaction.Mutate(ctx, request)
	if err == nil && tx.owner.mode == "after-write" {
		return nil, queryViewInjectedFailure
	}
	return result, err
}

func (tx *queryViewFaultTransaction) Commit(ctx stdcontext.Context) error {
	if tx.owner.mode == "before-commit" {
		tx.owner.probeCommitAttempt()
		return queryViewInjectedFailure
	}
	return tx.guardedExampleTransaction.Commit(ctx)
}

func (p *queryViewFaultProvider) probeCommitAttempt() { p.commits++ }

func TestGeneratedRollbackQueryViews(t *testing.T) {
	for _, mode := range []string{"after-write", "before-commit"} {
		t.Run(mode, func(t *testing.T) {
			base := os.Getenv("TEAQL_FIX_E_DB")
			if base == "" {
				t.Fatal("owned SQLite path required")
			}
			t.Setenv("TEAQL_FIX_E_DB", base+".rollback-"+mode)
			f := newProjectionFixture(t)
			created, createdCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a rollback query-view fixture").Purpose("Retain query output metadata after actual database rollback").NewEntity(f.ctx)
			if createdCreationErr != nil {
				t.Fatal(createdCreationErr)
			}
			created.UpdateContactPhone("5550800").UpdateInternalNote(nil)
			if _, err := created.AuditAs("Create the rollback fixture through generated mutation").Save(f.ctx); err != nil {
				t.Fatal(err)
			}
			request := lib.Q.PurchaseOrdersMinimal().WithIdIs(created.Id()).SelectContactPhone().SelectInternalNote().SelectCreateTime().SelectUpdateTime().
				SelectTenantWith(lib.Q.TenantsMinimal().Limit(1).Comment("Retain the sparse selected root identity")).Limit(1).
				Comment("Load a complete mutable row with readonly query output")
			unchangedAliases(request.GetQuery())
			row, err := request.Purpose("A failed save must not revoke a prior query view").ExecuteForOne(f.ctx)
			if err != nil || row == nil {
				t.Fatal("missing rollback view", err)
			}
			unchangedProperties(t, row)
			properties := row.Base().QueryProperties()
			key, ledger := row.EntityKey(), row.EntityRoot()
			f.ctx.WithBusinessClock(runtime.NewFixedBusinessClock(f.second))
			row.UpdateContactPhone("5550801")
			fault := &queryViewFaultProvider{f.probe, mode}
			f.ctx.InsertResource("dataService", fault)
			before := f.counts()
			_, failure := row.AuditAs("Rollback a controlled failed query-view write").Save(f.ctx)
			if !errors.Is(failure, queryViewInjectedFailure) {
				t.Fatal("expected the actual injected provider failure", failure)
			}
			effects := f.counts()
			for name := range effects {
				effects[name] -= before[name]
			}
			want := map[string]int{"begins": 1, "queries": 0, "mutations": 1, "commits": 0, "rollbacks": 1, "audits": 0, "policies": 1}
			if mode == "before-commit" {
				want["commits"] = 1
			}
			if !reflect.DeepEqual(effects, want) {
				t.Fatal("wrong rollback boundary", effects, want)
			}
			var storedPhone string
			var storedVersion int64
			db := f.ctx.GetResource("db").(*sql.DB)
			if err := db.QueryRow("SELECT contact_phone,version FROM purchase_order_data WHERE id=?", row.Id()).Scan(&storedPhone, &storedVersion); err != nil {
				t.Fatal(err)
			}
			version, known := ledger.OriginalVersion(key)
			pending := ledger.Change(key)
			retained := reflect.DeepEqual(properties, row.Base().QueryProperties())
			body, _ := projectionJSON(t, row)
			_, nullPresent := body["_null"]
			observation := map[string]any{"round": f.round, "case": mode, "id": row.Id(), "effects": effects, "storedPhone": storedPhone, "storedVersion": storedVersion,
				"viewVersion": row.Version(), "originalVersion": version, "originalKnown": known, "pendingPhone": pending["contact_phone"].V,
				"queryPropertiesPreserved": retained, "selectedNullPresent": nullPresent, "passed": retained}
			encoded, _ := json.Marshal(observation)
			fmt.Println("GO_ROLLBACK_VIEW_CASE=" + string(encoded))
			if storedPhone != "5550800" || storedVersion != 1 || row.Version() != 1 || !known || version != 1 || pending["contact_phone"].V != "5550801" {
				t.Fatal("rollback lost storage, version or retry intent", observation)
			}
			if !retained || !nullPresent {
				t.Fatal("rollback lost output-only query properties", observation)
			}
			unchangedProperties(t, row)
			_, _, missing := lib.E.PurchaseOrder(row).Tenant().Name().TryEval()
			projectionMissing(t, missing, "/tenant/name")
			if id, ok := lib.E.PurchaseOrder(row).Tenant().Id().Eval(); !ok || id != 1 {
				t.Fatal("rollback lost loaded relation identity")
			}
			fault.mode = ""
			before = f.counts()
			saved, err := row.AuditAs("Retry the exact rolled-back view without reloading").Save(f.ctx)
			if err != nil {
				t.Fatal(err)
			}
			if saved != row || row.EntityRoot() != ledger || row.Version() != 2 || len(row.Base().QueryProperties()) != 0 || len(ledger.Change(key)) != 0 {
				t.Fatal("retry did not return fresh authoritative state")
			}
			for _, name := range []string{"begins", "mutations", "commits", "audits", "policies"} {
				if f.counts()[name] != before[name]+1 {
					t.Fatal("retry skipped the real commit boundary", name)
				}
			}
			if err := db.QueryRow("SELECT contact_phone,version FROM purchase_order_data WHERE id=?", row.Id()).Scan(&storedPhone, &storedVersion); err != nil {
				t.Fatal(err)
			}
			if storedPhone != "5550801" || storedVersion != 2 {
				t.Fatal("retry did not persist its original intent")
			}
			encoded, _ = json.Marshal(map[string]any{"round": f.round, "case": mode, "id": row.Id(), "phone": storedPhone, "version": storedVersion, "staleProperties": len(row.Base().QueryProperties()), "passed": true})
			fmt.Println("GO_ROLLBACK_VIEW_RETRY=" + string(encoded))
		})
	}
}

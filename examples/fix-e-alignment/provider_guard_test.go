package fixe

import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// This controlled framework probe preserves SQLite's real transactional ID and
// floor capabilities. It counts entry before the actual business provider.
type guardedExampleProvider struct{ *restorationProvider }
type guardedExampleTransaction struct {
	*restorationTransaction
	ids   runtime.InternalIdGenerator
	floor runtime.InternalIdFloorReserver
}

func (p *guardedExampleProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	p.begins++
	tx, err := p.transaction.Begin(ctx)
	if err != nil {
		return nil, err
	}
	ids, hasIDs := tx.(runtime.InternalIdGenerator)
	floor, hasFloor := tx.(runtime.InternalIdFloorReserver)
	if !hasIDs || !hasFloor {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("SQLite fixture lost transactional ID capabilities")
	}
	return &guardedExampleTransaction{restorationTransaction: &restorationTransaction{Transaction: tx, probe: p.restorationProvider}, ids: ids, floor: floor}, nil
}
func (tx *guardedExampleTransaction) GenerateId(entity string) (uint64, error) {
	return tx.ids.GenerateId(entity)
}
func (tx *guardedExampleTransaction) EnsureIdFloor(ctx stdcontext.Context, entity string, floor uint64) error {
	return tx.floor.EnsureIdFloor(ctx, entity, floor)
}

type lateExampleIDs struct {
	delegate runtime.InternalIdGenerator
	calls    int
}

func (ids *lateExampleIDs) GenerateId(entity string) (uint64, error) {
	ids.calls++
	if ids.delegate == nil {
		return 0, fmt.Errorf("creation fixture requires the selected context allocator")
	}
	return ids.delegate.GenerateId(entity)
}

// ID reservation no longer runs inside save. Exercise an actual late SPI at
// transaction admission instead, retaining the original no-write/retry guard.
type beginCallbackProvider struct {
	data_service.TransactionExecutor
	afterBegin func()
}

func (p *beginCallbackProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	tx, err := p.TransactionExecutor.Begin(ctx)
	if err == nil && p.afterBegin != nil {
		p.afterBegin()
	}
	return tx, err
}

func TestGeneratedProviderGuardRejectsLateSPIMutationAndRetriesSameObject(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-provider-guard-%d?mode=memory&cache=shared", os.Getpid())
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := ctx.GetResource("db").(*sql.DB)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, 4, 3, 9, 0, 0, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	row, rowCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a synthetic late-SPI fixture").Purpose("Prove actual provider requests cannot escape review").NewEntity(ctx)
	if rowCreationErr != nil {
		t.Fatal(rowCreationErr)
	}
	row.UpdateContactPhone("5550910").UpdateInternalNote(nil)
	original := ctx.GetResource("dataService")
	late := &beginCallbackProvider{TransactionExecutor: original.(data_service.TransactionExecutor), afterBegin: func() { row.UpdateContactPhone("5550911") }}
	probe := &guardedExampleProvider{&restorationProvider{QueryExecutor: original.(data_service.QueryExecutor), transaction: late}}
	ctx.InsertResource("dataService", probe)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	ids := &lateExampleIDs{delegate: ctx.InternalIdGenerator()}
	ctx.WithInternalIdGenerator(ids)
	reservedID := row.Id()
	if reservedID == 0 {
		t.Fatal("creation did not reserve its identity")
	}
	var before int
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data WHERE contact_phone IN ('5550910','5550911')").Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, err = row.AuditAs("Reject a custom SPI changing business values after bridge replay").Save(ctx)
	var rejected *runtime.RuntimeError
	if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "mutation_changed_after_preparation" || rejected.CheckResults[0].InstancePath() != "/contactPhone" {
		t.Fatal("late SPI edit did not produce a precise provider-boundary failure", err)
	}
	if ids.calls != 0 || row.Id() != reservedID || probe.begins != 1 || probe.mutations != 0 || probe.commits != 0 || probe.rollbacks != 1 || audit.events != 0 {
		t.Fatal("late mutation reached business storage or committed audit", ids.calls, *probe.restorationProvider, audit.events)
	}
	var absent int
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data WHERE contact_phone IN ('5550910','5550911')").Scan(&absent); err != nil {
		t.Fatal(err)
	}
	// Prior unreset rounds may own the same synthetic phone. This failed attempt
	// must not change that count; the later retry must add exactly one row.
	if absent != before {
		t.Fatal("failed provider-boundary attempt inserted a domain row", before, absent)
	}
	rejection := map[string]int{"begins": probe.begins, "mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks, "auditEvents": audit.events, "allocatorCalls": ids.calls}
	probe.begins, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0
	late.afterBegin = nil
	row, err = row.AuditAs("Review and retry the preserved late SPI edit").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ids.calls != 0 || row.Id() != reservedID || probe.begins != 1 || probe.mutations != 1 || probe.commits != 1 || probe.rollbacks != 0 || audit.events != 1 {
		t.Fatal("retry did not use a fresh governed transaction", ids.calls, *probe.restorationProvider, audit.events)
	}
	loaded, err := lib.Q.PurchaseOrders().WithIdIs(row.Id()).Limit(1).Comment("Reload the approved SPI retry").Purpose("Verify generated Q/E agrees with physical committed values").ExecuteForOne(ctx)
	if err != nil || loaded == nil {
		t.Fatal("retry row missing", err)
	}
	phone, present := lib.E.PurchaseOrder(loaded).ContactPhone().Eval()
	version, versionLoaded := lib.E.PurchaseOrder(loaded).Version().Eval()
	if !present || !versionLoaded || phone != "5550911" || version != 1 {
		t.Fatal("Q/E disagrees with reviewed retry", phone, version)
	}
	var physicalPhone string
	var physicalVersion int64
	if err := db.QueryRow("SELECT contact_phone,version FROM purchase_order_data WHERE id=?", row.Id()).Scan(&physicalPhone, &physicalVersion); err != nil {
		t.Fatal(err)
	}
	if physicalPhone != phone || physicalVersion != version {
		t.Fatal("physical row differs from Q/E")
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data WHERE contact_phone IN ('5550910','5550911')").Scan(&absent); err != nil || absent != before+1 {
		t.Fatal("failed attempt leaked a domain row", err, before, absent)
	}
	frame := map[string]any{"language": "go", "passed": true, "id": row.Id(), "version": version, "phone": phone, "rejection": rejection, "sameObjectRetry": true, "returnedEEqualsPhysical": true, "lateCallback": "transaction.begin", "reservedIdentityPreserved": row.Id() == reservedID, "saveAllocations": ids.calls}
	bytes, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_PREPARED_PROVIDER=" + string(bytes))
}

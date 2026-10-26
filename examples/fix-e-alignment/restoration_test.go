package fixe

import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// This probe wraps the existing provider only after creation and loading. It
// counts actual provider entry, rather than inferring no-op behavior from SQL
// text or from a generated dirty flag. The original provider owns the database.
type restorationProvider struct {
	data_service.QueryExecutor
	transaction                                    data_service.TransactionExecutor
	begins, queries, mutations, commits, rollbacks int
}

func (p *restorationProvider) Query(ctx stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.queries++
	return p.QueryExecutor.Query(ctx, request)
}

func (p *restorationProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	p.begins++
	tx, err := p.transaction.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &restorationTransaction{Transaction: tx, probe: p}, nil
}

type restorationTransaction struct {
	data_service.Transaction
	probe *restorationProvider
}

func (tx *restorationTransaction) Query(ctx stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	tx.probe.queries++
	return tx.Transaction.Query(ctx, request)
}
func (tx *restorationTransaction) Mutate(ctx stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	tx.probe.mutations++
	return tx.Transaction.Mutate(ctx, request)
}
func (tx *restorationTransaction) Commit(ctx stdcontext.Context) error {
	tx.probe.commits++
	return tx.Transaction.Commit(ctx)
}
func (tx *restorationTransaction) Rollback(ctx stdcontext.Context) error {
	tx.probe.rollbacks++
	return tx.Transaction.Rollback(ctx)
}

type restorationAudit struct{ events int }

func (s *restorationAudit) OnSafeEvent(*runtime.UserContext, *runtime.SafeAuditEvent) error {
	s.events++
	return nil
}

func TestGeneratedRestoredLoadedScalarSaveIsNoop(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-restoration-%d?mode=memory&cache=shared", os.Getpid())
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
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	order := lib.Q.PurchaseOrders().Comment("Prepare a synthetic restoration fixture").
		Purpose("Prove restored values do not enter the mutation pipeline").NewEntity(ctx)
	order.UpdateContactPhone("5550800").UpdateInternalNote(nil)
	saved, err := order.AuditAs("Create an order with a known null note").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := lib.Q.PurchaseOrders().WithIdIs(saved.Id()).Limit(1).
		Comment("Load the authoritative baseline").Purpose("Compare restored values against persisted data").ExecuteForList(ctx)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("baseline missing: %v", err)
	}
	row := rows.Data[0]
	version, present := lib.E.PurchaseOrder(row).Version().Eval()
	if !present {
		t.Fatal("optimistic version was not loaded")
	}
	provider := ctx.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	policyCalls := 0
	ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy {
		policyCalls++
		return nil
	}))
	row.UpdateContactPhone("5550801").UpdateContactPhone("5550800")
	note := "temporary note"
	row.UpdateInternalNote(&note).UpdateInternalNote(nil)
	_, saveErr := row.AuditAs("Save a graph restored to its loaded baseline").Save(ctx)
	var storedPhone string
	var storedVersion int64
	var storedNote sql.NullString
	if err := db.QueryRow("SELECT contact_phone, internal_note, version FROM purchase_order_data WHERE id = ?", saved.Id()).Scan(&storedPhone, &storedNote, &storedVersion); err != nil {
		t.Fatal(err)
	}
	state := map[string]any{"language": "go", "case": "known-loaded-baseline-restored", "id": saved.Id(),
		"versionBefore": version, "versionAfter": storedVersion, "phoneAfter": storedPhone, "nullAfter": !storedNote.Valid,
		"providerBegins": probe.begins, "providerQueries": probe.queries, "providerMutations": probe.mutations,
		"providerCommits": probe.commits, "providerRollbacks": probe.rollbacks, "policyCalls": policyCalls, "auditEvents": audit.events,
		"saveSucceeded": saveErr == nil}
	passed := saveErr == nil && storedVersion == version && storedPhone == "5550800" && !storedNote.Valid &&
		probe.begins == 0 && probe.queries == 0 && probe.mutations == 0 && probe.commits == 0 && probe.rollbacks == 0 && policyCalls == 0 && audit.events == 0
	state["passed"] = passed
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_MUTATION_RESTORATION=" + string(data))
	if !passed {
		t.Fatalf("restored save was not inert: %s; error=%v", data, saveErr)
	}
}

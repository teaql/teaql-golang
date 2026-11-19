package fixe

// This controlled adapter regression intentionally exercises AttachEntityRoot
// and the runtime ledger, not a suggested application API-discovery shortcut.
// Business loading, edits and persistence use the retained generated Q/E/save
// contract. Never read or hand-edit generated library implementation here.
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

func TestGeneratedLedgerCompositionPreservesBothOwnersAndRejectsBeforeIO(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-ledger-composition-%d?mode=memory&cache=shared", os.Getpid())
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database)
	context, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := context.GetResource("db").(*sql.DB)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	context.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, time.April, 1, 9, 0, 0, 0, time.UTC)))
	if err := lib.EnsureSchema(context); err != nil {
		t.Fatal(err)
	}
	row := lib.Q.PurchaseOrders().Comment("Create a synthetic ledger-composition fixture").Purpose("Verify typed graph-version boundaries").NewEntity(context)
	row.UpdateContactPhone("5550870").UpdateInternalNote(nil)
	row, err = row.AuditAs("Create the versioned composition fixture").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	load := func() *purchase_order.PurchaseOrder {
		row, err := lib.Q.PurchaseOrders().WithIdIs(row.Id()).Limit(1).Comment("Load a bounded independent order snapshot").Purpose("Compare original versions without sharing mutation ownership").ExecuteForOne(context)
		if err != nil || row == nil {
			t.Fatalf("missing order snapshot: %v", err)
		}
		return row
	}
	version := func(entity *purchase_order.PurchaseOrder) int64 {
		v, loaded := lib.E.PurchaseOrder(entity).Version().Eval()
		if !loaded {
			t.Fatal("missing optimistic version")
		}
		return v
	}
	stale, current := load(), load()
	if version(stale) != 1 || version(current) != 1 || stale.EntityRoot() == current.EntityRoot() {
		t.Fatal("query roots or initial versions were not independent")
	}
	provider := context.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	context.InsertResource("dataService", probe)
	audit, policyCalls := &restorationAudit{}, 0
	context.WithAppAuditEventSink(audit).WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policyCalls++; return nil }))
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0, 0
		audit.events, policyCalls = 0, 0
	}
	counts := func(writes int) map[string]int {
		if probe.begins != writes || probe.queries != 0 || probe.mutations != writes || probe.commits != writes || probe.rollbacks != 0 || audit.events != writes || policyCalls != writes {
			t.Fatalf("unexpected boundary: begin=%d query=%d mutate=%d commit=%d rollback=%d audit=%d policy=%d", probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, audit.events, policyCalls)
		}
		return map[string]int{"providerBegins": probe.begins, "providerQueries": probe.queries, "providerMutations": probe.mutations, "providerCommits": probe.commits, "providerRollbacks": probe.rollbacks, "auditEvents": audit.events, "policyCalls": policyCalls}
	}
	check := func(expectedVersion int64, expectedNote *string) {
		var phone string
		var note sql.NullString
		var actual int64
		if err := db.QueryRow("SELECT contact_phone,internal_note,version FROM purchase_order_data WHERE id=?", row.Id()).Scan(&phone, &note, &actual); err != nil {
			t.Fatal(err)
		}
		if phone != "5550871" || actual != expectedVersion || note.Valid != (expectedNote != nil) || note.Valid && note.String != *expectedNote {
			t.Fatalf("incorrect durable row: phone=%s note=%v version=%d", phone, note, actual)
		}
	}
	current.UpdateContactPhone("5550871")
	current, err = current.AuditAs("Calibrate an actual generated update").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	calibration := counts(1)
	check(2, nil)
	source := load()
	sourceRoot := source.EntityRoot()
	sourceNote := "independent pending note"
	source.UpdateInternalNote(&sourceNote)
	stale.UpdateContactPhone("stale pending")
	sourceIntent, targetIntent := sourceRoot.Change(source.EntityKey()), stale.EntityRoot().Change(stale.EntityKey())
	source.AttachEntityRoot(stale.EntityRoot())
	if source.EntityRoot() != sourceRoot || !reflect.DeepEqual(sourceRoot.Change(source.EntityKey()), sourceIntent) || !reflect.DeepEqual(stale.EntityRoot().Change(stale.EntityKey()), targetIntent) {
		t.Fatal("rejected attachment replaced an owner or copied/consumed its pending intent")
	}
	reset()
	_, err = stale.AuditAs("Reject conflicting versions before graph effects").Save(context)
	var rejected *runtime.RuntimeError
	if !errors.As(err, &rejected) || rejected.Type != "Check" || len(rejected.CheckResults) != 1 {
		t.Fatal("missing structured composition rejection", err)
	}
	result := rejected.CheckResults[0]
	if result.RuleID != "conflicting_original_version" || result.EntityType != "purchase_order" || result.InstancePath() != "/version" || result.InputValue != int64(2) || result.SystemValue != int64(1) {
		t.Fatalf("incorrect typed conflict: %+v", result)
	}
	rejection := counts(0)
	check(2, nil)
	reset()
	source, err = source.AuditAs("Save the independent source after failed composition").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	counts(1)
	check(3, &sourceNote)
	left, right := load(), load()
	mergedNote := "merged compatible pending"
	left.UpdateContactPhone("temporary restored phone").UpdateContactPhone("5550871")
	right.UpdateInternalNote(&mergedNote)
	left.EntityRoot().MergeFrom(right.EntityRoot())
	reset()
	left, err = left.AuditAs("Persist compatible raw-ledger intent alongside a restored generated edit").Save(context)
	if err != nil {
		t.Fatal(err)
	}
	counts(1)
	check(4, &mergedNote)
	value, loaded := lib.E.PurchaseOrder(left).InternalNote().Eval()
	if !loaded || value == nil || *value != mergedNote {
		t.Fatal("returned E view lost the imported pending value")
	}
	if version(stale) != 1 || version(left) != 4 {
		t.Fatal("independent view or committed version was replaced")
	}
	if right.EntityRoot().Change(right.EntityKey())["internal_note"] != core.ValText(mergedNote) {
		t.Fatal("saving the composed target consumed the independent source's intent")
	}
	frame := map[string]any{"language": "go", "id": row.Id(), "targetVersion": 1, "sourceVersion": 2, "finalVersion": version(left), "calibration": calibration, "rejection": rejection,
		"sourceOwnerPreserved": true, "bothIntentsPreserved": true, "independentSourceSaved": true, "compatibleRawAndRestoredSave": true, "passed": true}
	bytes, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_LEDGER_COMPOSITION=" + string(bytes))
}

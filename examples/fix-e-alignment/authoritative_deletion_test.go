package fixe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// A database trigger deliberately makes the tombstone different from -(v+1).
// The second deletion must never behave like the separate Recover operation.
func TestGeneratedAuthoritativeTombstoneAndRepeatedDeletion(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-tombstone-%d?mode=memory&cache=shared", os.Getpid())
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
	clock := time.Date(2032, 4, 5, 9, 0, 0, 0, time.UTC)
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(clock))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	var frames []map[string]any
	for _, mode := range []string{"ordinary", "trigger"} {
		t.Run(mode, func(t *testing.T) {
			row := lib.Q.PurchaseOrders().Comment("Prepare an authoritative-deletion fixture").Purpose("Verify deleted state is read from storage and cannot be toggled into recovery").NewEntity(ctx)
			phone := "5550930"
			if mode == "trigger" {
				phone = "5550931"
			}
			row.UpdateContactPhone(phone).UpdateInternalNote(nil)
			row, err = row.AuditAs("Create the synthetic tombstone fixture").Save(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "trigger" {
				_, err := db.Exec(`CREATE TRIGGER IF NOT EXISTS core_ml13_tombstone
AFTER UPDATE OF version ON purchase_order_data
WHEN NEW.version < 0 AND OLD.version > 0 AND NEW.contact_phone = '5550931'
BEGIN
  UPDATE purchase_order_data SET version = -900 - NEW.id, internal_note = 'database tombstone', update_time = 1964854800000 WHERE id = NEW.id;
END`)
				if err != nil {
					t.Fatal(err)
				}
			}
			original := ctx.GetResource("dataService")
			probe := &guardedExampleProvider{&restorationProvider{QueryExecutor: original.(data_service.QueryExecutor), transaction: original.(data_service.TransactionExecutor)}}
			ctx.InsertResource("dataService", probe)
			defer ctx.InsertResource("dataService", original)
			audit := &restorationAudit{}
			ctx.WithAppAuditEventSink(audit)
			row, err = row.MarkForDeletion().AuditAs("Persist one deletion and return its authoritative tombstone").Save(ctx)
			if err != nil {
				t.Fatal(err)
			}
			version, loaded := lib.E.PurchaseOrder(row).Version().Eval()
			expected := int64(-2)
			if mode == "trigger" {
				expected = -900 - int64(row.Id())
			}
			if !loaded || version != expected || probe.begins != 1 || probe.mutations != 1 || probe.commits != 1 || probe.rollbacks != 0 || audit.events != 1 || row.IsMarkedAsDelete() {
				t.Fatal("Save did not accept the authoritative tombstone", version, expected, *probe.restorationProvider, audit.events)
			}
			if original, ok := row.EntityRoot().OriginalVersion(row.EntityKey()); !ok || original != version {
				t.Fatal("authoritative optimistic baseline was not retained", original, version)
			}
			note, notePresent, noteErr := lib.E.PurchaseOrder(row).InternalNote().TryEval()
			updated, timeLoaded := lib.E.PurchaseOrder(row).UpdateTime().Eval()
			if noteErr != nil || !timeLoaded {
				t.Fatal("returned tombstone lost selected field state")
			}
			if mode == "ordinary" && (note != nil || notePresent) {
				t.Fatal("known null became a value", note)
			}
			if mode == "trigger" && (!notePresent || note == nil || *note != "database tombstone" || updated.UnixMilli() != 1964854800000) {
				t.Fatal("Save returned computed state instead of trigger effects", note, updated)
			}
			hidden, err := lib.Q.PurchaseOrders().WithIdIs(row.Id()).Limit(1).Comment("Check ordinary visibility after deletion").Purpose("Verify deletion does not return a live row").ExecuteForList(ctx)
			if err != nil || len(hidden.Data) != 0 {
				t.Fatal("deleted row is ordinarily visible", err)
			}
			deleted, err := lib.Q.PurchaseOrders().WithIdIs(row.Id()).DeletedRowsOnly().Limit(1).Comment("Read the stored tombstone explicitly").Purpose("Compare returned E state with a fresh generated query").ExecuteForOne(ctx)
			if err != nil || deleted == nil {
				t.Fatal("authoritative deleted row cannot be queried", err)
			}
			fresh, freshLoaded := lib.E.PurchaseOrder(deleted).Version().Eval()
			if !freshLoaded || fresh != version {
				t.Fatal("returned and queried tombstones differ", version, fresh)
			}
			probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, audit.events = 0, 0, 0, 0, 0, 0
			_, repeatErr := row.MarkForDeletion().AuditAs("Reject deleting the same tombstone again without an explicit recovery").Save(ctx)
			var physical int64
			if err := db.QueryRow("SELECT version FROM purchase_order_data WHERE id=?", row.Id()).Scan(&physical); err != nil {
				t.Fatal(err)
			}
			if physical != version || physical >= 0 || audit.events != 0 || probe.begins != 0 || probe.queries != 0 || probe.mutations != 0 || probe.commits != 0 || probe.rollbacks != 0 {
				t.Fatal("repeated deletion changed or revived a tombstone", repeatErr, "before", version, "after", physical, "mutations", probe.mutations, "commits", probe.commits, "audits", audit.events)
			}
			if repeatErr == nil {
				t.Fatal("an already-deleted root must require an explicit recovery, not another deletion")
			}
			frames = append(frames, map[string]any{"case": mode, "id": row.Id(), "version": version, "storedVersion": physical,
				"triggerEffectsPreserved": mode != "trigger" || note != nil && *note == "database tombstone", "repeatRejected": true,
				"repeatBegins": probe.begins, "repeatQueries": probe.queries, "repeatMutations": probe.mutations, "repeatCommits": probe.commits, "repeatRollbacks": probe.rollbacks, "repeatAudits": audit.events})
		})
	}
	if !t.Failed() {
		value, err := json.Marshal(map[string]any{"language": "go", "passed": true, "cases": frames})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("CORE_AUTHORITATIVE_TOMBSTONE=" + string(value))
	}
}

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
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// Keep SQLite's ID and floor capabilities while failing a selected business
// command. Begin/Commit/Rollback and actual provider entry stay observable.
type cancellationProvider struct {
	*guardedExampleProvider
	failAt int
}
type cancellationTransaction struct {
	*guardedExampleTransaction
	owner *cancellationProvider
}

func (p *cancellationProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	tx, err := p.guardedExampleProvider.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &cancellationTransaction{tx.(*guardedExampleTransaction), p}, nil
}
func (tx *cancellationTransaction) Mutate(ctx stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	if tx.owner.failAt != 0 && tx.owner.mutations+1 == tx.owner.failAt {
		tx.owner.mutations++
		return nil, errors.New("controlled cancellation-fixture second-command failure")
	}
	return tx.guardedExampleTransaction.Mutate(ctx, request)
}

func TestGeneratedCancelledNewRootAndChildLifecycle(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-cancelled-graph-%d?mode=memory&cache=shared", os.Getpid())
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
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, 4, 4, 9, 0, 0, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	original := ctx.GetResource("dataService")
	probe := &cancellationProvider{guardedExampleProvider: &guardedExampleProvider{&restorationProvider{
		QueryExecutor: original.(data_service.QueryExecutor), transaction: original.(data_service.TransactionExecutor),
	}}}
	ctx.InsertResource("dataService", probe)
	ids := &lateExampleIDs{context: ctx}
	ctx.WithInternalIdGenerator(ids)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks = 0, 0, 0, 0, 0
		ids.calls, audit.events, probe.failAt = 0, 0, 0
	}
	counts := func() map[string]int {
		return map[string]int{"begins": probe.begins, "queries": probe.queries, "mutations": probe.mutations,
			"commits": probe.commits, "rollbacks": probe.rollbacks, "ids": ids.calls, "auditEvents": audit.events}
	}
	var frames []map[string]any
	t.Run("cancelled-root-never-enters-provider", func(t *testing.T) {
		row, rowCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a root cancelled before its first save").Purpose("Verify ML-12 rejects non-existent durable roots").NewEntity(ctx)
		if rowCreationErr != nil {
			t.Fatal(rowCreationErr)
		}
		row.UpdateContactPhone("5550920").UpdateInternalNote(nil)
		reset()
		row.MarkForDeletion()
		if got := counts(); got["begins"] != 0 || got["mutations"] != 0 || got["ids"] != 0 {
			t.Fatal("MarkForDeletion performed provider work", got)
		}
		_, err := row.AuditAs("Reject this cancelled new root").Save(ctx)
		if err == nil || probe.begins != 0 || probe.queries != 0 || probe.mutations != 0 || probe.commits != 0 || probe.rollbacks != 0 || ids.calls != 0 || audit.events != 0 {
			t.Fatal("cancelled root did not reject before provider/ID entry", err, counts())
		}
		if !row.IsNew() || !row.IsMarkedAsDelete() || !row.EntityRoot().IsNew(row.EntityKey()) || !row.EntityRoot().IsDeleted(row.EntityKey()) {
			t.Fatal("rejected root lost its cancellation intent")
		}
		frames = append(frames, map[string]any{"case": "cancelled-root", "rejected": true, "counts": counts()})
	})
	for _, variant := range []string{"complete-child", "incomplete-child", "already-composed-child", "rollback-retry"} {
		t.Run(variant, func(t *testing.T) {
			parent, parentCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a parent graph with one cancelled new child").Purpose("Verify cancellation does not reject valid sibling work").NewEntity(ctx)
			if parentCreationErr != nil {
				t.Fatal(parentCreationErr)
			}
			parent.UpdateContactPhone("5550921").UpdateInternalNote(nil)
			cancelled, cancelledCreationErr := lib.Q.OrderItems().Comment("Prepare the child which will never be persisted").Purpose("Verify no Checker or allocation is needed for cancelled data").NewEntity(ctx)
			if cancelledCreationErr != nil {
				t.Fatal(cancelledCreationErr)
			}
			if variant != "incomplete-child" {
				cancelled.UpdateProductName("Cancelled synthetic widget").UpdateUnitPrice(decimal.NewFromInt(10))
			}
			cancelled.MarkForDeletion()
			if variant == "already-composed-child" {
				cancelled.AttachEntityRoot(parent.EntityRoot())
			}
			kept, keptCreationErr := lib.Q.OrderItems().Comment("Prepare the live sibling").Purpose("Verify the parent and intended child persist together").NewEntity(ctx)
			if keptCreationErr != nil {
				t.Fatal(keptCreationErr)
			}
			kept.UpdateProductName("Retained synthetic widget").UpdateUnitPrice(decimal.NewFromInt(12))
			parent.OrderItemList().Add(cancelled)
			parent.OrderItemList().Add(kept)
			reset()
			var rejected map[string]int
			if variant == "rollback-retry" {
				var parentsBefore, childrenBefore int
				if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&parentsBefore); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRow("SELECT COUNT(*) FROM order_item_data").Scan(&childrenBefore); err != nil {
					t.Fatal(err)
				}
				probe.failAt = 2
				_, err := parent.AuditAs("Fail the surviving child after the parent write").Save(ctx)
				if err == nil || probe.begins != 1 || probe.mutations != 2 || probe.commits != 0 || probe.rollbacks != 1 || audit.events != 0 {
					t.Fatal("injected graph failure was not atomic", err, counts())
				}
				var parentsAfter, childrenAfter int
				if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&parentsAfter); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRow("SELECT COUNT(*) FROM order_item_data").Scan(&childrenAfter); err != nil {
					t.Fatal(err)
				}
				if parentsAfter != parentsBefore || childrenAfter != childrenBefore || parent.Id() != 0 || kept.Id() != 0 || !parent.IsNew() || !kept.IsNew() || !cancelled.IsMarkedAsDelete() {
					t.Fatal("failed graph lost retry state or left physical rows", parentsBefore, parentsAfter, childrenBefore, childrenAfter)
				}
				rejected = counts()
				reset()
			}
			saved, err := parent.AuditAs("Save the parent and live sibling, omitting its cancelled child").Save(ctx)
			if err != nil {
				t.Fatal("cancelled child rejected the parent save", err, counts())
			}
			if saved != parent || probe.begins != 1 || probe.mutations != 2 || probe.commits != 1 || probe.rollbacks != 0 || ids.calls != 2 || audit.events != 2 || cancelled.Id() != 0 || !cancelled.IsNew() || !cancelled.IsMarkedAsDelete() {
				t.Fatal("cancelled child was allocated, written or incorrectly committed", counts(), cancelled.Id())
			}
			written := counts()
			loaded, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).SelectOrderItemListWith(lib.Q.OrderItems().Limit(10).Comment("Load persisted live children")).Limit(1).
				Comment("Reload the parent after cancellation").Purpose("Compare generated Q/E with physical child rows").ExecuteForOne(ctx)
			if err != nil || loaded == nil {
				t.Fatal("committed parent missing", err)
			}
			size, present := lib.E.PurchaseOrder(loaded).OrderItemList().Size().Eval()
			if !present || size != 1 || len(loaded.OrderItemList().Items()) != 1 {
				t.Fatal("Q/E did not retain exactly one live child", size, err)
			}
			name, present := lib.E.OrderItem(loaded.OrderItemList().Items()[0]).ProductName().Eval()
			if !present || name != "Retained synthetic widget" {
				t.Fatal("Q/E returned the cancelled sibling", name)
			}
			var rows, version int64
			if err := db.QueryRow("SELECT COUNT(*), MIN(version) FROM order_item_data WHERE purchase_order = ?", parent.Id()).Scan(&rows, &version); err != nil || rows != 1 || version != 1 {
				t.Fatal("physical child rows disagree", rows, version, err)
			}
			reset()
			if _, err := parent.AuditAs("Resave the unchanged surviving graph").Save(ctx); err != nil || probe.begins != 0 || probe.mutations != 0 || ids.calls != 0 || audit.events != 0 {
				t.Fatal("repeated save revived a cancelled child or rewrote unchanged nodes", err, counts())
			}
			frames = append(frames, map[string]any{"case": variant, "id": parent.Id(), "childId": kept.Id(), "childCount": size,
				"cancelledId": cancelled.Id(), "cancelledPending": cancelled.IsNew() && cancelled.IsMarkedAsDelete(), "write": written, "rejected": rejected, "repeat": counts()})
		})
	}
	if !t.Failed() {
		value, err := json.Marshal(map[string]any{"language": "go", "passed": true, "cases": frames})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("CORE_CANCELLED_NEW_GRAPH=" + string(value))
	}
}

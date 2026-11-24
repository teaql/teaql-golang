package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// Deliberately stateful customer Fix: calling it twice would change the value
// after policy review. It is installed only after normal generated creation.
type preparedExampleChecker struct{ calls int }

func (checker *preparedExampleChecker) CheckAndFix(_ *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	if input.Entity == "purchase_order" && input.Operation == core.MutationUpdate {
		checker.calls++
		input.Values["internal_note"] = core.ValText(fmt.Sprintf("prepared-fix-%d", checker.calls))
	}
	return nil
}

type preparedExamplePolicy struct {
	review func(*runtime.MutationPlan) runtime.MutationDecision
}

func (*preparedExamplePolicy) Identity() runtime.MutationPolicyIdentity {
	return runtime.NewMutationPolicyIdentity("prepared-example", "1", "sha256:prepared-example")
}
func (policy *preparedExamplePolicy) Review(_ *runtime.UserContext, plan *runtime.MutationPlan) runtime.MutationDecision {
	return policy.review(plan)
}

func TestGeneratedPreparedSnapshotMatchesReviewAndRejectsLateChanges(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		database = fmt.Sprintf("file:teaql-prepared-snapshot-%d?mode=memory&cache=shared", os.Getpid())
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
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, 4, 2, 9, 0, 0, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	row, rowCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a synthetic reviewed-write fixture").Purpose("Compare custom Fix, policy review and persisted values").NewEntity(ctx)
	if rowCreationErr != nil {
		t.Fatal(rowCreationErr)
	}
	row.UpdateContactPhone("5550890").UpdateInternalNote(nil)
	row, err = row.AuditAs("Create the reviewed-write fixture with the normal generated Checker").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	row, err = lib.Q.PurchaseOrders().WithIdIs(row.Id()).Limit(1).Comment("Load the complete synthetic order").Purpose("Use generated Q/E and audited save for snapshot verification").ExecuteForOne(ctx)
	if err != nil || row == nil {
		t.Fatal("missing prepared snapshot order", err)
	}
	provider := ctx.GetResource("dataService")
	probe := &restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	checker, audit := &preparedExampleChecker{}, &restorationAudit{}
	ctx.WithCheckerRegistry(checker).WithAppAuditEventSink(audit)
	var reviewed []string
	changeLiveIntent := false
	ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy {
		return &preparedExamplePolicy{review: func(plan *runtime.MutationPlan) runtime.MutationDecision {
			reviewed = append(reviewed, plan.Operations[0].ChangedValues["internal_note"].V.(string))
			if changeLiveIntent {
				row.UpdateContactPhone("5550899")
			}
			return runtime.AllowMutation()
		}}
	}))
	check := func(version int64, phone, note string) {
		var actualVersion int64
		var actualPhone, actualNote string
		if err := db.QueryRow("SELECT version,contact_phone,internal_note FROM purchase_order_data WHERE id=?", row.Id()).Scan(&actualVersion, &actualPhone, &actualNote); err != nil {
			t.Fatal(err)
		}
		if actualVersion != version || actualPhone != phone || actualNote != note {
			t.Fatalf("physical row differs from review: version=%d phone=%s note=%s", actualVersion, actualPhone, actualNote)
		}
	}
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, audit.events = 0, 0, 0, 0, 0, 0
	}
	assertCounts := func(writes int) {
		if probe.begins != writes || probe.queries != 0 || probe.mutations != writes || probe.commits != writes || probe.rollbacks != 0 || audit.events != writes {
			t.Fatal("unexpected provider or audit activity", *probe, audit.events)
		}
	}
	counts := func() map[string]int {
		return map[string]int{"begins": probe.begins, "queries": probe.queries, "mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks, "auditEvents": audit.events}
	}
	row.UpdateContactPhone("5550891")
	row, err = row.AuditAs("Persist exactly the first reviewed custom Fix").Save(ctx)
	if err != nil || checker.calls != 1 || len(reviewed) != 1 || reviewed[0] != "prepared-fix-1" {
		t.Fatal("Fix did not execute exactly once before review", err, checker.calls, reviewed)
	}
	assertCounts(1)
	firstCounts := counts()
	check(2, "5550891", reviewed[0])
	note, loaded := lib.E.PurchaseOrder(row).InternalNote().Eval()
	if !loaded || note == nil || *note != reviewed[0] {
		t.Fatal("returned E view differs from policy review")
	}
	reset()
	changeLiveIntent = true
	row.UpdateContactPhone("5550892")
	_, err = row.AuditAs("Reject a policy callback that changes live intent").Save(ctx)
	var rejected *runtime.RuntimeError
	if !errors.As(err, &rejected) || len(rejected.CheckResults) != 1 || rejected.CheckResults[0].RuleID != "mutation_changed_after_preparation" || rejected.CheckResults[0].InstancePath() != "/contactPhone" || rejected.CheckResults[0].ModelPath() != "contact_phone" {
		t.Fatal("missing precise live-intent rejection", err)
	}
	if checker.calls != 2 || len(reviewed) != 2 {
		t.Fatal("failed save replayed Fix or skipped review")
	}
	assertCounts(0)
	rejectedCounts := counts()
	check(2, "5550891", reviewed[0])
	reset()
	changeLiveIntent = false
	row, err = row.AuditAs("Review and retry the preserved pending intent").Save(ctx)
	if err != nil || checker.calls != 3 || len(reviewed) != 3 || reviewed[2] != "prepared-fix-3" {
		t.Fatal("retry did not get its own fresh preparation", err, checker.calls, reviewed)
	}
	assertCounts(1)
	check(3, "5550899", reviewed[2])
	frame := map[string]any{"language": "go", "passed": true, "id": row.Id(), "fixCalls": checker.calls, "reviewed": reviewed,
		"firstVersion": 2, "finalVersion": 3, "rejectedBeforeProvider": true, "rejectedBeforeAudit": true, "sameObjectRetry": true, "returnedEEqualsReview": true,
		"first": firstCounts, "rejected": rejectedCounts, "retry": counts()}
	bytes, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CORE_PREPARED_SNAPSHOT=" + string(bytes))
}

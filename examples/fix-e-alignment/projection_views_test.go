package fixe

// Generated Q/E/audited-save operations only. The existing provider wrappers
// observe real SQLite transactions; direct SQL remains an independent oracle.
import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type projectionFixture struct {
	t             *testing.T
	ctx           *runtime.UserContext
	probe         *guardedExampleProvider
	audit         *restorationAudit
	policyCalls   int
	round         int
	first, second time.Time
}

func newProjectionFixture(t *testing.T) *projectionFixture {
	t.Helper()
	base := os.Getenv("TEAQL_FIX_E_DB")
	round, err := strconv.Atoi(os.Getenv("TEAQL_PARTIAL_ROOT_ROUND"))
	if base == "" || err != nil || (round != 1 && round != 2) {
		t.Fatal("an owned SQLite path and explicit round 1 or 2 are required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".projection-views.sqlite")
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
	first := time.Date(2032, 2, 29, 10, 15, 30, 0, time.UTC).Add(time.Duration(round-1) * 24 * time.Hour)
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(first))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	provider := ctx.GetResource("dataService")
	f := &projectionFixture{t: t, ctx: ctx, round: round, first: first, second: first.Add(2 * time.Hour), audit: &restorationAudit{},
		probe: &guardedExampleProvider{&restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}}}
	ctx.InsertResource("dataService", f.probe)
	ctx.WithAppAuditEventSink(f.audit).WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { f.policyCalls++; return nil }))
	return f
}

func (f *projectionFixture) counts() map[string]int {
	return map[string]int{"begins": f.probe.begins, "queries": f.probe.queries, "mutations": f.probe.mutations,
		"commits": f.probe.commits, "rollbacks": f.probe.rollbacks, "audits": f.audit.events, "policies": f.policyCalls}
}

func (f *projectionFixture) seed() *purchase_order.PurchaseOrder {
	f.t.Helper()
	row, rowCreationErr := lib.Q.PurchaseOrders().Comment("Prepare an independent query-view fixture").
		Purpose("Verify projection boundaries and independent mutation ledgers").NewEntity(f.ctx)
	if rowCreationErr != nil {
		f.t.Fatal(rowCreationErr)
	}
	row.UpdateContactPhone("5550700").UpdateInternalNote(nil)
	before := f.counts()
	saved, err := row.AuditAs("Create the projection fixture through generated mutation").Save(f.ctx)
	if err != nil {
		f.t.Fatal(err)
	}
	after := f.counts()
	for _, name := range []string{"begins", "mutations", "commits", "audits", "policies"} {
		if after[name] <= before[name] {
			f.t.Fatal("uncalibrated real provider boundary", name, after, before)
		}
	}
	return saved
}

func (f *projectionFixture) full(id uint64) *purchase_order.PurchaseOrder {
	f.t.Helper()
	row, err := lib.Q.PurchaseOrders().WithIdIs(id).Limit(1).
		Comment("Load a complete independent order view").Purpose("Compare explicit query views for one stored identity").ExecuteForOne(f.ctx)
	if err != nil || row == nil {
		f.t.Fatal("missing complete order view", err)
	}
	return row
}

func projectionJSON(t *testing.T, value any) (map[string]any, []byte) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	return body, encoded
}

func projectionMissing(t *testing.T, failure error, path string) {
	t.Helper()
	var notLoaded *runtime.NotLoadedError
	if !errors.As(failure, &notLoaded) || notLoaded.ToWire().AttemptedInstancePath != path {
		t.Fatal("incorrect sparse-view diagnostic", path, failure)
	}
}

func (f *projectionFixture) emit(name string, id uint64, facts map[string]any) {
	f.t.Helper()
	encoded, err := json.Marshal(map[string]any{"round": f.round, "case": name, "orderId": strconv.FormatUint(id, 10), "passed": true, "facts": facts})
	if err != nil {
		f.t.Fatal(err)
	}
	fmt.Println("GO_PROJECTION_VIEW_CASE=" + string(encoded))
}

func TestGeneratedIndependentProjectionViews(t *testing.T) {
	t.Run("later_full_query_keeps_sparse_view", func(t *testing.T) {
		f := newProjectionFixture(t)
		original := f.seed()
		sparse, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(original.Id()).Limit(1).
			Comment("Load only the sparse order identity").Purpose("A later complete query must not widen an earlier view").ExecuteForOne(f.ctx)
		if err != nil || sparse == nil {
			t.Fatal("missing sparse order", err)
		}
		_, _, err = lib.E.PurchaseOrder(sparse).ContactPhone().TryEval()
		projectionMissing(t, err, "/contactPhone")
		_, beforeJSON := projectionJSON(t, sparse)
		wide := f.full(original.Id())
		if sparse == wide || sparse.Id() != wide.Id() {
			t.Fatal("the queries must return different views of the same identity")
		}
		before := f.counts()
		_, _, err = lib.E.PurchaseOrder(sparse).ContactPhone().TryEval()
		projectionMissing(t, err, "/contactPhone")
		_, _, err = lib.E.PurchaseOrder(sparse).InternalNote().TryEval()
		projectionMissing(t, err, "/internalNote")
		if phone, ok := lib.E.PurchaseOrder(wide).ContactPhone().Eval(); !ok || phone != "5550700" {
			t.Fatal("complete phone missing")
		}
		if note, _, err := lib.E.PurchaseOrder(wide).InternalNote().TryEval(); err != nil || note != nil {
			t.Fatal("selected null is not loaded", err)
		}
		narrow, afterJSON := projectionJSON(t, sparse)
		complete, _ := projectionJSON(t, wide)
		if _, ok := narrow["contactPhone"]; ok {
			t.Fatal("JSON widened the sparse phone")
		}
		if _, ok := narrow["internalNote"]; ok {
			t.Fatal("JSON widened the sparse null")
		}
		if value, ok := complete["internalNote"]; !ok || value != nil {
			t.Fatal("JSON lost the selected null")
		}
		if !bytes.Equal(beforeJSON, afterJSON) || !reflect.DeepEqual(f.counts(), before) || len(sparse.EntityRoot().Change(sparse.EntityKey())) != 0 || len(wide.EntityRoot().Change(wide.EntityKey())) != 0 {
			t.Fatal("E/JSON performed provider work, dirtied state, or changed the sparse view")
		}
		if version, ok := lib.E.PurchaseOrder(wide).Version().Eval(); !ok || version != 1 {
			t.Fatal("wrong complete version")
		}
		f.emit("independent-query-projection-boundary", original.Id(), map[string]any{"sameIdentity": true, "separateViews": true, "sparseJsonUnchanged": true, "sparseJsonOmitsFields": true, "wideNullIsLoaded": true, "implicitProviderCalls": 0, "hydrationDirtyKeys": 0, "version": 1})
	})
	t.Run("same_identity_independent_pending_edits", func(t *testing.T) {
		f := newProjectionFixture(t)
		original := f.seed()
		left, other := f.full(original.Id()), f.full(original.Id())
		if left == other || left.EntityRoot() == other.EntityRoot() {
			t.Fatal("independent views share mutable state")
		}
		left.UpdateContactPhone("5550701")
		note := "pending-other-view"
		other.UpdateInternalNote(&note)
		if other.EntityRoot().Change(other.EntityKey())["internal_note"].V != note {
			t.Fatal("missing independent pending edit")
		}
		f.ctx.WithBusinessClock(runtime.NewFixedBusinessClock(f.second))
		if _, err := left.AuditAs("Save only the first independently queried view").Save(f.ctx); err != nil {
			t.Fatal(err)
		}
		stored := f.full(original.Id())
		if phone, ok := lib.E.PurchaseOrder(stored).ContactPhone().Eval(); !ok || phone != "5550701" {
			t.Fatal("first edit was not saved")
		}
		if value, _, err := lib.E.PurchaseOrder(stored).InternalNote().TryEval(); err != nil || value != nil {
			t.Fatal("saving one view leaked the other's edit", err)
		}
		if version, ok := lib.E.PurchaseOrder(stored).Version().Eval(); !ok || version != 2 {
			t.Fatal("first view did not advance once")
		}
		if phone, ok := lib.E.PurchaseOrder(other).ContactPhone().Eval(); !ok || phone != "5550700" {
			t.Fatal("other view's baseline was silently replaced")
		}
		intent, before := other.EntityRoot().Change(other.EntityKey()), f.counts()
		_, failure := other.AuditAs("Reject a stale independent view while retaining its pending edit").Save(f.ctx)
		effects := f.counts()
		for key := range effects {
			effects[key] -= before[key]
		}
		var conflict *runtime.RuntimeError
		errors.As(failure, &conflict)
		kind := ""
		if conflict != nil {
			kind = conflict.Type
		}
		observed, _ := json.Marshal(map[string]any{"round": f.round, "orderId": strconv.FormatUint(original.Id(), 10), "errorType": fmt.Sprintf("%T", failure), "kind": kind, "message": fmt.Sprint(failure), "effects": effects})
		fmt.Println("GO_STALE_VIEW_OBSERVATION=" + string(observed))
		// The existing prepared-provider guard rejects a zero-affected stale
		// update as incomplete reviewed work. ML-10/12 require optimistic
		// admission and no false commit, not a particular exception label.
		if failure == nil || conflict == nil || conflict.Type != "Check" || len(conflict.CheckResults) != 1 || conflict.CheckResults[0].RuleID != "prepared_mutation_incomplete" {
			t.Fatal("missing existing structured prepared-plan rejection", failure)
		}
		if effects["begins"] != 1 || effects["mutations"] != 1 || effects["rollbacks"] != 1 || effects["commits"] != 0 || effects["audits"] != 0 {
			t.Fatal("stale save did not roll back without commit/audit", effects)
		}
		after := other.EntityRoot().Change(other.EntityKey())
		for key, value := range intent {
			if !reflect.DeepEqual(after[key], value) {
				t.Fatal("rejection lost original pending intent", key)
			}
		}
		derived := []string{}
		for key := range after {
			if _, known := intent[key]; !known {
				derived = append(derived, key)
			}
		}
		sort.Strings(derived)
		if !reflect.DeepEqual(derived, []string{"update_time"}) {
			t.Fatal("unexpected post-Fix pending fields", derived)
		}
		if millis, ok := after["update_time"].TryTimestamp(); !ok || millis != f.second.UnixMilli() {
			t.Fatal("wrong BusinessClock Fix retained")
		}
		if version, ok := lib.E.PurchaseOrder(other).Version().Eval(); !ok || version != 1 {
			t.Fatal("stale version silently rebased")
		}
		if value, _, err := lib.E.PurchaseOrder(other).InternalNote().TryEval(); err != nil || value == nil || *value != note {
			t.Fatal("other view's edit was discarded", err)
		}
		stored = f.full(original.Id())
		if phone, ok := lib.E.PurchaseOrder(stored).ContactPhone().Eval(); !ok || phone != "5550701" {
			t.Fatal("stale save altered the committed phone")
		}
		if value, _, err := lib.E.PurchaseOrder(stored).InternalNote().TryEval(); err != nil || value != nil {
			t.Fatal("stale pending note was committed")
		}
		if version, ok := lib.E.PurchaseOrder(stored).Version().Eval(); !ok || version != 2 {
			t.Fatal("stale save changed the persisted version")
		}
		f.emit("same-identity-independent-mutation-ledgers", original.Id(), map[string]any{"separateViews": true, "otherPendingEditRetained": true, "noCrossViewWrite": true, "staleException": "Check:prepared_mutation_incomplete", "storedVersion": 2, "otherVersion": 1, "derivedFixFields": derived, "rejectionRolledBack": true, "commitAuditsOnRejection": 0})
	})
	t.Run("repeated_identity_keeps_explicit_projections", func(t *testing.T) {
		f := newProjectionFixture(t)
		original := f.seed()
		root, err := lib.Q.TenantsMinimal().WithIdIs(1).
			SelectPurchaseOrderListWith(lib.Q.PurchaseOrdersMinimal().WithIdIs(original.Id()).
				SelectTenantWith(lib.Q.Tenants().Limit(1).Comment("Load the full nested alias of the root")).Limit(1).
				Comment("Load the selected sparse order and its full forward root")).Limit(1).
			Comment("Keep the outer root sparse despite a full nested alias").Purpose("Identity deduplication must not union explicit query projections").ExecuteForOne(f.ctx)
		if err != nil || root == nil || len(root.PurchaseOrderList().Items()) != 1 {
			t.Fatal("missing repeated-identity query graph", err)
		}
		parent := root.PurchaseOrderList().Items()[0]
		before := f.counts()
		targetID, ok := lib.E.PurchaseOrder(parent).Tenant().Id().Eval()
		if !ok || targetID != root.Id() {
			t.Fatal("nested target is not the same identity")
		}
		name, ok := lib.E.PurchaseOrder(parent).Tenant().Name().Eval()
		if !ok || name == "" {
			t.Fatal("complete target name missing")
		}
		_, _, err = lib.E.Tenant(root).Name().TryEval()
		projectionMissing(t, err, "/name")
		body, _ := projectionJSON(t, root)
		if _, ok := body["name"]; ok {
			t.Fatal("outer JSON projection was widened")
		}
		parents, ok := body["purchaseOrderList"].([]any)
		if !ok || len(parents) != 1 {
			t.Fatal("JSON lost the selected reverse list")
		}
		parentBody, ok := parents[0].(map[string]any)
		if !ok {
			t.Fatal("JSON parent is not an object")
		}
		target, ok := parentBody["tenant"].(map[string]any)
		if !ok || target["name"] != name {
			t.Fatal("JSON lost the full nested target", string(mustProjectionJSON(t, root)))
		}
		if !reflect.DeepEqual(f.counts(), before) || len(root.EntityRoot().Change(root.EntityKey())) != 0 {
			t.Fatal("E/JSON read performed provider work or created mutation intent")
		}
		f.emit("intra-query-repeated-identity-projection-boundary", original.Id(), map[string]any{"sameIdentity": true, "sparseRootNameNotLoaded": true, "nestedTargetNameLoaded": true, "jsonRespectsBothViews": true, "implicitProviderCalls": 0})
	})
}

func mustProjectionJSON(t *testing.T, value any) []byte {
	_, bytes := projectionJSON(t, value)
	return bytes
}

package fixe

import (
	stdcontext "context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/order_item"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

const graphFixEpoch int64 = 1988247845000

type graphFixClock struct {
	mu    sync.Mutex
	next  int64
	reads int
}

func (c *graphFixClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := c.next
	c.next++
	c.reads++
	return time.UnixMilli(v).UTC()
}
func (c *graphFixClock) state() (int64, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.next, c.reads
}

type graphFixVisit struct {
	Entity                 string `json:"entity"`
	Operation              int    `json:"operation"`
	Time                   int64  `json:"time"`
	ProviderCallsDuringFix int    `json:"providerCallsDuringFix"`
}
type graphFixChecker struct {
	delegate runtime.CheckerRegistry
	effects  func() map[string]int
	visits   []graphFixVisit
	reject   bool
}

func (c *graphFixChecker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	before := c.effects()
	result := c.delegate.CheckAndFix(ctx, input)
	calls := 0
	if !reflect.DeepEqual(before, c.effects()) {
		calls = 1
	}
	c.visits = append(c.visits, graphFixVisit{input.Entity, int(input.Operation), input.Now.UnixMilli(), calls})
	if c.reject && input.Entity == "order_item" {
		result = append(result, runtime.CheckResult{RuleID: "core236_fix_rejected", CanonicalLocation: runtime.Location().Property("product_name"), Message: "controlled graph Fix rejection"})
	}
	return result
}

type graphFixProvider struct {
	*guardedExampleProvider
	beforeBegin func()
	cancel      stdcontext.CancelFunc
	mutations   int
}

func (p *graphFixProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	if p.beforeBegin != nil {
		p.beforeBegin()
	}
	tx, err := p.guardedExampleProvider.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &graphFixTransaction{Transaction: tx, provider: p}, nil
}

type graphFixTransaction struct {
	data_service.Transaction
	provider *graphFixProvider
}

func (tx *graphFixTransaction) GenerateId(entity string) (uint64, error) {
	return tx.Transaction.(runtime.InternalIdGenerator).GenerateId(entity)
}
func (tx *graphFixTransaction) EnsureIdFloor(ctx stdcontext.Context, entity string, floor uint64) error {
	return tx.Transaction.(runtime.InternalIdFloorReserver).EnsureIdFloor(ctx, entity, floor)
}
func (tx *graphFixTransaction) Mutate(ctx stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	result, err := tx.Transaction.Mutate(ctx, request)
	tx.provider.mutations++
	if err == nil && tx.provider.cancel != nil && tx.provider.mutations == 2 {
		tx.provider.cancel()
		return nil, stdcontext.Canceled
	}
	return result, err
}

type graphFixFixture struct {
	ctx      *runtime.UserContext
	db       *sql.DB
	provider *graphFixProvider
	checker  *graphFixChecker
	ids      *lateExampleIDs
}

func newGraphFixFixture(t *testing.T, suffix string) *graphFixFixture {
	t.Helper()
	base := os.Getenv("CORE236_GO_GRAPH_DB")
	if base == "" {
		t.Fatal("retained database required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+suffix)
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
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(graphFixEpoch).UTC()))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	original := ctx.GetResource("dataService")
	p := &graphFixProvider{guardedExampleProvider: &guardedExampleProvider{&restorationProvider{QueryExecutor: original.(data_service.QueryExecutor), transaction: original.(data_service.TransactionExecutor)}}}
	f := &graphFixFixture{ctx: ctx, db: db, provider: p, ids: &lateExampleIDs{delegate: ctx.InternalIdGenerator()}}
	f.checker = &graphFixChecker{delegate: lib.Module().CheckerRegistry, effects: f.effects}
	ctx.InsertResource("dataService", p)
	ctx.WithInternalIdGenerator(f.ids).WithCheckerRegistry(f.checker)
	return f
}
func (f *graphFixFixture) effects() map[string]int {
	p := f.provider
	return map[string]int{"queries": p.queries, "begins": p.begins, "mutations": p.restorationProvider.mutations, "commits": p.commits, "rollbacks": p.rollbacks, "ids": f.ids.calls}
}
func (f *graphFixFixture) clear() {
	p := f.provider
	p.queries = 0
	p.begins = 0
	p.restorationProvider.mutations = 0
	p.commits = 0
	p.rollbacks = 0
	p.mutations = 0
	f.ids.calls = 0
	f.checker.visits = nil
}
func graphFixMake(ctx *runtime.UserContext, count int) (*purchase_order.PurchaseOrder, []*order_item.OrderItem) {
	parent, parentCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a generated graph clock owner").Purpose("Verify actual generated Fix and audited persistence").NewEntity(ctx)
	if parentCreationErr != nil {
		panic(parentCreationErr)
	}
	parent.UpdateContactPhone("5550400").UpdateInternalNote(nil)
	children := make([]*order_item.OrderItem, count)
	for i := range children {
		child, childCreationErr := lib.Q.OrderItems().Comment("Prepare a generated graph child").Purpose("Verify one stable graph instant").NewEntity(ctx)
		if childCreationErr != nil {
			panic(childCreationErr)
		}
		child.UpdateProductName(fmt.Sprintf("Graph item %d", i)).UpdateUnitPrice(decimal.RequireFromString("12.34"))
		parent.OrderItemList().Add(child)
		children[i] = child
	}
	return parent, children
}
func graphFixFrame(name string, facts any) {
	b, err := json.Marshal(map[string]any{"case": name, "facts": facts})
	if err != nil {
		panic(err)
	}
	fmt.Println("GO_GRAPH_FIX_CASE=" + string(b))
}
func graphFixVisits(t *testing.T, visits []graphFixVisit, expected int64, count int) {
	t.Helper()
	if len(visits) != count {
		t.Fatalf("Fix visits=%v want %d", visits, count)
	}
	for _, v := range visits {
		if v.Time != expected || v.ProviderCallsDuringFix != 0 {
			t.Fatalf("inconsistent Fix observation: %+v", v)
		}
	}
}
func graphFixRead(t *testing.T, f *graphFixFixture, parent *purchase_order.PurchaseOrder, children []*order_item.OrderItem, created, updated, version int64) {
	t.Helper()
	childCreated := make([]int64, len(children))
	for i := range childCreated {
		childCreated[i] = created
	}
	graphFixReadCreated(t, f, parent, children, created, childCreated, updated, version)
}
func graphFixReadCreated(t *testing.T, f *graphFixFixture, parent *purchase_order.PurchaseOrder, children []*order_item.OrderItem, created int64, childCreated []int64, updated, version int64, childUpdates ...[]int64) {
	t.Helper()
	owner, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).Limit(1).Comment("Reload the persisted graph owner").Purpose("Validate generated E timestamps and versions").ExecuteForOne(f.ctx)
	if err != nil || owner == nil {
		t.Fatal("owner reload failed", err)
	}
	c, _ := lib.E.PurchaseOrder(owner).CreateTime().Eval()
	u, _ := lib.E.PurchaseOrder(owner).UpdateTime().Eval()
	v, _ := lib.E.PurchaseOrder(owner).Version().Eval()
	r, _ := lib.E.PurchaseOrder(owner).TenantId().Eval()
	phone, _ := lib.E.PurchaseOrder(owner).ContactPhone().Eval()
	if c.UnixMilli() != created || u.UnixMilli() != updated || v != version || r != 1 || phone == "" {
		t.Fatalf("persisted parent differs: %v %v %d %d", c, u, v, r)
	}
	for i, child := range children {
		childUpdated := updated
		if len(childUpdates) > 0 {
			childUpdated = childUpdates[0][i]
		}
		row, err := lib.Q.OrderItems().WithIdIs(child.Id()).Limit(1).Comment("Reload the persisted graph child").Purpose("Validate business input and generated E values").ExecuteForOne(f.ctx)
		if err != nil || row == nil {
			t.Fatal("child reload failed", err)
		}
		c, _ := lib.E.OrderItem(row).CreateTime().Eval()
		u, _ := lib.E.OrderItem(row).UpdateTime().Eval()
		v, _ := lib.E.OrderItem(row).Version().Eval()
		r, _ := lib.E.OrderItem(row).TenantId().Eval()
		name, _ := lib.E.OrderItem(row).ProductName().Eval()
		if c.UnixMilli() != childCreated[i] || u.UnixMilli() != childUpdated || v != version || r != 1 || name != fmt.Sprintf("Graph item %d", i) {
			t.Fatalf("persisted child differs: %v %v %d %d %q", c, u, v, r, name)
		}
	}
}
func graphFixReceipt(t *testing.T, ctx *runtime.UserContext, create bool) map[string]any {
	t.Helper()
	rows := ctx.LastFixEvidence()
	if len(rows) == 0 {
		t.Fatal("missing graph Fix provenance")
	}
	fields := map[string]bool{}
	for _, row := range rows {
		fields[row.EntityType+":"+row.ModelPath] = true
		if row.Source != runtime.FixEvidenceClock && row.Source != runtime.FixEvidenceContext {
			t.Fatal("invalid provenance source")
		}
		if row.SourceLabel != "graphClock" && row.SourceLabel != "activeRoot:Tenant" {
			t.Fatal("unsafe provenance label", row.SourceLabel)
		}
	}
	paths := []string{"update_time"}
	if create {
		paths = []string{"tenant", "create_time", "update_time"}
	}
	if len(fields) != 2*len(paths) {
		t.Fatal("incomplete Fix provenance", fields)
	}
	for _, entity := range []string{"PurchaseOrder", "OrderItem"} {
		for _, p := range paths {
			if !fields[entity+":"+p] {
				t.Fatal("missing provenance", entity, p)
			}
		}
	}
	original := rows[0].SourceLabel
	rows[0].SourceLabel = "caller-mutated-copy"
	if ctx.LastFixEvidence()[0].SourceLabel != original {
		t.Fatal("mutable evidence alias escaped")
	}
	return map[string]any{"records": len(rows), "distinctFields": len(fields), "detached": true, "valueFree": true}
}

func TestCore236GraphFixLifecycle(t *testing.T) {
	f := newGraphFixFixture(t, ".lifecycle.sqlite")
	round := 0
	fmt.Sscan(os.Getenv("CORE236_GO_GRAPH_ROUND"), &round)
	clock := &graphFixClock{next: graphFixEpoch + int64(round)*1000}
	f.ctx.WithBusinessClock(clock)
	parent, children := graphFixMake(f.ctx, 2)
	created, _ := clock.state()
	if _, err := parent.AuditAs("Create a generated parent and two children").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, f.checker.visits, created, 3)
	receipt := graphFixReceipt(t, f.ctx, true)
	graphFixRead(t, f, parent, children, created, created, 1)
	graphFixFrame("create", map[string]any{"visits": f.checker.visits, "clockCaptures": 1, "receipt": receipt, "parentId": parent.Id(), "childIds": []uint64{children[0].Id(), children[1].Id()}, "created": created})
	change := func(phone, price string) {
		parent.UpdateContactPhone(phone)
		for _, child := range children {
			child.UpdateUnitPrice(decimal.RequireFromString(price))
		}
		f.clear()
	}
	change("5550401", "23.45")
	updated, _ := clock.state()
	if _, err := parent.AuditAs("Update at the next graph clock instant").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, f.checker.visits, updated, 3)
	graphFixRead(t, f, parent, children, created, updated, 2)
	graphFixReceipt(t, f.ctx, false)
	graphFixFrame("update", map[string]any{"visits": f.checker.visits, "created": created, "updated": updated, "version": 2})
	retry, retryChildren := graphFixMake(f.ctx, 2)
	f.clear()
	f.checker.reject = true
	failed, _ := clock.state()
	_, err := retry.AuditAs("Reject in Checker before any provider entry").Save(f.ctx)
	var check *runtime.RuntimeError
	if !errors.As(err, &check) || check.Type != "Check" {
		t.Fatal("missing controlled Checker rejection", err)
	}
	graphFixVisits(t, f.checker.visits, failed, 2)
	if !reflect.DeepEqual(f.effects(), map[string]int{"queries": 0, "begins": 0, "mutations": 0, "commits": 0, "rollbacks": 0, "ids": 0}) {
		t.Fatal("provider work before rejected Fix", f.effects())
	}
	f.checker.reject = false
	f.clear()
	retried, _ := clock.state()
	// A successful parent Fix can remain on the object after a later child
	// rejection. A failed child Fix need not publish its candidate values.
	// Snapshot each actual input before retry: preserve present insert values,
	// fill only absent values, and require every Checker to receive the new Now.
	retryCreated, createPresent := lib.E.PurchaseOrder(retry).CreateTime().Eval()
	retryUpdated, updatePresent := lib.E.PurchaseOrder(retry).UpdateTime().Eval()
	if !createPresent {
		retryCreated = time.UnixMilli(retried)
	}
	if !updatePresent {
		retryUpdated = time.UnixMilli(retried)
	}
	childCreated, childUpdated := make([]int64, len(retryChildren)), make([]int64, len(retryChildren))
	for i, child := range retryChildren {
		c, cp := lib.E.OrderItem(child).CreateTime().Eval()
		u, up := lib.E.OrderItem(child).UpdateTime().Eval()
		childCreated[i], childUpdated[i] = retried, retried
		if cp {
			childCreated[i] = c.UnixMilli()
		}
		if up {
			childUpdated[i] = u.UnixMilli()
		}
	}
	if _, err := retry.AuditAs("Retry the same rejected graph with a fresh invocation").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, f.checker.visits, retried, 3)
	graphFixReadCreated(t, f, retry, retryChildren, retryCreated.UnixMilli(), childCreated, retryUpdated.UnixMilli(), 1, childUpdated)
	graphFixFrame("checker-retry", map[string]any{"failedTime": failed, "retriedTime": retried, "rejectedProviderCalls": 0, "sameGraphRetry": true, "existingInsertTimesPreserved": true})
	if _, err := f.db.Exec("CREATE TRIGGER core236_go_fix_abort BEFORE UPDATE ON order_item_data BEGIN SELECT RAISE(ABORT, 'CORE236_GO_FIX_PROVIDER_ABORT'); END;"); err != nil {
		t.Fatal(err)
	}
	change("5550402", "34.56")
	providerTime, _ := clock.state()
	if _, err := parent.AuditAs("Rollback a real SQLite child update failure").Save(f.ctx); err == nil || !strings.Contains(err.Error(), "CORE236_GO_FIX_PROVIDER_ABORT") {
		t.Fatal("expected SQLite failure", err)
	}
	graphFixVisits(t, f.checker.visits, providerTime, 3)
	graphFixRead(t, f, parent, children, created, updated, 2)
	if _, err := f.db.Exec("DROP TRIGGER core236_go_fix_abort"); err != nil {
		t.Fatal(err)
	}
	f.clear()
	providerRetry, _ := clock.state()
	if _, err := parent.AuditAs("Retry after provider rollback").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, f.checker.visits, providerRetry, 3)
	graphFixRead(t, f, parent, children, created, providerRetry, 3)
	graphFixFrame("provider-retry", map[string]any{"failedTime": providerTime, "retriedTime": providerRetry, "rollbackVerified": true})
	// Cancel an owned provider operation after its actual write. Do not replace
	// UserContext.Context while database/sql can still observe that Context.
	cancellable, cancel := stdcontext.WithCancel(f.ctx.Context)
	defer cancel()
	f.provider.cancel = cancel
	change("5550403", "45.67")
	cancelled, _ := clock.state()
	_, err = parent.AuditAs("Cancel after a real child write but before commit").Save(f.ctx)
	cancel()
	f.provider.cancel = nil
	if !errors.Is(err, stdcontext.Canceled) || !errors.Is(cancellable.Err(), stdcontext.Canceled) {
		t.Fatal("cancellation cause lost", err)
	}
	graphFixVisits(t, f.checker.visits, cancelled, 3)
	graphFixRead(t, f, parent, children, created, providerRetry, 3)
	f.clear()
	cancelRetry, _ := clock.state()
	if _, err := parent.AuditAs("Retry the same graph after cancellation rollback").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, f.checker.visits, cancelRetry, 3)
	graphFixRead(t, f, parent, children, created, cancelRetry, 4)
	receipt = graphFixReceipt(t, f.ctx, false)
	graphFixFrame("cancel-retry", map[string]any{"cancelledTime": cancelled, "retriedTime": cancelRetry, "rollbackVerified": true, "originalCauseRetained": true, "cancellationScope": "provider-operation-after-write", "callerContextReplaced": false, "receipt": receipt, "updated": cancelRetry, "version": 4})
	_, reads := clock.state()
	if reads != 8 {
		t.Fatalf("unexpected clock reads: %d", reads)
	}
	graphFixFrame("lifecycle-complete", map[string]any{"round": round, "clockCaptures": reads, "parentId": parent.Id(), "childIds": []uint64{children[0].Id(), children[1].Id()}, "retryParentId": retry.Id(), "retryChildIds": []uint64{retryChildren[0].Id(), retryChildren[1].Id()}, "created": created, "updated": cancelRetry, "retryParentCreated": retryCreated.UnixMilli(), "retryParentUpdated": retryUpdated.UnixMilli(), "retryChildCreated": childCreated, "retryChildUpdated": childUpdated})
}

func TestCore236GraphFixRepeatedObject(t *testing.T) {
	f := newGraphFixFixture(t, ".duplicate.sqlite")
	clock := &graphFixClock{next: graphFixEpoch}
	f.ctx.WithBusinessClock(clock)
	parent, children := graphFixMake(f.ctx, 1)
	parent.OrderItemList().Add(children[0])
	_, err := parent.AuditAs("Save a child instance reached twice in one relation list").Save(f.ctx)
	violations := []map[string]string{}
	var check *runtime.RuntimeError
	if errors.As(err, &check) {
		for _, v := range check.CheckResults {
			violations = append(violations, map[string]string{"ruleId": v.RuleID, "path": v.ModelPath()})
		}
	}
	_, reads := clock.state()
	graphFixFrame("repeated-object", map[string]any{"passed": err == nil, "visits": f.checker.visits, "clockCaptures": reads, "error": fmt.Sprint(err), "violations": violations, "effects": f.effects()})
	if err != nil {
		t.Fatal("repeated reference must not create another write", err)
	}
	graphFixVisits(t, f.checker.visits, graphFixEpoch, 2)
	graphFixRead(t, f, parent, children, graphFixEpoch, graphFixEpoch, 1)
	parent.UpdateContactPhone("5550404")
	children[0].UpdateUnitPrice(decimal.RequireFromString("23.45"))
	f.clear()
	if _, err := parent.AuditAs("Update one repeated instance once").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, f.checker.visits, graphFixEpoch+1, 2)
	graphFixRead(t, f, parent, children, graphFixEpoch, graphFixEpoch+1, 2)
	graphFixFrame("repeated-object-complete", map[string]any{"parentId": parent.Id(), "childId": children[0].Id(), "version": 2, "oncePerSave": true})
}

func TestCore236GraphFixIndependentContexts(t *testing.T) {
	a, b := newGraphFixFixture(t, ".first.sqlite"), newGraphFixFixture(t, ".second.sqlite")
	clock := &graphFixClock{next: graphFixEpoch}
	a.ctx.WithBusinessClock(clock)
	b.ctx.WithBusinessClock(clock)
	left, leftChildren := graphFixMake(a.ctx, 1)
	right, rightChildren := graphFixMake(b.ctx, 1)
	ready, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	a.provider.beforeBegin = func() { close(ready); <-release }
	go func() {
		_, err := left.AuditAs("Hold the first real generated graph before provider entry").Save(a.ctx)
		done <- err
	}()
	released := false
	defer func() {
		if !released {
			close(release)
			<-done
		}
	}()
	select {
	case <-ready:
	case err := <-done:
		released = true
		t.Fatal("first graph failed before hold", err)
	case <-time.After(20 * time.Second):
		t.Fatal("first graph did not reach provider")
	}
	if _, err := right.AuditAs("Commit another real graph on an independent Context").Save(b.ctx); err != nil {
		t.Fatal(err)
	}
	if a.ctx.FixTime().UnixMilli() != graphFixEpoch {
		t.Fatal("foreign context replaced the held graph snapshot")
	}
	close(release)
	released = true
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	graphFixVisits(t, a.checker.visits, graphFixEpoch, 2)
	graphFixVisits(t, b.checker.visits, graphFixEpoch+1, 2)
	graphFixRead(t, a, left, leftChildren, graphFixEpoch, graphFixEpoch, 1)
	graphFixRead(t, b, right, rightChildren, graphFixEpoch+1, graphFixEpoch+1, 1)
	graphFixFrame("independent-contexts", map[string]any{"first": a.checker.visits, "second": b.checker.visits, "separateSnapshots": true, "firstReceipt": graphFixReceipt(t, a.ctx, true), "secondReceipt": graphFixReceipt(t, b.ctx, true)})
}

func TestCore236GraphFixDistinctViews(t *testing.T) {
	f := newGraphFixFixture(t, ".distinct.sqlite")
	f.ctx.WithBusinessClock(&graphFixClock{next: graphFixEpoch})
	parent, children := graphFixMake(f.ctx, 1)
	if _, err := parent.AuditAs("Create the distinct-view guard calibration").Save(f.ctx); err != nil {
		t.Fatal(err)
	}
	owner, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).Limit(1).Comment("Load an independently owned parent view").Purpose("Verify pointer deduplication never hides distinct views").ExecuteForOne(f.ctx)
	if err != nil || owner == nil {
		t.Fatal("owner load failed", err)
	}
	load := func() *order_item.OrderItem {
		row, err := lib.Q.OrderItems().WithIdIs(children[0].Id()).Limit(1).Comment("Load a separate child view").Purpose("Verify same ID does not mean same instance").ExecuteForOne(f.ctx)
		if err != nil || row == nil {
			t.Fatal("child load failed", err)
		}
		return row
	}
	a, b := load(), load()
	if a == b || a.Id() != b.Id() {
		t.Fatal("distinct same-identity calibration failed")
	}
	a.UpdateUnitPrice(decimal.RequireFromString("21.00"))
	b.UpdateUnitPrice(decimal.RequireFromString("34.00"))
	owner.OrderItemList().Add(a)
	owner.OrderItemList().Add(b)
	f.clear()
	_, err = owner.AuditAs("Reject two different pending views of one child identity").Save(f.ctx)
	var check *runtime.RuntimeError
	if !errors.As(err, &check) || check.Type != "Check" || len(check.CheckResults) == 0 {
		t.Fatal("conflicting views not rejected", err)
	}
	effects := f.effects()
	if effects["commits"] != 0 || (effects["begins"] > 0 && effects["rollbacks"] != 1) {
		t.Fatal("distinct views were committed or not rolled back", effects)
	}
	graphFixRead(t, f, parent, children, graphFixEpoch, graphFixEpoch, 1)
	row := load()
	price, ok := lib.E.OrderItem(row).UnitPrice().Eval()
	if !ok || !price.Equal(decimal.RequireFromString("12.34")) {
		t.Fatal("rejected view altered storage", price)
	}
	graphFixFrame("distinct-views", map[string]any{"sameIdentity": true, "differentInstances": true, "rejectedBeforeCommit": true, "rejectedBeforeProvider": effects["begins"] == 0, "effects": effects, "version": 1, "rule": check.CheckResults[0].RuleID, "preProviderAdmissionQualified": false})
}

func TestCore236GraphFixOrigin(t *testing.T) {
	pc := reflect.ValueOf(runtime.CheckAndFixGraphEntity).Pointer()
	file, _ := goruntime.FuncForPC(pc).FileLine(pc)
	root := os.Getenv("CORE236_GO_RUNTIME_ROOT")
	if root == "" || !strings.HasPrefix(file, root+"/") {
		t.Fatal("runtime escaped selected distribution", file, root)
	}
	bytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(bytes)
	if hex.EncodeToString(digest[:]) != os.Getenv("CORE236_GO_RUNTIME_SHA256") {
		t.Fatal("compiled runtime origin hash mismatch")
	}
	graphFixFrame("runtime-origin", map[string]any{"file": file, "sha256": hex.EncodeToString(digest[:]), "sourceFallback": false})
}

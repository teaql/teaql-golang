package tracechain_test

import (
	"fmt"
	"reflect"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/platform"
)

// The first transaction is held at COMMIT. Observe the other generated Save
// actually waiting at the Context gate, rather than treating a pre-call ready
// channel as proof that the save requests overlap. This does not change the
// runtime or claim simultaneous SQLite writers.
func awaitBlockedGeneratedSave(t *testing.T) {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	buffer := make([]byte, 64*1024)
	for {
		length := goruntime.Stack(buffer, true)
		for _, stack := range strings.Split(string(buffer[:length]), "\n\n") {
			if strings.Contains(stack, "[semacquire]") &&
				strings.Contains(stack, "(*UserContext).ExecutePreparedGraphSave(") &&
				strings.Contains(stack, "(*CustomerOrder).Save(") {
				t.Log("second generated Save observed waiting at the same Context graph gate before first COMMIT")
				return
			}
		}
		select {
		case <-timer.C:
			t.Fatal("second generated Save never entered the Context gate while the first COMMIT was held")
		default:
			goruntime.Gosched()
		}
	}
}

// Q/E/save are the application path. Base/RelationEntity/EntityRoot below are
// observation-only generated infrastructure: no expected trace, SQL, relation,
// or successful mutation is injected into the production path.
func TestGeneratedSharedReadonlyReferencesKeepIndependentMutationOwnership(t *testing.T) {
	e := openEnvironment(t)
	for _, label := range []string{"left", "right"} {
		order := newOrder(t, e, "shared readonly fixture "+label)
		if _, err := order.AuditAs("prepare shared reference fixture").Save(e.context); err != nil {
			t.Fatal(err)
		}
	}
	e.reset()
	rows, err := lib.Q.CustomerOrders().WithPlatformIs(1).OrderByIdAsc().Limit(2).
		SelectId().SelectVersion().SelectOrderNumber().SelectDescription().
		SelectPlatformWith(lib.Q.Platforms().Limit(1)).
		Comment("load two roots with a shared readonly platform").
		Purpose("verify snapshot sharing without mutation ownership sharing").ExecuteForList(e.context)
	if err != nil || len(rows.Data) != 2 {
		t.Fatalf("bounded generated relation query failed: rows=%v error=%v", rows, err)
	}
	left, right := rows.Data[0], rows.Data[1]
	leftRaw, leftLoaded := left.Base().GetDynamic("platformEntity")
	rightRaw, rightLoaded := right.Base().GetDynamic("platformEntity")
	leftRecord, leftRecordOK := leftRaw.V.(core.Record)
	rightRecord, rightRecordOK := rightRaw.V.(core.Record)
	if !leftLoaded || !rightLoaded || !leftRecordOK || !rightRecordOK || len(leftRecord) == 0 ||
		reflect.ValueOf(leftRecord).UnsafePointer() != reflect.ValueOf(rightRecord).UnsafePointer() {
		t.Fatal("fixture did not exercise actual shared readonly relation records")
	}
	if left.EntityRoot() == right.EntityRoot() {
		t.Fatal("independent loaded roots share mutation ownership through the query ledger")
	}
	t.Log("shared relation record identity confirmed; root mutation ledgers are distinct")

	orders := []*customer_order.CustomerOrder{left, right}
	reasons := []string{"review left shared-reference graph", "review right shared-reference graph"}
	branches := []string{"correct left item", "correct right item"}
	ids, versions := make([]uint64, 2), make([]int64, 2)
	items := make([]*order_item.OrderItem, 2)
	references := make([]*platform.Platform, 2)
	var referenceVersion int64
	for index, order := range orders {
		expression := customer_order.NewCustomerOrderExpression(order)
		ids[index], _ = expression.Id().Eval()
		versions[index], _ = expression.Version().Eval()
		if foreignKey, present := expression.PlatformId().Eval(); !present || foreignKey != 1 {
			t.Fatal("generated E did not observe the shared platform foreign key")
		}
		related, loaded := order.RelationEntity("platformEntity")
		reference, typed := related.(*platform.Platform)
		if !loaded || !typed || reference == nil {
			t.Fatal("generated relation factory did not hydrate the readonly platform")
		}
		references[index] = reference
		referenceExpression := platform.NewPlatformExpression(reference)
		if id, present := referenceExpression.Id().Eval(); !present || id != 1 {
			t.Fatal("generated E lost the loaded reference identity")
		}
		currentVersion, present := referenceExpression.Version().Eval()
		if !present || currentVersion < 1 || (index > 0 && currentVersion != referenceVersion) {
			t.Fatal("loaded reference version was not preserved")
		}
		referenceVersion = currentVersion
		if reference.EntityRoot() == order.EntityRoot() || reference.EntityRoot().IsNew(reference.EntityKey()) ||
			len(reference.EntityRoot().Changes()) != 0 || len(reference.EntityRoot().TraceChain(reference.EntityKey())) != 0 {
			t.Fatal("loading an unmodified reference adopted mutation ownership or queued an insert")
		}
		if order.EntityRoot().IsNew(order.EntityKey()) || len(order.EntityRoot().Changes()) != 0 {
			t.Fatal("loading an existing root queued a mutation")
		}
		if keys := order.EntityRoot().Keys(); len(keys) != 1 || !reflect.DeepEqual(keys[0], order.EntityKey()) {
			t.Fatalf("loaded root ledger contains another graph or a readonly reference: %+v", keys)
		}
		if index == 1 && versions[1] == versions[0] {
			order.UpdateDescription("prepare a distinct original version")
			if _, err := order.AuditAs("prepare independent optimistic versions").Save(e.context); err != nil {
				t.Fatal(err)
			}
			versions[1], _ = customer_order.NewCustomerOrderExpression(order).Version().Eval()
		}
		if version, present := order.EntityRoot().OriginalVersion(order.EntityKey()); !present || version != versions[index] {
			t.Fatal("the loaded root lost its own optimistic version")
		}
		order.UpdateDescription(fmt.Sprintf("isolated graph %d after version %d", index, versions[index]))
		items[index] = newItem(e, fmt.Sprintf("shared reference item %d version %d", index, versions[index])).Comment(branches[index])
		order.OrderItemList().Add(items[index])
	}
	if references[0] == references[1] {
		t.Fatal("Go relation factories unexpectedly shared a mutable entity wrapper")
	}
	e.reset()
	commitsBefore := e.observer.commits
	firstAtCommit, release := make(chan struct{}), make(chan struct{})
	var atCommit, releaseOnce sync.Once
	releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseFirst()
	e.observer.beforeCommit = func() { atCommit.Do(func() { close(firstAtCommit); <-release }) }
	results := make(chan error, 2)
	go func() { _, err := left.AuditAs(reasons[0]).Save(e.context); results <- err }()
	select {
	case <-firstAtCommit:
	case <-time.After(5 * time.Second):
		t.Fatal("first shared-reference graph did not reach real COMMIT")
	}
	secondInvoked := make(chan struct{})
	go func() {
		audited := right.AuditAs(reasons[1])
		close(secondInvoked)
		_, saveErr := audited.Save(e.context)
		results <- saveErr
	}()
	<-secondInvoked
	awaitBlockedGeneratedSave(t)
	if len(e.sink.snapshot()) != 0 {
		t.Error("an uncommitted graph emitted a committed audit")
	}
	releaseFirst()
	for index := 0; index < 2; index++ {
		select {
		case err := <-results:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("overlapping generated saves did not finish")
		}
	}
	requests, events := e.observer.snapshot(), e.sink.snapshot()
	if len(requests) != 4 || len(events) != 4 || e.observer.commits-commitsBefore != 2 {
		t.Fatalf("expected two independent root+child commits, commands=%d audits=%d commits=%d", len(requests), len(events), e.observer.commits-commitsBefore)
	}
	wantByRoot := make(map[uint64][]*core.TraceNode)
	for index, order := range orders {
		itemID, present := order_item.NewOrderItemExpression(items[index]).Id().Eval()
		if !present || itemID == 0 {
			t.Fatal("the generated child save lost its allocated identity")
		}
		wantByRoot[ids[index]] = []*core.TraceNode{
			reasonNode("Customer Order", ids[index], reasons[index]),
			reasonNode("Order Item", itemID, branches[index]),
		}
		if order.EntityRoot() != items[index].EntityRoot() || len(order.EntityRoot().Changes()) != 0 {
			t.Fatal("explicit mutable child did not join and clear only its owning root ledger")
		}
		for _, key := range order.EntityRoot().Keys() {
			if key.Entity == "Platform" || (key.Entity == "Customer Order" && !reflect.DeepEqual(key.ID, core.ValU64(ids[index]))) {
				t.Fatalf("commit retained another root or readonly reference: %+v", key)
			}
		}
	}
	assertObserved := func(trace []*core.TraceNode) {
		t.Helper()
		if len(trace) == 0 || trace[0].EntityId == nil {
			t.Fatal("mutation evidence has no owning root identity")
		}
		want, exists := wantByRoot[*trace[0].EntityId]
		if !exists || len(trace) > len(want) {
			t.Fatalf("another graph entered the emitted lineage: %+v", trace)
		}
		assertLineage(t, trace, want[:len(trace)])
	}
	for index, request := range requests {
		assertObserved(request.TraceChain())
		assertLineage(t, events[index].TraceChain, request.TraceChain())
		governance := events[index].MutationGovernance
		if governance == nil || len(governance.Operations) != 2 {
			t.Fatal("the reviewed plan contains phantom inserts or another graph's pending changes")
		}
		for _, operation := range governance.Operations {
			switch operation.Entity {
			case "Customer Order":
				id, present := operation.ID.TryU64()
				if !present || id != *request.TraceChain()[0].EntityId || operation.Kind != core.MutationUpdate {
					t.Fatal("the reviewed plan mixed root identity or mutation kind")
				}
			case "Order Item":
				if operation.Kind != core.MutationInsert {
					t.Fatal("the reviewed plan lost its explicit new child")
				}
			default:
				t.Fatalf("readonly or unrelated entity entered the reviewed plan: %s", operation.Entity)
			}
		}
		switch mutation := request.(type) {
		case *data_service.UpdateMutation:
			if mutation.Cmd.Entity != "Customer Order" || len(request.TraceChain()) != 1 {
				t.Fatal("the unmodified reference or another entity emitted UPDATE")
			}
		case *data_service.InsertMutation:
			if mutation.Cmd.Entity != "Order Item" || len(request.TraceChain()) != 2 {
				t.Fatal("the unmodified reference or another entity emitted INSERT")
			}
		default:
			t.Fatalf("unexpected mutation type %T", request)
		}
	}
	statements := e.sqlEvidence.Snapshot()
	if len(statements) < 4 {
		t.Fatal("no real physical SQL evidence for the generated saves")
	}
	for _, statement := range statements {
		assertObserved(statement.MutationLineage)
		if len(statement.TraceChain) != 4 || statement.TraceChain[0].Name != "Customer Order" {
			t.Fatalf("physical SQL replaced the owning operation path: %+v", statement.TraceChain)
		}
	}
	for index := range orders {
		persisted, err := lib.Q.CustomerOrders().WithIdIs(ids[index]).Limit(1).
			SelectId().SelectVersion().SelectDescription().SelectPlatformWith(lib.Q.Platforms().Limit(1)).
			Comment("reload independently committed root").Purpose("verify values and readonly reference version").ExecuteForList(e.context)
		if err != nil || len(persisted.Data) != 1 {
			t.Fatalf("root reload failed: %v", err)
		}
		value, present := customer_order.NewCustomerOrderExpression(persisted.Data[0]).Description().Eval()
		version, versionPresent := customer_order.NewCustomerOrderExpression(persisted.Data[0]).Version().Eval()
		if !present || value != fmt.Sprintf("isolated graph %d after version %d", index, versions[index]) || !versionPresent || version != versions[index]+1 {
			t.Fatalf("independent root values/version were lost: value=%q version=%d", value, version)
		}
		related, _ := persisted.Data[0].RelationEntity("platformEntity")
		reference, typed := related.(*platform.Platform)
		if !typed {
			t.Fatal("reloaded root lost its readonly relation")
		}
		if version, present := platform.NewPlatformExpression(reference).Version().Eval(); !present || version != referenceVersion {
			t.Fatal("save wrote or bumped the shared readonly reference")
		}
		itemID, _ := order_item.NewOrderItemExpression(items[index]).Id().Eval()
		children, err := lib.Q.OrderItems().WithIdIs(itemID).Limit(1).
			Comment("reload independently committed child").Purpose("verify correct graph ownership").ExecuteForList(e.context)
		if err != nil || len(children.Data) != 1 {
			t.Fatalf("child reload failed: %v", err)
		}
		childExpression := order_item.NewOrderItemExpression(children.Data[0])
		owner, ownerPresent := childExpression.CustomerOrderId().Eval()
		name, namePresent := childExpression.Name().Eval()
		if !ownerPresent || owner != ids[index] || !namePresent || name != fmt.Sprintf("shared reference item %d version %d", index, versions[index]) {
			t.Fatal("child was saved into the other graph")
		}
	}
	t.Logf("TC-MUT-12 SHARED READONLY PASSED: commands=4 audits=4 commits=2 shared_record=true reference_version=%d root_versions=%d,%d", referenceVersion, versions[0], versions[1])
}

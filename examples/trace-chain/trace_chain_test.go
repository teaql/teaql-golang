package tracechain_test

import (
	stdcontext "context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/payment_attempt"
)

// This observer wraps the real generated SQLite executor. It does not invent
// trace nodes, mutate commands, replace Checker, or simulate a successful write.
type graphObserver struct {
	data_service.TransactionExecutor
	mu           sync.Mutex
	requests     []data_service.MutationRequest
	commits      int
	rollbacks    int
	queries      int
	begins       int
	beforeCommit func()
}

func (p *graphObserver) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	p.mu.Lock()
	p.begins++
	p.mu.Unlock()
	tx, err := p.TransactionExecutor.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &observedTransaction{Transaction: tx, observer: p}, nil
}

func (p *graphObserver) Query(ctx stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.mu.Lock()
	p.queries++
	p.mu.Unlock()
	return p.TransactionExecutor.(data_service.QueryExecutor).Query(ctx, request)
}

func (p *graphObserver) QueryStream(ctx stdcontext.Context, request *data_service.QueryRequest, chunkSize int, yield func(*data_service.StreamChunk) error) error {
	p.mu.Lock()
	p.queries++
	p.mu.Unlock()
	return p.TransactionExecutor.(data_service.StreamQueryExecutor).QueryStream(ctx, request, chunkSize, yield)
}

type observedTransaction struct {
	data_service.Transaction
	observer *graphObserver
}

func (tx *observedTransaction) Mutate(ctx stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	captured, err := data_service.CaptureMutationRequest(request)
	if err != nil {
		return nil, err
	}
	tx.observer.mu.Lock()
	tx.observer.requests = append(tx.observer.requests, captured)
	tx.observer.mu.Unlock()
	return tx.Transaction.Mutate(ctx, request)
}

func (tx *observedTransaction) GenerateId(entity string) (uint64, error) {
	return tx.Transaction.(interface{ GenerateId(string) (uint64, error) }).GenerateId(entity)
}

func (tx *observedTransaction) EnsureIdFloor(ctx stdcontext.Context, entity string, floor uint64) error {
	return tx.Transaction.(interface {
		EnsureIdFloor(stdcontext.Context, string, uint64) error
	}).EnsureIdFloor(ctx, entity, floor)
}

func (tx *observedTransaction) Commit(ctx stdcontext.Context) error {
	if tx.observer.beforeCommit != nil {
		tx.observer.beforeCommit()
	}
	err := tx.Transaction.Commit(ctx)
	var committed *data_service.MutationCommittedError
	if err != nil && !errors.As(err, &committed) {
		return err
	}
	tx.observer.mu.Lock()
	tx.observer.commits++
	tx.observer.mu.Unlock()
	return err
}

type failingSafeEvents struct {
	calls int
	err   error
}

func (s *failingSafeEvents) OnSafeEvent(*runtime.UserContext, *runtime.SafeAuditEvent) error {
	s.calls++
	return s.err
}

func TestGeneratedCommittedAuditFailureKeepsLedgerAndContextUsable(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "committed before consumer failure")
	order.OrderItemList().Add(newItem(e, "committed child"))
	failure := &failingSafeEvents{err: errors.New("safe audit consumer unavailable")}
	e.context.WithAppAuditEventSink(failure)
	_, err := order.AuditAs("commit graph despite consumer failure").Save(e.context)
	var committed *data_service.MutationCommittedError
	if !errors.As(err, &committed) || !errors.Is(err, failure.err) || failure.calls != 2 || e.observer.commits != 1 || e.observer.rollbacks != 0 {
		t.Fatalf("committed graph was treated as retryable failure: error=%v calls=%d commits=%d rollbacks=%d", err, failure.calls, e.observer.commits, e.observer.rollbacks)
	}
	if e.context.GetResource("dataService") != e.observer {
		t.Fatal("after-commit failure retained transaction-scoped resources")
	}
	id, present := customer_order.NewCustomerOrderExpression(order).Id().Eval()
	if !present || id == 0 {
		t.Fatal("committed graph did not retain the allocated identity")
	}
	rows, err := lib.Q.CustomerOrders().WithIdIs(id).Limit(1).
		Comment("inspect committed graph").Purpose("verify post-commit consumer failure").ExecuteForList(e.context)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("committed root unavailable through generated Q: %v", err)
	}
	description, present := customer_order.NewCustomerOrderExpression(rows.Data[0]).Description().Eval()
	if !present || description != "committed before consumer failure" {
		t.Fatal("generated E did not observe the committed values")
	}
	e.context.WithAppAuditEventSink(e.sink)
	e.reset()
	if _, err := order.AuditAs("save unchanged committed graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	if len(e.observer.snapshot()) != 0 || len(e.sink.snapshot()) != 0 {
		t.Fatal("post-commit failure left mutations queued for duplicate persistence")
	}
	order.UpdateDescription("updated after committed consumer failure")
	if _, err := order.AuditAs("update committed graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	rows, err = lib.Q.CustomerOrders().WithIdIs(id).Limit(1).
		Comment("inspect later update").Purpose("verify cleaned optimistic version").ExecuteForList(e.context)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("post-commit update could not reload: %v", err)
	}
	description, present = customer_order.NewCustomerOrderExpression(rows.Data[0]).Description().Eval()
	version, versionPresent := customer_order.NewCustomerOrderExpression(rows.Data[0]).Version().Eval()
	if !present || description != "updated after committed consumer failure" || !versionPresent || version != 2 {
		t.Fatalf("post-commit ledger/version cleanup failed: version=%d description=%q", version, description)
	}
	if len(e.observer.snapshot()) != 1 || len(e.sink.snapshot()) != 1 {
		t.Fatal("subsequent generated save replayed previously committed children")
	}
}

func (tx *observedTransaction) Rollback(ctx stdcontext.Context) error {
	err := tx.Transaction.Rollback(ctx)
	tx.observer.mu.Lock()
	tx.observer.rollbacks++
	tx.observer.mu.Unlock()
	return err
}

func (p *graphObserver) reset() { p.mu.Lock(); p.requests = nil; p.mu.Unlock() }
func (p *graphObserver) snapshot() []data_service.MutationRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]data_service.MutationRequest(nil), p.requests...)
}

type safeEvents struct {
	mu     sync.Mutex
	events []*runtime.SafeAuditEvent
}

func (s *safeEvents) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}
func (s *safeEvents) reset() { s.mu.Lock(); s.events = nil; s.mu.Unlock() }
func (s *safeEvents) snapshot() []*runtime.SafeAuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*runtime.SafeAuditEvent(nil), s.events...)
}

type environment struct {
	context     *runtime.UserContext
	observer    *graphObserver
	sink        *safeEvents
	sqlEvidence *runtime.SQLExecutionEvidenceStore
	db          *sql.DB
}

func openEnvironment(t *testing.T) *environment {
	t.Helper()
	directory := os.Getenv("TEAQL_TRACE_CHAIN_DATABASE_DIRECTORY")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	databasePath := filepath.Join(directory, t.Name()+".sqlite")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TRACE_CHAIN_SERVICE_CORE_DATABASE_URL", databasePath)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := ctx.GetResource("db").(*sql.DB)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	observer := &graphObserver{TransactionExecutor: ctx.GetResource("dataService").(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", observer)
	sink := &safeEvents{}
	ctx.WithAppAuditEventSink(sink)
	evidence := runtime.NewSQLExecutionEvidenceStore()
	ctx.WithRuntimeTelemetrySink(evidence)
	return &environment{context: ctx, observer: observer, sink: sink, sqlEvidence: evidence, db: db}
}

func (e *environment) reset() { e.observer.reset(); e.sink.reset(); e.sqlEvidence.EnableAll() }

func newOrder(t *testing.T, e *environment, description string) *customer_order.CustomerOrder {
	t.Helper()
	order := lib.Q.CustomerOrders().Comment("initialize order fixture").Purpose("verify generated graph trace").NewEntity(e.context)
	order.UpdatePlatformId(1).UpdateOrderNumber(fmt.Sprintf("TC-%d", time.Now().UnixNano())).UpdateDescription(description)
	return order
}

func newItem(e *environment, name string) *order_item.OrderItem {
	item := lib.Q.OrderItems().Comment("initialize item fixture").Purpose("verify generated graph trace").NewEntity(e.context)
	item.UpdateName(name)
	return item
}

func assertLineage(t *testing.T, got []*core.TraceNode, want []*core.TraceNode) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lineage\ngot: %#v\nwant: %#v", got, want)
	}
}

func reasonNode(entity string, id uint64, reason string) *core.TraceNode {
	return &core.TraceNode{Kind: "auditReason", Name: entity, EntityType: entity, EntityId: &id, Comment: reason}
}

func TestGeneratedCleanParentRetainsChangedChildLineage(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "unchanged parent")
	item := newItem(e, "initial child")
	order.OrderItemList().Add(item)
	if _, err := order.AuditAs("initialize clean-parent graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	orderID, _ := customer_order.NewCustomerOrderExpression(order).Id().Eval()
	itemID, _ := order_item.NewOrderItemExpression(item).Id().Eval()
	e.reset()
	item.UpdateName("changed child").Comment("correct child details")
	if _, err := order.AuditAs("review clean-parent graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	requests := e.observer.snapshot()
	if len(requests) != 1 {
		t.Fatalf("clean parent or other child emitted redundant SQL: commands=%d", len(requests))
	}
	update, ok := requests[0].(*data_service.UpdateMutation)
	if !ok || update.Cmd.Entity != "Order Item" {
		t.Fatal("changed child did not use its own update request")
	}
	want := []*core.TraceNode{
		reasonNode("Customer Order", orderID, "review clean-parent graph"),
		reasonNode("Order Item", itemID, "correct child details"),
	}
	assertLineage(t, requests[0].TraceChain(), want)
	events := e.sink.snapshot()
	if len(events) != 1 {
		t.Fatal("clean parent produced a committed mutation audit")
	}
	assertLineage(t, events[0].TraceChain, want)
	for _, metadata := range e.sqlEvidence.Snapshot() {
		assertLineage(t, metadata.MutationLineage, want)
		if metadata.TraceChain[0].Name != "Customer Order" {
			t.Fatal("skipping clean-parent SQL reset the graph's trace root")
		}
	}
	parents, err := lib.Q.CustomerOrders().WithIdIs(orderID).Limit(1).
		Comment("inspect clean parent").Purpose("verify its version was not bumped").ExecuteForList(e.context)
	if err != nil || len(parents.Data) != 1 {
		t.Fatalf("parent query failed: %v", err)
	}
	version, present := customer_order.NewCustomerOrderExpression(parents.Data[0]).Version().Eval()
	if !present || version != 1 {
		t.Fatal("clean parent version changed")
	}
	children, err := lib.Q.OrderItems().WithIdIs(itemID).Limit(1).
		Comment("inspect changed child").Purpose("verify isolated child persistence").ExecuteForList(e.context)
	if err != nil || len(children.Data) != 1 {
		t.Fatalf("child query failed: %v", err)
	}
	name, present := order_item.NewOrderItemExpression(children.Data[0]).Name().Eval()
	version, versionPresent := order_item.NewOrderItemExpression(children.Data[0]).Version().Eval()
	if !present || name != "changed child" || !versionPresent || version != 2 {
		t.Fatal("changed child persistence was skipped with the clean parent")
	}
}

func TestGeneratedGraphMutationLineage(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "draft")
	first, removed := newItem(e, "retained item"), newItem(e, "unavailable item")
	pay := lib.Q.Payments().Comment("initialize payment").Purpose("verify graph").NewEntity(e.context)
	pay.UpdateReferenceCode("payment reference")
	attempt := lib.Q.PaymentAttempts().Comment("initialize attempt").Purpose("verify graph").NewEntity(e.context)
	attempt.UpdateReferenceCode("attempt reference")
	ship := lib.Q.Shipments().Comment("initialize shipment").Purpose("verify graph").NewEntity(e.context)
	ship.UpdateReferenceCode("shipment reference")
	pay.PaymentAttemptList().Add(attempt)
	order.OrderItemList().Add(first)
	order.OrderItemList().Add(removed)
	order.PaymentList().Add(pay)
	order.ShipmentList().Add(ship)
	if _, err := order.AuditAs("prepare graph fixture").Save(e.context); err != nil {
		t.Fatal(err)
	}
	orderID, ok := customer_order.NewCustomerOrderExpression(order).Id().Eval()
	if !ok || orderID == 0 {
		t.Fatal("allocated root ID missing")
	}
	firstID, _ := order_item.NewOrderItemExpression(first).Id().Eval()
	removedID, _ := order_item.NewOrderItemExpression(removed).Id().Eval()
	payID, _ := payment.NewPaymentExpression(pay).Id().Eval()
	attemptID, _ := payment_attempt.NewPaymentAttemptExpression(attempt).Id().Eval()
	if firstID == 0 || removedID == 0 || payID == 0 || attemptID == 0 {
		t.Fatal("allocated child ID missing")
	}
	created := e.observer.snapshot()
	if len(created) != 6 {
		t.Fatalf("created commands=%d", len(created))
	}
	for _, request := range created {
		assertLineage(t, request.TraceChain(), []*core.TraceNode{reasonNode("Customer Order", orderID, "prepare graph fixture")})
		insert, ok := request.(*data_service.InsertMutation)
		if !ok {
			t.Fatal("new graph did not emit insert")
		}
		if id, ok := insert.Cmd.Values["id"].TryU64(); !ok || id == 0 {
			t.Fatal("lineage captured before assigned identity")
		}
	}
	if orderID != payID {
		t.Fatal("fixture must exercise same numeric ID across distinct entity types")
	}

	e.reset()
	e.observer.beforeCommit = func() {
		if len(e.sink.snapshot()) != 0 {
			t.Error("audit escaped before commit")
		}
	}
	order.UpdateDescription("submitted")
	first.UpdateName("available item")
	removed.MarkForDeletion().Comment("remove unavailable item")
	pay.UpdateReferenceCode("authorized reference").Comment("authorize payment")
	attempt.UpdateReferenceCode("confirmed attempt")
	ship.UpdateReferenceCode("dispatched reference").Comment("dispatch shipment")
	if _, err := order.AuditAs("submit order").Save(e.context); err != nil {
		t.Fatal(err)
	}
	root := reasonNode("Customer Order", orderID, "submit order")
	requests := e.observer.snapshot()
	if len(requests) != 6 {
		t.Fatalf("commands=%d, want 6", len(requests))
	}
	for _, request := range requests {
		if request.Comment() == nil || *request.Comment() != "submit order" {
			t.Fatal("request root intent lost")
		}
		want := []*core.TraceNode{root}
		switch req := request.(type) {
		case *data_service.UpdateMutation:
			switch req.Cmd.Entity {
			case "Customer Order", "Order Item":
			case "Payment", "Payment Attempt":
				want = append(want, reasonNode("Payment", payID, "authorize payment"))
			case "Shipment":
				id, _ := req.Cmd.Id.TryU64()
				want = append(want, reasonNode("Shipment", id, "dispatch shipment"))
			default:
				t.Fatalf("unexpected entity %s", req.Cmd.Entity)
			}
		case *data_service.DeleteMutation:
			id, _ := req.Cmd.Id.TryU64()
			if req.Cmd.Entity != "Order Item" || id != removedID {
				t.Fatal("wrong deleted child")
			}
			want = append(want, reasonNode("Order Item", removedID, "remove unavailable item"))
		default:
			t.Fatalf("unexpected command %T", request)
		}
		assertLineage(t, request.TraceChain(), want)
	}
	events := e.sink.snapshot()
	if len(events) != 6 {
		t.Fatalf("committed safe audits=%d", len(events))
	}
	// The deleted item's loaded name is private. Trusted requests retain the
	// full reason; exported safe evidence must preserve its shape but mask it.
	expectedSafeLineage := func(request data_service.MutationRequest) []*core.TraceNode {
		want := core.CloneTraceNodes(request.TraceChain())
		if _, deleting := request.(*data_service.DeleteMutation); deleting {
			want[len(want)-1].Comment = "remove [REDACTED]"
		}
		return want
	}
	for index, event := range events {
		assertLineage(t, event.TraceChain, expectedSafeLineage(requests[index]))
		if event.AuditReason == nil || *event.AuditReason != "submit order" {
			t.Fatal("safe audit lost root intent")
		}
	}
	writes := 0
	for _, metadata := range e.sqlEvidence.Snapshot() {
		if metadata.Operation == data_service.OpQuery {
			continue
		}
		writes++
		if metadata.AuditReason == nil || *metadata.AuditReason != "submit order" {
			t.Fatal("SQL replaced request root reason with descendant comment")
		}
		if len(metadata.TraceChain) != 4 || metadata.TraceChain[0].Kind != "operation" || metadata.TraceChain[1].Kind != "entity" || metadata.TraceChain[2].Kind != "provider" || metadata.TraceChain[3].Kind != "sql" {
			t.Fatalf("non-canonical SQL path: %+v", metadata.TraceChain)
		}
		assertLineage(t, metadata.MutationLineage, expectedSafeLineage(requests[writes-1]))
	}
	if writes != 6 {
		t.Fatalf("physical writes=%d", writes)
	}
	loaded, err := lib.Q.CustomerOrders().WithIdIs(orderID).
		SelectOrderItemListWith(order_item.NewOrderItemMinimalRequest()).Limit(1).
		Comment("verify submitted graph").Purpose("assert committed generated Q/E").ExecuteForList(e.context)
	if err != nil || len(loaded.Data) != 1 {
		t.Fatalf("query=%v rows=%v", err, loaded)
	}
	description, present := customer_order.NewCustomerOrderExpression(loaded.Data[0]).Description().Eval()
	count, listPresent := customer_order.NewCustomerOrderExpression(loaded.Data[0]).OrderItemList().Size().Eval()
	if !present || description != "submitted" || !listPresent || count != 1 {
		t.Fatalf("description=%q items=%d", description, count)
	}
}

func TestGeneratedCheckerStopsEntireGraphBeforeWrite(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "invalid child")
	invalid := lib.Q.OrderItems().Comment("initialize invalid child").Purpose("verify generated Checker").NewEntity(e.context)
	order.OrderItemList().Add(invalid)
	e.reset()
	_, err := order.AuditAs("reject invalid graph").Save(e.context)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "name") {
		t.Fatalf("missing checker diagnostic: %v", err)
	}
	if e.observer.commits != 0 || len(e.observer.snapshot()) != 0 || len(e.sink.snapshot()) != 0 || len(e.sqlEvidence.Snapshot()) != 0 {
		t.Fatal("invalid graph reached provider/audit")
	}
}

func TestGeneratedDriverFailureRollsBackWithoutAudit(t *testing.T) {
	e := openEnvironment(t)
	// Test-only schema fault: persistence remains generated Mutation, never INSERT.
	if _, err := e.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS trace_unique_shipment_reference ON shipment_data(reference_code)"); err != nil {
		t.Fatal(err)
	}
	code := fmt.Sprintf("conflict-%d", time.Now().UnixNano())
	first := lib.Q.Shipments().Comment("prepare collision").Purpose("verify rollback").NewEntity(e.context)
	parent := newOrder(t, e, "prepare collision")
	first.UpdateReferenceCode(code)
	parent.ShipmentList().Add(first)
	if _, err := parent.AuditAs("prepare unique fixture").Save(e.context); err != nil {
		t.Fatal(err)
	}
	e.reset()
	order := newOrder(t, e, "must rollback")
	item := newItem(e, "must rollback too")
	order.OrderItemList().Add(item)
	conflict := lib.Q.Shipments().Comment("prepare conflicting shipment").Purpose("verify rollback").NewEntity(e.context)
	conflict.UpdateReferenceCode(code).Comment("dispatch conflict")
	order.ShipmentList().Add(conflict)
	_, err := order.AuditAs("submit rollback graph").Save(e.context)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("expected real driver UNIQUE failure, got %v", err)
	}
	if len(e.observer.snapshot()) < 3 || e.observer.rollbacks != 1 || len(e.sink.snapshot()) != 0 {
		t.Fatal("did not rollback earlier physical mutations without audit")
	}
	rootCommand, ok := e.observer.snapshot()[0].(*data_service.InsertMutation)
	if !ok {
		t.Fatal("first physical operation was not the root insert")
	}
	allocatedID, ok := rootCommand.Cmd.Values["id"].TryU64()
	if !ok {
		t.Fatal("root insert did not allocate an ID")
	}
	rows, queryErr := lib.Q.CustomerOrders().WithIdIs(allocatedID).Limit(1).
		Comment("verify rolled back graph").Purpose("assert no committed root").ExecuteForList(e.context)
	if queryErr != nil || len(rows.Data) != 0 {
		t.Fatalf("rolled-back root survived: %v", queryErr)
	}
	e.reset()
	conflict.UpdateReferenceCode(code + "-retry")
	if _, err := order.AuditAs("retry after rollback").Save(e.context); err != nil {
		t.Fatalf("rolled-back generated state cannot retry: %v", err)
	}
	if len(e.sink.snapshot()) != 3 {
		t.Fatal("retry must emit only three new committed audits")
	}
}

func TestGeneratedThreeLevelQueryTrace(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "three level query")
	pay := lib.Q.Payments().Comment("initialize payment").Purpose("verify query trace").NewEntity(e.context)
	pay.UpdateReferenceCode("query payment")
	attempt := lib.Q.PaymentAttempts().Comment("initialize attempt").Purpose("verify query trace").NewEntity(e.context)
	attempt.UpdateReferenceCode("query attempt")
	pay.PaymentAttemptList().Add(attempt)
	order.PaymentList().Add(pay)
	if _, err := order.AuditAs("prepare query graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := payment_attempt.NewPaymentAttemptExpression(attempt).Id().Eval()
	e.reset()
	rows, err := lib.Q.PaymentAttemptsMinimal().WithIdIs(id).SelectReferenceCode().SelectPaymentWith(
		lib.Q.PaymentsMinimal().SelectReferenceCode().SelectCustomerOrderWith(lib.Q.CustomerOrdersMinimal().SelectPlatformWith(lib.Q.Platforms())),
	).Limit(1).Comment("load three relation levels").Purpose("verify physical query provenance").ExecuteForList(e.context)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("query: %v", err)
	}
	code, present := payment_attempt.NewPaymentAttemptExpression(rows.Data[0]).Payment().ReferenceCode().Eval()
	if !present || code != "query payment" {
		t.Fatal("typed loaded E traversal failed")
	}
	metadata := e.sqlEvidence.Snapshot()
	if len(metadata) != 4 {
		t.Fatalf("physical queries=%d", len(metadata))
	}
	wantRelations := []string{"paymentEntity", "customerOrderEntity", "platformEntity"}
	wantOwners := []string{"Payment Attempt", "Payment", "Customer Order"}
	for index, statement := range metadata {
		if statement.Operation != data_service.OpQuery || statement.Comment == nil || *statement.Comment != "load three relation levels" || statement.Purpose == nil || *statement.Purpose != "verify physical query provenance" {
			t.Fatal("derived request lost root intent")
		}
		frames := statement.TraceChain
		if len(frames) != 4+index || frames[0].Name != "Payment Attempt" || frames[0].Kind != "operation" || frames[1].Kind != "request" || frames[len(frames)-2].Kind != "provider" || frames[len(frames)-1].Name != "select" {
			t.Fatalf("bad query path: %+v", frames)
		}
		for relation := 0; relation < index; relation++ {
			node := frames[2+relation]
			if node.Kind != "relation" || node.Name != wantRelations[relation] || node.Comment != wantOwners[relation]+"."+wantRelations[relation] {
				t.Fatalf("lost relation level %d: %+v", relation, node)
			}
		}
	}
	if len(e.sink.snapshot()) != 0 {
		t.Fatal("read-only query emitted mutation audit")
	}
}

type countingPolicy struct {
	runtime.DefaultRequestPolicy
	calls int
}

// The generated AuditAs constructor rejects invalid input with the runtime's
// typed panic. Negative tests catch only that error, never NotLoaded or success.
func rejectedIntent(run func() error) (err error) {
	defer func() {
		if failure := recover(); failure != nil {
			if required, ok := failure.(*core.RequestIntentError); ok {
				err = required
			} else {
				panic(failure)
			}
		}
	}()
	return run()
}

func (p *countingPolicy) EnforceSelect(*runtime.UserContext, *core.SelectQuery) error {
	p.calls++
	return nil
}

func TestGeneratedMissingRootIntentWithLoggingOff(t *testing.T) {
	e := openEnvironment(t)
	e.context.DisableSqlLog()
	policy := &countingPolicy{}
	e.context.SetRequestPolicy(policy)
	order := newOrder(t, e, "missing root intent")
	child := newItem(e, "annotated child").Comment("cannot replace root intent")
	order.OrderItemList().Add(child)
	for _, text := range []string{"", "\u00a0 \t"} {
		err := rejectedIntent(func() error { _, failure := order.AuditAs(text).Save(e.context); return failure })
		var required *core.RequestIntentError
		if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" || required.Field != "comment" {
			t.Fatalf("mutation diagnostic: %v", err)
		}
		_, err = lib.Q.CustomerOrders().Limit(1).Comment(text).Purpose("verify rejection").ExecuteForList(e.context)
		if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" {
			t.Fatalf("query diagnostic: %v", err)
		}
	}
	if policy.calls != 0 || e.observer.queries != 0 || e.observer.begins != 0 || len(e.observer.snapshot()) != 0 || e.observer.commits != 0 || len(e.sink.snapshot()) != 0 {
		t.Fatal("missing intent reached downstream work")
	}
}

func TestGeneratedCompleteLedgerOverrideReplacesInheritance(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "ledger override")
	pay := lib.Q.Payments().Comment("initialize payment").Purpose("verify ledger override").NewEntity(e.context)
	pay.UpdateReferenceCode("override payment")
	order.PaymentList().Add(pay)
	sibling := newItem(e, "override sibling")
	order.OrderItemList().Add(sibling)
	if _, err := order.AuditAs("prepare override").Save(e.context); err != nil {
		t.Fatal(err)
	}
	rootID, _ := customer_order.NewCustomerOrderExpression(order).Id().Eval()
	payID, _ := payment.NewPaymentExpression(pay).Id().Eval()
	siblingID, _ := order_item.NewOrderItemExpression(sibling).Id().Eval()
	full := []*core.TraceNode{reasonNode("Customer Order", rootID, "submit override graph"), reasonNode("Payment", payID, "complete reviewed authorization")}
	pay.EntityRoot().SetTraceChain(pay.EntityKey(), full)
	full[1].Comment = "must not corrupt ledger"
	order.UpdateDescription("reviewed")
	pay.UpdateReferenceCode("reviewed reference").Comment("fallback must be replaced")
	sibling.UpdateName("reviewed sibling")
	e.reset()
	if _, err := order.AuditAs("submit override graph").Save(e.context); err != nil {
		t.Fatal(err)
	}
	want := []*core.TraceNode{reasonNode("Customer Order", rootID, "submit override graph"), reasonNode("Payment", payID, "complete reviewed authorization")}
	wantByEntity := map[string][]*core.TraceNode{
		"Customer Order": {reasonNode("Customer Order", rootID, "submit override graph")},
		"Order Item":     {reasonNode("Customer Order", rootID, "submit override graph")},
		"Payment":        want,
	}
	ids := map[string]uint64{"Customer Order": rootID, "Order Item": siblingID, "Payment": payID}
	requests := e.observer.snapshot()
	events := e.sink.snapshot()
	if len(requests) != 3 || len(events) != 3 {
		t.Fatalf("wrong override graph size: commands=%d audits=%d", len(requests), len(events))
	}
	commands := make(map[string]data_service.MutationRequest)
	for _, request := range requests {
		update, ok := request.(*data_service.UpdateMutation)
		if !ok {
			t.Fatalf("override graph emitted non-update command %T", request)
		}
		id, _ := update.Cmd.Id.TryU64()
		if _, exists := commands[update.Cmd.Entity]; exists || id != ids[update.Cmd.Entity] || id == 0 {
			t.Fatalf("unexpected or duplicate command identity: %s#%d", update.Cmd.Entity, id)
		}
		assertLineage(t, request.TraceChain(), wantByEntity[update.Cmd.Entity])
		commands[update.Cmd.Entity] = request
	}
	seenEvents := make(map[string]bool)
	for _, event := range events {
		request, ok := commands[event.Entity]
		if !ok || seenEvents[event.Entity] {
			t.Fatalf("unexpected or duplicate audit entity: %s", event.Entity)
		}
		seenEvents[event.Entity] = true
		assertLineage(t, event.TraceChain, request.TraceChain())
	}
	writes := make(map[string]int)
	for _, metadata := range e.sqlEvidence.Snapshot() {
		if metadata.Operation == data_service.OpQuery {
			continue
		}
		if metadata.Operation != data_service.OpUpdate || metadata.ExecutionOutcome != "success" || metadata.AffectedRows == nil || *metadata.AffectedRows != 1 {
			t.Fatalf("override graph did not observe a successful physical update: %+v", metadata)
		}
		if len(metadata.TraceChain) != 4 || metadata.TraceChain[0].Kind != "operation" || metadata.TraceChain[0].Name != "Customer Order" || metadata.TraceChain[1].Kind != "entity" {
			t.Fatalf("override SQL lost its originating graph route: %+v", metadata.TraceChain)
		}
		entity := metadata.TraceChain[1].Name
		request, ok := commands[entity]
		if !ok {
			t.Fatalf("unexpected physical override entity: %s", entity)
		}
		assertLineage(t, metadata.MutationLineage, request.TraceChain())
		if metadata.AuditReason == nil || *metadata.AuditReason != "submit override graph" {
			t.Fatal("ledger-specific lineage replaced the SQL request's root intent")
		}
		writes[entity]++
	}
	for entity := range wantByEntity {
		if writes[entity] != 1 {
			t.Fatalf("physical updates for %s=%d, want 1", entity, writes[entity])
		}
	}
	if len(pay.EntityRoot().TraceChain(pay.EntityKey())) != 0 {
		t.Fatal("committed ledger override leaked into next operation")
	}
	t.Logf("LEDGER OVERRIDE PASSED: commands=%d writes=%d audits=%d; Payment override replaces fallback; Order Item inherits only root", len(requests), len(writes), len(events))
}

func TestGeneratedConcurrentGraphsOnOneContext(t *testing.T) {
	e := openEnvironment(t)
	left, right := newOrder(t, e, "left"), newOrder(t, e, "right")
	for index, order := range []*customer_order.CustomerOrder{left, right} {
		pay := lib.Q.Payments().Comment("initialize concurrent payment").Purpose("verify operation isolation").NewEntity(e.context)
		pay.UpdateReferenceCode(fmt.Sprintf("concurrent reference %d", index)).Comment(fmt.Sprintf("local payment branch %d", index))
		order.PaymentList().Add(pay)
	}
	firstAtCommit, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	e.observer.beforeCommit = func() { once.Do(func() { close(firstAtCommit); <-release }) }
	results := make(chan error, 2)
	go func() { _, err := left.AuditAs("submit left graph").Save(e.context); results <- err }()
	select {
	case <-firstAtCommit:
	case <-time.After(5 * time.Second):
		t.Fatal("first graph never reached commit")
	}
	secondInvoked := make(chan struct{})
	go func() {
		close(secondInvoked)
		_, err := right.AuditAs("submit right graph").Save(e.context)
		results <- err
	}()
	<-secondInvoked
	if len(e.sink.snapshot()) != 0 {
		t.Error("uncommitted concurrent graph emitted audit")
	}
	close(release)
	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent graph did not finish")
		}
	}
	requests, events := e.observer.snapshot(), e.sink.snapshot()
	if len(requests) != 4 || len(events) != 4 || e.observer.commits != 2 {
		t.Fatal("independent roots joined or lost a transaction")
	}
	for index, request := range requests {
		trace := request.TraceChain()
		root := *request.Comment()
		if trace[0].Comment != root {
			t.Fatal("shared context contaminated root reason")
		}
		if len(trace) == 2 {
			branch := "local payment branch 0"
			if root == "submit right graph" {
				branch = "local payment branch 1"
			}
			if trace[1].Comment != branch {
				t.Fatal("sibling graph contaminated local reason")
			}
		}
		assertLineage(t, events[index].TraceChain, trace)
	}
	// Independently query both committed roots through Q and extract through E.
	for _, order := range []*customer_order.CustomerOrder{left, right} {
		id, _ := customer_order.NewCustomerOrderExpression(order).Id().Eval()
		rows, err := lib.Q.CustomerOrders().WithIdIs(id).Limit(1).Comment("verify concurrent root").Purpose("assert independent graph commit").ExecuteForList(e.context)
		if err != nil || len(rows.Data) != 1 {
			t.Fatalf("concurrent query: %v", err)
		}
	}
}

func TestGeneratedSuccessfulWriteFailedReadbackRollsBack(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "original")
	if _, err := order.AuditAs("prepare readback fixture").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := customer_order.NewCustomerOrderExpression(order).Id().Eval()
	// A real SQLite trigger removes the updated row, producing an empty
	// authoritative readback after a successfully executed UPDATE.
	if _, err := e.db.Exec(`CREATE TRIGGER IF NOT EXISTS trace_readback_empty AFTER UPDATE ON customer_order_data
WHEN NEW.description = 'force-readback-empty' BEGIN DELETE FROM customer_order_data WHERE id = NEW.id; END`); err != nil {
		t.Fatal(err)
	}
	e.reset()
	order.UpdateDescription("force-readback-empty")
	_, err := order.AuditAs("verify failed readback").Save(e.context)
	if err == nil {
		t.Fatal("empty authoritative readback was accepted")
	}
	if e.observer.rollbacks != 1 || len(e.sink.snapshot()) != 0 {
		t.Fatal("failed readback emitted committed audit")
	}
	metadata := e.sqlEvidence.Snapshot()
	if len(metadata) != 2 || metadata[0].Operation != data_service.OpUpdate || metadata[0].ExecutionOutcome != "success" || metadata[0].AffectedRows == nil || *metadata[0].AffectedRows != 1 || metadata[1].Operation != data_service.OpQuery || metadata[1].ResultCount == nil || *metadata[1].ResultCount != 0 {
		t.Fatalf("write/readback evidence conflated: %+v", metadata)
	}
	if metadata[1].AuditReason == nil || *metadata[1].AuditReason != "verify failed readback" || metadata[1].TraceChain[len(metadata[1].TraceChain)-1].Name != "select" {
		t.Fatal("readback lost request intent or has duplicate SQL path")
	}
	rows, err := lib.Q.CustomerOrders().WithIdIs(id).Limit(1).Comment("verify restored row").Purpose("assert generated rollback state").ExecuteForList(e.context)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("rollback did not restore row: %v", err)
	}
	description, present := customer_order.NewCustomerOrderExpression(rows.Data[0]).Description().Eval()
	if !present || description != "original" {
		t.Fatal("failed readback committed the updated value")
	}
	e.reset()
	order.UpdateDescription("retried")
	if _, err := order.AuditAs("retry after readback failure").Save(e.context); err != nil {
		t.Fatal(err)
	}
	if len(e.sink.snapshot()) != 1 {
		t.Fatal("retry must produce one committed audit")
	}
}

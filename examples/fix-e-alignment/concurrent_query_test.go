package fixe

import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

// This application-owned probe observes real runtime boundaries. It never
// manufactures trace nodes or assigns a QueryRequestScope to a request.
type concurrentQueryProvider struct {
	data_service.QueryExecutor
	calls atomic.Int64
}

func (p *concurrentQueryProvider) Query(ctx stdcontext.Context, req *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.calls.Add(1)
	return p.QueryExecutor.Query(ctx, req)
}

type concurrentQueryAudit struct{ events atomic.Int64 }

func (s *concurrentQueryAudit) OnSafeEvent(*runtime.UserContext, *runtime.SafeAuditEvent) error {
	s.events.Add(1)
	return nil
}

type concurrentQueryEntry struct {
	ctx   *runtime.UserContext
	scope *core.QueryRequestScope
}

type concurrentQueryPolicy struct {
	runtime.DefaultRequestPolicy
	entered chan concurrentQueryEntry
	release chan struct{}
	once    sync.Once
}

func (p *concurrentQueryPolicy) unblock() { p.once.Do(func() { close(p.release) }) }

func (p *concurrentQueryPolicy) EnforceSelect(ctx *runtime.UserContext, q *core.SelectQuery) error {
	if err := p.DefaultRequestPolicy.EnforceSelect(ctx, q); err != nil {
		return err
	}
	scope := q.QueryRequestScope()
	if scope == nil || scope.Parent() != nil || (scope.Comment() != "alpha root" && scope.Comment() != "beta root") {
		return nil
	}
	select {
	case p.entered <- concurrentQueryEntry{ctx, scope}:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-p.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func concurrentQueryNodes(nodes []*core.TraceNode) []map[string]string {
	result := make([]map[string]string, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, map[string]string{"kind": node.Kind, "name": node.Name, "comment": node.Comment})
	}
	return result
}

func TestGeneratedConcurrentQueryIntentStaysRequestOwned(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("a retained SQLite path is required")
	}
	round, err := strconv.Atoi(os.Getenv("CORE236_GO_GRAPH_ROUND"))
	if err != nil || (round != 1 && round != 2) {
		t.Fatal("one explicit retained round is required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+".concurrent-query.sqlite")
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
	ctx.Context, err = contextWithQueryDeadline(ctx.Context, t)
	if err != nil {
		t.Fatal(err)
	}
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, 2, 29, 10, 15, 30, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	ids := make(map[string]uint64)
	childIDs := make(map[string]uint64)
	phones := map[string]string{"alpha": "5550701", "beta": "5550702"}
	for _, label := range []string{"alpha", "beta"} {
		order := lib.Q.PurchaseOrders().Comment("Prepare " + label + " query fixture").Purpose("Verify independent concurrent request ancestry").NewEntity(ctx)
		order.UpdateContactPhone(phones[label]).UpdateInternalNote(nil)
		child := lib.Q.OrderItems().Comment("Prepare " + label + " child").Purpose("Verify generated nested Q and loaded E").NewEntity(ctx)
		child.UpdateProductName(label + " child").UpdateUnitPrice(decimal.RequireFromString("1.00"))
		order.OrderItemList().Add(child)
		order, err = order.AuditAs("Seed " + label + " graph for request isolation").Save(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ids[label], childIDs[label] = order.Id(), child.Id()
		if order.Id() == 0 || child.Id() == 0 {
			t.Fatal("saved graph identity missing")
		}
	}
	original := ctx.GetResource("dataService").(data_service.QueryExecutor)
	provider := &concurrentQueryProvider{QueryExecutor: original}
	ctx.InsertResource("dataService", provider)
	audit := &concurrentQueryAudit{}
	ctx.WithAppAuditEventSink(audit)
	evidence := runtime.NewSQLExecutionEvidenceStore()
	ctx.WithRuntimeTelemetrySink(evidence)
	policy := &concurrentQueryPolicy{entered: make(chan concurrentQueryEntry, 2), release: make(chan struct{})}
	ctx.WithRequestPolicy(policy)
	load := func(label string, id uint64) (*purchase_order.PurchaseOrder, error) {
		return lib.Q.PurchaseOrders().WithIdIs(id).
			SelectOrderItemListWith(lib.Q.OrderItems().Limit(10).Comment(label + " children")).
			Limit(1).Comment(label + " root").Purpose(label + " purpose").ExecuteForOne(ctx)
	}
	type queryResult struct {
		label string
		row   *purchase_order.PurchaseOrder
		err   error
	}
	results := make(chan queryResult, 2)
	done := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	for _, label := range []string{"alpha", "beta"} {
		go func(label string) {
			defer workers.Done()
			row, e := load(label, ids[label])
			results <- queryResult{label, row, e}
		}(label)
	}
	go func() { workers.Wait(); close(done) }()
	defer func() { policy.unblock(); <-done }()
	entries := make([]concurrentQueryEntry, 0, 2)
	for len(entries) < 2 {
		select {
		case entry := <-policy.entered:
			entries = append(entries, entry)
		case failure := <-results:
			t.Fatalf("query completed before overlap: %s %v", failure.label, failure.err)
		case <-ctx.Done():
			t.Fatal("both queries did not reach the live checkpoint", ctx.Err())
		}
	}
	if len(results) != 0 || entries[0].ctx != ctx || entries[1].ctx != ctx || entries[0].scope.Root() == entries[1].scope.Root() {
		t.Fatal("expected two distinct live request roots on the identical Context")
	}
	checkpoints := make([]map[string]any, 0, 2)
	for _, entry := range entries {
		checkpoints = append(checkpoints, map[string]any{"comment": entry.scope.Comment(), "purpose": entry.scope.Purpose(), "rootIsSelf": entry.scope.Root() == entry.scope,
			"parentAbsent": entry.scope.Parent() == nil, "nodes": concurrentQueryNodes(entry.scope.Recover())})
	}
	policy.unblock()
	<-done
	loaded := make(map[string]*purchase_order.PurchaseOrder)
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err != nil || r.row == nil {
			t.Fatal("generated query failed", r.label, r.err)
		}
		loaded[r.label] = r.row
	}
	after, err := load("after", ids["alpha"])
	if err != nil || after == nil {
		t.Fatal("following independent request failed", err)
	}
	loaded["after"] = after
	beforeE := provider.calls.Load()
	values := make(map[string]any)
	for _, label := range []string{"alpha", "beta", "after"} {
		row := loaded[label]
		phone, hasPhone := lib.E.PurchaseOrder(row).ContactPhone().Eval()
		count, hasCount := lib.E.PurchaseOrder(row).OrderItemList().Size().Eval()
		product, hasProduct := lib.E.PurchaseOrder(row).OrderItemList().First().ProductName().Eval()
		owner := label
		if label == "after" {
			owner = "alpha"
		}
		if !hasPhone || !hasCount || !hasProduct || phone != phones[owner] || count != 1 || product != owner+" child" || row.Id() != ids[owner] {
			t.Fatal("nested generated Q/E crossed request data", label, phone, count, product)
		}
		values[label] = map[string]any{"id": row.Id(), "phone": phone, "count": count, "product": product}
	}
	events := make([]map[string]any, 0)
	for _, m := range evidence.Snapshot() {
		comment, purpose := "", ""
		if m.Comment != nil {
			comment = *m.Comment
		}
		if m.Purpose != nil {
			purpose = *m.Purpose
		}
		events = append(events, map[string]any{"operation": m.Operation, "outcome": m.ExecutionOutcome, "comment": comment, "purpose": purpose,
			"business": concurrentQueryNodes(m.RequestTraceChain), "physical": concurrentQueryNodes(m.TraceChain), "resultCount": m.ResultCount, "sql": m.ParameterizedSQL})
	}
	frame := map[string]any{"round": round, "language": "go", "overlapEntered": len(entries), "sameContext": true, "distinctRoots": true,
		"checkpoints": checkpoints, "ids": ids, "childIds": childIDs, "values": values, "events": events,
		"providerQueries": provider.calls.Load(), "eProviderReads": provider.calls.Load() - beforeE, "auditEvents": audit.events.Load()}
	printFrame := func(marker string) {
		bytes, e := json.Marshal(frame)
		if e != nil {
			t.Fatal(e)
		}
		fmt.Println(marker + string(bytes))
	}
	printFrame("GO_CONCURRENT_QUERY_OBSERVATION=")
	if len(events) != 6 || provider.calls.Load() != 6 || provider.calls.Load()-beforeE != 0 || audit.events.Load() != 0 {
		t.Fatal("expected six real SELECTs, no implicit E reads, and no query mutation audits")
	}
	for _, label := range []string{"alpha", "beta", "after"} {
		var selected []map[string]any
		for _, e := range events {
			if e["purpose"] == label+" purpose" {
				selected = append(selected, e)
			}
		}
		if len(selected) != 2 {
			t.Fatal("root/child SQL envelope intent lost", label, selected)
		}
		for _, e := range selected {
			if e["purpose"] != label+" purpose" || e["operation"] != data_service.OpQuery || e["outcome"] != "success" {
				t.Fatal("SQL evidence lost request intent", e)
			}
		}
		root := []map[string]string{{"kind": "purpose", "name": "purchase_order", "comment": label + " purpose"}, {"kind": "comment", "name": "purchase_order", "comment": label + " root"}}
		if !reflect.DeepEqual(selected[0]["business"], root) {
			t.Fatal("root request lineage differs", selected[0])
		}
		child := append(append([]map[string]string{}, root...), map[string]string{"kind": "comment", "name": "SelectOrderItemListWith", "comment": "SelectOrderItemListWith(...)"}, map[string]string{"kind": "comment", "name": "order_item", "comment": label + " children"})
		if !reflect.DeepEqual(selected[1]["business"], child) {
			t.Fatal("nested operation/local comment lineage differs", selected[1])
		}
		if selected[0]["comment"] != label+" root" || selected[1]["comment"] != label+" children" {
			t.Fatal("SQL envelope must use its local request comment", selected)
		}
		physicalRoot := []map[string]string{{"kind": "operation", "name": "purchase_order", "comment": "query"}, {"kind": "request", "name": "purchase_order", "comment": ""}}
		for index, event := range selected {
			path := append([]map[string]string{}, physicalRoot...)
			if index == 1 {
				path = append(path, map[string]string{"kind": "relation", "name": "orderItemList", "comment": "purchase_order.orderItemList"})
			}
			path = append(path, map[string]string{"kind": "provider", "name": "sqlite", "comment": ""}, map[string]string{"kind": "sql", "name": "select", "comment": ""})
			if !reflect.DeepEqual(event["physical"], path) {
				t.Fatal("physical SQL path differs", event)
			}
		}
	}
	frame["passed"] = true
	printFrame("GO_CONCURRENT_QUERY=")
}

func contextWithQueryDeadline(ctx stdcontext.Context, t *testing.T) (stdcontext.Context, error) {
	if ctx == nil {
		return nil, fmt.Errorf("runtime Context must embed an execution context")
	}
	bounded, cancel := stdcontext.WithTimeout(ctx, 20*time.Second)
	t.Cleanup(cancel)
	return bounded, nil
}

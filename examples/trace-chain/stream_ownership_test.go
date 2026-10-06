package tracechain_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
)

func TestGeneratedStreamsOverlapWithIndependentSavesAndSafeTermination(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	e := openEnvironment(t)
	parent := newOrder(t, e, "stream fixture")
	if _, err := parent.AuditAs("allocate stream fixture").Save(e.context); err != nil {
		t.Fatal(err)
	}
	parentID, _ := customer_order.NewCustomerOrderExpression(parent).Id().Eval()
	names := []string{fmt.Sprintf("PRIVATE-STREAM-%d-left", parentID), fmt.Sprintf("PRIVATE-STREAM-%d-right", parentID)}
	for _, name := range names {
		for i := 0; i < 6; i++ {
			parent.OrderItemList().Add(newItem(e, name))
		}
	}
	if _, err := parent.AuditAs("seed streamed items").Save(e.context); err != nil {
		t.Fatal(err)
	}
	// Two physical connections prove real overlap, not just two started goroutines.
	e.db.SetMaxOpenConns(2)
	e.reset()
	entered := make(chan int, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	var workers sync.WaitGroup
	defer func() { unblock(); workers.Wait() }()
	done := make(chan error, 2)
	rows := make([][]*order_item.OrderItem, 2)
	for index, name := range names {
		workers.Add(1)
		go func(index int, name string) {
			defer workers.Done()
			err := lib.Q.OrderItems().WithNameIs(name).OrderByIdAsc().Limit(10).
				Comment(fmt.Sprintf("stream %d %s", index, name)).Purpose("inspect "+name).
				ExecuteForStream(e.context, 2, func(item *order_item.OrderItem) error {
					if len(rows[index]) == 0 {
						entered <- index
						<-release
					}
					rows[index] = append(rows[index], item)
					return nil
				})
			done <- err
		}(index, name)
	}
	for i := 0; i < 2; i++ {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			unblock()
			t.Fatal("two streams did not reach their first callback")
		}
	}
	inUse := e.db.Stats().InUse
	unblock()
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if inUse != 2 || e.db.Stats().InUse != 0 {
		t.Fatal("two cursors did not overlap and close")
	}
	statements := e.sqlEvidence.Snapshot()
	if len(statements) != 2 {
		t.Fatalf("expected two physical stream facts, got %d", len(statements))
	}
	comments := make(map[string]bool)
	for _, entry := range statements {
		if entry.Comment == nil {
			t.Fatal("stream lost request comment")
		}
		comments[*entry.Comment] = true
		encoded, _ := json.Marshal(entry)
		for _, name := range names {
			if strings.Contains(string(encoded), name) {
				t.Fatal("stream intent leaked masked field")
			}
		}
		path := entry.TraceChain
		if entry.ExecutionOutcome != "success" || entry.ResultCount == nil || *entry.ResultCount != 6 || len(path) != 4 || path[0].Name != "Order Item" || path[2].Name != "sqlite" || path[3].Name != "select" {
			t.Fatal("stream terminal fact lost path or row count")
		}
	}
	if !comments["stream 0 [REDACTED]"] || !comments["stream 1 [REDACTED]"] {
		t.Fatal("concurrent streams did not retain distinct request intent")
	}
	ids := make([]uint64, 0, 12)
	for index, group := range rows {
		if len(group) != 6 {
			t.Fatal("stream lost rows")
		}
		for _, item := range group {
			expr := order_item.NewOrderItemExpression(item)
			id, _ := expr.Id().Eval()
			name, _ := expr.Name().Eval()
			version, _ := expr.Version().Eval()
			if name != names[index] || version != 1 {
				t.Fatal("stream results crossed requests")
			}
			ids = append(ids, id)
		}
	}
	first, second := rows[0][0], rows[0][1]
	if first.EntityRoot() == second.EntityRoot() || first.EntityRoot() == rows[1][0].EntityRoot() {
		t.Fatal("stream roots share a mutation ledger")
	}
	first.UpdateName("stream first saved")
	second.UpdateName("stream second saved")
	e.reset()
	if _, err := first.AuditAs("save first stream root").Save(e.context); err != nil {
		t.Fatal(err)
	}
	if len(e.observer.snapshot()) != 1 || len(e.sink.snapshot()) != 1 || len(second.EntityRoot().Changes()) == 0 {
		t.Fatal("first save crossed root ownership")
	}
	pending, err := lib.Q.OrderItems().WithIdIs(ids[1]).Limit(1).Comment("inspect pending row").Purpose("verify independent stream save").ExecuteForOne(e.context)
	if err != nil || pending == nil {
		t.Fatalf("pending row missing: %v", err)
	}
	expr := order_item.NewOrderItemExpression(pending)
	pendingName, _ := expr.Name().Eval()
	pendingVersion, _ := expr.Version().Eval()
	if pendingName != names[0] || pendingVersion != 1 {
		t.Fatal("second root saved prematurely")
	}
	if _, err := second.AuditAs("save second stream root").Save(e.context); err != nil {
		t.Fatal(err)
	}
	for index, id := range ids[:2] {
		item, err := lib.Q.OrderItems().WithIdIs(id).Limit(1).Comment("reload saved stream row").Purpose("confirm values and versions").ExecuteForOne(e.context)
		if err != nil || item == nil {
			t.Fatalf("saved row missing: %v", err)
		}
		expr := order_item.NewOrderItemExpression(item)
		version, _ := expr.Version().Eval()
		name, _ := expr.Name().Eval()
		if version != 2 || name != []string{"stream first saved", "stream second saved"}[index] {
			t.Fatal("independent save did not persist")
		}
	}
	e.reset()
	stop := errors.New("consumer stops after one entity")
	callbacks := 0
	err = lib.Q.OrderItems().WithNameIs(names[1]).OrderByIdAsc().Limit(10).
		Comment("cancel "+names[1]).Purpose("cancel intent "+names[1]).ExecuteForStream(e.context, 1, func(*order_item.OrderItem) error { callbacks++; return stop })
	if !errors.Is(err, stop) || callbacks != 1 || e.db.Stats().InUse != 0 {
		t.Fatal("consumer stop did not release stream cursor")
	}
	terminal := e.sqlEvidence.Snapshot()
	if len(terminal) != 1 || terminal[0].ExecutionOutcome != "cancelled" || terminal[0].ResultCount == nil || *terminal[0].ResultCount != 1 {
		t.Fatal("missing precise cancellation fact")
	}
	encoded, _ := json.Marshal(terminal)
	if strings.Contains(string(encoded), names[1]) {
		t.Fatal("cancellation leaked private intent")
	}
	observed, _ := json.Marshal(map[string]any{"ids": ids, "sql": statements, "cancel": terminal[0], "overlapping_connections": inUse})
	t.Log("STREAM_OBSERVED " + string(observed))
	t.Log("GENERATED STREAM PASSED: real overlapping cursors, independent ledgers, safe terminal facts")
}

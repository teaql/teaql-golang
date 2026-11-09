package tracechain_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
)

func TestGeneratedCancelledQueuedSaveRetainsPendingMutation(t *testing.T) {
	for _, logging := range []bool{true, false} {
		label := "logging-on"
		if !logging {
			label = "logging-off"
		}
		t.Run(label, func(t *testing.T) {
			e := openEnvironment(t)
			if !logging {
				e.context.DisableSqlLog()
			}
			accepted := newOrder(t, e, "uncancelled owner graph")
			pending := newOrder(t, e, "cancelled queued graph")
			acceptedID, _ := customer_order.NewCustomerOrderExpression(accepted).Id().Eval()
			pendingID, pendingIDLoaded := customer_order.NewCustomerOrderExpression(pending).Id().Eval()
			versionBefore, versionLoadedBefore := customer_order.NewCustomerOrderExpression(pending).Version().Eval()
			changesBefore := len(pending.EntityRoot().Changes())
			if !pendingIDLoaded || pendingID == 0 || pendingID == acceptedID || !pending.EntityRoot().IsNew(pending.EntityKey()) || changesBefore == 0 {
				t.Fatal("fixture did not create two independent pending generated entities")
			}
			e.reset()
			originalContext := e.context.Context
			atCommit, release := make(chan struct{}), make(chan struct{})
			var commitOnce, releaseOnce sync.Once
			releaseFirst := func() {
				releaseOnce.Do(func() {
					e.context.Context = originalContext
					close(release)
				})
			}
			e.observer.beforeCommit = func() { commitOnce.Do(func() { close(atCommit); <-release }) }
			ownerDone, waiterDone := make(chan error, 1), make(chan error, 1)
			ownerJoined, waiterJoined, waiterStarted := false, false, false
			defer func() {
				releaseFirst()
				if !ownerJoined {
					select {
					case <-ownerDone:
					case <-time.After(5 * time.Second):
						t.Error("owned generated owner did not terminate")
					}
				}
				if waiterStarted && !waiterJoined {
					select {
					case <-waiterDone:
					case <-time.After(5 * time.Second):
						t.Error("owned generated waiter did not terminate")
					}
				}
			}()
			go func() { _, err := accepted.AuditAs("commit the uncancelled graph").Save(e.context); ownerDone <- err }()
			select {
			case <-atCommit:
			case <-time.After(5 * time.Second):
				t.Fatal("owner never reached real COMMIT")
			}
			requestsBefore, sqlBefore := len(e.observer.snapshot()), len(e.sqlEvidence.Snapshot())
			if requestsBefore == 0 || (logging && sqlBefore == 0) || len(e.sink.snapshot()) != 0 {
				t.Fatal("fixture did not observe an uncommitted owner mutation")
			}
			queued, cancel := context.WithCancel(context.Background())
			defer cancel()
			e.context.Context = queued
			waiterStarted = true
			go func() { _, err := pending.AuditAs("cancel the queued graph").Save(e.context); waiterDone <- err }()
			awaitBlockedGeneratedSave(t)
			cancel()
			select {
			case err := <-waiterDone:
				waiterJoined = true
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("generated queued save returned %v, want context.Canceled", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("cancelled generated Save waited for the owner's COMMIT")
			}
			versionAfter, versionLoadedAfter := customer_order.NewCustomerOrderExpression(pending).Version().Eval()
			if versionAfter != versionBefore || versionLoadedAfter != versionLoadedBefore || !pending.EntityRoot().IsNew(pending.EntityKey()) || len(pending.EntityRoot().Changes()) != changesBefore {
				t.Fatal("cancelled admission changed the pending entity's version or mutation intent")
			}
			if len(e.observer.snapshot()) != requestsBefore || len(e.sqlEvidence.Snapshot()) != sqlBefore || len(e.sink.snapshot()) != 0 || e.observer.begins != 1 || e.observer.commits != 0 {
				t.Fatal("cancelled graph performed IO, entered a transaction or published a committed audit")
			}
			releaseFirst()
			select {
			case err := <-ownerDone:
				ownerJoined = true
				if err != nil {
					t.Fatalf("queued cancellation affected the uncancelled graph: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("uncancelled graph did not finish")
			}
			absent, err := lib.Q.CustomerOrders().WithIdIs(pendingID).Limit(1).
				Comment("verify the cancelled graph was not persisted").Purpose("observe graph isolation through generated Q").ExecuteForList(e.context)
			if err != nil || len(absent.Data) != 0 {
				t.Fatalf("cancelled root was persisted by another graph: %v", err)
			}
			if _, err := pending.AuditAs("retry the cancelled graph independently").Save(e.context); err != nil {
				t.Fatal(err)
			}
			rows, err := lib.Q.CustomerOrders().WithIdIs(pendingID).Limit(1).
				Comment("reload the independently retried graph").Purpose("verify persisted identity and version through Q and E").ExecuteForList(e.context)
			if err != nil || len(rows.Data) != 1 {
				t.Fatalf("retry was not persisted: %v", err)
			}
			id, _ := customer_order.NewCustomerOrderExpression(rows.Data[0]).Id().Eval()
			version, loaded := customer_order.NewCustomerOrderExpression(rows.Data[0]).Version().Eval()
			events := e.sink.snapshot()
			if id != pendingID || !loaded || version != 1 || pending.EntityRoot().IsNew(pending.EntityKey()) || len(pending.EntityRoot().Changes()) != 0 || e.observer.begins != 2 || e.observer.commits != 2 || e.observer.rollbacks != 0 || len(events) != 2 {
				t.Fatal("same-entity retry lost identity, pending-state cleanup or independent transaction/audit counts")
			}
			for i, reason := range []string{"commit the uncancelled graph", "retry the cancelled graph independently"} {
				root := acceptedID
				if i == 1 {
					root = pendingID
				}
				if len(events[i].TraceChain) != 1 || events[i].TraceChain[0].EntityId == nil || *events[i].TraceChain[0].EntityId != root || events[i].TraceChain[0].Comment != reason {
					t.Fatal("queued cancellation mixed the independent committed audit lineages")
				}
			}
			b, err := json.Marshal(map[string]interface{}{"logging": logging, "acceptedId": acceptedID, "retriedId": pendingID, "queuedBeforeCancellation": true, "cancelledBeforeOwnerCommit": true, "cancelledProviderEntries": 0, "cancelledSQLStatements": 0, "pendingIntentRetained": true, "retriedSameEntity": true, "retriedVersion": version, "commits": e.observer.commits, "audits": len(events)})
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("GO_QUEUED_SAVE_CANCELLATION=%s\n", b)
		})
	}
}

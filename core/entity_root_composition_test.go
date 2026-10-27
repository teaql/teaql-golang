package core

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestLedgerCompositionRejectsConflictingVersionsWithoutCopyingIntent(t *testing.T) {
	target, source := NewEntityRoot(), NewEntityRoot()
	key, added := NewEntityKey("PurchaseOrder", ValU64(7)), NewEntityKey("OrderItem", ValI64(-1))
	target.SetOriginalVersion(key, 3)
	target.Set(key, "phone", ValText("target pending"))
	source.SetOriginalVersion(key, 4)
	source.Set(key, "phone", ValText("source pending"))
	source.MarkAsNew(added)
	source.Set(added, "quantity", ValI64(2))
	before := target.Change(key)
	target.MergeFrom(source)
	if version, known := target.OriginalVersion(key); !known || version != 3 {
		t.Fatalf("conflicting composition overwrote optimistic baseline: version=%d known=%v", version, known)
	}
	if !reflect.DeepEqual(target.Change(key), before) || len(target.Change(added)) != 0 || target.IsNew(added) {
		t.Fatal("conflicting composition partially copied intent")
	}
	if source.Change(key)["phone"].V != "source pending" || !source.IsNew(added) {
		t.Fatal("source intent was consumed by rejected composition")
	}
}

func TestLedgerCompositionExplicitErrorIsAtomicAndRetryable(t *testing.T) {
	target, source := NewEntityRoot(), NewEntityRoot()
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	target.SetOriginalVersion(key, 3)
	source.SetOriginalVersion(key, 4)
	target.Set(key, "phone", ValText("target"))
	source.Set(key, "phone", ValText("source"))
	var conflict *LedgerVersionConflict
	if !errors.As(target.TryMergeFrom(source), &conflict) || conflict.TargetVersion != 3 || conflict.SourceVersion != 4 || conflict.Key != key {
		t.Fatal("missing exact typed conflict")
	}
	if target.CompositionError() != nil || target.Change(key)["phone"].V != "target" || source.Change(key)["phone"].V != "source" {
		t.Fatal("failable composition changed either owner")
	}
	source.SetOriginalVersion(key, 3)
	if err := target.TryMergeFrom(source); err != nil || target.Change(key)["phone"].V != "source" {
		t.Fatal("corrected explicit composition cannot retry", err)
	}
}

func TestLedgerCompositionVoidErrorCannotBeMutatedClearedOrImportedAsValid(t *testing.T) {
	target, source := NewEntityRoot(), NewEntityRoot()
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	target.SetOriginalVersion(key, 3)
	source.SetOriginalVersion(key, 4)
	target.MergeFrom(source)
	var conflict *LedgerVersionConflict
	if !errors.As(target.CompositionError(), &conflict) {
		t.Fatal("void composition lost the fault")
	}
	conflict.TargetVersion = 999
	target.ClearCommitted()
	if !errors.As(target.CompositionError(), &conflict) || conflict.TargetVersion != 3 {
		t.Fatal("caller or ClearCommitted erased the fault")
	}
	other := NewEntityRoot()
	other.MergeFrom(target)
	if other.CompositionError() == nil {
		t.Fatal("a poisoned source was imported as valid")
	}
}

func TestLedgerCompositionValidDeletionKeepsItsLifecycleAndOwnedTrace(t *testing.T) {
	target, source := NewEntityRoot(), NewEntityRoot()
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	target.SetOriginalVersion(key, 3)
	source.SetOriginalVersion(key, 3)
	target.Set(key, "phone", ValText("pending"))
	source.MarkAsDeleted(key)
	source.SetTraceChain(key, []*TraceNode{{Name: "PurchaseOrder", Comment: "remove an order"}})
	if err := target.TryMergeFrom(source); err != nil {
		t.Fatal(err)
	}
	if !target.IsDeleted(key) || len(target.Change(key)) != 0 || !source.IsDeleted(key) {
		t.Fatal("valid deletion lost its effective lifecycle")
	}
	trace := source.TraceChain(key)
	trace[0].Comment = "outside mutation"
	if target.TraceChain(key)[0].Comment != "remove an order" {
		t.Fatal("imported trace aliases the source")
	}
}

func TestLedgerCompositionOppositeDirectionsDoNotDeadlockOrSharePendingMaps(t *testing.T) {
	left, right := NewEntityRoot(), NewEntityRoot()
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	left.SetOriginalVersion(key, 3)
	right.SetOriginalVersion(key, 3)
	left.Set(key, "left", ValI64(1))
	right.Set(key, "right", ValI64(2))
	var workers sync.WaitGroup
	for _, pair := range [][2]*EntityRoot{{left, right}, {right, left}} {
		workers.Add(1)
		go func(pair [2]*EntityRoot) {
			defer workers.Done()
			for n := 0; n < 64; n++ {
				if err := pair[0].TryMergeFrom(pair[1]); err != nil {
					t.Error(err)
					return
				}
			}
		}(pair)
	}
	done := make(chan struct{})
	go func() { workers.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("opposing composition deadlocked")
	}
	left.Set(key, "left", ValI64(99))
	if right.Change(key)["left"].V != int64(1) {
		t.Fatal("pending maps were shared")
	}
}

func TestLedgerCompositionSeededTrialsPreserveEveryRejectedGraph(t *testing.T) {
	const seed int64 = 0x434f4d504f5345
	random := rand.New(rand.NewSource(seed))
	for trial := 0; trial < 256; trial++ {
		target, source := NewEntityRoot(), NewEntityRoot()
		key, extra := NewEntityKey("PurchaseOrder", ValU64(7)), NewEntityKey("OrderItem", ValU64(7))
		version := int64(random.Intn(11) + 1)
		sourceVersion := version
		if trial%3 == 0 {
			sourceVersion++
		}
		target.SetOriginalVersion(key, version)
		source.SetOriginalVersion(key, sourceVersion)
		source.SetOriginalVersion(extra, 19)
		target.Set(key, "phone", ValText("target"))
		for n := 0; n < 16; n++ {
			source.Set(key, fmt.Sprintf("field_%d", random.Intn(7)), ValI64(random.Int63n(100)))
			source.Set(extra, "quantity", ValI64(int64(n)))
		}
		before, sourceBefore := target.Change(key), source.Change(key)
		err := target.TryMergeFrom(source)
		if version != sourceVersion {
			if err == nil || !reflect.DeepEqual(target.Change(key), before) || len(target.Change(extra)) != 0 {
				t.Fatalf("non-atomic rejection: seed=%x trial=%d", seed, trial)
			}
		} else if err != nil || !reflect.DeepEqual(target.Change(extra), source.Change(extra)) {
			t.Fatalf("lost valid composition: seed=%x trial=%d err=%v", seed, trial, err)
		}
		if !reflect.DeepEqual(source.Change(key), sourceBefore) {
			t.Fatalf("source changed: seed=%x trial=%d", seed, trial)
		}
	}
}

func TestDiscardRestoredVersionGuardPreservesIntentAfterBaselineReplacement(t *testing.T) {
	root := NewEntityRoot()
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", ValText("same"))
	observed := root.Change(key)
	root.SetOriginalVersion(key, 4)
	removed, err := root.DiscardFieldsAtVersion(key, 3, observed)
	var conflict *LedgerVersionConflict
	if !errors.As(err, &conflict) || conflict.SourceVersion != 3 || conflict.TargetVersion != 4 || len(removed) != 0 || root.Change(key)["phone"].V != "same" {
		t.Fatal("stale comparison erased intent or lost its typed versions", err)
	}
	removed, err = root.DiscardFieldsAtVersion(key, 4, observed)
	if err != nil || len(removed) != 1 || len(root.Change(key)) != 0 {
		t.Fatal("matching current baseline cannot cancel", err)
	}
	root.SetOriginalVersion(key, 5)
	if version, _ := root.OriginalVersion(key); version != 5 || root.CompositionError() != nil {
		t.Fatal("ordinary authoritative commit/rollback refresh was prohibited")
	}
}

func TestDiscardRestoredVersionProvenanceUsesTheTypedOwner(t *testing.T) {
	key := NewEntityKey("PurchaseOrder", ValU64(7))
	snapshot := NewOwnedLoadedScalarSnapshot(key, 3, Record{"phone": ValText("same")}, "phone")
	if version, known := snapshot.OriginalVersionFor(key); !known || version != 3 {
		t.Fatal("missing authoritative version")
	}
	for _, other := range []EntityKey{NewEntityKey("OrderItem", ValU64(7)), NewEntityKey("PurchaseOrder", ValI64(7)), NewEntityKey("PurchaseOrder", ValU64(8))} {
		if _, known := snapshot.OriginalVersionFor(other); known {
			t.Fatal("typed owner was ignored")
		}
	}
	if _, known := NewLoadedScalarSnapshot(Record{"phone": ValText("same")}, "phone").OriginalVersionFor(key); known {
		t.Fatal("privacy-only snapshot became authoritative")
	}
}

func TestLedgerCompositionKeepsEqualVersionsAndDifferentTypesIndependent(t *testing.T) {
	target, source := NewEntityRoot(), NewEntityRoot()
	order, item := NewEntityKey("PurchaseOrder", ValU64(7)), NewEntityKey("OrderItem", ValU64(7))
	target.SetOriginalVersion(order, 3)
	source.SetOriginalVersion(order, 3)
	source.SetOriginalVersion(item, 9)
	source.Set(order, "phone", ValText("composed"))
	source.Set(item, "quantity", ValI64(2))
	target.MergeFrom(source)
	if version, _ := target.OriginalVersion(order); version != 3 {
		t.Fatal("same-version order was lost")
	}
	if version, _ := target.OriginalVersion(item); version != 9 {
		t.Fatal("same numeric ID collapsed entity types")
	}
	if target.Change(order)["phone"].V != "composed" || target.Change(item)["quantity"].V != int64(2) {
		t.Fatal("valid composition lost intent")
	}
	source.Set(order, "phone", ValText("later source edit"))
	if target.Change(order)["phone"].V != "composed" {
		t.Fatal("copied ledger shares the source's pending map")
	}
}

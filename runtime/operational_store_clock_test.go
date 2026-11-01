package runtime_test

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

func storeClockObservation(t *testing.T, name string, clock *referenceBusinessClock, reads int, facts map[string]bool) {
	t.Helper()
	passed := clock.calls == reads
	for _, value := range facts {
		passed = passed && value
	}
	data, err := json.Marshal(map[string]any{"case": name, "passed": passed, "businessReads": clock.calls, "facts": facts})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_OPERATIONAL_STORE_CASE=" + string(data))
	if !passed {
		t.Fatal("operational store contract failed", name)
	}
}

func TestOperationalStoreExpiration(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	store, one := runtime.NewInMemoryDataStore(), uint64(1)
	store.Put(ctx, "ephemeral", core.ValText("synthetic temporary value"), &one)
	value, present := store.Get(ctx, "ephemeral")
	initial := present && value.V == "synthetic temporary value"
	clock.value = clock.value.AddDate(100, 0, 0)
	_, liveAfterForward := store.Get(ctx, "ephemeral")
	time.Sleep(1100 * time.Millisecond)
	_, expired := store.Get(ctx, "ephemeral")
	clock.value = clock.value.AddDate(-200, 0, 0)
	_, revived := store.Get(ctx, "ephemeral")
	storeClockObservation(t, "store-real-expiry-no-revival", clock, 1, map[string]bool{
		"initialHit": initial, "businessForwardKeepsLive": liveAfterForward,
		"reallyExpired": !expired, "backwardCannotRevive": !revived,
	})
}

func TestOperationalStoreNullEmptyFalseZeroPresence(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	store := runtime.NewInMemoryDataStore()
	values := []core.Value{core.ValNull(), core.ValText(""), core.ValBool(false), core.ValI64(0), {V: []int{}}}
	for n, expected := range values {
		key := fmt.Sprintf("presence-%d", n)
		store.Put(ctx, key, expected, nil)
		actual, present := store.Get(ctx, key)
		if !present || !reflect.DeepEqual(actual, expected) {
			t.Fatal("a loaded empty/null/false/zero value became a cache miss")
		}
		store.Remove(ctx, key)
		if _, present := store.Get(ctx, key); present {
			t.Fatal("removed value remained present")
		}
	}
	storeClockObservation(t, "store-presence-and-remove", clock, 1, map[string]bool{"nullIsHit": true, "emptyIsHit": true, "falseIsHit": true, "zeroIsHit": true, "removeIsMiss": true})
}

func TestOperationalStoreZeroAndNilTTLPreserveEntries(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	store, zero, one := runtime.NewInMemoryDataStore(), uint64(0), uint64(1)
	store.Put(ctx, "nil", core.ValText("nil timeout"), nil)
	store.Put(ctx, "zero", core.ValText("zero timeout"), &zero)
	store.Put(ctx, "control", core.ValText("finite timeout"), &one)
	time.Sleep(1100 * time.Millisecond)
	nilValue, nilHit := store.Get(ctx, "nil")
	zeroValue, zeroHit := store.Get(ctx, "zero")
	_, finiteHit := store.Get(ctx, "control")
	storeClockObservation(t, "store-nil-zero-ttl", clock, 1, map[string]bool{"nilTTLRemains": nilHit && nilValue.V == "nil timeout", "zeroTTLRemains": zeroHit && zeroValue.V == "zero timeout", "finiteControlExpires": !finiteHit})
}

func TestOperationalStoreOverwriteReplacesExpiry(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	store, one := runtime.NewInMemoryDataStore(), uint64(1)
	store.Put(ctx, "clear-expiry", core.ValText("temporary"), &one)
	store.Put(ctx, "clear-expiry", core.ValText("persistent replacement"), nil)
	store.Put(ctx, "set-expiry", core.ValText("persistent"), nil)
	store.Put(ctx, "set-expiry", core.ValText("temporary replacement"), &one)
	time.Sleep(1100 * time.Millisecond)
	value, kept := store.Get(ctx, "clear-expiry")
	_, expired := store.Get(ctx, "set-expiry")
	storeClockObservation(t, "store-overwrite-expiry", clock, 1, map[string]bool{"newPersistentValueRetained": kept && value.V == "persistent replacement", "newFiniteValueExpires": !expired})
}

func TestOperationalStoreUnreadableBusinessClock(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	clock.broken = true
	calibrated := false
	func() {
		defer func() { calibrated = recover() == "synthetic unavailable business time" }()
		_ = ctx.BusinessTime()
	}()
	if !calibrated {
		t.Fatal("unreadable business clock was not calibrated")
	}
	store := runtime.NewInMemoryDataStore()
	store.Put(ctx, "unreadable", core.ValI64(9), nil)
	value, present := store.Get(ctx, "unreadable")
	store.Remove(ctx, "unreadable")
	storeClockObservation(t, "store-unreadable-business", clock, 2, map[string]bool{"calibrated": calibrated, "putGetRemoveIndependent": present && value.V == int64(9)})
}

func TestOperationalStoreHugePositiveTTLDoesNotWrap(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	store, huge := runtime.NewInMemoryDataStore(), uint64(math.MaxUint64)
	store.Put(ctx, "huge", core.ValI64(11), &huge)
	value, present := store.Get(ctx, "huge")
	storeClockObservation(t, "store-positive-ttl-overflow", clock, 1, map[string]bool{"positiveTTLDoesNotImmediatelyExpire": present && value.V == int64(11)})
}

func TestOperationalStoreZeroValueAndIndependentInstances(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	var zero runtime.InMemoryDataStore
	zero.Put(ctx, "standalone", core.ValI64(13), nil)
	value, present := zero.Get(ctx, "standalone")
	other := runtime.NewInMemoryDataStore()
	_, leaked := other.Get(ctx, "standalone")
	storeClockObservation(t, "store-zero-value-and-isolation", clock, 1, map[string]bool{"zeroValueUsable": present && value.V == int64(13), "instancesIndependent": !leaked})
}

func TestOperationalStoreConcurrentReadWriteRemove(t *testing.T) {
	ctx := runtime.NewUserContext()
	store := runtime.NewInMemoryDataStore()
	var group sync.WaitGroup
	var failed int32
	for n := 0; n < 8; n++ {
		group.Add(1)
		go func(worker int) {
			defer group.Done()
			key := fmt.Sprintf("worker-%d", worker)
			for step := 0; step < 64; step++ {
				store.Put(ctx, key, core.ValI64(int64(step)), nil)
				if value, ok := store.Get(ctx, key); !ok || value.V != int64(step) {
					atomic.AddInt32(&failed, 1)
				}
				store.Remove(ctx, key)
				if _, ok := store.Get(ctx, key); ok {
					atomic.AddInt32(&failed, 1)
				}
			}
		}(n)
	}
	group.Wait()
	storeClockObservation(t, "store-concurrent-access", &referenceBusinessClock{}, 0, map[string]bool{"all512SequencesCorrect": failed == 0})
}

func TestOperationalLocalCacheClockSeparation(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	key := fmt.Sprintf("clock-local-cache-%s-%d", t.Name(), time.Now().UnixNano())
	t.Cleanup(func() { ctx.RemoveFromLocalCache(key) })
	ctx.PutToLocalCache(key, []int{}, 1)
	initial, hit := ctx.GetFromLocalCache(key).([]int)
	clock.value = clock.value.AddDate(100, 0, 0)
	_, futureHit := ctx.GetFromLocalCache(key).([]int)
	time.Sleep(1100 * time.Millisecond)
	expired := ctx.GetFromLocalCache(key) == nil
	clock.value = clock.value.AddDate(-200, 0, 0)
	notRevived := ctx.GetFromLocalCache(key) == nil
	storeClockObservation(t, "local-cache-real-expiry", clock, 1, map[string]bool{"emptyHitIsNotMissing": hit && len(initial) == 0, "futureKeepsLive": futureHit, "reallyExpired": expired, "backwardCannotRevive": notRevived})
}

func TestOperationalIDSetCacheClockSeparation(t *testing.T) {
	ctx, _, clock := referenceClockFixture(t)
	store := runtime.NewInMemoryIDSetStore()
	ctx.SetIDSetStore(store)
	deadline := time.Now().Add(80 * time.Millisecond)
	if err := ctx.IDSetStore().PutIDSet(ctx, &runtime.RetainedIDSet{QueryKey: "empty", IDs: []uint64{}, ExpiresAt: deadline}); err != nil {
		t.Fatal(err)
	}
	initial, err := ctx.IDSetStore().GetIDSet(ctx, "empty")
	if err != nil || initial == nil || len(initial.IDs) != 0 {
		t.Fatal("an empty retained ID set was treated as missing", err)
	}
	clock.value = clock.value.AddDate(100, 0, 0)
	future, err := ctx.IDSetStore().GetIDSet(ctx, "empty")
	if err != nil || future == nil {
		t.Fatal("business future invalidated a live ID set", err)
	}
	time.Sleep(time.Until(deadline) + 25*time.Millisecond)
	expired, err := ctx.IDSetStore().GetIDSet(ctx, "empty")
	if err != nil {
		t.Fatal(err)
	}
	clock.value = clock.value.AddDate(-200, 0, 0)
	revived, err := ctx.IDSetStore().GetIDSet(ctx, "empty")
	if err != nil {
		t.Fatal(err)
	}
	storeClockObservation(t, "id-set-cache-real-expiry", clock, 1, map[string]bool{"emptyHitIsNotMissing": true, "futureKeepsLive": true, "reallyExpired": expired == nil, "backwardCannotRevive": revived == nil})
}

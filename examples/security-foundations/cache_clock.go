package main

import (
	"fmt"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

func verifyCacheClockDomains() {
	future := time.Date(2132, 2, 29, 10, 15, 30, 0, time.UTC)
	ctx := runtime.NewUserContext().WithBusinessClock(runtime.NewFixedBusinessClock(future))
	require(ctx.BusinessTime().Equal(future), "cache example business clock was not installed")
	store, one := runtime.NewInMemoryDataStore(), uint64(1)
	store.Put(ctx, "empty", core.Value{V: []int{}}, &one)
	value, hit := store.Get(ctx, "empty")
	empty, typed := value.V.([]int)
	require(hit && typed && len(empty) == 0, "stored empty value became a miss")
	key := fmt.Sprintf("teaql-security-cache-example-%d", time.Now().UnixNano())
	defer ctx.RemoveFromLocalCache(key)
	ctx.PutToLocalCache(key, []int{}, 1)
	_, localHit := ctx.GetFromLocalCache(key).([]int)
	require(localHit, "local cache empty value became a miss")
	broken := &unavailableBusinessClock{}
	ctx.SetBusinessClock(broken)
	calibrated := false
	func() {
		defer func() { calibrated = recover() == "synthetic unavailable business time" }()
		_ = ctx.BusinessTime()
	}()
	require(calibrated && broken.reads == 1, "cache example unreadable clock was not calibrated")
	time.Sleep(1100 * time.Millisecond)
	_, stored := store.Get(ctx, "empty")
	require(!stored && ctx.GetFromLocalCache(key) == nil, "cache expiry used the unavailable business clock or ignored TTL")
	ctx.SetBusinessClock(runtime.NewFixedBusinessClock(future.AddDate(-200, 0, 0)))
	_, revived := store.Get(ctx, "empty")
	require(!revived && ctx.GetFromLocalCache(key) == nil && broken.reads == 1, "business rewind revived cache data")
	fmt.Println(`GO_CACHE_CLOCK_EXAMPLE={"passed":true,"storeRealExpiry":true,"localCacheRealExpiry":true,"emptyHitIsNotMissing":true,"unreadableBusinessCalibrated":true,"backwardCannotRevive":true,"additionalBusinessClockReads":0}`)
}

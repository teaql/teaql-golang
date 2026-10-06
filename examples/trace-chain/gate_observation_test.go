package tracechain_test

import "testing"

func TestBlockedGeneratedSaveStackRequiresActualContextMutexWait(t *testing.T) {
	frames := "sync.(*Mutex).lockSlow(0x1)\n" +
		"runtime.(*UserContext).ExecutePreparedGraphSave(0x2)\n" +
		"customer_order.(*CustomerOrder).Save(0x3)\n"
	for _, state := range []string{"[semacquire]", "[sync.Mutex.Lock]"} {
		if !blockedGeneratedSaveStack("goroutine 42 " + state + ":\n" + frames) {
			t.Fatalf("real mutex wait was rejected for %s", state)
		}
	}
	for _, stack := range []string{
		"goroutine 42 [runnable]:\nruntime.(*UserContext).ExecutePreparedGraphSave(0x2)\ncustomer_order.(*CustomerOrder).Save(0x3)\n",
		"goroutine 42 [chan receive]:\ncustomer_order.(*CustomerOrder).Save(0x3)\n",
		"sync.(*Mutex).lockSlow(0x1)\nruntime.(*UserContext).ExecuteGraphSave(0x2)\ncustomer_order.(*CustomerOrder).Save(0x3)\n",
		"sync.(*Mutex).lockSlow(0x1)\nruntime.(*UserContext).ExecutePreparedGraphSave(0x2)\nother.(*Other).Save(0x3)\n",
	} {
		if blockedGeneratedSaveStack(stack) {
			t.Fatalf("unproved overlap was accepted: %s", stack)
		}
	}
}

package runtime

import (
	"testing"
)

func TestAtomicCounterIdGenerator(t *testing.T) {
	generator := NewAtomicCounterIdGenerator(100)
	
	id1, err := generator.GenerateId("Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id1 != 101 {
		t.Errorf("expected 101, got %v", id1)
	}
	
	id2, err := generator.GenerateId("Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id2 != 102 {
		t.Errorf("expected 102, got %v", id2)
	}
}

func TestSnowflakeIdGenerator(t *testing.T) {
	generator := NewSnowflakeIdGenerator(1, 1)
	
	id1, err := generator.GenerateId("Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	id2, err := generator.GenerateId("Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if id2 <= id1 {
		t.Errorf("expected id2 > id1, got id1=%v, id2=%v", id1, id2)
	}
}

func TestLocalIdGeneratorSingleton(t *testing.T) {
	gen1 := LocalIdGenerator()
	gen2 := LocalIdGenerator()
	
	id1, err := gen1.GenerateId("Entity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	id2, err := gen2.GenerateId("Entity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if id2 != id1+1 {
		t.Errorf("expected id2 == id1+1, got id1=%v, id2=%v", id1, id2)
	}
}

func TestSnowflakeIdGenerator_PanicWorkerId(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	NewSnowflakeIdGenerator(SnowflakeMaxWorkerId+1, 0)
}

func TestSnowflakeIdGenerator_PanicDatacenterId(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	NewSnowflakeIdGenerator(0, SnowflakeMaxDatacenterId+1)
}

func TestSnowflakeIdGenerator_SequenceRollover(t *testing.T) {
	generator := DefaultSnowflakeIdGenerator()
	
	var lastId uint64
	for i := 0; i < 10000; i++ {
		id, err := generator.GenerateId("Order")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id <= lastId && i > 0 {
			t.Errorf("ids should be strictly increasing, got %v then %v", lastId, id)
		}
		lastId = id
	}
}

func TestSnowflakeIdGenerator_ClockBackwards(t *testing.T) {
	generator := DefaultSnowflakeIdGenerator()
	
	id1, err := generator.GenerateId("Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Simulate clock moving backward by setting lastTimestamp ahead of current time
	generator.mu.Lock()
	generator.state.lastTimestamp = currentMillis() + 10 // 10ms in the future
	generator.mu.Unlock()
	
	id2, err := generator.GenerateId("Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if id2 <= id1 {
		t.Errorf("expected id2 > id1 despite clock moving backwards")
	}
}

func TestSnowflakeIdGenerator_BeforeEpoch(t *testing.T) {
	generator := DefaultSnowflakeIdGenerator()
	
	// Simulate epoch being in the future
	generator.mu.Lock()
	generator.epochMillis = currentMillis() + 100000 // far in the future
	generator.mu.Unlock()
	
	_, err := generator.GenerateId("Order")
	if err == nil {
		t.Fatalf("expected error when clock is before epoch")
	}
	if err.Error() != "system clock is before snowflake epoch" {
		t.Errorf("unexpected error message: %v", err)
	}
}

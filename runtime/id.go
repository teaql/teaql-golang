package runtime

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type InternalIdGenerator interface {
	GenerateId(entity string) (uint64, error)
}

// ---------------------------------------------------------------------------
// AtomicCounterIdGenerator
// ---------------------------------------------------------------------------

type AtomicCounterIdGenerator struct {
	counter uint64
}

func NewAtomicCounterIdGenerator(start uint64) *AtomicCounterIdGenerator {
	return &AtomicCounterIdGenerator{
		counter: start,
	}
}

func DefaultAtomicCounterIdGenerator() *AtomicCounterIdGenerator {
	return NewAtomicCounterIdGenerator(1000)
}

func (g *AtomicCounterIdGenerator) GenerateId(entity string) (uint64, error) {
	return atomic.AddUint64(&g.counter, 1), nil
}

// ---------------------------------------------------------------------------
// SnowflakeIdGenerator
// ---------------------------------------------------------------------------

const (
	SnowflakeDefaultEpochMillis = uint64(1_288_834_974_657)
	SnowflakeWorkerIdBits       = uint64(5)
	SnowflakeDatacenterIdBits   = uint64(5)
	SnowflakeSequenceBits       = uint64(12)

	SnowflakeMaxWorkerId     = (1 << SnowflakeWorkerIdBits) - 1
	SnowflakeMaxDatacenterId = (1 << SnowflakeDatacenterIdBits) - 1
	SnowflakeSequenceMask    = (1 << SnowflakeSequenceBits) - 1

	SnowflakeWorkerIdShift     = SnowflakeSequenceBits
	SnowflakeDatacenterIdShift = SnowflakeSequenceBits + SnowflakeWorkerIdBits
	SnowflakeTimestampShift    = SnowflakeSequenceBits + SnowflakeWorkerIdBits + SnowflakeDatacenterIdBits
)

type snowflakeState struct {
	lastTimestamp uint64
	sequence      uint64
}

type SnowflakeIdGenerator struct {
	epochMillis   uint64
	workerId      uint64
	datacenterId  uint64
	state         *snowflakeState
	mu            sync.Mutex
}

func NewSnowflakeIdGenerator(workerId, datacenterId uint64) *SnowflakeIdGenerator {
	if workerId > SnowflakeMaxWorkerId {
		panic(fmt.Sprintf("worker id %d out of range", workerId))
	}
	if datacenterId > SnowflakeMaxDatacenterId {
		panic(fmt.Sprintf("datacenter id %d out of range", datacenterId))
	}

	return &SnowflakeIdGenerator{
		epochMillis:  SnowflakeDefaultEpochMillis,
		workerId:     workerId,
		datacenterId: datacenterId,
		state:        &snowflakeState{},
	}
}

func DefaultSnowflakeIdGenerator() *SnowflakeIdGenerator {
	return NewSnowflakeIdGenerator(0, 0)
}

func currentMillis() uint64 {
	return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

func waitUntilNextMillis(lastTimestamp uint64) uint64 {
	for {
		timestamp := currentMillis()
		if timestamp > lastTimestamp {
			return timestamp
		}
		time.Sleep(time.Millisecond)
	}
}

func (g *SnowflakeIdGenerator) GenerateId(entity string) (uint64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := currentMillis()

	if timestamp < g.state.lastTimestamp {
		timestamp = waitUntilNextMillis(g.state.lastTimestamp)
	}

	if timestamp == g.state.lastTimestamp {
		g.state.sequence = (g.state.sequence + 1) & SnowflakeSequenceMask
		if g.state.sequence == 0 {
			timestamp = waitUntilNextMillis(g.state.lastTimestamp)
		}
	} else {
		g.state.sequence = 0
	}

	g.state.lastTimestamp = timestamp

	if timestamp < g.epochMillis {
		return 0, fmt.Errorf("system clock is before snowflake epoch")
	}
	relativeTimestamp := timestamp - g.epochMillis

	id := (relativeTimestamp << SnowflakeTimestampShift) |
		(g.datacenterId << SnowflakeDatacenterIdShift) |
		(g.workerId << SnowflakeWorkerIdShift) |
		g.state.sequence

	return id, nil
}

var (
	localIdGeneratorInstance *AtomicCounterIdGenerator
	localIdGeneratorOnce     sync.Once
)

func LocalIdGenerator() *AtomicCounterIdGenerator {
	localIdGeneratorOnce.Do(func() {
		localIdGeneratorInstance = DefaultAtomicCounterIdGenerator()
	})
	return localIdGeneratorInstance
}

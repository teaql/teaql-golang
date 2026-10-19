package runtime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type InternalIdGenerator interface {
	GenerateId(entity string) (uint64, error)
}

type InternalIdFloorReserver interface {
	EnsureIdFloor(context.Context, string, uint64) error
}

// InstallDefaultInternalIdGenerator is called by generated runtime wiring.
// An application-owned generator installed later always has precedence.
func (c *UserContext) InstallDefaultInternalIdGenerator(generator InternalIdGenerator) {
	if missingSPI(generator) {
		panic("default internal ID generator must not be nil")
	}
	c.graphSaveMu.Lock()
	c.resources["idGenerator"] = generator
	c.graphSaveMu.Unlock()
}

// SetInternalIdGenerator installs an application-owned ID provider. It has
// precedence over the generated default and graph transaction capabilities.
func (c *UserContext) SetInternalIdGenerator(generator InternalIdGenerator) {
	if missingSPI(generator) {
		panic("internal ID generator must not be nil")
	}
	c.graphSaveMu.Lock()
	c.internalIDGenerator = generator
	c.graphSaveMu.Unlock()
}

func (c *UserContext) WithInternalIdGenerator(generator InternalIdGenerator) *UserContext {
	c.SetInternalIdGenerator(generator)
	return c
}

func (c *UserContext) InternalIdGenerator() InternalIdGenerator {
	if configured := c.GetResource("idGenerator"); configured != nil {
		if generator, ok := configured.(InternalIdGenerator); ok {
			return generator
		}
		panic("idGenerator resource does not implement InternalIdGenerator")
	}
	return LocalIdGenerator()
}

func (c *UserContext) GenerateInternalId(entity string) (uint64, error) {
	if entity == "" {
		return 0, fmt.Errorf("internal ID entity must not be empty")
	}
	id, err := c.InternalIdGenerator().GenerateId(entity)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("internal ID generator returned zero for %s", entity)
	}
	return id, nil
}

func (c *UserContext) EnsureInternalIdFloor(ctx context.Context, entity string, floor uint64) error {
	if entity == "" {
		return fmt.Errorf("internal ID entity must not be empty")
	}
	if reserver, ok := c.InternalIdGenerator().(InternalIdFloorReserver); ok {
		return reserver.EnsureIdFloor(ctx, entity, floor)
	}
	return fmt.Errorf("internal ID generator cannot reserve explicit ID floor for %s", entity)
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
	for {
		current := atomic.LoadUint64(&g.counter)
		if current == ^uint64(0) {
			return 0, fmt.Errorf("internal ID space exhausted for %s", entity)
		}
		if atomic.CompareAndSwapUint64(&g.counter, current, current+1) {
			return current + 1, nil
		}
	}
}

func (g *AtomicCounterIdGenerator) EnsureIdFloor(_ context.Context, _ string, floor uint64) error {
	for {
		current := atomic.LoadUint64(&g.counter)
		if current >= floor {
			return nil
		}
		if atomic.CompareAndSwapUint64(&g.counter, current, floor) {
			return nil
		}
	}
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
	epochMillis  uint64
	workerId     uint64
	datacenterId uint64
	state        *snowflakeState
	mu           sync.Mutex
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

package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type graphTransactionProbe struct {
	mu      sync.Mutex
	begins  int
	commits int
	active  int
	max     int
}

func (p *graphTransactionProbe) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Transaction: true}
}

func (p *graphTransactionProbe) Begin(context.Context) (data_service.Transaction, error) {
	p.mu.Lock()
	p.begins++
	p.active++
	if p.active > p.max {
		p.max = p.active
	}
	p.mu.Unlock()
	return &graphTransactionProbeTx{probe: p}, nil
}

type graphTransactionProbeTx struct{ probe *graphTransactionProbe }

func (t *graphTransactionProbeTx) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true, Mutation: true, Transaction: true}
}
func (t *graphTransactionProbeTx) Query(context.Context, *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return &data_service.QueryResult{}, nil
}
func (t *graphTransactionProbeTx) Mutate(context.Context, data_service.MutationRequest) (*data_service.MutationResult, error) {
	return &data_service.MutationResult{}, nil
}
func (t *graphTransactionProbeTx) Commit(context.Context) error {
	t.probe.mu.Lock()
	t.probe.commits++
	t.probe.active--
	t.probe.mu.Unlock()
	return nil
}
func (t *graphTransactionProbeTx) Rollback(context.Context) error {
	t.probe.mu.Lock()
	t.probe.active--
	t.probe.mu.Unlock()
	return nil
}

func TestIndependentGraphSavesAreSerialized(t *testing.T) {
	probe := &graphTransactionProbe{}
	userContext := NewUserContext()
	userContext.InsertResource("dataService", probe)
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	errors := make(chan error, 2)

	go func() {
		errors <- userContext.ExecuteGraphSave(func() error {
			close(firstStarted)
			<-releaseFirst
			return nil
		})
	}()
	<-firstStarted
	go func() { errors <- userContext.ExecuteGraphSave(func() error { return nil }) }()
	time.Sleep(10 * time.Millisecond)

	probe.mu.Lock()
	if probe.begins != 1 || probe.max != 1 {
		t.Fatalf("independent save joined or overlapped active graph: begins=%d max=%d", probe.begins, probe.max)
	}
	probe.mu.Unlock()
	close(releaseFirst)
	for i := 0; i < 2; i++ {
		if err := <-errors; err != nil {
			t.Fatal(err)
		}
	}
	probe.mu.Lock()
	defer probe.mu.Unlock()
	if probe.begins != 2 || probe.commits != 2 || probe.max != 1 {
		t.Fatalf("expected two serialized transactions, got begins=%d commits=%d max=%d", probe.begins, probe.commits, probe.max)
	}
}

func TestGraphSaveCapturesOneFixTime(t *testing.T) {
	probe := &graphTransactionProbe{}
	userContext := NewUserContext()
	userContext.InsertResource("dataService", probe)
	var first, second time.Time
	if err := userContext.ExecuteGraphSave(func() error {
		first = userContext.FixTime()
		time.Sleep(time.Millisecond)
		second = userContext.FixTime()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if first.IsZero() || !first.Equal(second) {
		t.Fatalf("one graph must share one fix time: first=%v second=%v", first, second)
	}
}

type fixTimeCapturingRegistry struct{ values []time.Time }

func (r *fixTimeCapturingRegistry) CheckAndFix(_ *UserContext, input *CheckAndFixInput) []CheckResult {
	r.values = append(r.values, input.Now)
	return nil
}

func TestCheckAndFixDefaultsToTheOneGraphFixTime(t *testing.T) {
	probe := &graphTransactionProbe{}
	registry := &fixTimeCapturingRegistry{}
	userContext := NewUserContext()
	userContext.InsertResource("dataService", probe)
	userContext.SetCheckerRegistry(registry)
	if err := userContext.ExecuteGraphSave(func() error {
		for index := 0; index < 2; index++ {
			if err := userContext.CheckAndFix(&CheckAndFixInput{
				Entity: "Task", Operation: core.MutationInsert, Values: core.Record{},
			}); err != nil {
				return err
			}
			time.Sleep(time.Millisecond)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(registry.values) != 2 || registry.values[0].IsZero() ||
		!registry.values[0].Equal(registry.values[1]) {
		t.Fatalf("checker calls must share one non-zero graph time: %v", registry.values)
	}
}

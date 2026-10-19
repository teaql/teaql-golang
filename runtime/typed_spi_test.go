package runtime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

type spiMetadata struct{}

func (spiMetadata) Entity(string) *core.EntityDescriptor  { return nil }
func (spiMetadata) AllEntities() []*core.EntityDescriptor { return nil }

type spiIDs struct {
	next      uint64
	floor     uint64
	err       error
	generated int
}

func (g *spiIDs) GenerateId(string) (uint64, error) {
	g.generated++
	if g.err != nil {
		return 0, g.err
	}
	g.next++
	return g.next, nil
}

func (g *spiIDs) EnsureIdFloor(_ context.Context, _ string, floor uint64) error {
	if g.err != nil {
		return g.err
	}
	g.floor = floor
	if g.next < floor {
		g.next = floor
	}
	return nil
}

func TestTypedSchemaExecutorPrecedenceAbsenceAndFailure(t *testing.T) {
	unconfigured := NewUserContext()
	if err := unconfigured.ExecuteSchema(SchemaExecutorFunc(func(context.Context, MetadataStore) error { return nil })); err == nil || !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("missing metadata was not diagnosed: %v", err)
	}
	ctx := NewUserContext()
	ctx.Metadata = spiMetadata{}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("typed-nil schema executor was accepted")
			}
		}()
		ctx.WithSchemaExecutor(SchemaExecutorFunc(nil))
	}()
	if err := ctx.ExecuteSchema(nil); err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("missing schema executor: %v", err)
	}
	defaultCalls, customCalls := 0, 0
	defaultSchema := SchemaExecutorFunc(func(context.Context, MetadataStore) error {
		defaultCalls++
		return nil
	})
	if err := ctx.ExecuteSchema(defaultSchema); err != nil || defaultCalls != 1 {
		t.Fatalf("default schema: calls=%d err=%v", defaultCalls, err)
	}
	ctx.WithSchemaExecutor(SchemaExecutorFunc(func(context.Context, MetadataStore) error {
		customCalls++
		return nil
	}))
	for i := 0; i < 2; i++ {
		if err := ctx.ExecuteSchema(defaultSchema); err != nil {
			t.Fatal(err)
		}
	}
	if defaultCalls != 1 || customCalls != 2 {
		t.Fatalf("schema precedence: default=%d custom=%d", defaultCalls, customCalls)
	}
	want := errors.New("custom schema rejected")
	ctx.WithSchemaExecutor(SchemaExecutorFunc(func(context.Context, MetadataStore) error { return want }))
	if err := ctx.ExecuteSchema(defaultSchema); !errors.Is(err, want) || defaultCalls != 1 {
		t.Fatalf("schema failure fell through: %v", err)
	}
	if _, ok := reflect.TypeOf(ctx).MethodByName("SetSchemaProvider"); ok {
		t.Fatal("misleading no-op Schema provider alias is still public")
	}
	if _, ok := reflect.TypeOf(ctx).MethodByName("WithSchemaProvider"); ok {
		t.Fatal("misleading compatibility Schema provider alias is still public")
	}
}

func TestTypedInternalIDPrecedenceFloorAndNoFallback(t *testing.T) {
	ctx := NewUserContext()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("untyped idGenerator resource was accepted")
			}
		}()
		ctx.InsertResource("idGenerator", struct{}{})
	}()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("typed-nil internal ID generator was accepted")
			}
		}()
		var absent *spiIDs
		ctx.WithInternalIdGenerator(absent)
	}()
	defaultIDs := &spiIDs{next: 10}
	ctx.InsertResource("idGenerator", defaultIDs)
	customIDs := &spiIDs{next: 499}
	ctx.WithInternalIdGenerator(customIDs)
	if err := ctx.EnsureInternalIdFloor(context.Background(), "School", 900); err != nil {
		t.Fatal(err)
	}
	id, err := ctx.GenerateInternalId("School")
	if err != nil || id != 901 || defaultIDs.generated != 0 || customIDs.floor != 900 {
		t.Fatalf("custom ID precedence: id=%d err=%v default=%d floor=%d", id, err, defaultIDs.generated, customIDs.floor)
	}
	want := errors.New("custom allocator denied")
	customIDs.err = want
	if _, err := ctx.GenerateInternalId("School"); !errors.Is(err, want) || defaultIDs.generated != 0 {
		t.Fatalf("allocation failure fell through: %v", err)
	}
	if _, ok := reflect.TypeOf(ctx).MethodByName("GenerateId"); ok {
		t.Fatal("untyped no-op GenerateId alias is still public")
	}
}

func TestAtomicInternalIDFloorAndExhaustion(t *testing.T) {
	generator := NewAtomicCounterIdGenerator(10)
	if err := generator.EnsureIdFloor(context.Background(), "School", 100); err != nil {
		t.Fatal(err)
	}
	id, err := generator.GenerateId("School")
	if err != nil || id != 101 {
		t.Fatalf("atomic floor was not respected: id=%d err=%v", id, err)
	}
	if err := generator.EnsureIdFloor(context.Background(), "School", ^uint64(0)); err != nil {
		t.Fatal(err)
	}
	if _, err := generator.GenerateId("School"); err == nil || !strings.Contains(err.Error(), "exhausted") {
		t.Fatalf("atomic generator wrapped at capacity: %v", err)
	}
}

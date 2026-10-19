package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type exampleIDs struct {
	next  uint64
	calls int
	floor uint64
	fail  bool
}

func (g *exampleIDs) GenerateId(string) (uint64, error) {
	g.calls++
	if g.fail {
		return 0, errors.New("custom allocation rejected")
	}
	g.next++
	return g.next, nil
}

func (g *exampleIDs) EnsureIdFloor(_ context.Context, _ string, floor uint64) error {
	g.floor = floor
	if g.next < floor {
		g.next = floor
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	school := core.NewEntityDescriptor("School").TableName("school_data")
	ctx := runtime.NewRuntimeModule().Entity(school).IntoContext()
	defaultCalls, customCalls := 0, 0
	defaultSchema := runtime.SchemaExecutorFunc(func(context.Context, runtime.MetadataStore) error {
		defaultCalls++
		return nil
	})
	ctx.WithSchemaExecutor(runtime.SchemaExecutorFunc(func(_ context.Context, metadata runtime.MetadataStore) error {
		customCalls++
		if metadata.Entity("School") == nil {
			return errors.New("School metadata missing")
		}
		return nil
	}))
	must(ctx.ExecuteSchema(defaultSchema))
	must(ctx.ExecuteSchema(defaultSchema))
	if customCalls != 2 || defaultCalls != 0 {
		panic("custom schema executor was bypassed")
	}

	defaultIDs := &exampleIDs{next: 10}
	customIDs := &exampleIDs{next: 499}
	ctx.InstallDefaultInternalIdGenerator(defaultIDs)
	ctx.WithInternalIdGenerator(customIDs)
	id, err := ctx.GenerateInternalId("School")
	must(err)
	if id != 500 || defaultIDs.calls != 0 {
		panic("custom internal ID generator was bypassed")
	}
	must(ctx.EnsureInternalIdFloor(context.Background(), "School", 900))
	id, err = ctx.GenerateInternalId("School")
	must(err)
	if id != 901 || customIDs.floor != 900 || defaultIDs.calls != 0 {
		panic("explicit ID floor did not reach custom generator")
	}
	customIDs.fail = true
	if _, err := ctx.GenerateInternalId("School"); err == nil || err.Error() != "custom allocation rejected" || defaultIDs.calls != 0 {
		panic("custom allocation failure fell through to default generator")
	}
	fmt.Println("PASS Go typed Schema and Internal ID example")
}

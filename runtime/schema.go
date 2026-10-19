package runtime

import (
	"context"
	"fmt"
	"reflect"
)

// SchemaExecutor is an application-owned schema capability. It is independent
// of query and mutation data services; generated modules still own bootstrap.
type SchemaExecutor interface {
	EnsureSchema(context.Context, MetadataStore) error
}

type SchemaExecutorFunc func(context.Context, MetadataStore) error

func (f SchemaExecutorFunc) EnsureSchema(ctx context.Context, metadata MetadataStore) error {
	return f(ctx, metadata)
}

func missingSPI(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (c *UserContext) WithSchemaExecutor(executor SchemaExecutor) *UserContext {
	if missingSPI(executor) {
		panic("schema executor must not be nil")
	}
	c.schemaExecutor = executor
	return c
}

// ExecuteSchema selects the context-installed capability before the generated
// module's default. The caller runs generated bootstrap only after success.
func (c *UserContext) ExecuteSchema(defaultExecutor SchemaExecutor) error {
	executor := c.schemaExecutor
	if missingSPI(executor) {
		executor = defaultExecutor
	}
	if missingSPI(executor) {
		return fmt.Errorf("schema executor is not configured")
	}
	if missingSPI(c.Metadata) {
		return fmt.Errorf("schema metadata is not configured")
	}
	return executor.EnsureSchema(c.Context, c.Metadata)
}

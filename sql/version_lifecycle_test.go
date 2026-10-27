package sql

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestVersionLifecycleSQLRetainsTypedCompileCause(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	_, err := dialect.CompileUpdate(entity(), core.NewUpdateCommand("Order", core.ValU64(7)).
		WithExpectedVersion(math.MaxInt64).Value("name", core.ValText("changed")))
	var compile *SqlCompileError
	var version *core.MutationVersionError
	if !errors.As(err, &compile) || !errors.As(err, &version) || !version.Exhausted {
		t.Fatalf("SQL compile lost its typed lifecycle cause: %v", err)
	}
}

// These tests exercise the production SQL compiler, not a duplicated numeric
// oracle. The adjacent legal endpoints prevent a blanket rejection workaround.
func TestVersionLifecycleSQLUpdateRejectsNonLiveVersions(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	for _, version := range []int64{0, -1, math.MinInt64} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			query, err := dialect.CompileUpdate(entity(), core.NewUpdateCommand("Order", core.ValU64(7)).
				WithExpectedVersion(version).Value("name", core.ValText("changed")))
			if err == nil || query != nil {
				t.Fatalf("non-live update version %d must reject without compiled SQL: query=%+v err=%v", version, query, err)
			}
		})
	}
}

func TestVersionLifecycleSQLUpdateRejectsExhaustion(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	query, err := dialect.CompileUpdate(entity(), core.NewUpdateCommand("Order", core.ValU64(7)).
		WithExpectedVersion(math.MaxInt64).Value("name", core.ValText("changed")))
	if err == nil || query != nil {
		t.Fatalf("exhausted update must reject, never wrap into a tombstone: query=%+v err=%v", query, err)
	}
}

func TestVersionLifecycleSQLBatchRejectsInvalidOrExhaustedVersions(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	for _, version := range []int64{0, -1, math.MaxInt64} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			good, bad := int64(1), version
			command := core.NewBatchUpdateCommand("Order", []string{"name"})
			command.BatchIds = []core.Value{core.ValU64(7), core.ValU64(8)}
			command.BatchValues = []core.Record{{"name": core.ValText("first")}, {"name": core.ValText("second")}}
			command.BatchExpectedVersions = []*int64{&good, &bad}
			query, err := dialect.CompileBatchUpdate(entity(), command)
			if err == nil || query != nil {
				t.Fatalf("invalid member %d must reject the entire compiled batch: query=%+v err=%v", version, query, err)
			}
		})
	}
}

func TestVersionLifecycleSQLRecoveryRejectsUnrepresentableSuccessor(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	for _, version := range []int64{math.MinInt64, math.MinInt64 + 1} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			query, err := dialect.CompileRecover(entity(), core.NewRecoverCommand("Order", core.ValU64(7), version))
			if err == nil || query != nil {
				t.Fatalf("exhausted recovery %d must reject without wrapped SQL: query=%+v err=%v", version, query, err)
			}
		})
	}
}

func TestVersionLifecycleSQLAdjacentEndpointsRemainValid(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	updated, err := dialect.CompileUpdate(entity(), core.NewUpdateCommand("Order", core.ValU64(7)).
		WithExpectedVersion(math.MaxInt64-1).Value("name", core.ValText("last update")))
	if err != nil || updated == nil {
		t.Fatalf("last representable update must compile: %v", err)
	}
	if got, ok := updated.Params[1].V.(int64); !ok || got != math.MaxInt64 {
		t.Fatalf("last live update must bind MAX exactly: %+v", updated.Params)
	}
	deleted, err := dialect.CompileDelete(entity(), core.NewDeleteCommand("Order", core.ValU64(7)).WithExpectedVersion(math.MaxInt64))
	if err != nil || deleted == nil {
		t.Fatalf("maximum live version still supports soft deletion: %v", err)
	}
	if got, ok := deleted.Params[0].V.(int64); !ok || got != math.MinInt64 {
		t.Fatalf("deleting MAX must bind MIN exactly: %+v", deleted.Params)
	}
	recovered, err := dialect.CompileRecover(entity(), core.NewRecoverCommand("Order", core.ValU64(7), math.MinInt64+2))
	if err != nil || recovered == nil {
		t.Fatalf("last representable recovery must compile: %v", err)
	}
	if got, ok := recovered.Params[0].V.(int64); !ok || got != math.MaxInt64 {
		t.Fatalf("last representable recovery must bind MAX exactly: %+v", recovered.Params)
	}
}

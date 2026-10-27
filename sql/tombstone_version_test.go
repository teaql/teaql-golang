package sql

import (
	"math"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestTombstoneSoftDeleteCannotCompileAsRecovery(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	for _, version := range []int64{math.MinInt64, -902, -2, -1, 0} {
		command := core.NewDeleteCommand("Order", core.ValU64(7)).WithExpectedVersion(version)
		if query, err := dialect.CompileDelete(entity(), command); err == nil || query != nil {
			t.Fatal("non-live soft delete compiled", version, query)
		}
	}
	for _, version := range []int64{1, math.MaxInt64} {
		query, err := dialect.CompileDelete(entity(), core.NewDeleteCommand("Order", core.ValU64(7)).WithExpectedVersion(version))
		if err != nil {
			t.Fatal(err)
		}
		expected := int64(-2)
		if version == math.MaxInt64 {
			expected = math.MinInt64
		}
		if value, ok := query.Params[0].TryI64(); !ok || value != expected {
			t.Fatal("live deletion failed exact signed boundary", version, query.Params)
		}
	}
}

func TestTombstoneBlindDeleteRequiresLiveRowAndHardDeleteKeepsItsSeparateMeaning(t *testing.T) {
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	query, err := dialect.CompileDelete(entity(), core.NewDeleteCommand("Order", core.ValU64(7)))
	if err != nil || !strings.Contains(query.Sql, `"version" >= 1`) {
		t.Fatal("blind deletion can rewrite a tombstone", query, err)
	}
	query, err = dialect.CompileDelete(entity(), core.NewDeleteCommand("Order", core.ValU64(7)).WithExpectedVersion(-902).HardDelete())
	if err != nil || !strings.HasPrefix(query.Sql, "DELETE FROM ") {
		t.Fatal("explicit physical removal was confused with soft deletion", query, err)
	}
}

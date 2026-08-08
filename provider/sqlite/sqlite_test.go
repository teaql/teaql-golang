package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

func entity() *core.EntityDescriptor {
	return core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
}

func TestSqliteDialectCompilesMutationsAndSchema(t *testing.T) {
	dialect := &SqliteDialect{}
	defDialect := &teaql_sql.DefaultSqlDialect{Dialect: dialect}

	insert, err := defDialect.CompileInsert(entity(), core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("name", core.ValText("A")))
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO orders (id, name) VALUES (?, ?)", insert.Sql)

	update, err := defDialect.CompileUpdate(entity(), core.NewUpdateCommand("Order", core.ValU64(1)).WithExpectedVersion(3).Value("name", core.ValText("B")))
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE orders SET name = ?, version = ? WHERE id = ? AND version = ?", update.Sql)

	del, err := defDialect.CompileDelete(entity(), core.NewDeleteCommand("Order", core.ValU64(1)).WithExpectedVersion(3))
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE orders SET version = ? WHERE id = ? AND version = ?", del.Sql)

	recover, err := defDialect.CompileRecover(entity(), core.NewRecoverCommand("Order", core.ValU64(1), -4))
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE orders SET version = ? WHERE id = ? AND version = ?", recover.Sql)

	create, err := defDialect.CompileCreateTable(entity())
	assert.NoError(t, err)
	assert.Equal(t, "CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY NOT NULL, version INTEGER NOT NULL, name VARCHAR(255))", create)
}

func TestSqliteExecutorExecuteAndFetchAll(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	executor := NewSqliteMutationExecutor(db)
	ctx := context.Background()

	dialect := &SqliteDialect{}
	defDialect := &teaql_sql.DefaultSqlDialect{Dialect: dialect}

	create, err := defDialect.CompileCreateTable(entity())
	assert.NoError(t, err)
	_, err = executor.ExecuteSql(ctx, &teaql_sql.CompiledQuery{Sql: create})
	assert.NoError(t, err)
	insert, err := defDialect.CompileInsert(entity(), core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("version", core.ValI64(1)).Value("name", core.ValText("draft")))
	assert.NoError(t, err)

	affected, err := executor.ExecuteSql(ctx, insert)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), affected)

	selectQuery, err := defDialect.CompileSelect(entity(), core.NewSelectQuery("Order").WithFilter(core.ExprEq("id", core.ValU64(1))).OrderAsc("id"))
	assert.NoError(t, err)

	rows, err := executor.FetchAllSql(ctx, selectQuery)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(rows))
	assert.Equal(t, core.ValI64(1), rows[0]["id"])
	assert.Equal(t, core.ValI64(1), rows[0]["version"])
	assert.Equal(t, core.ValText("draft"), rows[0]["name"])
}

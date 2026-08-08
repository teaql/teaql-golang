package postgres

import (
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

func TestPostgresDialectCompilesMutationsAndSchema(t *testing.T) {
	dialect := &PostgresDialect{}
	defDialect := &teaql_sql.DefaultSqlDialect{Dialect: dialect}

	insert, err := defDialect.CompileInsert(entity(), core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("name", core.ValText("A")))
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO orders (id, name) VALUES ($1, $2)", insert.Sql)

	update, err := defDialect.CompileUpdate(entity(), core.NewUpdateCommand("Order", core.ValU64(1)).WithExpectedVersion(3).Value("name", core.ValText("B")))
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE orders SET name = $1, version = $2 WHERE id = $3 AND version = $4", update.Sql)

	del, err := defDialect.CompileDelete(entity(), core.NewDeleteCommand("Order", core.ValU64(1)).WithExpectedVersion(3))
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE orders SET version = $1 WHERE id = $2 AND version = $3", del.Sql)

	recover, err := defDialect.CompileRecover(entity(), core.NewRecoverCommand("Order", core.ValU64(1), -4))
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE orders SET version = $1 WHERE id = $2 AND version = $3", recover.Sql)

	create, err := defDialect.CompileCreateTable(entity())
	assert.NoError(t, err)
	assert.Equal(t, "CREATE TABLE IF NOT EXISTS orders (id BIGINT PRIMARY KEY NOT NULL, version BIGINT NOT NULL, name VARCHAR(255))", create)
}

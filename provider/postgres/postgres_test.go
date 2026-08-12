package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"

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

func TestPgMutationExecutorLogic(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER, name TEXT)")
	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO test VALUES (1, 'foo')")
	assert.NoError(t, err)

	exec := NewPgMutationExecutor(db)
	ctx := context.Background()

	// ExecuteSql
	queryExec := &teaql_sql.CompiledQuery{
		Sql:    "INSERT INTO test VALUES (?, ?)",
		Params: []core.Value{core.ValI64(2), core.ValText("bar")},
	}
	affected, err := exec.ExecuteSql(ctx, queryExec)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), affected)

	// FetchAllSql
	queryFetch := &teaql_sql.CompiledQuery{
		Sql:    "SELECT * FROM test",
		Params: nil,
	}
	records, err := exec.FetchAllSql(ctx, queryFetch)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(records))

	// BeginSql
	txExec, err := exec.BeginSql(ctx)
	assert.NoError(t, err)

	affected, err = txExec.ExecuteSql(ctx, queryExec)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), affected)

	records, err = txExec.FetchAllSql(ctx, queryFetch)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(records)) // +1 from txExec

	assert.NoError(t, txExec.CommitSql(ctx))

	txExec2, _ := exec.BeginSql(ctx)
	assert.NoError(t, txExec2.RollbackSql(ctx))

	// Errors
	queryErr := &teaql_sql.CompiledQuery{Sql: "INVALID"}
	_, err = exec.ExecuteSql(ctx, queryErr)
	assert.Error(t, err)
	_, err = exec.FetchAllSql(ctx, queryErr)
	assert.Error(t, err)
	_, err = txExec2.ExecuteSql(ctx, queryErr)
	assert.Error(t, err)
	_, err = txExec2.FetchAllSql(ctx, queryErr)
	assert.Error(t, err)

	queryBindErr := &teaql_sql.CompiledQuery{Sql: "SELECT 1", Params: []core.Value{{V: struct{}{}}}}
	_, err = exec.ExecuteSql(ctx, queryBindErr)
	assert.Error(t, err)
	_, err = exec.FetchAllSql(ctx, queryBindErr)
	assert.Error(t, err)
	_, err = txExec2.ExecuteSql(ctx, queryBindErr)
	assert.Error(t, err)
	_, err = txExec2.FetchAllSql(ctx, queryBindErr)
	assert.Error(t, err)

	db.Close()
	_, err = exec.ExecuteSql(ctx, queryExec)
	assert.Error(t, err)
	_, err = exec.FetchAllSql(ctx, queryFetch)
	assert.Error(t, err)
}

func TestBindPgValue(t *testing.T) {
	val, err := bindPgValue(core.ValU64(1))
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = bindPgValue(core.ValNull())
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = bindPgValue(core.ValBool(true))
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	val, err = bindPgValue(core.ValF64(1.23))
	assert.NoError(t, err)
	assert.Equal(t, 1.23, val)

	val, err = bindPgValue(core.ValDecimal(decimal.RequireFromString("123.450")))
	assert.NoError(t, err)
	assert.Equal(t, "123.45", val)

	date := time.Date(2026, time.January, 6, 0, 0, 0, 0, time.UTC)
	val, err = bindPgValue(core.ValDate(date))
	assert.NoError(t, err)
	assert.Equal(t, date, val)
}

func TestDecodePgRow(t *testing.T) {
	record, err := decodePgRow([]string{"col1"}, nil, []interface{}{nil})
	assert.NoError(t, err)
	assert.Nil(t, record["col1"].V)

	record, err = decodePgRow([]string{"col1"}, nil, []interface{}{true})
	assert.NoError(t, err)
	assert.Equal(t, true, record["col1"].V)

	record, err = decodePgRow([]string{"col1"}, nil, []interface{}{int64(1)})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), record["col1"].V)

	record, err = decodePgRow([]string{"col1"}, nil, []interface{}{1.23})
	assert.NoError(t, err)
	assert.Equal(t, 1.23, record["col1"].V)

	record, err = decodePgRow([]string{"col1"}, nil, []interface{}{[]byte("test")})
	assert.NoError(t, err)
	assert.Equal(t, "test", record["col1"].V)

	record, err = decodePgRow([]string{"col1"}, nil, []interface{}{"str"})
	assert.NoError(t, err)
	assert.Equal(t, "str", record["col1"].V)

	record, err = decodePgRow([]string{"col1"}, nil, []interface{}{struct{}{}})
	assert.NoError(t, err)
	assert.Equal(t, "{}", record["col1"].V)
}

func TestPostgresDialect(t *testing.T) {
	d := &PostgresDialect{}

	assert.Equal(t, teaql_sql.DatabaseKindPostgreSQL, d.Kind())
	assert.Equal(t, `"select"`, d.QuoteIdent("select"))
	assert.Equal(t, "$1", d.Placeholder(1))

	_, err := d.CompileGbkFunction(nil, nil, nil)
	assert.Error(t, err)
	assert.NotNil(t, d.SchemaSetupSqls())

	// SchemaTypeSql
	tests := []struct {
		dt  core.DataType
		exp string
	}{
		{core.TypeBool, "BOOLEAN"},
		{core.TypeI64, "BIGINT"},
		{core.TypeU64, "BIGINT"},
		{core.TypeF64, "DOUBLE PRECISION"},
		{core.TypeDecimal, "NUMERIC"},
		{core.TypeText, "VARCHAR(255)"},
		{core.TypeLargeText, "TEXT"},
		{core.TypeJson, "JSONB"},
		{core.TypeDate, "DATE"},
		{core.TypeTimestamp, "TIMESTAMPTZ"},
	}
	for _, tt := range tests {
		res, err := d.SchemaTypeSql(tt.dt, nil)
		assert.NoError(t, err)
		assert.Equal(t, tt.exp, res)
	}

	_, err = d.SchemaTypeSql(core.DataType(999), nil)
	assert.Error(t, err)

	ent := &core.EntityDescriptor{TabName: "test"}
	prop := &core.PropertyDescriptor{Name: "col", ColName: "col", DataType: core.TypeI64}
	res, err := d.CompileAddColumn(ent, prop)
	assert.NoError(t, err)
	assert.Equal(t, "ALTER TABLE test ADD COLUMN col BIGINT", res)
}

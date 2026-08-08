package mysql

import (
	"testing"
	"context"
	"database/sql"
	
	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

func TestMysqlMutationExecutorLogic(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE test (id INTEGER, name TEXT)")
	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO test VALUES (1, 'foo')")
	assert.NoError(t, err)

	exec := NewMysqlMutationExecutor(db)
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

func TestBindMysqlValue(t *testing.T) {
	val, err := bindMysqlValue(core.ValU64(1))
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = bindMysqlValue(core.ValNull())
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = bindMysqlValue(core.ValBool(true))
	assert.NoError(t, err)
	assert.Equal(t, true, val)

	val, err = bindMysqlValue(core.ValF64(1.23))
	assert.NoError(t, err)
	assert.Equal(t, 1.23, val)
}

func TestDecodeMysqlRow(t *testing.T) {
	record, err := decodeMysqlRow([]string{"col1"}, nil, []interface{}{nil})
	assert.NoError(t, err)
	assert.Nil(t, record["col1"].V)

	record, err = decodeMysqlRow([]string{"col1"}, nil, []interface{}{true})
	assert.NoError(t, err)
	assert.Equal(t, true, record["col1"].V)

	record, err = decodeMysqlRow([]string{"col1"}, nil, []interface{}{int64(1)})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), record["col1"].V)

	record, err = decodeMysqlRow([]string{"col1"}, nil, []interface{}{1.23})
	assert.NoError(t, err)
	assert.Equal(t, 1.23, record["col1"].V)

	record, err = decodeMysqlRow([]string{"col1"}, nil, []interface{}{[]byte("test")})
	assert.NoError(t, err)
	assert.Equal(t, "test", record["col1"].V)

	record, err = decodeMysqlRow([]string{"col1"}, nil, []interface{}{"str"})
	assert.NoError(t, err)
	assert.Equal(t, "str", record["col1"].V)

	record, err = decodeMysqlRow([]string{"col1"}, nil, []interface{}{struct{}{}})
	assert.NoError(t, err)
	assert.Equal(t, "{}", record["col1"].V)
}

func entity() *core.EntityDescriptor {
	return core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
}

func TestMysqlDialectCompilesMutationsAndSchema(t *testing.T) {
	dialect := &MysqlDialect{}
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
	assert.Equal(t, "CREATE TABLE IF NOT EXISTS orders (id BIGINT PRIMARY KEY NOT NULL, version BIGINT NOT NULL, name VARCHAR(255))", create)
}

func TestMysqlDialect(t *testing.T) {
	d := &MysqlDialect{}

	assert.Equal(t, teaql_sql.DatabaseKindMySQL, d.Kind())
	assert.Equal(t, "`select`", d.QuoteIdent("select"))
	assert.Equal(t, "?", d.Placeholder(0))

	_, err := d.CompileGbkFunction(nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, d.SchemaSetupSqls())

	// SchemaTypeSql
	tests := []struct {
		dt  core.DataType
		exp string
	}{
		{core.TypeBool, "BOOLEAN"},
		{core.TypeI64, "BIGINT"},
		{core.TypeU64, "BIGINT"},
		{core.TypeF64, "DOUBLE"},
		{core.TypeDecimal, "DECIMAL(38, 10)"},
		{core.TypeText, "VARCHAR(255)"},
		{core.TypeLargeText, "LONGTEXT"},
		{core.TypeJson, "JSON"},
		{core.TypeDate, "DATE"},
		{core.TypeTimestamp, "DATETIME(6)"},
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



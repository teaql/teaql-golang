package sql

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

func TestCompiledQuerySqlWithComment(t *testing.T) {
	comment := "this is a test"
	q := CompiledQuery{
		Sql:     "SELECT * FROM users",
		Comment: &comment,
	}
	assert.Equal(t, "/* this is a test */ SELECT * FROM users", q.SqlWithComment())

	comment2 := "attempt to close */ comment early"
	q.Comment = &comment2
	assert.Equal(t, "/* attempt to close * / comment early */ SELECT * FROM users", q.SqlWithComment())
	
	// without comment
	q2 := CompiledQuery{Sql: "SELECT 1"}
	assert.Equal(t, "SELECT 1", q2.SqlWithComment())
}

func TestCompiledQueryDebugSqlPostgres(t *testing.T) {
	q := CompiledQuery{
		Sql:    "SELECT * FROM users WHERE name = $1 AND age > $2 AND id = $10",
		Params: []core.Value{core.ValText("John O'Connor"), core.ValI64(18), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValI64(999)},
	}
	
	debug := q.DebugSql(DatabaseKindPostgreSQL)
	assert.Equal(t, "SELECT * FROM users WHERE name = 'John O''Connor' AND age > 18 AND id = 999", debug)
	
	q2 := CompiledQuery{
		Sql: "SELECT '$1' $99", // string escaping and out of bounds
		Params: []core.Value{},
	}
	assert.Equal(t, "SELECT '$1' $99", q2.DebugSql(DatabaseKindPostgreSQL))
}

func TestCompiledQueryDebugSqlPositional(t *testing.T) {
	q := CompiledQuery{
		Sql:    "SELECT * FROM users WHERE name = ? AND age > ?",
		Params: []core.Value{core.ValText("John O'Connor"), core.ValI64(18)},
	}
	
	debugSqlite := q.DebugSql(DatabaseKindSQLite)
	assert.Equal(t, "SELECT * FROM users WHERE name = 'John O''Connor' AND age > 18", debugSqlite)
	
	debugMysql := q.DebugSql(DatabaseKindMySQL)
	assert.Equal(t, "SELECT * FROM users WHERE name = 'John O''Connor' AND age > 18", debugMysql)
	
	// extra placeholders
	q2 := CompiledQuery{Sql: "SELECT ?, ?", Params: []core.Value{core.ValI64(1)}}
	assert.Equal(t, "SELECT 1, ?", q2.DebugSql(DatabaseKindSQLite))
	
	// single quote open
	q3 := CompiledQuery{Sql: "SELECT '", Params: []core.Value{core.ValI64(1)}}
	assert.Equal(t, "SELECT '", q3.DebugSql(DatabaseKindSQLite))
}

func TestCompiledQueryDebugSqlPositionalIgnoresStringQuotes(t *testing.T) {
	q := CompiledQuery{
		Sql:    "SELECT * FROM users WHERE name = '?' AND age > ?",
		Params: []core.Value{core.ValI64(18)},
	}
	
	debugSqlite := q.DebugSql(DatabaseKindSQLite)
	assert.Equal(t, "SELECT * FROM users WHERE name = '?' AND age > 18", debugSqlite)
	
	q2 := CompiledQuery{
		Sql:    "SELECT * FROM users WHERE name = '?' '' AND age > ?",
		Params: []core.Value{core.ValI64(18)},
	}
	assert.Equal(t, "SELECT * FROM users WHERE name = '?' '' AND age > 18", q2.DebugSql(DatabaseKindSQLite))
	
	// extra placeholders in sql without enough params
	q3 := CompiledQuery{Sql: "SELECT ?, ?", Params: []core.Value{core.ValI64(1)}}
	assert.Equal(t, "SELECT 1, ?", q3.DebugSql(DatabaseKindSQLite))
}

func TestCompiledQueryDebugSqlPostgresEdge(t *testing.T) {
	// index parsing
	q := CompiledQuery{Sql: "SELECT $1a", Params: []core.Value{core.ValI64(1)}}
	assert.Equal(t, "SELECT 1a", q.DebugSql(DatabaseKindPostgreSQL))

	q2 := CompiledQuery{Sql: "SELECT $", Params: []core.Value{}}
	assert.Equal(t, "SELECT $", q2.DebugSql(DatabaseKindPostgreSQL))
	
	q3 := CompiledQuery{Sql: "SELECT 'a''", Params: []core.Value{}}
	assert.Equal(t, "SELECT 'a''", q3.DebugSql(DatabaseKindPostgreSQL))
}

func TestCompiledQueryDebugSqlUnknown(t *testing.T) {
	q := CompiledQuery{
		Sql: "SELECT 1",
	}
	assert.Equal(t, "SELECT 1", q.DebugSql(DatabaseKind(-1)))
}

func TestSqlLiteral(t *testing.T) {
	tests := []struct {
		val  core.Value
		kind DatabaseKind
		want string
	}{
		{core.ValNull(), DatabaseKindPostgreSQL, "NULL"},
		{core.ValBool(true), DatabaseKindPostgreSQL, "TRUE"},
		{core.ValBool(false), DatabaseKindPostgreSQL, "FALSE"},
		{core.ValI64(42), DatabaseKindPostgreSQL, "42"},
		{core.ValU64(42), DatabaseKindPostgreSQL, "42"},
		{core.ValF64(3.14), DatabaseKindPostgreSQL, "3.140000"},
		{core.ValDecimal(decimal.NewFromFloat(1.23)), DatabaseKindPostgreSQL, "1.23"},
		{core.ValText("foo"), DatabaseKindPostgreSQL, "'foo'"},
		{core.Value{V: []core.Value{core.ValI64(1), core.ValI64(2)}}, DatabaseKindPostgreSQL, "ARRAY[1, 2]"},
		{core.Value{V: []core.Value{core.ValI64(1), core.ValI64(2)}}, DatabaseKindSQLite, "(1, 2)"},
		{core.Value{V: core.Record{"id": core.ValI64(1)}}, DatabaseKindPostgreSQL, "'map[id:1]'"}, // rough json
		{core.Value{V: timeValue()}, DatabaseKindPostgreSQL, "'{}'"}, // fallback
	}
	
	for _, tt := range tests {
		assert.Equal(t, tt.want, sqlLiteral(tt.val, tt.kind))
	}
}

// Just a dummy struct to trigger default fallback
type dummyStruct struct{}
func timeValue() dummyStruct { return dummyStruct{} }

func TestErrors(t *testing.T) {
	assert.Equal(t, "unknown entity: User", ErrUnknownEntity("User").Error())
	assert.Equal(t, "unknown field: name", ErrUnknownField("name").Error())
	assert.Equal(t, "IN requires at least one value", ErrEmptyInList().Error())
	assert.Equal(t, "entity User has no id property", ErrMissingIdProperty("User").Error())
	assert.Equal(t, "entity User has no version property", ErrMissingVersionProperty("User").Error())
	assert.Equal(t, "Insert requires at least one writable field", ErrEmptyMutation("Insert").Error())
	assert.Equal(t, "recover requires a negative version, got 1", ErrInvalidRecoverVersion(1).Error())
	assert.Equal(t, "unsupported schema type: 99", ErrUnsupportedSchemaType(core.DataType(99)).Error())
	assert.Equal(t, "invalid args", ErrInvalidFunctionArguments("invalid args").Error())
	assert.Equal(t, "subquery does not support operator: =", ErrInvalidSubQueryOperator("=").Error())
}

package sql

import (
	"testing"
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
}

func TestCompiledQueryDebugSqlPostgres(t *testing.T) {
	q := CompiledQuery{
		Sql:    "SELECT * FROM users WHERE name = $1 AND age > $2 AND id = $10",
		Params: []core.Value{core.ValText("John O'Connor"), core.ValI64(18), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValNull(), core.ValI64(999)},
	}
	
	debug := q.DebugSql(DatabaseKindPostgreSQL)
	assert.Equal(t, "SELECT * FROM users WHERE name = 'John O''Connor' AND age > 18 AND id = 999", debug)
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
}

func TestCompiledQueryDebugSqlPositionalIgnoresStringQuotes(t *testing.T) {
	q := CompiledQuery{
		Sql:    "SELECT * FROM users WHERE name = '?' AND age > ?",
		Params: []core.Value{core.ValI64(18)},
	}
	
	debugSqlite := q.DebugSql(DatabaseKindSQLite)
	assert.Equal(t, "SELECT * FROM users WHERE name = '?' AND age > 18", debugSqlite)
}

package mysql

import (
	stdcontext "context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

func TestMysqlErrorsWithMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	exec := NewMysqlMutationExecutor(db)
	context := stdcontext.Background()
	query := &teaql_sql.CompiledQuery{Sql: "SELECT * FROM test"}

	// 1. FetchAllSql - rows error
	mock.ExpectQuery("SELECT \\* FROM test").WillReturnError(fmt.Errorf("query error"))
	_, err = exec.FetchAllSql(context, query)
	if err == nil {
		t.Errorf("Expected query error")
	}

	// Tx FetchAllSql - query error
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM test").WillReturnError(fmt.Errorf("tx query error"))
	mock.ExpectRollback()

	txExec, _ := exec.BeginSql(context)
	_, err = txExec.FetchAllSql(context, query)
	if err == nil {
		t.Errorf("Expected tx query error")
	}
	txExec.RollbackSql(context)

	// Tx ExecuteSql - query error
	mock.ExpectBegin()
	mock.ExpectExec("INSERT").WillReturnError(fmt.Errorf("tx exec error"))
	mock.ExpectRollback()

	txExec2, _ := exec.BeginSql(context)
	_, err = txExec2.ExecuteSql(context, &teaql_sql.CompiledQuery{Sql: "INSERT"})
	if err == nil {
		t.Errorf("Expected tx exec error")
	}
	txExec2.RollbackSql(context)

	dialect := &MysqlDialect{}
	ent := &core.EntityDescriptor{TabName: "test"}
	prop := &core.PropertyDescriptor{Name: "col", ColName: "col", DataType: core.DataType(999)}
	_, err = dialect.CompileAddColumn(ent, prop)
	if err == nil {
		t.Errorf("Expected error from CompileAddColumn with bad type")
	}
}

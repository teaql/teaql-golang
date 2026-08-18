package sqlite

import (
	stdcontext "context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	teaql_sql "github.com/teaql/teaql-golang/sql"
)

type mockDriver struct{}

func (m *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{}, nil
}

type mockConn struct{}

func (m *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{}, nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin error")
}

func (m *mockConn) BeginTx(context stdcontext.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, errors.New("begin error")
}

type mockStmt struct{}

func (m *mockStmt) Close() error { return nil }

func (m *mockStmt) NumInput() int { return -1 }

func (m *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &mockResult{}, nil
}

func (m *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &mockRows{}, nil
}

type mockResult struct{}

func (m *mockResult) LastInsertId() (int64, error) { return 0, nil }

func (m *mockResult) RowsAffected() (int64, error) {
	return 0, errors.New("rows affected error")
}

type mockRows struct{}

func (m *mockRows) Columns() []string {
	// We can't trigger error on Columns() directly via driver interface in a simple way without causing a panic or it might not propagate nicely, but let's try returning nil
	return []string{"id"}
}

func (m *mockRows) Close() error { return nil }

func (m *mockRows) Next(dest []driver.Value) error {
	return errors.New("next error")
}

func init() {
	sql.Register("mock_sqlite", &mockDriver{})
}

func TestSqliteExecutorErrors(t *testing.T) {
	db, err := sql.Open("mock_sqlite", "")
	if err != nil {
		t.Fatal(err)
	}

	exec := NewSqliteMutationExecutor(db)
	context := stdcontext.Background()

	// 1. BeginTx error
	_, err = exec.BeginSql(context)
	if err == nil {
		t.Errorf("Expected begin error")
	}

	// 2. RowsAffected error
	query := &teaql_sql.CompiledQuery{Sql: "INSERT"}
	_, err = exec.ExecuteSql(context, query)
	if err == nil {
		t.Errorf("Expected RowsAffected error")
	}

	// 3. Scan error (triggered by Next returning error which makes Scan fail or return error, wait, driver.Rows Next returning error stops iteration. To make Scan fail, we should return a type that can't be scanned, but actually rows.Scan error in FetchAllSql is what we want. If rows.Next() fails, rows.Scan isn't called. We need rows.Next() to succeed and then rows.Scan to fail.
	// Actually, a simpler way is to close the rows before scanning. Wait, we can't do that easily.)
}

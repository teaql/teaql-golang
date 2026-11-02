package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"testing"
)

type featureQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func assertConnectionFeatures(t *testing.T, q featureQueryer) {
	t.Helper()
	for _, c := range []struct{ query, want string }{
		{`SELECT soundex('Robert')`, "R163"},
		{`SELECT teaql_decimal_valid('9007199254740993.01')`, "1"},
		{`SELECT teaql_decimal_valid('not-a-decimal')`, "0"},
		{`SELECT '9007199254740993.01' COLLATE TEAQL_DECIMAL > '9007199254740992.99'`, "1"},
		{`SELECT TEAQL_DECIMAL_SUM(v) FROM (SELECT '9007199254740993.01' AS v UNION ALL SELECT '0.09')`, "9007199254740993.10"},
	} {
		var actual string
		if err := q.QueryRowContext(context.Background(), c.query).Scan(&actual); err != nil || actual != c.want {
			t.Fatalf("connection feature query %s: got %q, error=%v, want=%q", c.query, actual, err, c.want)
		}
	}
	var empty sql.NullString
	if err := q.QueryRowContext(context.Background(), `SELECT TEAQL_DECIMAL_SUM(v) FROM (SELECT '0.01' AS v) WHERE 0`).Scan(&empty); err != nil || empty.Valid {
		t.Fatalf("empty exact Decimal SUM must remain NULL: %+v %v", empty, err)
	}
}

func featureDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenDatabase(filepath.Join(t.TempDir(), "features.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func TestOpenDatabaseFeaturesSurvivePhysicalDiscard(t *testing.T) {
	db := featureDatabase(t)
	assertConnectionFeatures(t, db)
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Raw(func(any) error { return driver.ErrBadConn }); !errors.Is(err, driver.ErrBadConn) {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
		t.Fatal(err)
	}
	assertConnectionFeatures(t, db)
}

func TestOpenDatabaseFeaturesOnEveryConcurrentConnection(t *testing.T) {
	db := featureDatabase(t)
	db.SetMaxOpenConns(3)
	// EnsureSchema's compatibility initializer must not reset an explicit pool size.
	if err := EnsureSoundex(db); err != nil {
		t.Fatal(err)
	}
	if db.Stats().MaxOpenConnections != 3 {
		t.Fatal("initializer reset pool configuration")
	}
	for n := 0; n < 3; n++ {
		conn, err := db.Conn(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := conn.Close(); err != nil {
				t.Error(err)
			}
		})
		assertConnectionFeatures(t, conn)
	}
}

func TestOpenDatabaseFeaturesAfterIdleEviction(t *testing.T) {
	db := featureDatabase(t)
	assertConnectionFeatures(t, db)
	db.SetMaxIdleConns(0)
	assertConnectionFeatures(t, db)
	if db.Stats().MaxIdleClosed == 0 {
		t.Fatal("physical connections were not evicted")
	}
}

func TestOpenDatabaseFeaturesAfterTransactionCancellation(t *testing.T) {
	db := featureDatabase(t)
	assertConnectionFeatures(t, db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT 1"); err != nil {
		t.Fatal(err)
	}
	cancel()
	// Let database/sql own the cancellation rollback instead of racing it with
	// an explicit rollback. The max-one pool waits until that transaction exits.
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	assertConnectionFeatures(t, conn)
}

func TestOpenDatabaseDoesNotModifyDefaultDriverOrCreateSchema(t *testing.T) {
	db := featureDatabase(t)
	assertConnectionFeatures(t, db)
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table'`).Scan(&count); err != nil || count != 0 {
		t.Fatal("open changed schema", count, err)
	}
	ordinary, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "ordinary.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer ordinary.Close()
	if err := ordinary.QueryRow(`SELECT teaql_decimal_valid('1.00')`).Scan(&count); err == nil {
		t.Fatal("default SQLite driver was globally modified")
	}
}

package sqlitefeatures

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	stdRuntime "runtime"
	"strings"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/shopspring/decimal"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
)

type lifecycleQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
type lifecycleFeature struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Error  string `json:"error,omitempty"`
}

func lifecycleFeatures(q lifecycleQueryer) []lifecycleFeature {
	result := []lifecycleFeature{}
	for _, c := range []struct{ name, sql, want string }{
		{"soundex", `SELECT soundex('Robert')`, "R163"},
		{"decimal-valid", `SELECT teaql_decimal_valid('9007199254740993.01')`, "1"},
		{"decimal-invalid", `SELECT teaql_decimal_valid('not-a-decimal')`, "0"},
		{"decimal-collation", `SELECT '9007199254740993.01' COLLATE TEAQL_DECIMAL > '9007199254740992.99'`, "1"},
		{"decimal-sum", `SELECT TEAQL_DECIMAL_SUM(v) FROM (SELECT '9007199254740993.01' AS v UNION ALL SELECT '0.09')`, "9007199254740993.10"},
	} {
		var value string
		err := q.QueryRowContext(context.Background(), c.sql).Scan(&value)
		f := lifecycleFeature{Name: c.name, Passed: err == nil && value == c.want}
		if err != nil {
			f.Error = err.Error()
		} else if value != c.want {
			f.Error = "unexpected scalar result"
		}
		result = append(result, f)
	}
	var empty sql.NullString
	err := q.QueryRowContext(context.Background(), `SELECT TEAQL_DECIMAL_SUM(v) FROM (SELECT '0.01' AS v) WHERE 0`).Scan(&empty)
	f := lifecycleFeature{Name: "decimal-empty-sum", Passed: err == nil && !empty.Valid}
	if err != nil {
		f.Error = err.Error()
	}
	return append(result, f)
}
func lifecycleFrame(name string, facts any) {
	b, err := json.Marshal(map[string]any{"case": name, "facts": facts})
	if err != nil {
		panic(err)
	}
	fmt.Println("SQLITE_LIFECYCLE_CASE=" + string(b))
}
func lifecycleGraph(ctx *runtime.UserContext, phone string) (uint64, uint64, error) {
	owner := lib.Q.PurchaseOrders().Comment("Prepare a physical SQLite lifecycle graph").Purpose("Verify generated saves retain exact scalar support after reconnect").NewEntity(ctx).UpdateContactPhone(phone).UpdateInternalNote(nil)
	child := lib.Q.OrderItems().Comment("Prepare the exact Decimal lifecycle child").Purpose("Verify generated Q and E use the replacement connection").NewEntity(ctx).UpdateProductName("Connection lifecycle item").UpdateUnitPrice(decimal.RequireFromString("12.34"))
	owner.OrderItemList().Add(child)
	_, err := owner.AuditAs("Persist one exact graph through the configured SQLite pool").Save(ctx)
	if err != nil {
		return 0, 0, err
	}
	loaded, err := lib.Q.PurchaseOrders().WithIdIs(owner.Id()).Limit(1).Comment("Reload the lifecycle owner").Purpose("Verify committed identity, root and timestamp through generated E").ExecuteForOne(ctx)
	if err != nil || loaded == nil {
		return 0, 0, fmt.Errorf("reload owner: %w", err)
	}
	name, ok := lib.E.PurchaseOrder(loaded).ContactPhone().Eval()
	if !ok || name != phone {
		return 0, 0, fmt.Errorf("owner business value changed")
	}
	root, ok := lib.E.PurchaseOrder(loaded).TenantId().Eval()
	if !ok || root != 1 {
		return 0, 0, fmt.Errorf("context root changed")
	}
	created, ok := lib.E.PurchaseOrder(loaded).CreateTime().Eval()
	if !ok || created.UnixMilli() != 1988247845000 {
		return 0, 0, fmt.Errorf("business clock changed")
	}
	item, err := lib.Q.OrderItems().WithIdIs(child.Id()).Limit(1).Comment("Reload exact Decimal after the physical connection transition").Purpose("Verify typed Decimal did not pass through binary floating point").ExecuteForOne(ctx)
	if err != nil || item == nil {
		return 0, 0, fmt.Errorf("reload child: %w", err)
	}
	amount, ok := lib.E.OrderItem(item).UnitPrice().Eval()
	if !ok || !amount.Equal(decimal.RequireFromString("12.34")) {
		return 0, 0, fmt.Errorf("exact Decimal changed")
	}
	return owner.Id(), child.Id(), nil
}
func TestCore236SQLiteConnectionLifecycle(t *testing.T) {
	for _, name := range []string{"replacement", "fanout", "idle-eviction", "transaction-cancellation"} {
		t.Run(name, func(t *testing.T) {
			base := os.Getenv("CORE236_SQLITE_LIFECYCLE_DB")
			if base == "" {
				t.Fatal("retained database required")
			}
			t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", base+"."+name+".sqlite")
			ctx, err := lib.ServiceRuntimeFromEnv()
			if err != nil {
				t.Fatal(err)
			}
			db := ctx.GetResource("db").(*sql.DB)
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(1988247845000).UTC()))
			if err := lib.EnsureSchema(ctx); err != nil {
				t.Fatal(err)
			}
			initial := lifecycleFeatures(db)
			for _, f := range initial {
				if !f.Passed {
					t.Fatalf("initial calibration failed: %+v", f)
				}
			}
			beforeParent, beforeChild, err := lifecycleGraph(ctx, "5550720")
			if err != nil {
				t.Fatal("initial generated graph", err)
			}
			c, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			var first any
			if err := c.Raw(func(v any) error { first = v; return nil }); err != nil {
				t.Fatal(err)
			}
			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
			var observations [][]lifecycleFeature
			changed := false
			switch name {
			case "replacement":
				c, err := db.Conn(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if err := c.Raw(func(any) error { return driver.ErrBadConn }); !errors.Is(err, driver.ErrBadConn) {
					t.Fatal("driver did not discard the owned physical connection", err)
				}
				if err := c.Close(); err != nil && !errors.Is(err, sql.ErrConnDone) {
					t.Fatal(err)
				}
			case "fanout":
				db.SetMaxOpenConns(3)
				var connections []*sql.Conn
				defer func() {
					for _, c := range connections {
						_ = c.Close()
					}
				}()
				for n := 0; n < 3; n++ {
					c, err := db.Conn(context.Background())
					if err != nil {
						t.Fatal(err)
					}
					connections = append(connections, c)
					observations = append(observations, lifecycleFeatures(c))
				}
				for _, c := range connections {
					if err := c.Close(); err != nil {
						t.Fatal(err)
					}
				}
				connections = nil
			case "idle-eviction":
				db.SetMaxIdleConns(0)
			case "transaction-cancellation":
				owned, cancel := context.WithCancel(context.Background())
				defer cancel()
				tx, err := db.BeginTx(owned, nil)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := tx.ExecContext(owned, "SELECT 1"); err != nil {
					t.Fatal(err)
				}
				cancel()
				if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Fatal(err)
				}
			}
			c, err = db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			var second any
			if err := c.Raw(func(v any) error { second = v; return nil }); err != nil {
				t.Fatal(err)
			}
			changed = first != second
			observations = append(observations, lifecycleFeatures(c))
			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
			stdRuntime.KeepAlive(first)
			afterParent, afterChild, saveErr := lifecycleGraph(ctx, "5550721")
			stats := db.Stats()
			lifecycleFrame(name, map[string]any{"initial": initial, "observations": observations, "physicalConnectionChanged": changed, "initialParentId": beforeParent, "initialChildId": beforeChild, "afterParentId": afterParent, "afterChildId": afterChild, "generatedSavePassed": saveErr == nil, "saveError": fmt.Sprint(saveErr), "maxOpen": stats.MaxOpenConnections, "idleClosed": stats.MaxIdleClosed, "callerContextReplaced": false})
			if name == "replacement" || name == "idle-eviction" {
				if !changed {
					t.Error("physical replacement not established")
				}
			}
			for _, rows := range observations {
				for _, f := range rows {
					if !f.Passed {
						t.Errorf("replacement feature failed: %+v", f)
					}
				}
			}
			if saveErr != nil {
				t.Error("generated save after connection transition", saveErr)
			}
		})
	}
}
func TestCore236SQLiteRuntimeOrigin(t *testing.T) {
	pc := reflect.ValueOf(provider.NewSqliteMutationExecutor).Pointer()
	file, _ := stdRuntime.FuncForPC(pc).FileLine(pc)
	root := os.Getenv("CORE236_SQLITE_RUNTIME_ROOT")
	if root == "" || !strings.HasPrefix(file, root+"/") {
		t.Fatal("provider escaped selected distribution", file)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(b)
	if hex.EncodeToString(digest[:]) != os.Getenv("CORE236_SQLITE_PROVIDER_SHA256") {
		t.Fatal("provider origin hash mismatch")
	}
	lifecycleFrame("runtime-origin", map[string]any{"file": file, "sha256": hex.EncodeToString(digest[:]), "sourceFallback": false})
}

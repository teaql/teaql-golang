package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

func TestGeneratedCheckerExactDecimalRanges(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		t.Fatal("retained SQLite path required")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database+".checker-ranges.sqlite")
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
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, time.February, 29, 10, 15, 30, 0, time.UTC)))
	if err := lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	parent, parentCreationErr := lib.Q.PurchaseOrders().Comment("Prepare exact modeled-bound controls").Purpose("Keep Decimal boundaries precise through generated saves").NewEntity(ctx)
	if parentCreationErr != nil {
		t.Fatal(parentCreationErr)
	}
	parent.UpdateContactPhone("5550965").UpdateInternalNote(nil)
	if _, err := parent.AuditAs("Create a valid parent for bounded numeric probes").Save(ctx); err != nil {
		t.Fatal(err)
	}
	provider := ctx.GetResource("dataService")
	probe := &guardedExampleProvider{&restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}}
	ctx.InsertResource("dataService", probe)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	var before int
	if err := db.QueryRow("SELECT COUNT(*) FROM order_item_data").Scan(&before); err != nil {
		t.Fatal(err)
	}
	results := []map[string]any{}
	valid := 0
	for _, test := range []struct {
		name, value, rule string
		valid             bool
	}{
		{"below-minimum", "0", "min", false},
		{"fraction-below-minimum", "0.009999999999999999999999", "min", false},
		{"fraction-above-maximum", "1000000.000000000000000001", "max", false},
		{"above-maximum", "1000001", "max", false},
		{"inclusive-minimum", "0.01", "", true},
		{"inclusive-maximum", "1000000", "", true},
	} {
		child, childCreationErr := lib.Q.OrderItems().Comment("Prepare an exact numeric boundary input").Purpose("Verify comparisons never round through float64").NewEntity(ctx)
		if childCreationErr != nil {
			t.Fatal(childCreationErr)
		}
		child.UpdatePurchaseOrderId(parent.Id()).UpdateProductName(test.name).UpdateUnitPrice(decimal.RequireFromString(test.value))
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, audit.events = 0, 0, 0, 0, 0, 0
		saved, err := child.AuditAs("Apply the model-derived Decimal range checker before persistence").Save(ctx)
		effects := map[string]int{"begins": probe.begins, "queries": probe.queries, "mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks, "audits": audit.events}
		row := map[string]any{"case": test.name, "value": test.value, "allowed": test.valid, "effects": effects}
		if test.valid {
			if err != nil || saved == nil {
				t.Fatal(test.name, err)
			}
			if probe.begins != 1 || probe.mutations != 1 || probe.commits != 1 || audit.events != 1 || probe.rollbacks != 0 {
				t.Fatal(test.name, effects)
			}
			var price string
			var version int64
			if err := db.QueryRow("SELECT unit_price,version FROM order_item_data WHERE id=?", saved.Id()).Scan(&price, &version); err != nil {
				t.Fatal(err)
			}
			if !decimal.RequireFromString(price).Equal(decimal.RequireFromString(test.value)) || version != 1 {
				t.Fatal("exact persisted control", price, version)
			}
			value, loaded := lib.E.OrderItem(saved).UnitPrice().Eval()
			if !loaded || !value.Equal(decimal.RequireFromString(test.value)) {
				t.Fatal("generated E boundary", value, loaded)
			}
			row["id"], row["version"] = saved.Id(), version
			valid++
		} else {
			var failure *runtime.RuntimeError
			if !errors.As(err, &failure) || failure.Type != "Check" || len(failure.CheckResults) != 1 {
				t.Fatal(test.name, err)
			}
			result := failure.CheckResults[0]
			if result.EntityType != "OrderItem" || result.RuleID != test.rule || result.ModelPath() != "unit_price" || fmt.Sprint(result.InputValue) != test.value {
				t.Fatal(test.name, result)
			}
			for name, value := range effects {
				if value != 0 {
					t.Fatalf("%s: rejected save performed %s=%d", test.name, name, value)
				}
			}
			row["violation"] = result.ToWire(runtime.JsonFieldCamelCase)
		}
		var after int
		if err := db.QueryRow("SELECT COUNT(*) FROM order_item_data").Scan(&after); err != nil {
			t.Fatal(err)
		}
		if after != before+valid {
			t.Fatalf("%s leaked invalid rows: %d != %d", test.name, after, before+valid)
		}
		results = append(results, row)
	}
	data, err := json.Marshal(map[string]any{"language": "go", "passed": true, "cases": results, "validSaves": 2, "rejectedSaves": 4, "sameDatabase": true})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_CHECKER_EXACT_RANGES=" + string(data))
}

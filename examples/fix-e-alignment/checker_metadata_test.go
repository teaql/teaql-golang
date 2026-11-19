package fixe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	lib "checker-fix-advanced-core-workspace/lib"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type metadataExampleChecker struct {
	delegate   runtime.CheckerRegistry
	specialize bool
}

func (c *metadataExampleChecker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := c.delegate.CheckAndFix(ctx, input)
	if c.specialize && input.Entity == "order_item" {
		for i := range results {
			results[i].EntityType = "ReviewedItem"
			results[i].SourceInstancePath = "/order_item_list/0/product~1name~0alias"
		}
	}
	return results
}

// Exercise actual generated Q/E/audited saves, not constructed RuntimeErrors.
func TestGeneratedCheckerDeclaringTypesAndNestedLocations(t *testing.T) {
	database := os.Getenv("TEAQL_FIX_E_DB")
	if database == "" {
		t.Fatal("a retained SQLite path is required for the independent oracle")
	}
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database)
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
	provider := ctx.GetResource("dataService")
	probe := &guardedExampleProvider{&restorationProvider{QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}}
	ctx.InsertResource("dataService", probe)
	audit := &restorationAudit{}
	ctx.WithAppAuditEventSink(audit)
	checker := &metadataExampleChecker{delegate: lib.Module().CheckerRegistry}
	ctx.WithCheckerRegistry(checker)
	newOwner := func() *purchase_order.PurchaseOrder {
		created, createdCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a Checker metadata regression").
			Purpose("Require declaring entity identity on actual generated-save errors").NewEntity(ctx)
		if createdCreationErr != nil {
			t.Fatal(createdCreationErr)
		}
		return created
	}
	good := newOwner().UpdateContactPhone("5550961").UpdateInternalNote(nil)
	if _, err := good.AuditAs("Calibrate successful Checker admission and actual provider instrumentation").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if probe.begins != 1 || probe.mutations != 1 || probe.commits != 1 || audit.events != 1 {
		t.Fatal("positive save did not calibrate provider/audit probes", *probe.restorationProvider, audit.events)
	}
	type physicalState struct {
		Parents, Children int
		Phone             string
		Version           int64
	}
	physical := func() physicalState {
		var state physicalState
		if err := db.QueryRow("SELECT COUNT(*) FROM purchase_order_data").Scan(&state.Parents); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT COUNT(*) FROM order_item_data").Scan(&state.Children); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT contact_phone,version FROM purchase_order_data WHERE id=?", good.Id()).Scan(&state.Phone, &state.Version); err != nil {
			t.Fatal(err)
		}
		return state
	}
	before := physical()
	observations := []map[string]any{}
	metadataFailures := []map[string]any{}
	humanLabelFailures := []map[string]any{}
	projections := 0
	invalid := func(scenario string, save func() error, expectedType, expectedPath, expectedRule string) {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, audit.events = 0, 0, 0, 0, 0, 0
		err := save()
		var failure *runtime.RuntimeError
		if !errors.As(err, &failure) || failure.Type != "Check" || len(failure.CheckResults) == 0 {
			t.Fatalf("%s: expected model Checker rejection, got %v", scenario, err)
		}
		effects := map[string]int{"begins": probe.begins, "queries": probe.queries, "mutations": probe.mutations,
			"commits": probe.commits, "rollbacks": probe.rollbacks, "audits": audit.events}
		for name, value := range effects {
			if value != 0 {
				t.Fatalf("%s performed %s=%d", scenario, name, value)
			}
		}
		if got := physical(); !reflect.DeepEqual(got, before) {
			t.Fatalf("%s changed storage: %+v != %+v", scenario, got, before)
		}
		wires := map[string][]runtime.WireCheckResult{}
		matched := false
		for _, result := range failure.CheckResults {
			if result.RuleID == expectedRule && result.ModelPath() == expectedPath {
				matched = true
			}
			if result.EntityType != expectedType {
				metadataFailures = append(metadataFailures, map[string]any{"case": scenario, "expected": expectedType, "actual": result.EntityType})
			}
			// This frozen library predates generated message labels. The natural
			// renderer's readable canonical fallback belongs in the sentence;
			// complete native/model/JSON paths remain structured diagnostics.
			segments := result.ObjectLocation().Segments
			leaf := ""
			if len(segments) > 0 && segments[len(segments)-1].Property != nil {
				leaf = strings.ReplaceAll(*segments[len(segments)-1].Property, "_", " ")
			}
			if result.RuleID != "invalid_type" && (leaf == "" || !strings.Contains(result.Message, leaf)) {
				humanLabelFailures = append(humanLabelFailures, map[string]any{"case": scenario, "expected": leaf, "actual": result.Message})
			}
			for _, profile := range []runtime.JsonFieldNamingProfile{runtime.JsonFieldCamelCase, runtime.JsonFieldSnakeCase, runtime.JsonFieldPascalCase} {
				wires[string(profile)] = append(wires[string(profile)], result.ToWire(profile))
				projections++
			}
		}
		if !matched {
			t.Fatalf("%s: missing %s at %s: %+v", scenario, expectedRule, expectedPath, failure.CheckResults)
		}
		observation := map[string]any{"case": scenario, "expectedEntityType": expectedType, "effects": effects, "wires": wires}
		observations = append(observations, observation)
		data, err := json.Marshal(observation)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("GO_CHECKER_METADATA_CASE=" + string(data))
	}
	missing := newOwner()
	invalid("parent_required", func() error { _, err := missing.AuditAs("Reject an omitted required field").Save(ctx); return err }, "PurchaseOrder", "contact_phone", "required")
	long := newOwner().UpdateContactPhone(strings.Repeat("x", 33))
	invalid("parent_length", func() error { _, err := long.AuditAs("Reject a modeled length violation").Save(ctx); return err }, "PurchaseOrder", "contact_phone", "max_length")
	complete, err := lib.Q.PurchaseOrders().WithIdIs(good.Id()).Limit(1).Comment("Load the complete owner before null admission").
		Purpose("Verify an explicitly invalid canonical value is not omission").ExecuteForOne(ctx)
	if err != nil || complete == nil {
		t.Fatal("missing complete owner", err)
	}
	// A required Go string cannot be nil. The public canonical mutation ledger
	// must still reject explicit Null rather than passing it to SQLite.
	complete.EntityRoot().Set(complete.EntityKey(), "contact_phone", core.ValNull())
	invalid("update_null", func() error { _, err := complete.AuditAs("Reject an explicit canonical null").Save(ctx); return err }, "PurchaseOrder", "contact_phone", "required")
	partial, err := lib.Q.PurchaseOrdersMinimal().WithIdIs(good.Id()).Limit(1).Comment("Load an identity-only owner").
		Purpose("Reject a modified incomplete object before persistence").ExecuteForOne(ctx)
	if err != nil || partial == nil {
		t.Fatal("missing partial owner", err)
	}
	partial.UpdateContactPhone("5550962")
	invalid("partial_owner", func() error { _, err := partial.AuditAs("Reject an incomplete modified owner").Save(ctx); return err }, "PurchaseOrder", "tenant", "invalid_type")
	badChild := func(scenario, product, price, rule, field, wantType string) {
		parent := newOwner().UpdateContactPhone("5550963").UpdateInternalNote(nil)
		child, childCreationErr := lib.Q.OrderItems().Comment("Prepare an invalid indexed child").Purpose("Preserve the child declaring type when prefixing errors").NewEntity(ctx)
		if childCreationErr != nil {
			t.Fatal(childCreationErr)
		}
		if product != "" {
			child.UpdateProductName(product)
		}
		child.UpdateUnitPrice(decimal.RequireFromString(price))
		parent.OrderItemList().Add(child)
		invalid(scenario, func() error {
			_, err := parent.AuditAs("Reject an invalid child graph before its first write").Save(ctx)
			return err
		}, wantType, "order_item_list[0]."+field, rule)
	}
	badChild("child_required", "", "10.00", "required", "product_name", "OrderItem")
	badChild("child_length", strings.Repeat("x", 81), "10.00", "max_length", "product_name", "OrderItem")
	badChild("child_min", "Synthetic Widget", "0", "min", "unit_price", "OrderItem")
	badChild("child_max", "Synthetic Widget", "1000001", "max", "unit_price", "OrderItem")
	if err := ctx.SetLocaleCode("zh-CN"); err != nil {
		t.Fatal(err)
	}
	localized := newOwner()
	invalid("localized_required", func() error {
		_, err := localized.AuditAs("Reject an omission with a localized message").Save(ctx)
		return err
	}, "PurchaseOrder", "contact_phone", "required")
	if err := ctx.SetLocaleCode("en"); err != nil {
		t.Fatal(err)
	}
	checker.specialize = true
	if err := ctx.SetLocaleCode("zh-CN"); err != nil {
		t.Fatal(err)
	}
	badChild("specialized_child", "", "10.00", "required", "product_name", "ReviewedItem")
	passed := len(metadataFailures) == 0 && len(humanLabelFailures) == 0
	data, err := json.Marshal(map[string]any{"language": "go", "passed": passed, "normalGeneratedSave": true,
		"observations": observations, "invalidSaves": len(observations), "wireProjections": projections, "metadataFailures": metadataFailures,
		"humanLabelFailures": humanLabelFailures, "readableFieldLabels": len(humanLabelFailures) == 0,
		"positiveProviderCalibration": true, "sameContextAndDatabase": true, "physicalId": good.Id(), "physicalVersion": before.Version})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GO_CHECKER_METADATA=" + string(data))
	if !passed {
		t.Fatalf("Checker diagnostics lost declaring types=%+v or readable field labels=%+v", metadataFailures, humanLabelFailures)
	}
}

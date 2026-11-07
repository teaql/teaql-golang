package fixe

import (
	stdcontext "context"
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

type rootBoundaryProvider struct {
	*guardedExampleProvider
	rootLookups int
}

func (p *rootBoundaryProvider) Query(ctx stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	if request.Query != nil && strings.EqualFold(request.Query.Entity, "tenant") {
		p.rootLookups++
	}
	return p.guardedExampleProvider.Query(ctx, request)
}

func (p *rootBoundaryProvider) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	tx, err := p.guardedExampleProvider.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &rootBoundaryTransaction{Transaction: tx, owner: p}, nil
}

type rootBoundaryTransaction struct {
	data_service.Transaction
	owner *rootBoundaryProvider
}

func (tx *rootBoundaryTransaction) Query(ctx stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	if request.Query != nil && strings.EqualFold(request.Query.Entity, "tenant") {
		tx.owner.rootLookups++
	}
	return tx.Transaction.Query(ctx, request)
}

func (tx *rootBoundaryTransaction) GenerateId(entity string) (uint64, error) {
	return tx.Transaction.(runtime.InternalIdGenerator).GenerateId(entity)
}

func (tx *rootBoundaryTransaction) EnsureIdFloor(ctx stdcontext.Context, entity string, floor uint64) error {
	return tx.Transaction.(runtime.InternalIdFloorReserver).EnsureIdFloor(ctx, entity, floor)
}

type rootBoundaryChecker struct {
	delegate runtime.CheckerRegistry
	effects  func() map[string]int
	t        *testing.T
	seen     []map[string]any
}

func (c *rootBoundaryChecker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	before := c.effects()
	results := c.delegate.CheckAndFix(ctx, input)
	if !reflect.DeepEqual(before, c.effects()) {
		c.t.Fatal("generated Fix performed provider, allocator, policy or audit work")
	}
	root, present := input.Values["tenant"]
	id, valid := root.TryU64()
	c.seen = append(c.seen, map[string]any{"entity": input.Entity, "operation": input.Operation,
		"rootPresent": present && valid, "root": id, "instant": input.Now.UnixMilli(), "ioDuringFix": 0})
	return results
}

func TestGeneratedContextRootInjectionAndOwnershipBoundaries(t *testing.T) {
	base := os.Getenv("TEAQL_FIX_E_DB")
	if base == "" {
		t.Fatal("a retained SQLite path is required")
	}
	database := base + ".context-root.sqlite"
	t.Setenv("CHECKER_FIX_ADVANCED_CORE_DATABASE_URL", database)
	bootstrap, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := bootstrap.GetResource("db").(*sql.DB)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := lib.EnsureSchema(bootstrap); err != nil {
		t.Fatal(err)
	}
	active, err := bootstrap.RequireActiveRoot("Tenant")
	if err != nil || active.ID == 0 {
		t.Fatal("bootstrap did not provide a real typed root", err)
	}
	provider := bootstrap.GetResource("dataService")
	probe := &rootBoundaryProvider{guardedExampleProvider: &guardedExampleProvider{&restorationProvider{
		QueryExecutor: provider.(data_service.QueryExecutor), transaction: provider.(data_service.TransactionExecutor)}}}
	audit := &restorationAudit{}
	policyCalls := 0
	ids := &lateExampleIDs{}
	effects := func() map[string]int {
		return map[string]int{"begins": probe.begins, "queries": probe.queries,
			"mutations": probe.mutations, "commits": probe.commits, "rollbacks": probe.rollbacks,
			"rootLookups": probe.rootLookups, "idAllocations": ids.calls, "policies": policyCalls, "audits": audit.events}
	}
	reset := func() {
		probe.begins, probe.queries, probe.mutations, probe.commits, probe.rollbacks, probe.rootLookups = 0, 0, 0, 0, 0, 0
		ids.calls, policyCalls, audit.events = 0, 0, 0
	}
	checker := &rootBoundaryChecker{delegate: lib.Module().CheckerRegistry, effects: effects, t: t}
	contextFor := func(root *runtime.EntityReference) *runtime.UserContext {
		ctx := lib.Module().IntoContext()
		if _, err := ctx.RequireActiveRoot("Tenant"); err == nil {
			t.Fatal("module installation must not choose an active root")
		}
		ctx.InsertResource("dataService", probe)
		ctx.WithCheckerRegistry(checker).WithAppAuditEventSink(audit)
		ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2032, 2, 29, 10, 15, 30, 0, time.UTC)))
		ctx.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(string) runtime.MutationPolicy { policyCalls++; return nil }))
		if root != nil {
			ctx.WithActiveRoot(*root)
		}
		ids.context = ctx
		ctx.WithInternalIdGenerator(ids)
		return ctx
	}
	newOrder := func(ctx *runtime.UserContext) *purchase_order.PurchaseOrder {
		created, createdCreationErr := lib.Q.PurchaseOrders().Comment("Prepare a generated root-binding boundary").
			Purpose("Derive ownership only from trusted context").NewEntity(ctx)
		if createdCreationErr != nil {
			t.Fatal(createdCreationErr)
		}
		return created.UpdateContactPhone("5550980").UpdateInternalNote(nil)
	}
	ctx := contextFor(&active)
	parent := newOrder(ctx)
	child, childCreationErr := lib.Q.OrderItems().Comment("Prepare the context-bound child").Purpose("Verify root injection at every graph node").NewEntity(ctx)
	if childCreationErr != nil {
		t.Fatal(childCreationErr)
	}
	child.UpdateProductName("Context root child").UpdateUnitPrice(decimal.RequireFromString("10.00"))
	parent.OrderItemList().Add(child)
	reset()
	parent, err = parent.AuditAs("Persist a graph with trusted context ownership").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	calibration := effects()
	if calibration["mutations"] != 2 || calibration["idAllocations"] != 2 || calibration["commits"] != 1 || calibration["audits"] != 2 || calibration["policies"] == 0 || calibration["rootLookups"] != 0 {
		t.Fatal("positive graph did not calibrate the real observers without root lookup", calibration)
	}
	rootID, present := lib.E.PurchaseOrder(parent).TenantId().Eval()
	childRoot, childPresent := lib.E.OrderItem(child).TenantId().Eval()
	if !present || !childPresent || rootID != active.ID || childRoot != active.ID {
		t.Fatal("the graph did not receive the active typed reference")
	}
	for _, row := range []any{parent, child} {
		if _, exists := reflect.TypeOf(row).MethodByName("UpdateTenantId"); exists {
			t.Fatal("the context-bound root has an ordinary public mutation method")
		}
	}
	physical := func() string {
		rows, err := db.Query("SELECT 'owner',id,version,tenant FROM purchase_order_data UNION ALL SELECT 'child',id,version,tenant FROM order_item_data ORDER BY 1,2")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		values := []any{}
		for rows.Next() {
			var kind string
			var id, version, tenant int64
			if err := rows.Scan(&kind, &id, &version, &tenant); err != nil {
				t.Fatal(err)
			}
			values = append(values, []any{kind, id, version, tenant})
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		bytes, err := json.Marshal(values)
		if err != nil {
			t.Fatal(err)
		}
		return string(bytes)
	}
	observations, failures := []map[string]any{}, []string{}
	reject := func(name string, context *runtime.UserContext, row *purchase_order.PurchaseOrder, rule string) {
		baseline := physical()
		reset()
		_, failure := row.AuditAs("Reject the synthetic " + name + " root violation").Save(context)
		var check *runtime.RuntimeError
		matched := false
		if errors.As(failure, &check) {
			for _, result := range check.CheckResults {
				if result.RuleID == rule && result.ModelPath() == "tenant" {
					matched = true
				}
			}
		}
		if rule == "context_root_missing" {
			matched = matched && check != nil && len(check.CheckResults) == 1 && check.CheckResults[0].EntityType == "PurchaseOrder"
		}
		counts := effects()
		zero := true
		for _, count := range counts {
			if count != 0 {
				zero = false
			}
		}
		unchanged := physical() == baseline
		passed := matched && zero && unchanged
		if !passed {
			failures = append(failures, name)
		}
		observation := map[string]any{"case": name, "passed": passed, "matchedRule": matched, "expectedRule": rule, "instancePath": "/tenant", "effects": counts, "physicalStateUnchanged": unchanged, "errorType": fmt.Sprintf("%T", failure)}
		observations = append(observations, observation)
		bytes, _ := json.Marshal(observation)
		fmt.Println("GO_CONTEXT_ROOT_CASE=" + string(bytes))
	}
	missing := contextFor(nil)
	reject("missing-root", missing, newOrder(missing), "context_root_missing")
	wrongType := contextFor(&runtime.EntityReference{Entity: "UnrelatedRoot", ID: active.ID})
	reject("type-mismatch", wrongType, newOrder(wrongType), "context_root_missing")
	zero := contextFor(&runtime.EntityReference{Entity: "Tenant", ID: 0})
	reject("zero-root", zero, newOrder(zero), "context_root_missing")
	valid := contextFor(&active)
	conflicting := newOrder(valid)
	conflicting.EntityRoot().Set(conflicting.EntityKey(), "tenant", core.ValU64(active.ID+1))
	reject("explicit-root-mismatch", valid, conflicting, "context_root_mismatch")
	load := func(ctx *runtime.UserContext) *purchase_order.PurchaseOrder {
		row, err := lib.Q.PurchaseOrders().WithIdIs(parent.Id()).Limit(1).Comment("Load the complete ownership baseline").
			Purpose("Verify update cannot move ownership").ExecuteForOne(ctx)
		if err != nil || row == nil {
			t.Fatal("saved ownership baseline missing", err)
		}
		return row
	}
	loaded := load(valid)
	loaded.EntityRoot().Set(loaded.EntityKey(), "tenant", core.ValU64(active.ID+1))
	reject("update-root-mismatch", valid, loaded, "context_root_mismatch")
	loaded = load(valid).UpdateContactPhone("5550981")
	other := contextFor(&runtime.EntityReference{Entity: "Tenant", ID: active.ID + 1})
	reject("update-wrong-context", other, loaded, "context_root_mismatch")
	frame := map[string]any{"language": "go", "passed": len(failures) == 0, "failures": failures, "cases": observations,
		"parentId": parent.Id(), "childId": child.Id(), "activeRoot": active.ID, "noPublicRootUpdater": true, "calibration": calibration, "fixObservations": checker.seen,
		"queryAuthorizationInferred": false, "newEntityBeforeReturnAccepted": false}
	bytes, _ := json.Marshal(frame)
	fmt.Println("GO_CONTEXT_ROOT_RESULT=" + string(bytes))
	if len(failures) != 0 {
		t.Fatal("context-root boundaries failed", failures)
	}
}

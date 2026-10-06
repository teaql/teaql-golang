package tracechain_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	lib "trace-chain-service-core-workspace/lib"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/platform"
)

func assertProjectedIdentity(t *testing.T, order *customer_order.CustomerOrder, id uint64, version int64) {
	t.Helper()
	if !order.IsLoaded("id") || !order.IsLoaded("version") {
		t.Fatalf("typed entity projection lost mandatory id/version: id_loaded=%t version_loaded=%t", order.IsLoaded("id"), order.IsLoaded("version"))
	}
	expression := customer_order.NewCustomerOrderExpression(order)
	actualID, idPresent := expression.Id().Eval()
	actualVersion, versionPresent := expression.Version().Eval()
	if !idPresent || actualID != id || !versionPresent || actualVersion != version {
		t.Fatalf("projected identity/version mismatch: id=%d version=%d, want %d/%d", actualID, actualVersion, id, version)
	}
	if original, present := order.EntityRoot().OriginalVersion(order.EntityKey()); !present || original != version || order.EntityRoot().IsNew(order.EntityKey()) {
		t.Fatal("projection did not retain an existing entity's optimistic version")
	}
}

func assertProjectedReference(t *testing.T, order *customer_order.CustomerOrder) {
	t.Helper()
	related, loaded := order.RelationEntity("platformEntity")
	reference, typed := related.(*platform.Platform)
	if !loaded || !typed || reference == nil || !reference.IsLoaded("id") || !reference.IsLoaded("version") {
		t.Fatal("narrow forward relation projection lost the reference id/version")
	}
	expression := platform.NewPlatformExpression(reference)
	id, idPresent := expression.Id().Eval()
	version, versionPresent := expression.Version().Eval()
	name, namePresent := expression.Name().Eval()
	if !idPresent || id != 1 || !versionPresent || version < 1 || !namePresent || name == "" {
		t.Fatal("projected forward reference did not expose real identity/version/name through E")
	}
	if reference.EntityRoot() == order.EntityRoot() || len(reference.EntityRoot().Changes()) != 0 {
		t.Fatal("projecting a reference adopted mutable ownership")
	}
}

func TestGeneratedRelationSelectionPreservesFullEntityProjection(t *testing.T) {
	e := openEnvironment(t)
	order := newOrder(t, e, "full projection fixture")
	order.OrderItemList().Add(newItem(e, "projected child"))
	if _, err := order.AuditAs("prepare full projection fixture").Save(e.context); err != nil {
		t.Fatal(err)
	}
	id, _ := customer_order.NewCustomerOrderExpression(order).Id().Eval()
	version, _ := customer_order.NewCustomerOrderExpression(order).Version().Eval()
	number, _ := customer_order.NewCustomerOrderExpression(order).OrderNumber().Eval()
	e.reset()
	request := lib.Q.CustomerOrders().WithIdIs(id).Limit(1).
		SelectPlatformWith(lib.Q.Platforms().SelectName().Limit(1)).
		SelectOrderItemListWith(lib.Q.OrderItems().SelectName().Limit(2))
	if len(request.Query.Projection) != 0 {
		t.Fatalf("relation selection narrowed the default select-all projection: %v", request.Query.Projection)
	}
	rows, err := request.Comment("load complete root with narrow relations").
		Purpose("verify relation loading preserves identity and all root fields").ExecuteForList(e.context)
	if err != nil || len(rows.Data) != 1 {
		t.Fatalf("relation query failed: rows=%v error=%v", rows, err)
	}
	loaded := rows.Data[0]
	assertProjectedIdentity(t, loaded, id, version)
	assertProjectedReference(t, loaded)
	expression := customer_order.NewCustomerOrderExpression(loaded)
	actualNumber, numberPresent := expression.OrderNumber().Eval()
	description, descriptionPresent := expression.Description().Eval()
	childID, childPresent := expression.OrderItemList().First().Id().Eval()
	if !numberPresent || actualNumber != number || !descriptionPresent || description != "full projection fixture" || !childPresent || childID == 0 {
		t.Fatal("default root fields or narrow reverse-child identity were lost")
	}
	child := loaded.OrderItemList().Items()[0]
	if !child.IsLoaded("version") || !child.IsLoaded("customer_order_id") {
		t.Fatal("narrow reverse child lost optimistic version or parent foreign key")
	}
	if child.EntityRoot() != loaded.EntityRoot() {
		t.Fatal("projected owned child did not retain its root ledger")
	}
	e.reset()
	loaded.UpdateDescription(fmt.Sprintf("full projection after version %d", version))
	if _, err := loaded.AuditAs("update completely projected root").Save(e.context); err != nil {
		t.Fatal(err)
	}
	requests, events := e.observer.snapshot(), e.sink.snapshot()
	if len(requests) != 1 || len(events) != 1 {
		t.Fatalf("readonly relations generated writes: commands=%d audits=%d", len(requests), len(events))
	}
	update, isUpdate := requests[0].(*data_service.UpdateMutation)
	if !isUpdate || update.Cmd.Entity != "Customer Order" || update.Cmd.ExpectedVersion == nil || *update.Cmd.ExpectedVersion != version {
		t.Fatal("full projection save lost the real root optimistic version")
	}
	assertLineage(t, update.TraceChain(), []*core.TraceNode{reasonNode("Customer Order", id, "update completely projected root")})
	assertLineage(t, events[0].TraceChain, update.TraceChain())
	reloaded, err := lib.Q.CustomerOrders().WithIdIs(id).SelectDescription().Limit(1).
		Comment("reload full projection update").Purpose("verify optimistic version and persisted value").ExecuteForOne(e.context)
	if err != nil || reloaded == nil {
		t.Fatalf("updated root could not be reloaded: %v", err)
	}
	assertProjectedIdentity(t, reloaded, id, version+1)
	actualDescription, _ := customer_order.NewCustomerOrderExpression(reloaded).Description().Eval()
	if actualDescription != fmt.Sprintf("full projection after version %d", version) {
		t.Fatal("full projection update did not reach SQLite")
	}
	t.Log("ENTITY PROJECTION FULL PASSED: complete root, narrow forward/reverse identities, one audited optimistic update")
}

func TestGeneratedNarrowEntityProjectionProtectsIdentityWithoutPermittingPartialMutation(t *testing.T) {
	for _, mode := range []string{"list", "minimal-list", "page", "stream"} {
		t.Run(mode, func(t *testing.T) {
			e := openEnvironment(t)
			order := newOrder(t, e, "narrow projection fixture")
			if _, err := order.AuditAs("prepare narrow projection fixture").Save(e.context); err != nil {
				t.Fatal(err)
			}
			id, _ := customer_order.NewCustomerOrderExpression(order).Id().Eval()
			version, _ := customer_order.NewCustomerOrderExpression(order).Version().Eval()
			e.reset()
			request := lib.Q.CustomerOrders()
			if mode == "minimal-list" {
				request = lib.Q.CustomerOrdersMinimal()
			}
			request.WithIdIs(id).SelectDescription().OrderByIdAsc().Limit(1)
			if mode != "stream" {
				request.SelectPlatformWith(lib.Q.Platforms().SelectName().Limit(1))
			}
			projectionBefore := append([]string(nil), request.Query.Projection...)
			comment := "load narrow entity through " + mode
			executable := request.Comment(comment).Purpose("verify mandatory identity and fail-closed partial mutation")
			var rows *core.SmartList[*customer_order.CustomerOrder]
			var err error
			switch mode {
			case "page":
				rows, err = executable.ExecuteForPage(e.context, 0, 1)
				if err == nil && (rows.TotalCount == nil || *rows.TotalCount != 1) {
					t.Fatal("page count does not match the bounded authorized filter")
				}
			case "stream":
				rows = core.NewSmartList([]*customer_order.CustomerOrder{})
				err = executable.ExecuteForStream(e.context, 1, func(row *customer_order.CustomerOrder) error {
					rows.Data = append(rows.Data, row)
					return nil
				})
			default:
				rows, err = executable.ExecuteForList(e.context)
			}
			if err != nil || len(rows.Data) != 1 {
				t.Fatalf("narrow %s failed: rows=%v error=%v", mode, rows, err)
			}
			if !reflect.DeepEqual(request.Query.Projection, projectionBefore) {
				t.Fatal("execution mutated the caller-owned projection builder")
			}
			loaded := rows.Data[0]
			assertProjectedIdentity(t, loaded, id, version)
			if mode != "stream" {
				assertProjectedReference(t, loaded)
			}
			if loaded.IsLoaded("order_number") {
				t.Fatal("mandatory identity protection silently widened an ordinary field")
			}
			statements := e.sqlEvidence.Snapshot()
			if len(statements) == 0 {
				t.Fatal("no actual SQL evidence for the projected entity")
			}
			countQueries := 0
			for _, statement := range statements {
				if statement.Comment == nil || *statement.Comment != comment || len(statement.TraceChain) < 4 || statement.TraceChain[0].Name != "Customer Order" {
					t.Fatal("projection preparation lost request-owned query intent/path")
				}
				if strings.Contains(statement.ParameterizedSQL, "__teaql_total") {
					countQueries++
					selectPart := strings.Split(statement.ParameterizedSQL, " FROM ")[0]
					if strings.Contains(selectPart, "version") || strings.Contains(selectPart, "description") || strings.Contains(selectPart, ",") {
						t.Fatal("identity protection contaminated the exact-count projection")
					}
				}
			}
			if (mode == "page" && countQueries != 1) || (mode != "page" && countQueries != 0) {
				t.Fatal("unexpected exact-count execution")
			}
			e.reset()
			beginsBefore := e.observer.begins
			loaded.UpdateDescription("must not persist an incompletely loaded entity")
			_, err = loaded.AuditAs("reject partial entity mutation").Save(e.context)
			var checkError *runtime.RuntimeError
			if !errors.As(err, &checkError) || checkError.Type != "Check" || !strings.Contains(err.Error(), "fully loaded entity") {
				t.Fatalf("narrow identity protection bypassed the full-load mutation guard: %v", err)
			}
			if len(e.observer.snapshot()) != 0 || len(e.sink.snapshot()) != 0 || e.observer.begins != beginsBefore {
				t.Fatal("partial mutation reached the transaction/provider/committed audit")
			}
			t.Logf("ENTITY PROJECTION NARROW PASSED: mode=%s identity=true partial_save=rejected count_queries=%d sql_paths=%d", mode, countQueries, len(statements))
		})
	}
}

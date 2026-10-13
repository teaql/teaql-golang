package main

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	_ "github.com/teaql/teaql-golang/telemetry/standard"
	"os"
	"path/filepath"
	lib "request-tree-order-service-core-workspace/lib"
	"request-tree-order-service-core-workspace/lib/customer"
	"request-tree-order-service-core-workspace/lib/customer_order"
	"request-tree-order-service-core-workspace/lib/order_item"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func require(ok bool, reason string) {
	if !ok {
		panic(reason)
	}
}
func value[T any](result T, present bool) T { require(present, "typed E value missing"); return result }

type audits struct{ count int }

func (a *audits) OnSafeEvent(_ *runtime.UserContext, _ *runtime.SafeAuditEvent) error {
	a.count++
	return nil
}

func main() {
	database := os.Getenv("TEAQL_STANDARD_ORDER_DB")
	if database == "" {
		directory, err := os.MkdirTemp("", "teaql-go-standard-order-")
		must(err)
		database = filepath.Join(directory, "orders.sqlite")
	}
	must(os.Setenv("REQUEST_TREE_ORDER_SERVICE_CORE_DATABASE_URL", database))
	context, err := lib.ServiceRuntimeFromEnv()
	must(err)
	defer context.GetResource("db").(*sql.DB).Close()
	defer context.RuntimeTelemetry().Shutdown(stdcontext.Background())
	context.GetResource("db").(*sql.DB).SetMaxOpenConns(1)
	// Shared structural fixtures declare ordinary fields public. Make this
	// demo's synthetic customer-facing number private via the existing seam.
	descriptor := context.Metadata.Entity("customer_order_data")
	require(descriptor != nil, "standard order metadata missing")
	descriptor.AuditMaskFields([]string{"order_number"})
	ordinary := runtime.NewSQLExecutionEvidenceStore()
	audit := &audits{}
	context.WithDiagnosticSQLLogSink(ordinary).WithAppAuditEventSink(audit).DisableSqlLog()
	must(lib.EnsureSchema(context))
	first := audit.count
	require(first == 0 || first == 4, "first audited platform/constants seed")
	must(lib.EnsureSchema(context))
	require(audit.count == first, "repeated seed")
	secret := fmt.Sprintf("ORDER-PRIVATE-%d", time.Now().UnixNano())
	account := customer.NewCustomer().UpdatePlatformId(1).UpdateName("North Star Retail").UpdateVersion(1)
	account, err = account.AuditAs("seed standard customer").Save(context)
	must(err)
	for index := 1; index <= 4; index++ {
		state := uint64(1002)
		if index == 1 {
			state = 1001
		}
		number := fmt.Sprintf("%s-%03d", secret, index)
		amount := int64(index * 100)
		if index == 4 {
			number = "OTHER-004"
			amount = 900
		}
		order := customer_order.NewCustomerOrder().UpdateCustomerId(account.Id()).UpdateStatusId(state).
			UpdateOrderNumber(number).UpdateOrderDate(time.Date(2026, 10, index, 0, 0, 0, 0, time.UTC)).
			UpdateTotalAmount(decimal.NewFromInt(amount)).UpdateVersion(1)
		order, err = order.AuditAs("seed standard order").Save(context)
		must(err)
		for item := 1; item <= 5; item++ {
			line := order_item.NewOrderItem().UpdateCustomerOrderId(order.Id()).UpdateProductName(fmt.Sprintf("Box %d", item)).
				UpdateQuantity(int64(item)).UpdateUnitPrice(decimal.NewFromInt(20)).UpdateVersion(1)
			_, err = line.AuditAs("seed standard item").Save(context)
			must(err)
		}
	}
	for repetition := 1; repetition <= 4; repetition++ {
		counted := repetition <= 2
		request := lib.Q.CustomerOrdersMinimal().WithOrderNumberContaining(secret).
			FacetByStatusAs("statusCounts", lib.Q.OrderStatusesMinimal().SelectName().SelectColor().
				CountAs("orderCount").SumTotalAmountOfCustomerOrdersAs("orderTotalAmount", lib.Q.CustomerOrdersMinimal()).
				OrderByDisplayOrderAsc().Comment("load status facets")).
			SelectOrderNumber().SelectOrderDate().SelectStatusWith(lib.Q.OrderStatusesMinimal().SelectName().SelectColor().Comment("load status")).
			SelectOrderItemListWith(lib.Q.OrderItemsMinimal().SelectProductName().SelectQuantity().OrderByIdAsc().Limit(3).Comment("load item previews")).
			CountOrderItemsAs("itemCount").OrderByOrderDateDesc().OrderByIdDesc().Offset(1).Limit(1).
			Comment("load order list")
		if counted {
			request.CountAs("totalOrders")
		}
		page, err := request.Purpose("render order list page").ExecuteForList(context)
		must(err)
		if counted {
			require(page.TotalCount != nil && *page.TotalCount == 3 && value(page.Aggregations["totalOrders"].TryU64()) == 3, "exact named filtered count")
		} else {
			require(page.TotalCount == nil && len(page.Aggregations) == 0, "ordinary list must not implicitly count")
		}
		require(len(page.Data) == 1, "bounded page")
		row := page.Data[0]
		require(value(lib.E.CustomerOrder(row).OrderNumber().Eval()) == secret+"-002", "typed selected order")
		require(value(lib.E.CustomerOrder(row).Status().Name().Eval()) == "Paid", "typed loaded forward status")
		require(value(lib.E.CustomerOrder(row).OrderItemList().Size().Eval()) == 3, "bounded reverse preview")
		require(value(lib.E.CustomerOrder(row).OrderItemList().First().ProductName().Eval()) == "Box 1", "typed sorted preview")
		count, present := row.Base().GetDynamic("itemCount")
		require(present && value(count.TryU64()) == 5, "full child count")
		facets := page.Facets["statusCounts"]
		require(facets != nil && len(facets.Data) == 3, "facets carried by list")
		for index, facet := range facets.Data {
			require(value(facet["id"].TryU64()) == uint64(1001+index), "facet order")
			require(value(facet["orderCount"].TryU64()) == []uint64{1, 2, 0}[index], "filtered facet count")
			if index < 2 {
				require(value(facet["orderTotalAmount"].TryDecimal()).Equal(decimal.NewFromInt([]int64{100, 500}[index])), "filtered facet sum")
			} else {
				require(facet["orderTotalAmount"].IsNull(), "empty SUM stays null")
			}
		}
	}
	require(audit.count == first+25 && len(ordinary.Snapshot()) == 0, "independent audit and ordinary SQL logging")
	must(context.RuntimeTelemetry().Flush(stdcontext.Background()))
	fmt.Printf("GO_STANDARD_ORDER PASSED queries=4 counted=2 uncounted=2 total=3 preview=3 itemCount=5 facets=3 first_seed_audits=%d repeat_seed_audits=0 audits=%d ordinary_sql_logs=0 database=%s\n", first, audit.count, database)
}

package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	lib "order-management-service-core-workspace/lib"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/order_search_preset"
)

type appAudit struct{}

func (appAudit) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	fmt.Printf("[audit/app] kind=%v entity=%s; safe_fields=%d\n", event.Kind, event.Entity, len(event.Fields))
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	db := os.Getenv("TEAQL_ORDER_MANAGEMENT_DB")
	if db == "" {
		db = filepath.Join("..", ".local", "order.db")
	}
	if _, err := os.Stat(db); os.IsNotExist(err) {
		fmt.Printf("[database] %s was not found; TeaQL will create it\n", db)
	}
	must(os.MkdirAll(filepath.Dir(db), 0o755))
	must(os.Setenv("ORDER_MANAGEMENT_SERVICE_CORE_DATABASE_URL", db))
	context, err := lib.ServiceRuntimeFromEnv()
	must(err)
	defer context.GetResource("db").(*sql.DB).Close()
	context.WithAppAuditEventSink(appAudit{})
	must(lib.EnsureSchema(context))
	must(lib.EnsureSchema(context))
	fmt.Println("[schema] ensured 7 generated entity tables")

	platform, err := lib.Q.CommercePlatforms().WithIdIs(1).
		Comment("Load the generated commerce root").
		Purpose("Use the model-owned root for quick-start data").ExecuteForOne(context)
	must(err)
	if platform == nil {
		panic("generated commerce root was not provisioned")
	}
	platformID := platform.Id()
	orders, err := lib.Q.CustomerOrders().WithOrderNumberIs("WEB-2026-001").
		Comment("Check whether deterministic quick-start data exists").
		Purpose("Keep example mutations idempotent").ExecuteForList(context)
	must(err)
	if len(orders.Data) == 0 {
		now := time.Date(2026, 8, 13, 9, 0, 0, 0, time.UTC)
		buyer := customer.NewCustomer().
			UpdateName("Acme Retail").
			UpdateEmail("masked-in-quick-start").
			UpdateCommercePlatformId(platformID).
			UpdateCreateTime(now).
			UpdateUpdateTime(now)
		if _, err := buyer.AuditAs("Create quick-start customer").Save(context); err != nil {
			panic(err)
		}
		orderDate := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
		order := customer_order.NewCustomerOrder().
			UpdateOrderNumber("WEB-2026-001").
			UpdateOrderDate(orderDate).
			UpdateTotalAmount(decimal.RequireFromString("129.95")).
			UpdateCustomerId(buyer.Id()).
			UpdateCommercePlatformId(platformID)
		// The generated constant transition uses the stable status id from the model.
		order.UpdateStatusToPending()
		if _, err := order.AuditAs("Create deterministic quick-start order").Save(context); err != nil {
			panic(err)
		}
		fmt.Println("[seed] inserted deterministic customer and order; reused generated root and status")
	} else {
		fmt.Println("[seed] deterministic data already exists; no duplicate rows added")
	}

	listedOrders, err := lib.Q.CustomerOrders().WithOrderNumberContaining("WEB-").OrderByIdAsc().
		Comment("List WEB orders for the terminal quick start").
		Purpose("Show the operator a deterministic order list").ExecuteForList(context)
	must(err)
	fmt.Printf("[query] matched %d order(s)\n", len(listedOrders.Data))
	for _, order := range listedOrders.Data {
		fmt.Printf("  %s  %s  %s\n", order.OrderNumber(), order.OrderDate().Format("2006-01-02"), order.TotalAmount())
	}

	presets, err := lib.Q.OrderSearchPresets().WithRequestIdIs("quick-start-pending-orders").
		Comment("Check idempotent quick-start preset").Purpose("Persist the operator's reusable search").ExecuteForList(context)
	must(err)
	if len(presets.Data) == 0 {
		preset := order_search_preset.NewOrderSearchPreset().
			UpdateName("Pending web orders").
			UpdateFilterJson(`{"order_number":"WEB-"}`).
			UpdateRequestId("quick-start-pending-orders").
			UpdateOwnerUserId("quick-start-user").
			UpdateCommercePlatformId(platformID)
		if _, err := preset.AuditAs("Save idempotent quick-start search preset").Save(context); err != nil {
			panic(err)
		}
		fmt.Printf("[mutation] saved preset #%d\n", preset.Id())
	} else {
		fmt.Printf("[mutation] preset #%d already exists\n", presets.Data[0].Id())
	}
}

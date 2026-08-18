package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	lib "order-management-service-core-workspace/lib"
	"order-management-service-core-workspace/lib/commerce_platform"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/order_search_preset"
	"order-management-service-core-workspace/lib/order_status"
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
	db := filepath.Join("..", ".local", "order.db")
	if _, err := os.Stat(db); os.IsNotExist(err) {
		fmt.Printf("[database] %s was not found; TeaQL will create it\n", db)
	}
	must(os.MkdirAll(filepath.Dir(db), 0o755))
	must(os.Setenv("ORDER_MANAGEMENT_SERVICE_CORE_DATABASE_URL", db))
	context, err := lib.ServiceRuntimeFromEnv()
	must(err)
	context.WithAppAuditEventSink(appAudit{})
	fmt.Println("[schema] ensured 7 generated entity tables")

	platforms, err := lib.Q.CommercePlatforms().WithNameIs("Northwind Demo").
		Comment("Check whether deterministic quick-start data exists").
		Purpose("Initialize the local order-management example").ExecuteForList(context)
	must(err)
	var platformID uint64
	if len(platforms.Data) == 0 {
		now := time.Date(2026, 8, 13, 9, 0, 0, 0, time.UTC)
		platform := commerce_platform.NewCommercePlatform().
			UpdateName("Northwind Demo").
			UpdateCreateTime(now).
			UpdateUpdateTime(now)
		must(platform.AuditAs("Create quick-start commerce platform").Save(context))
		platformID = platform.Id()
		buyer := customer.NewCustomer().
			UpdateName("Acme Retail").
			UpdateEmail("masked-in-quick-start").
			UpdateCommercePlatformId(platform.Id()).
			UpdateCreateTime(now).
			UpdateUpdateTime(now)
		must(buyer.AuditAs("Create masked quick-start customer").Save(context))
		pending := order_status.NewOrderStatus().
			UpdateName("Pending").
			UpdateCode("PENDING").
			UpdateColor("#F97316").
			UpdateDisplayOrder(decimal.NewFromInt(10)).
			UpdateCommercePlatformId(platform.Id())
		must(pending.AuditAs("Create quick-start pending status").Save(context))
		orderDate := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
		order := customer_order.NewCustomerOrder().
			UpdateOrderNumber("WEB-2026-001").
			UpdateOrderDate(orderDate).
			UpdateTotalAmount(decimal.RequireFromString("129.95")).
			UpdateCustomerId(buyer.Id()).
			UpdateCommercePlatformId(platform.Id())
		// The generated constant transition uses the stable status id from the model.
		order.UpdateStatusToPending()
		must(order.AuditAs("Create deterministic quick-start order").Save(context))
		fmt.Println("[seed] inserted deterministic platform, customer, status, and order")
	} else {
		platformID = platforms.Data[0].Id()
		fmt.Println("[seed] deterministic data already exists; no duplicate rows added")
	}

	orders, err := lib.Q.CustomerOrders().WithOrderNumberContaining("WEB-").OrderByIdAsc().
		Comment("List WEB orders for the terminal quick start").
		Purpose("Show the operator a deterministic order list").ExecuteForList(context)
	must(err)
	fmt.Printf("[query] matched %d order(s)\n", len(orders.Data))
	for _, order := range orders.Data {
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
		must(preset.AuditAs("Save idempotent quick-start search preset").Save(context))
		fmt.Printf("[mutation] saved preset #%d\n", preset.Id())
	} else {
		fmt.Printf("[mutation] preset #%d already exists\n", presets.Data[0].Id())
	}
}

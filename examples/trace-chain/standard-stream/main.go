package main

import (
	stdcontext "context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/runtime"
	_ "github.com/teaql/teaql-golang/telemetry/standard"
	lib "request-tree-order-service-core-workspace/lib"
	"request-tree-order-service-core-workspace/lib/customer"
	"request-tree-order-service-core-workspace/lib/customer_order"
	"request-tree-order-service-core-workspace/lib/order_item"
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
func value[T any](v T, ok bool) T { require(ok, "typed E value missing"); return v }

type audits struct{ count int }

func (a *audits) OnSafeEvent(_ *runtime.UserContext, _ *runtime.SafeAuditEvent) error {
	a.count++
	return nil
}

func main() {
	database := os.Getenv("TEAQL_STANDARD_STREAM_DB")
	if database == "" {
		dir, err := os.MkdirTemp("", "teaql-go-stream-")
		must(err)
		database = filepath.Join(dir, "streams.sqlite")
	}
	must(os.Setenv("REQUEST_TREE_ORDER_SERVICE_CORE_DATABASE_URL", database))
	context, err := lib.ServiceRuntimeFromEnv()
	must(err)
	defer context.GetResource("db").(*sql.DB).Close()
	defer context.RuntimeTelemetry().Shutdown(stdcontext.Background())
	// The independent query inside a callback needs its own read connection.
	context.GetResource("db").(*sql.DB).SetMaxOpenConns(4)
	context.Metadata.Entity("order_item_data").AuditMaskFields([]string{"product_name"})
	context.Metadata.Entity("customer_order_data").AuditMaskFields([]string{"order_number"})
	ordinary := runtime.NewSQLExecutionEvidenceStore()
	sink := &audits{}
	context.WithDiagnosticSQLLogSink(ordinary).WithAppAuditEventSink(sink).DisableSqlLog()
	must(lib.EnsureSchema(context))
	first := sink.count
	require(first == 0 || first == 4, "bootstrap audit")
	must(lib.EnsureSchema(context))
	require(first == sink.count, "repeated seed")
	account, err := customer.NewCustomer().UpdatePlatformId(1).UpdateName("Stream Customer").UpdateVersion(1).AuditAs("seed stream customer").Save(context)
	must(err)
	secret := fmt.Sprintf("STREAM-PRIVATE-%d", time.Now().UnixNano())
	order := customer_order.NewCustomerOrder().UpdateCustomerId(account.Id()).UpdateStatusId(1002).
		UpdateOrderNumber(secret).UpdateOrderDate(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)).
		UpdateTotalAmount(decimal.NewFromInt(3)).UpdateVersion(1)
	names := []string{}
	for index := 1; index <= 3; index++ {
		name := fmt.Sprintf("%s-%d", secret, index)
		names = append(names, name)
		order.OrderItemList().Add(order_item.NewOrderItem().UpdateProductName(name).UpdateQuantity(1).
			UpdateUnitPrice(decimal.NewFromInt(1)).UpdateVersion(1))
	}
	_, err = order.AuditAs("seed standard stream graph").Save(context)
	must(err)
	require(sink.count == first+5, "independent committed audit")
	unsupportedCallbacks := 0
	err = lib.Q.OrderItems().WithProductNameIn(names).CountAs("items").Limit(3).
		Comment("reject unsupported counted stream").Purpose("verify unsupported counted stream").
		ExecuteForStream(context, 2, func(*order_item.OrderItem) error { unsupportedCallbacks++; return nil })
	require(err != nil && unsupportedCallbacks == 0, "stream silently ignored explicit count")
	err = lib.Q.CustomerOrders().WithIdIs(order.Id()).
		SelectCustomerWith(lib.Q.CustomersMinimal().SelectName().Comment("load selected customer")).Limit(1).
		Comment("reject unsupported relation stream").Purpose("verify unsupported relation stream").
		ExecuteForStream(context, 2, func(*customer_order.CustomerOrder) error { unsupportedCallbacks++; return nil })
	require(err != nil && unsupportedCallbacks == 0, "stream silently ignored selected relation")
	err = lib.Q.CustomerOrders().WithIdIs(order.Id()).
		FacetByStatusAs("statusCounts", lib.Q.OrderStatusesMinimal().SelectName().CountAs("orderCount").OrderByIdAsc().Limit(3).Comment("load status facets")).Limit(1).
		Comment("reject unsupported faceted stream").Purpose("verify unsupported faceted stream").
		ExecuteForStream(context, 2, func(*customer_order.CustomerOrder) error { unsupportedCallbacks++; return nil })
	require(err != nil && unsupportedCallbacks == 0, "stream silently ignored requested facets")
	privateError := errors.New("private callback failure " + secret)
	for _, kind := range []string{"complete", "early-stop", "token-cancel", "callback-error", "callback-panic"} {
		request := lib.Q.OrderItems().WithProductNameIn(names).OrderByIdAsc().Limit(3).
			Comment("load standard " + kind + " stream").Purpose("verify standard " + kind + " stream")
		original := context.Context
		cancelContext, cancel := stdcontext.WithCancel(original)
		if kind == "token-cancel" {
			context.Context = cancelContext
		}
		delivered := 0
		panicked := false
		func() {
			defer func() {
				if p := recover(); p != nil {
					require(p == privateError, "original callback panic changed")
					panicked = true
				}
			}()
			err = request.ExecuteForStream(context, 2, func(item *order_item.OrderItem) error {
				delivered++
				require(value(lib.E.OrderItem(item).ProductName().Eval()) == names[delivered-1], "typed stream product")
				require(value(lib.E.OrderItem(item).Quantity().Eval()) == 1, "typed stream quantity")
				if kind == "complete" && delivered == 1 {
					rows, queryErr := lib.Q.OrderItems().WithProductNameIn(names).Limit(1).
						Comment("load independent stream query").Purpose("verify stream activation isolation").ExecuteForList(context)
					must(queryErr)
					require(len(rows.Data) == 1, "independent query result")
				}
				switch kind {
				case "early-stop":
					return runtime.ErrStopStream
				case "token-cancel":
					cancel()
					return nil
				case "callback-error":
					return privateError
				case "callback-panic":
					panic(privateError)
				}
				return nil
			})
		}()
		context.Context = original
		cancel()
		if kind == "complete" {
			must(err)
			require(delivered == 3 && !panicked, "complete stream")
		} else {
			require(delivered == 1, "exclude prefetched unconsumed rows")
		}
		switch kind {
		case "early-stop":
			require(errors.Is(err, runtime.ErrStopStream), "original stop sentinel")
		case "token-cancel":
			require(errors.Is(err, stdcontext.Canceled), "original cancellation")
		case "callback-error":
			require(err == privateError, "original callback error")
		case "callback-panic":
			require(panicked, "original callback panic")
		}
	}
	require(len(ordinary.Snapshot()) == 0 && sink.count == first+5, "diagnostics/audit independence")
	must(context.RuntimeTelemetry().Flush(stdcontext.Background()))
	fmt.Printf("GO_STANDARD_STREAM PASSED complete=3 early=1 cancelled=1 callback_error=1 callback_panic=1 first_seed_audits=%d repeat_seed_audits=0 audits=%d ordinary_sql_logs=0 database=%s\n", first, sink.count, database)
}

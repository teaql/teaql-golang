package main

import (
	stdcontext "context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"github.com/teaql/teaql-golang/telemetry/standard"
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
func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func value[T any](v T, ok bool) T { require(ok, "typed E value missing"); return v }

type audits struct {
	count  int
	traces [][]*core.TraceNode
}

func (a *audits) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	a.count++
	a.traces = append(a.traces, core.CloneTraceNodes(event.TraceChain))
	return nil
}

func (a *audits) graphLineage(root string) {
	traces := a.traces[len(a.traces)-3:]
	require(len(traces[0]) == 1 && traces[0][0].Comment == root, "root audit lineage")
	require(len(traces[1]) == 2 && traces[1][0].Comment == root && traces[1][1].Comment == "approve explicit item", "explicit child lineage")
	require(len(traces[2]) == 1 && traces[2][0].Comment == root, "inherited child lineage")
}

// Transparent test-only commit fault: all writes and authoritative readbacks
// execute on the real SQLite transaction. No fake results or audit records.
type commitFault struct {
	data_service.TransactionExecutor
	failure error
}

func (f *commitFault) Query(ctx stdcontext.Context, q *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return f.TransactionExecutor.(data_service.QueryExecutor).Query(ctx, q)
}
func (f *commitFault) Begin(ctx stdcontext.Context) (data_service.Transaction, error) {
	tx, err := f.TransactionExecutor.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &faultTransaction{Transaction: tx, failure: f.failure}, nil
}

type faultTransaction struct {
	data_service.Transaction
	failure error
}

func (f *faultTransaction) Commit(stdcontext.Context) error { return f.failure }
func (f *faultTransaction) GenerateId(name string) (uint64, error) {
	return f.Transaction.(interface{ GenerateId(string) (uint64, error) }).GenerateId(name)
}
func (f *faultTransaction) EnsureIdFloor(ctx stdcontext.Context, name string, floor uint64) error {
	return f.Transaction.(interface {
		EnsureIdFloor(stdcontext.Context, string, uint64) error
	}).EnsureIdFloor(ctx, name, floor)
}

func graph(account uint64, number string) *customer_order.CustomerOrder {
	order := customer_order.NewCustomerOrder().UpdateCustomerId(account).UpdateStatusId(1002).
		UpdateOrderNumber(number).UpdateOrderDate(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)).
		UpdateTotalAmount(decimal.NewFromInt(100)).UpdateVersion(1)
	first := order_item.NewOrderItem().UpdateProductName("Explicit Item").UpdateQuantity(1).
		UpdateUnitPrice(decimal.NewFromInt(40)).UpdateVersion(1).AuditAs("approve explicit item")
	second := order_item.NewOrderItem().UpdateProductName("Inherited Item").UpdateQuantity(1).
		UpdateUnitPrice(decimal.NewFromInt(60)).UpdateVersion(1)
	order.OrderItemList().Add(first)
	order.OrderItemList().Add(second)
	return order
}

func main() {
	pressure := flag.Bool("queue-pressure", false, "exercise production telemetry queue capacity with generated Q/E")
	postShutdown := flag.Bool("post-shutdown", false, "verify business reads after owned telemetry shutdown")
	shutdownBacklog := flag.Bool("shutdown-backlog", false, "verify finite shutdown against a blocked receiver with queued records")
	flag.Parse()
	database := os.Getenv("TEAQL_STANDARD_GRAPH_DB")
	if database == "" {
		dir, err := os.MkdirTemp("", "teaql-go-graph-")
		must(err)
		database = filepath.Join(dir, "graphs.sqlite")
	}
	must(os.Setenv("REQUEST_TREE_ORDER_SERVICE_CORE_DATABASE_URL", database))
	context, err := lib.ServiceRuntimeFromEnv()
	must(err)
	defer context.GetResource("db").(*sql.DB).Close()
	defer context.RuntimeTelemetry().Shutdown(stdcontext.Background())
	context.GetResource("db").(*sql.DB).SetMaxOpenConns(1)
	context.Metadata.Entity("customer_order_data").AuditMaskFields([]string{"order_number"})
	sink := &audits{}
	ordinary := runtime.NewSQLExecutionEvidenceStore()
	context.WithAppAuditEventSink(sink).WithDiagnosticSQLLogSink(ordinary).DisableSqlLog()
	must(lib.EnsureSchema(context))
	first := sink.count
	require(first == 0 || first == 4, "bootstrap audit")
	must(lib.EnsureSchema(context))
	require(sink.count == first, "repeat seed")
	account, err := customer.NewCustomer().UpdatePlatformId(1).UpdateName("Graph Customer").UpdateVersion(1).
		AuditAs("seed graph customer").Save(context)
	must(err)
	secret := fmt.Sprintf("GRAPH-PRIVATE-%d", time.Now().UnixNano())
	order, err := graph(account.Id(), secret+"-SUCCESS").AuditAs("save standard graph").Save(context)
	must(err)
	require(sink.count == first+4, "three committed graph audit records")
	sink.graphLineage("save standard graph")
	rows, err := lib.Q.CustomerOrders().WithIdIs(order.Id()).
		SelectOrderItemListWith(lib.Q.OrderItems().OrderByIdAsc().Limit(3).Comment("load graph items")).
		Limit(1).Comment("reload standard graph").Purpose("verify standard graph Q and E").ExecuteForList(context)
	must(err)
	require(len(rows.Data) == 1 && value(lib.E.CustomerOrder(rows.Data[0]).OrderItemList().Size().Eval()) == 2, "saved graph Q/E")
	require(value(lib.E.CustomerOrder(rows.Data[0]).OrderNumber().Eval()) == secret+"-SUCCESS", "saved root value")
	original := context.GetResource("dataService")
	injected := errors.New("private synthetic commit transport failure")
	context.InsertResource("dataService", &commitFault{TransactionExecutor: original.(data_service.TransactionExecutor), failure: injected})
	_, err = graph(account.Id(), secret+"-ROLLBACK").AuditAs("rollback standard graph").Save(context)
	require(errors.Is(err, injected), "original commit failure")
	require(sink.count == first+4, "rolled-back graph cannot emit committed audit")
	context.InsertResource("dataService", original)
	missing, err := lib.Q.CustomerOrders().WithOrderNumberIs(secret + "-ROLLBACK").Limit(1).
		Comment("inspect graph rollback").Purpose("verify atomic graph failure").ExecuteForList(context)
	must(err)
	require(len(missing.Data) == 0, "rollback must remove all written rows")
	_, err = graph(account.Id(), secret+"-RECOVERY").AuditAs("recover standard graph").Save(context)
	must(err)
	require(sink.count == first+7 && len(ordinary.Snapshot()) == 0, "recovery audit and independent ordinary logging")
	sink.graphLineage("recover standard graph")
	if *pressure {
		for i := 0; i < 1500; i++ {
			rows, readErr := lib.Q.CustomerOrdersMinimal().WithIdIs(order.Id()).SelectOrderNumber().Limit(1).
				Comment("load queue-pressure order").Purpose("verify bounded telemetry delivery").ExecuteForList(context)
			must(readErr)
			require(len(rows.Data) == 1 && value(lib.E.CustomerOrder(rows.Data[0]).OrderNumber().Eval()) == secret+"-SUCCESS", "pressure Q/E changed")
		}
		require(sink.count == first+7, "pressure changed independent audit")
		fmt.Printf("GO_QUEUE_PRESSURE READY queries=1500 audits=%d\n", sink.count)
	}
	if *shutdownBacklog {
		require(*pressure && *postShutdown, "shutdown-backlog requires queue-pressure and post-shutdown")
		telemetry, ok := context.RuntimeTelemetry().(*standard.Telemetry)
		require(ok, "backlog mode requires enabled standard telemetry")
		fmt.Println("GO_SHUTDOWN_BACKLOG READY")
		started := time.Now()
		must(telemetry.Shutdown(stdcontext.Background()))
		health := telemetry.ExportHealth()
		for _, signal := range []string{"traces", "logs"} {
			q := health.Queues[signal]
			require(q.Outstanding == 0 && q.Exporting == 0 && q.Retired > 0, "backlog shutdown retained records")
			failed := health.TraceSpansFailed
			if signal == "logs" {
				failed = health.LogRecordsFailed
			}
			require(q.Admitted == failed+q.Retired, "backlog shutdown loss accounting")
		}
		encoded, encodeErr := json.Marshal(health)
		must(encodeErr)
		fmt.Printf("GO_BACKLOG_SHUTDOWN PASSED elapsedMs=%d health=%s\n", time.Since(started).Milliseconds(), encoded)
	} else {
		must(context.RuntimeTelemetry().Flush(stdcontext.Background()))
		// Test/operations diagnostics only; no business method configures SDK spans.
		if telemetry, ok := context.RuntimeTelemetry().(*standard.Telemetry); ok {
			health, healthErr := json.Marshal(telemetry.ExportHealth())
			must(healthErr)
			fmt.Printf("GO_STANDARD_EXPORT_HEALTH %s\n", health)
		}
		if *pressure {
			rows, readErr := lib.Q.CustomerOrdersMinimal().WithIdIs(order.Id()).SelectOrderNumber().Limit(1).
				Comment("load recovered queue order").Purpose("verify telemetry recovery after export").ExecuteForList(context)
			must(readErr)
			require(len(rows.Data) == 1 && value(lib.E.CustomerOrder(rows.Data[0]).OrderNumber().Eval()) == secret+"-SUCCESS", "recovered queue Q/E changed")
			must(context.RuntimeTelemetry().Flush(stdcontext.Background()))
			telemetry := context.RuntimeTelemetry().(*standard.Telemetry)
			health, healthErr := json.Marshal(telemetry.ExportHealth())
			must(healthErr)
			fmt.Printf("GO_QUEUE_RECOVERY PASSED rows=1 health=%s\n", health)
		}
	}
	if *postShutdown {
		telemetry, ok := context.RuntimeTelemetry().(*standard.Telemetry)
		require(ok, "post-shutdown mode requires enabled standard telemetry")
		must(telemetry.Shutdown(stdcontext.Background()))
		before := telemetry.ExportHealth()
		rows, readErr := lib.Q.CustomerOrdersMinimal().WithIdIs(order.Id()).SelectOrderNumber().Limit(1).
			Comment("load post-shutdown order").Purpose("verify business continuity after telemetry shutdown").ExecuteForList(context)
		must(readErr)
		require(len(rows.Data) == 1 && value(lib.E.CustomerOrder(rows.Data[0]).OrderNumber().Eval()) == secret+"-SUCCESS" && sink.count == first+7, "post-shutdown Q/E or audit changed")
		after := telemetry.ExportHealth()
		for _, signal := range []string{"traces", "logs"} {
			require(after.Queues[signal].PostShutdown-before.Queues[signal].PostShutdown == 2, "post-shutdown root/provider handoffs not counted")
			require(after.Queues[signal].Outstanding == 0, "shutdown retained quota")
		}
		health, healthErr := json.Marshal(after)
		must(healthErr)
		fmt.Printf("GO_POST_SHUTDOWN_READ PASSED rows=1 audits=%d health=%s\n", sink.count, health)
	}
	fmt.Printf("GO_STANDARD_GRAPH PASSED savedEntities=3 rolledBackEntities=3 recoveredEntities=3 first_seed_audits=%d repeat_seed_audits=0 audits=%d ordinary_sql_logs=0 database=%s\n", first, sink.count, database)
}

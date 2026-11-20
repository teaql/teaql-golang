package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
	orders "order-service-routing-core-workspace/lib"
	"order-service-routing-core-workspace/lib/purchase_order"
)

const (
	reference = "PO-ROUTED"
	requested = "Receive a valve kit"
	accepted  = "Receive two valve kits"
)

var (
	schemaRefused   = errors.New("purchasing schema is unavailable")
	numberingFailed = errors.New("purchasing ID service is unavailable")
	writeRefused    = errors.New("purchasing write path is temporarily unavailable")
)

type schemaValidator struct {
	db    *sql.DB
	calls int
	deny  bool
}

func (schema *schemaValidator) EnsureSchema(ctx context.Context, metadata runtime.MetadataStore) error {
	schema.calls++
	if schema.deny {
		return schemaRefused
	}
	order := metadata.Entity("PurchaseOrder")
	if order == nil || order.TabName != "purchase_order_data" {
		return errors.New("the generated purchase order schema is missing")
	}
	var table string
	return schema.db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name=?", order.TabName).Scan(&table)
}

type deskIDs struct {
	next, floor uint64
	calls       int
	fail        bool
}

func (ids *deskIDs) GenerateId(entity string) (uint64, error) {
	ids.calls++
	if ids.fail || entity != "PurchaseOrder" {
		return 0, numberingFailed
	}
	ids.next++
	return ids.next, nil
}

func (ids *deskIDs) EnsureIdFloor(_ context.Context, entity string, floor uint64) error {
	if entity != "PurchaseOrder" {
		return fmt.Errorf("unmanaged order ID floor for %s", entity)
	}
	ids.floor = floor
	if floor > ids.next {
		ids.next = floor
	}
	return nil
}

type observedDefaultIDs struct {
	runtime.InternalIdGenerator
	calls int
}

func (ids *observedDefaultIDs) GenerateId(entity string) (uint64, error) {
	ids.calls++
	return ids.InternalIdGenerator.GenerateId(entity)
}

type businessProvider struct {
	*teaqlsql.SqlDataServiceExecutor
	reads, begins, writes, commits, rollbacks int
	failWrite                                 bool
}

func (provider *businessProvider) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	provider.reads++
	return provider.SqlDataServiceExecutor.Query(ctx, request)
}

func (provider *businessProvider) Begin(ctx context.Context) (data_service.Transaction, error) {
	provider.begins++
	transaction, err := provider.SqlDataServiceExecutor.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &businessTransaction{Transaction: transaction, provider: provider}, nil
}

type businessTransaction struct {
	data_service.Transaction
	provider *businessProvider
}

func (transaction *businessTransaction) Mutate(ctx context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	transaction.provider.writes++
	if transaction.provider.failWrite {
		return nil, writeRefused
	}
	return transaction.Transaction.Mutate(ctx, request)
}

func (transaction *businessTransaction) Commit(ctx context.Context) error {
	if err := transaction.Transaction.Commit(ctx); err != nil {
		return err
	}
	transaction.provider.commits++
	return nil
}

func (transaction *businessTransaction) Rollback(ctx context.Context) error {
	transaction.provider.rollbacks++
	return transaction.Transaction.Rollback(ctx)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	phase := os.Getenv("TEAQL_ORDER_ROUTING_PHASE")
	if phase != "first" && phase != "restart" {
		return errors.New("set TEAQL_ORDER_ROUTING_PHASE to first or restart")
	}
	caller, err := orders.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	db := caller.GetResource("db").(*sql.DB)
	defer db.Close()
	// Provision the model once with the generated schema before the order desk attaches.
	if phase == "first" {
		if err := orders.EnsureSchema(caller); err != nil {
			return err
		}
	}

	schema := &schemaValidator{db: db}
	provider := &businessProvider{SqlDataServiceExecutor: caller.GetResource("dataService").(*teaqlsql.SqlDataServiceExecutor)}
	fallback := &observedDefaultIDs{InternalIdGenerator: caller.InternalIdGenerator()}
	ids := new(deskIDs)
	caller.InsertResource("dataService", provider)
	caller.WithSchemaExecutor(schema).WithInternalIdGenerator(ids)
	caller.InstallDefaultInternalIdGenerator(fallback)

	if err := orders.EnsureSchema(caller); err != nil {
		return err
	}
	reads := provider.reads
	schema.deny = true
	if err := orders.EnsureSchema(caller); !errors.Is(err, schemaRefused) ||
		schema.calls != 2 || provider.reads != reads || provider.begins != 0 {
		return fmt.Errorf("schema refusal entered business execution or bypassed the selected capability: %v", err)
	}

	if phase == "first" {
		if err := placeAndRevise(caller, provider, ids); err != nil {
			return err
		}
	}
	retained, err := readOrder(caller)
	if err != nil {
		return err
	}
	description, loaded := orders.E.PurchaseOrder(retained).Description().Eval()
	if !loaded || description != accepted || retained.Id() != 901 || retained.Version() != 2 || provider.reads <= reads {
		return errors.New("the selected read path did not return the expected order and revision")
	}
	ids.fail = true
	before := provider.begins
	rejected, err := prepareOrder(caller)
	if rejected != nil || err == nil || !strings.Contains(err.Error(), numberingFailed.Error()) ||
		fallback.calls != 0 || provider.begins != before {
		return fmt.Errorf("failed allocation reached mutation or the default generator: %v", err)
	}

	fmt.Print("ORDER_SERVICE_ROUTING=")
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"phase": phase, "id": retained.Id(), "version": retained.Version(), "description": description,
		"schemaCalls": schema.calls, "schemaRejectionBeforeBusiness": true, "businessRead": true,
		"defaultIdCalls": fallback.calls, "customIdCalls": ids.calls, "idFloor": ids.floor,
		"allocationRejected": true, "begins": provider.begins, "writes": provider.writes,
		"commits": provider.commits, "rollbacks": provider.rollbacks,
	})
}

func prepareOrder(caller *runtime.UserContext) (*purchase_order.PurchaseOrder, error) {
	return orders.Q.PurchaseOrders().Comment("Open a purchasing request").
		Purpose("Allocate the identity with this desk's Context").NewEntity(caller)
}

func readOrder(caller *runtime.UserContext) (*purchase_order.PurchaseOrder, error) {
	order, err := orders.Q.PurchaseOrders().WithReferenceIs(reference).Limit(1).
		Comment("Read the order through the selected query service").
		Purpose("Verify the durable purchase revision").ExecuteForOne(caller)
	if err == nil && order == nil {
		return nil, errors.New("the routed order is missing")
	}
	return order, err
}

func placeAndRevise(caller *runtime.UserContext, provider *businessProvider, ids *deskIDs) error {
	if err := caller.EnsureInternalIdFloor(context.Background(), "PurchaseOrder", 900); err != nil {
		return err
	}
	order, err := prepareOrder(caller)
	if err != nil {
		return err
	}
	if order.Id() != 901 || ids.floor != 900 {
		return errors.New("the purchase order missed the allocated range")
	}
	order.UpdatePlatformId(1).UpdateReference(reference).UpdateDescription(requested)
	order, err = order.AuditAs("Place an order for a valve kit").Save(caller)
	if err != nil {
		return err
	}

	order.UpdateDescription(accepted)
	provider.failWrite = true
	_, err = order.AuditAs("Revise the requested stock").Save(caller)
	provider.failWrite = false
	if err == nil || !strings.Contains(err.Error(), writeRefused.Error()) || provider.rollbacks != 1 {
		return fmt.Errorf("the configured write failure did not roll back: %v", err)
	}
	previous, err := readOrder(caller)
	if err != nil {
		return err
	}
	if previous.Version() != 1 || previous.Description() != requested {
		return errors.New("the failed write changed the stored order")
	}
	_, err = order.AuditAs("Retry the same stock revision after the outage").Save(caller)
	return err
}

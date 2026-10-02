package sql

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

// #41: executor/sink wiring, not generated Q/Mutation acceptance.
type traceRecorderContext struct {
	context.Context
	entries []ds.ExecutionMetadata
}

func TestMutationExecutorRetainsTypedLineageSeparateFromSQL(t *testing.T) {
	transport := &mockSqlTransport{executeSql: func(context.Context, *CompiledQuery) (uint64, error) { return 1, nil }}
	provider := &mockSchemaProvider{getEntity: func(string) *core.EntityDescriptor { return entity() }}
	executor := NewSqlDataServiceExecutor(&TestDialect{}, transport, provider)
	rootID, itemID := uint64(100), uint64(201)
	root := core.NewTypedTraceNode("auditReason", "CustomerOrder", "submit order")
	root.EntityId = &rootID
	child := core.NewTypedTraceNode("auditReason", "Order", "change item")
	child.EntityId = &itemID
	command := core.NewUpdateCommand("Order", core.ValU64(itemID)).Value("name", core.ValText("B")).WithExpectedVersion(1)
	command.TraceChain = []*core.TraceNode{root, child}
	request, err := ds.NewMutationRequest(&ds.UpdateMutation{Cmd: command}, "submit order")
	if err != nil {
		t.Fatal(err)
	}
	// Request capture must own both frames and their ID pointers.
	root.Comment = "mutated builder"
	rootID = 999
	caller := &traceRecorderContext{Context: context.Background()}
	result, err := executor.Mutate(caller, request)
	if err != nil {
		t.Fatal(err)
	}
	lineage := result.Metadata.MutationLineage
	if len(lineage) != 2 || lineage[0].Name != "CustomerOrder" || lineage[0].Comment != "submit order" || *lineage[0].EntityId != 100 ||
		lineage[1].Name != "Order" || *lineage[1].EntityId != 201 {
		t.Fatalf("lineage changed: %v", lineage)
	}
	if result.Metadata.TraceChain[0].Name != "CustomerOrder" || result.Metadata.TraceChain[1].Name != "Order" ||
		result.Metadata.TraceChain[3].Name != "update" {
		t.Fatalf("SQL route differs: %v", result.Metadata.TraceChain)
	}
	before := core.CloneTraceNodes(lineage)
	readError := errors.New("readback probe")
	recordMutationReadback(caller, &CompiledQuery{Sql: "SELECT * FROM order_data WHERE id=?"}, result.Metadata, time.Now(), 0, readError)
	read := caller.entries[len(caller.entries)-1]
	if read.ExecutionOutcome != "failure" || !reflect.DeepEqual(read.MutationLineage, before) {
		t.Fatal("failure lost lineage")
	}
	if result.Metadata.ExecutionOutcome != "success" {
		t.Fatal("readback overwrote successful write")
	}
	*read.MutationLineage[0].EntityId = 500
	if *result.Metadata.MutationLineage[0].EntityId != 100 {
		t.Fatal("readback shares mutable IDs with write")
	}
}

func (c *traceRecorderContext) RecordExecutionMetadata(entry ds.ExecutionMetadata) {
	c.entries = append(c.entries, entry)
}

func TestQueryExecutorUsesRustCanonicalNodeMeaning(t *testing.T) {
	transport := &mockSqlTransport{fetchAllSql: func(context.Context, *CompiledQuery) ([]core.Record, error) {
		return []core.Record{{"id": core.ValI64(1)}}, nil
	}}
	provider := &mockSchemaProvider{getEntity: func(string) *core.EntityDescriptor { return entity() }}
	executor := NewSqlDataServiceExecutor(&TestDialect{}, transport, provider)
	caller := &traceRecorderContext{Context: context.Background()}
	request := &ds.QueryRequest{
		Query:   core.NewSelectQuery("Order").Project("id"),
		Comment: fixtureIntentText("load order"), Purpose: fixtureIntentText("show order"),
		TraceChain: []*core.TraceNode{core.NewTypedTraceNode("relation", "items", "Order.items")},
	}
	result, err := executor.Query(caller, request)
	if err != nil {
		t.Fatal(err)
	}
	path := result.Metadata.TraceChain
	if len(path) != 5 || path[0].Kind != "operation" || path[0].Name != "Order" || path[0].Comment != "query" {
		t.Fatalf("operation must carry the semantic root, got %v", path)
	}
	if path[1].Name != "Order" || path[1].Comment != "" || path[2].Name != "items" || path[2].Comment != "Order.items" {
		t.Fatalf("request/relation meaning changed: %v", path)
	}
	if path[3].Comment != "" || path[4].Name != "select" || path[4].Comment != "" {
		t.Fatalf("provider/sql details must be empty: %v", path)
	}
	if len(caller.entries) != 1 {
		t.Fatal("executor did not send the actual metadata to the sink")
	}
}

func TestMutationReadbackHasOneSelectNodeAndPreservesWrite(t *testing.T) {
	caller := &traceRecorderContext{Context: context.Background()}
	reason := "submit order"
	source := ds.ExecutionMetadata{
		Backend: "sqlite", Operation: ds.OpUpdate, ExecutionOutcome: "success", AuditReason: &reason,
		TraceChain: []*core.TraceNode{
			core.NewTypedTraceNode("operation", "Order", "mutation"),
			core.NewTypedTraceNode("entity", "OrderItem", ""),
			core.NewTypedTraceNode("provider", "sqlite", ""),
			core.NewTypedTraceNode("sql", "update", ""),
		},
	}
	recordMutationReadback(caller, &CompiledQuery{Sql: "SELECT * FROM order_item_data WHERE id = ?"}, source, time.Now(), 1, nil)
	if len(caller.entries) != 1 {
		t.Fatal("readback not observed")
	}
	read := caller.entries[0]
	if len(read.TraceChain) != 4 || read.TraceChain[0].Name != "Order" || read.TraceChain[0].Comment != "query" ||
		read.TraceChain[1].Kind != "request" || read.TraceChain[3].Name != "select" {
		t.Fatalf("readback must be a separate canonical select, got %v", read.TraceChain)
	}
	if source.TraceChain[3].Name != "update" || source.ExecutionOutcome != "success" {
		t.Fatal("readback mutated the write evidence")
	}
}

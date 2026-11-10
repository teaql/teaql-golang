package runtime

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"testing"
)

type listSemanticsExecutor struct {
	scalar    core.Value
	countRows []core.Record
}

func (*listSemanticsExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}
func (e *listSemanticsExecutor) Query(_ context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	if len(request.Query.Aggregates) == 1 && request.Query.Aggregates[0].Alias == "totalOrders" {
		return &data_service.QueryResult{Rows: e.countRows}, nil
	}
	return &data_service.QueryResult{Rows: []core.Record{{"sum": e.scalar}}}, nil
}
func TestTypedListScalarAliasesNeverHydrateEntities(t *testing.T) {
	for _, scalar := range []core.Value{core.ValDecimal(decimal.NewFromInt(600)), core.ValNull()} {
		for _, counted := range []bool{false, true} {
			query, options := countedProbeQuery()
			if !counted {
				query.Aggregates = nil
				options.PageCountOperation = nil
			}
			query.Sum("amount", "sum")
			snapshot, plan, err := CaptureQueryPlan(query, options)
			if err != nil {
				t.Fatal(err)
			}
			executor := &listSemanticsExecutor{scalar: scalar, countRows: []core.Record{{"totalOrders": core.ValU64(3)}}}
			rows, err := ExecuteTypedList(NewUserContext(), NewRuntimeDataService(nil, executor), snapshot, plan, func(core.Record) (int, error) { t.Fatal("scalar alias reached entity hydration"); return 0, nil })
			if err != nil {
				t.Fatal(err)
			}
			if len(rows.Data) != 0 || !rows.IsLoaded || (rows.TotalCount != nil) != counted {
				t.Fatal("scalar collection boundary", rows)
			}
			value, exists := rows.Aggregations["sum"]
			if !exists || value.IsNull() != scalar.IsNull() {
				t.Fatal("scalar alias/null missing")
			}
			if counted && *rows.TotalCount != 3 {
				t.Fatal("count was lost")
			}
		}
	}
}
func TestExactListCountRejectsFractionsAndMalformedCardinality(t *testing.T) {
	for _, records := range [][]core.Record{
		{{"totalOrders": core.ValDecimal(decimal.RequireFromString("1.5"))}},
		{{"totalOrders": core.ValI64(-1)}},
		{{"totalOrders": core.ValU64(1)}, {"totalOrders": core.ValU64(2)}},
		nil,
	} {
		query, options := countedProbeQuery()
		snapshot, plan, err := CaptureQueryPlan(query, options)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ExecuteTypedList(NewUserContext(), NewRuntimeDataService(nil, &listSemanticsExecutor{countRows: records}), snapshot, plan, func(core.Record) (int, error) { return 0, nil })
		if err == nil {
			t.Fatal("invalid exact count accepted", records)
		}
	}
}

func TestTypedListRejectsGroupedProjections(t *testing.T) {
	query := core.NewSelectQuery("Order").WithGroupBy("amount").Limit(1).Comment("group orders").Purpose("verify entity boundary")
	snapshot, plan, err := CaptureQueryPlan(query, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ExecuteTypedList(NewUserContext(), NewRuntimeDataService(nil, &listSemanticsExecutor{}), snapshot, plan, func(core.Record) (int, error) { t.Fatal("projection reached entity hydration"); return 0, nil })
	if err == nil {
		t.Fatal("grouped projection accepted as entities")
	}
}

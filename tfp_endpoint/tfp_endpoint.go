package tfp_endpoint

import (
	stdcontext "context"
	"encoding/json"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type TfpEndpoint struct {
	queryExecutor    data_service.QueryExecutor
	mutationExecutor data_service.MutationExecutor
}

func NewTfpEndpoint(q data_service.QueryExecutor, m data_service.MutationExecutor) *TfpEndpoint {
	return &TfpEndpoint{
		queryExecutor:    q,
		mutationExecutor: m,
	}
}

// TFP Models for JSON parsing
type TfpOrderBy struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type TfpSelectQuery struct {
	Entity          string                 `json:"entity"`
	FilterCondition map[string]interface{} `json:"filterCondition,omitempty"`
	LimitValue      *uint64                `json:"limitValue,omitempty"`
	OffsetValue     *uint64                `json:"offsetValue,omitempty"`
	OrderItems      []TfpOrderBy           `json:"orderItems,omitempty"`
	SelectItems     []string               `json:"selectItems,omitempty"`
	GroupByItems    []string               `json:"groupByItems,omitempty"`
	CommentText     *string                `json:"commentText,omitempty"`
}

type TfpMutationQuery struct {
	Entity  string                 `json:"entity"`
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
	Id      interface{}            `json:"id,omitempty"`
	Comment *string                `json:"comment,omitempty"`
}

func (e *TfpEndpoint) HandleQuery(context stdcontext.Context, payload []byte) (map[string]interface{}, error) {
	var tfpQuery TfpSelectQuery
	if err := json.Unmarshal(payload, &tfpQuery); err != nil {
		return nil, fmt.Errorf("failed to parse JSON payload: %w", err)
	}

	q := core.NewSelectQuery(tfpQuery.Entity)
	if tfpQuery.LimitValue != nil {
		q.Limit(*tfpQuery.LimitValue)
	}
	if tfpQuery.OffsetValue != nil {
		q.Offset(*tfpQuery.OffsetValue)
	}
	for _, o := range tfpQuery.OrderItems {
		dir := core.SortAsc
		if o.Direction == "Desc" {
			dir = core.SortDesc
		}
		q.WithOrderBy(core.NewOrderBy(o.Field, dir))
	}
	q.Projects(tfpQuery.SelectItems...)
	for _, g := range tfpQuery.GroupByItems {
		q.WithGroupBy(g)
	}
	// Note: Filter is skipped in this mock mapper for simplicity
	// Inject implicit soft delete filter: version > 0
	q.WithFilter(core.ExprGt("version", core.Value{V: int64(0)}))
	if tfpQuery.CommentText != nil {
		q.Comment(*tfpQuery.CommentText)
	}

	req := &data_service.QueryRequest{
		Query:   q,
		Comment: tfpQuery.CommentText,
	}

	res, err := e.queryExecutor.Query(context, req)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	rows := make([]map[string]interface{}, 0, len(res.Rows))
	for _, r := range res.Rows {
		// Convert core.Record (map[string]core.Value) back to map[string]interface{}
		m := make(map[string]interface{})
		for k, v := range r {
			m[k] = v.ToJsonValue()
		}
		rows = append(rows, m)
	}

	response := map[string]interface{}{
		"data":       rows,
		"resultCode": 0,
		"status":     "YES",
	}

	return response, nil
}

func (e *TfpEndpoint) HandleMutation(context stdcontext.Context, payload []byte) (map[string]interface{}, error) {
	var tfpMut TfpMutationQuery
	if err := json.Unmarshal(payload, &tfpMut); err != nil {
		return nil, fmt.Errorf("failed to parse JSON payload: %w", err)
	}

	var mutReq data_service.MutationRequest
	trace := []*core.TraceNode{{EntityType: tfpMut.Entity}}
	if tfpMut.Comment != nil {
		trace[0].Comment = *tfpMut.Comment
	}

	// Convert payload to core.Record
	record := make(core.Record)
	for k, v := range tfpMut.Payload {
		record[k] = core.Value{V: v}
	}
	idVal := core.Value{V: tfpMut.Id}
	if tfpMut.Id == nil {
		idVal = core.ValNull()
	}

	var expectedVersion *int64
	if v, ok := tfpMut.Payload["version"]; ok {
		switch val := v.(type) {
		case float64:
			ver := int64(val)
			expectedVersion = &ver
		case int:
			ver := int64(val)
			expectedVersion = &ver
		case int64:
			expectedVersion = &val
		}
	}

	switch tfpMut.Action {
	case "Create":
		mutReq = &data_service.InsertMutation{
			Cmd: &core.InsertCommand{
				Entity:     tfpMut.Entity,
				Values:     record,
				TraceChain: trace,
			},
		}
	case "Update":
		mutReq = &data_service.UpdateMutation{
			Cmd: &core.UpdateCommand{
				Entity:          tfpMut.Entity,
				Id:              idVal,
				ExpectedVersion: expectedVersion,
				Values:          record,
				TraceChain:      trace,
			},
		}
	case "Delete":
		mutReq = &data_service.DeleteMutation{
			Cmd: &core.DeleteCommand{
				Entity:          tfpMut.Entity,
				Id:              idVal,
				ExpectedVersion: expectedVersion,
				SoftDelete:      true,
				TraceChain:      trace,
			},
		}
	case "Recover":
		mutReq = &data_service.RecoverMutation{
			Cmd: &core.RecoverCommand{
				Entity:     tfpMut.Entity,
				Id:         idVal,
				TraceChain: trace,
			},
		}
	default:
		return nil, fmt.Errorf("unknown mutation action: %s", tfpMut.Action)
	}

	res, err := e.mutationExecutor.Mutate(context, mutReq)
	if err != nil {
		return nil, fmt.Errorf("mutation execution failed: %w", err)
	}

	dataArr := []map[string]interface{}{}
	if res.GeneratedValues != nil {
		m := make(map[string]interface{})
		for k, v := range res.GeneratedValues {
			m[k] = v.ToJsonValue()
		}
		dataArr = append(dataArr, m)
	}

	response := map[string]interface{}{
		"affectedRows": res.AffectedRows,
		"resultCode":   0,
		"status":       "YES",
		"data":         dataArr,
	}

	return response, nil
}

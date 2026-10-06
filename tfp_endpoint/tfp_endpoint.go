package tfp_endpoint

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type TfpEndpoint struct {
	queryExecutor    data_service.QueryExecutor
	mutationExecutor data_service.MutationExecutor
	trusted          *TrustedFederalContext
}

func mapField(fields map[string]string, field string) (string, error) {
	mapped, ok := fields[field]
	if !ok {
		return "", &TfpError{Code: "TFP_FORBIDDEN_FIELD", Message: "field is not allowed: " + field}
	}
	return mapped, nil
}

const (
	maxSelectItems   = 256
	maxOrderItems    = 32
	maxGroupItems    = 64
	maxPayloadFields = 256
	maxFilterDepth   = 16
	maxFilterNodes   = 256
	maxInValues      = 100
)

func parseFilter(node map[string]interface{}, fields map[string]string, depth int, nodes *int) (*core.Expr, error) {
	*nodes++
	if depth > maxFilterDepth || *nodes > maxFilterNodes {
		return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "filter complexity exceeds protocol budget"}
	}
	if len(node) != 1 {
		return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "filter must contain one expression"}
	}
	for key, raw := range node {
		if key == "$and" || key == "$or" {
			items, ok := raw.([]interface{})
			if !ok || len(items) == 0 {
				return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "logical filter requires operands"}
			}
			parts := make([]*core.Expr, 0, len(items))
			for _, item := range items {
				child, ok := item.(map[string]interface{})
				if !ok {
					return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "invalid logical operand"}
				}
				expr, err := parseFilter(child, fields, depth+1, nodes)
				if err != nil {
					return nil, err
				}
				parts = append(parts, expr)
			}
			if key == "$and" {
				return core.ExprAndNode(parts...), nil
			}
			return core.ExprOrNode(parts...), nil
		}
		field, err := mapField(fields, key)
		if err != nil {
			return nil, err
		}
		predicate, ok := raw.(map[string]interface{})
		if !ok || len(predicate) != 1 {
			return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "invalid field predicate"}
		}
		for op, value := range predicate {
			switch op {
			case "$eq":
				return core.ExprEq(field, core.Value{V: value}), nil
			case "$gte":
				return core.ExprGte(field, core.Value{V: value}), nil
			case "$lte":
				return core.ExprLte(field, core.Value{V: value}), nil
			case "$contains":
				text, ok := value.(string)
				if !ok {
					return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "contains requires text"}
				}
				return core.ExprContain(field, text), nil
			case "$in":
				values, ok := value.([]interface{})
				if !ok || len(values) > maxInValues {
					return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "in requires array"}
				}
				list := make([]core.Value, len(values))
				for i, v := range values {
					list[i] = core.Value{V: v}
				}
				return core.ExprInList(field, list), nil
			default:
				return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "unsupported predicate operator"}
			}
		}
	}
	return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "empty filter"}
}

func rejectPrivilegedInput(payload []byte) error {
	var value interface{}
	if err := json.Unmarshal(payload, &value); err != nil {
		return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "invalid JSON"}
	}
	forbidden := map[string]bool{"tenant": true, "tenantId": true, "merchant": true, "merchantId": true, "user": true, "userId": true, "permissions": true, "requestPolicy": true, "purposePolicy": true, "trustedContext": true, "hardLimit": true, "hard_limit": true, "hardLimitValue": true, "hard_limit_value": true, "continuousPageFetch": true, "idSetPagination": true, "id_set_pagination": true, "paginationWithIdSet": true}
	var visit func(interface{}) error
	visit = func(current interface{}) error {
		switch item := current.(type) {
		case map[string]interface{}:
			for key, child := range item {
				if forbidden[key] {
					return &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "client cannot provide server-owned field"}
				}
				if err := visit(child); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, child := range item {
				if err := visit(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return visit(value)
}

func rejectUnknownTopLevel(payload []byte, allowed map[string]bool) error {
	var object map[string]interface{}
	if err := json.Unmarshal(payload, &object); err != nil {
		return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "invalid JSON"}
	}
	for field := range object {
		if !allowed[field] {
			return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "unknown TFP field: " + field}
		}
	}
	return nil
}

type TrustedFederalContext struct {
	TenantField       string
	TenantID          core.Value
	AuthenticatedUser string
	ApprovedPurpose   string
	AllowedEntities   map[string]bool
	ReadableFields    map[string]map[string]string
	WritableFields    map[string]map[string]string
	AllowedActions    map[string]map[string]bool
	MaxPageSize       uint64
	MaxOffset         uint64
}

type TfpError struct {
	Code    string
	Message string
}

func (e *TfpError) Error() string { return e.Code + ": " + e.Message }

func NewTfpEndpoint(q data_service.QueryExecutor, m data_service.MutationExecutor) *TfpEndpoint {
	return &TfpEndpoint{
		queryExecutor:    q,
		mutationExecutor: m,
	}
}

func (e *TfpEndpoint) WithTrustedContext(trusted TrustedFederalContext) *TfpEndpoint {
	e.trusted = &trusted
	return e
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
	AggregateItems  []interface{}          `json:"aggregateItems,omitempty"`
	CommentText     *string                `json:"commentText,omitempty"`
	PurposeText     *string                `json:"purposeText,omitempty"`
}

type TfpMutationQuery struct {
	Entity          string                 `json:"entity"`
	Action          string                 `json:"action"`
	Payload         map[string]interface{} `json:"payload"`
	Id              interface{}            `json:"id,omitempty"`
	ExpectedVersion *int64                 `json:"expectedVersion,omitempty"`
	Comment         *string                `json:"comment,omitempty"`
}

// Decode only the required intent before typed payload decoding, so wrong JSON
// types produce the same stable, value-free error as missing/null text.
func validateWireRequestIntent(payload []byte, query bool) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(payload, &object); err != nil {
		return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "invalid JSON"}
	}
	text := func(key string) *string {
		var value *string
		if err := json.Unmarshal(object[key], &value); err != nil {
			return nil
		}
		return value
	}
	if query {
		_, err := core.NewQueryIntent(text("commentText"), text("purposeText"))
		return err
	}
	_, err := core.NewMutationIntent(text("comment"))
	return err
}

func (e *TfpEndpoint) HandleQuery(context stdcontext.Context, payload []byte) (map[string]interface{}, error) {
	if e.trusted == nil {
		return nil, &TfpError{Code: "TFP_UNAUTHORIZED", Message: "trusted federation context is required"}
	}
	if err := rejectPrivilegedInput(payload); err != nil {
		return nil, err
	}
	if err := rejectUnknownTopLevel(payload, map[string]bool{"entity": true, "filterCondition": true, "limitValue": true, "offsetValue": true, "orderItems": true, "selectItems": true, "groupByItems": true, "aggregateItems": true, "commentText": true, "purposeText": true}); err != nil {
		return nil, err
	}
	var tfpQuery TfpSelectQuery
	if err := validateWireRequestIntent(payload, true); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(payload, &tfpQuery); err != nil {
		return nil, fmt.Errorf("failed to parse JSON payload: %w", err)
	}
	if _, err := core.NewQueryIntent(tfpQuery.CommentText, tfpQuery.PurposeText); err != nil {
		return nil, err
	}

	trusted := e.trusted
	if !trusted.AllowedEntities[tfpQuery.Entity] {
		return nil, &TfpError{Code: "TFP_FORBIDDEN_ENTITY", Message: "entity is not allowed"}
	}
	fields, ok := trusted.ReadableFields[tfpQuery.Entity]
	if !ok {
		return nil, &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "no readable field policy"}
	}
	if err := validateIntentText(*tfpQuery.CommentText); err != nil {
		return nil, err
	}
	if err := validateIntentText(*tfpQuery.PurposeText); err != nil {
		return nil, err
	}
	if tfpQuery.LimitValue == nil {
		return nil, &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "limitValue is required"}
	}
	if len(tfpQuery.SelectItems) > maxSelectItems || len(tfpQuery.OrderItems) > maxOrderItems || len(tfpQuery.GroupByItems) > maxGroupItems {
		return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "query shape exceeds protocol budget"}
	}
	q := core.NewSelectQuery(tfpQuery.Entity)
	if *tfpQuery.LimitValue == 0 || *tfpQuery.LimitValue > trusted.MaxPageSize {
		return nil, &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "invalid federation page size"}
	}
	q.Limit(*tfpQuery.LimitValue)
	if tfpQuery.OffsetValue != nil {
		maxOffset := trusted.MaxOffset
		if maxOffset == 0 {
			maxOffset = 10_000
		}
		if *tfpQuery.OffsetValue > maxOffset {
			return nil, &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "offset exceeds federation policy"}
		}
		q.Offset(*tfpQuery.OffsetValue)
	}
	seen := map[string]bool{}
	for _, o := range tfpQuery.OrderItems {
		dir := core.SortAsc
		if o.Direction == "Desc" {
			dir = core.SortDesc
		}
		if o.Direction != "Asc" && o.Direction != "Desc" {
			return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "unsupported order direction"}
		}
		mapped, err := mapField(fields, o.Field)
		if err != nil {
			return nil, err
		}
		if seen["order:"+mapped] {
			return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "duplicate mapped order field"}
		}
		seen["order:"+mapped] = true
		q.WithOrderBy(core.NewOrderBy(mapped, dir))
	}
	selected := make([]string, 0, len(tfpQuery.SelectItems))
	for _, f := range tfpQuery.SelectItems {
		mapped, err := mapField(fields, f)
		if err != nil {
			return nil, err
		}
		if seen["select:"+mapped] {
			return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "duplicate mapped select field"}
		}
		seen["select:"+mapped] = true
		selected = append(selected, mapped)
	}
	q.Projects(selected...)
	for _, g := range tfpQuery.GroupByItems {
		mapped, err := mapField(fields, g)
		if err != nil {
			return nil, err
		}
		if seen["group:"+mapped] {
			return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "duplicate mapped group field"}
		}
		seen["group:"+mapped] = true
		q.WithGroupBy(mapped)
	}
	if len(tfpQuery.AggregateItems) > 0 {
		return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "aggregation is not supported by this endpoint"}
	}
	if tfpQuery.FilterCondition != nil {
		nodes := 0
		filter, err := parseFilter(tfpQuery.FilterCondition, fields, 1, &nodes)
		if err != nil {
			return nil, err
		}
		q.WithFilter(filter)
	}
	q.AndFilter(core.ExprEq(trusted.TenantField, trusted.TenantID))
	q.AndFilter(core.ExprGt("version", core.Value{V: int64(0)}))
	if tfpQuery.CommentText != nil {
		q.Comment(*tfpQuery.CommentText)
	}

	req, err := data_service.CaptureQueryRequest(&data_service.QueryRequest{
		Query:   q,
		Comment: tfpQuery.CommentText,
		Purpose: tfpQuery.PurposeText,
	})
	if err != nil {
		return nil, err
	}

	res, err := e.queryExecutor.Query(context, req)
	if err != nil {
		return nil, &TfpError{Code: "TFP_EXECUTION_FAILED", Message: "Data service execution failed"}
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
	if e.trusted == nil {
		return nil, &TfpError{Code: "TFP_UNAUTHORIZED", Message: "trusted federation context is required"}
	}
	if err := rejectPrivilegedInput(payload); err != nil {
		return nil, err
	}
	if err := rejectUnknownTopLevel(payload, map[string]bool{"entity": true, "action": true, "payload": true, "id": true, "expectedVersion": true, "comment": true}); err != nil {
		return nil, err
	}
	var tfpMut TfpMutationQuery
	if err := validateWireRequestIntent(payload, false); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(payload, &tfpMut); err != nil {
		return nil, fmt.Errorf("failed to parse JSON payload: %w", err)
	}
	if _, err := core.NewMutationIntent(tfpMut.Comment); err != nil {
		return nil, err
	}

	trusted := e.trusted
	if !trusted.AllowedEntities[tfpMut.Entity] {
		return nil, &TfpError{Code: "TFP_FORBIDDEN_ENTITY", Message: "entity is not allowed"}
	}
	if !trusted.AllowedActions[tfpMut.Entity][tfpMut.Action] {
		return nil, &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "mutation action is not allowed"}
	}
	if err := validateIntentText(*tfpMut.Comment); err != nil {
		return nil, err
	}
	if len(tfpMut.Payload) > maxPayloadFields {
		return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "mutation payload exceeds protocol budget"}
	}
	if err := validateMutationLifecycle(tfpMut); err != nil {
		return nil, err
	}
	writable, ok := trusted.WritableFields[tfpMut.Entity]
	if !ok {
		return nil, &TfpError{Code: "TFP_POLICY_VIOLATION", Message: "no writable field policy"}
	}
	var mutReq data_service.MutationRequest
	trace := []*core.TraceNode{{EntityType: tfpMut.Entity}}
	if tfpMut.Comment != nil {
		trace[0].Comment = *tfpMut.Comment
	}

	// Convert payload to core.Record
	record := make(core.Record)
	for k, v := range tfpMut.Payload {
		mapped, exists := writable[k]
		if !exists {
			return nil, &TfpError{Code: "TFP_FORBIDDEN_FIELD", Message: "mutation field is not allowed: " + k}
		}
		record[mapped] = core.Value{V: v}
	}
	record[trusted.TenantField] = trusted.TenantID
	normalizedID := normalizeMutationID(tfpMut.Id)
	idVal := core.Value{V: normalizedID}
	if tfpMut.Id == nil {
		idVal = core.ValNull()
	}

	expectedVersion := tfpMut.ExpectedVersion

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
				Guards:          core.Record{trusted.TenantField: trusted.TenantID},
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
				Guards:          core.Record{trusted.TenantField: trusted.TenantID},
			},
		}
	case "Recover":
		mutReq = &data_service.RecoverMutation{
			Cmd: &core.RecoverCommand{
				Entity:          tfpMut.Entity,
				Id:              idVal,
				ExpectedVersion: *expectedVersion,
				TraceChain:      trace,
				Guards:          core.Record{trusted.TenantField: trusted.TenantID},
			},
		}
	default:
		return nil, &TfpError{Code: "TFP_INVALID_REQUEST", Message: "unknown mutation action"}
	}

	mutReq, err := data_service.NewMutationRequest(mutReq, *tfpMut.Comment)
	if err != nil {
		return nil, err
	}
	res, err := e.mutationExecutor.Mutate(context, mutReq)
	if err != nil {
		return nil, &TfpError{Code: "TFP_EXECUTION_FAILED", Message: "Data service execution failed"}
	}
	if tfpMut.Action == "Create" && res.AffectedRows != 1 {
		return nil, &TfpError{Code: "TFP_EXECUTION_FAILED", Message: "Data service execution failed"}
	}
	if tfpMut.Action != "Create" && res.AffectedRows != 1 {
		return nil, &TfpError{Code: "TFP_UNAVAILABLE", Message: "mutation target is unavailable"}
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

func validMutationID(id interface{}) bool {
	switch value := id.(type) {
	case string:
		return strings.TrimSpace(value) != ""
	case float64:
		return value > 0 && value == math.Trunc(value)
	case json.Number:
		_, err := value.Int64()
		return err == nil
	case map[string]interface{}:
		return len(value) == 1 && validMutationID(value["id"])
	default:
		return false
	}
}

func normalizeMutationID(id interface{}) interface{} {
	if wrapper, ok := id.(map[string]interface{}); ok && len(wrapper) == 1 {
		return wrapper["id"]
	}
	return id
}

func validateIntentText(value string) error {
	if len([]byte(value)) > 1024 {
		return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "intent text exceeds 1024 UTF-8 bytes"}
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "intent text contains control characters"}
		}
	}
	return nil
}

func validateMutationLifecycle(m TfpMutationQuery) error {
	empty := len(m.Payload) == 0
	switch m.Action {
	case "Create":
		if m.Id != nil || m.ExpectedVersion != nil {
			return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "create cannot provide id or expectedVersion"}
		}
	case "Update":
		if !validMutationID(m.Id) || m.ExpectedVersion == nil || *m.ExpectedVersion <= 0 {
			return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "update requires id and positive expectedVersion"}
		}
	case "Delete":
		if !validMutationID(m.Id) || m.ExpectedVersion == nil || *m.ExpectedVersion <= 0 || !empty {
			return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "delete requires id, positive expectedVersion and empty payload"}
		}
	case "Recover":
		if !validMutationID(m.Id) || m.ExpectedVersion == nil || *m.ExpectedVersion >= 0 || !empty {
			return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "recover requires id, negative expectedVersion and empty payload"}
		}
	default:
		return &TfpError{Code: "TFP_INVALID_REQUEST", Message: "unknown mutation action"}
	}
	return nil
}

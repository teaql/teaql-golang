package core

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

var searchModels = map[string]SearchModel{
	"Order":    {Fields: map[string]string{"id": "integer", "name": "string", "amount": "decimal"}, Relations: map[string]string{"customer": "Customer"}},
	"Customer": {Fields: map[string]string{"name": "string"}, Relations: map[string]string{}},
}

func TestDynamicSearchCalendarYearRange(t *testing.T) {
	if validSearchScalar("0000-01-01", "date") || !validSearchScalar("0001-01-01", "date") {
		t.Fatal("local search calendar dates must use years 0001 through 9999")
	}
}

func TestDynamicSearchDriftAndWarningValues(t *testing.T) {
	var emitted []DynamicSearchWarning
	result, err := NormalizeDynamicSearch([]byte(`{"filter":{"removed":"SECRET","missing.name":"SECRET","customer.removed":"SECRET","name":"valid","customer.name":{"$eq":"Ada"}},"orderBy":[{"field":"removed","direction":"asc"},{"field":"id","direction":"desc"}]}`), "Order", searchModels, 100, func(w DynamicSearchWarning) { emitted = append(emitted, w) })
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Filter) != 2 || result.Filter["name"]["$eq"] != "valid" || len(result.OrderBy) != 1 || len(emitted) != 4 {
		t.Fatalf("unexpected result: %#v", result)
	}
	data, _ := json.Marshal(emitted)
	if strings.Contains(string(data), "SECRET") {
		t.Fatal("value leaked")
	}
	for _, w := range emitted {
		if w.Code != "DYNAMIC_SEARCH_UNKNOWN_FIELD" || w.Entity != "Order" {
			t.Fatal(w)
		}
	}
}

func TestDynamicSearchFatalInputs(t *testing.T) {
	for _, source := range []string{`{`, `[]`, `null`, `{} {}`, `{"tenant":1}`, `{"hardLimit":1}`, `{"filter":{"removed":{"$bad":1}}}`, `{"filter":{"id":true}}`, `{"filter":{"id":1.5}}`, `{"filter":{"amount":"NaN"}}`, `{"filter":{"constructor.name":"x"}}`, `{"orderBy":[{"field":"name","direction":"sideways"}]}`} {
		t.Run(source, func(t *testing.T) {
			_, err := NormalizeDynamicSearch([]byte(source), "Order", searchModels, 100, func(DynamicSearchWarning) { t.Fatal("unexpected warning") })
			if err == nil {
				t.Fatal("expected fatal input")
			}
		})
	}
}

func TestDynamicSearchNativeCompositionPreservesScope(t *testing.T) {
	base := NewSelectQuery("Order").WithFilter(ExprEq("tenant", ValI64(1))).Limit(10).OrderAsc("id")
	before, _ := json.Marshal(base)
	query, warnings, err := MergeDynamicSearch(base, []byte(`{"filter":{"name":"valid","removed":"secret"}}`), searchModels,
		func(path string, predicate map[string]any) (*Expr, error) {
			return ExprEq(path, ValText(predicate["$eq"].(string))), nil
		},
		func(path, direction string) (*OrderBy, error) { return OrderAsc(path), nil }, func(DynamicSearchWarning) {})
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(base)
	if string(before) != string(after) {
		t.Fatal("base mutated")
	}
	if len(warnings) != 1 || !reflect.DeepEqual(query.Slice, base.Slice) || !reflect.DeepEqual(query.OrderBy, base.OrderBy) {
		t.Fatal("query controls lost")
	}
	if !reflect.DeepEqual(query.Filter, ExprAndNode(base.Filter, ExprEq("name", ValText("valid")))) {
		t.Fatal("scope lost")
	}
}

func TestDynamicSearchTemporalAndDecimal(t *testing.T) {
	models := map[string]SearchModel{"Entry": {Fields: map[string]string{"date": "date", "created": "timestamp", "amount": "decimal"}}}
	result, err := NormalizeDynamicSearch([]byte(`{"filter":{"date":"2024-02-29","created":1709164800000,"amount":"9007199254740993.01"}}`), "Entry", models, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Filter["amount"]["$eq"] != "9007199254740993.01" {
		t.Fatal("decimal precision changed")
	}
	for _, source := range []string{`{"filter":{"date":"2025-02-29"}}`, `{"filter":{"created":"2024-02-29"}}`} {
		if _, err := NormalizeDynamicSearch([]byte(source), "Entry", models, 100, nil); err == nil {
			t.Fatal("invalid temporal value accepted")
		}
	}
}

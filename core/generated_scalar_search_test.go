package core

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func generatedSearchFixture() map[string]GeneratedSearchModel {
	return map[string]GeneratedSearchModel{
		"Invoice": {
			SearchModel: SearchModel{
				Fields:    map[string]string{"reference_code": "string", "referenceCode": "string", "amount": "decimal", "active": "boolean", "customer": "integer", "date": "date", "created": "timestamp", "ratio": "number", "payload": "unsupported"},
				Relations: map[string]string{"customer": "Customer"},
			},
			QueryFields: map[string]string{"reference_code": "reference_code", "referenceCode": "reference_code", "amount": "amount", "active": "active", "customer": "customer_id", "date": "date", "created": "created", "ratio": "ratio", "payload": "payload"},
		},
		"Customer": {SearchModel: SearchModel{Fields: map[string]string{"name": "string"}}, QueryFields: map[string]string{"name": "name"}},
	}
}

func TestGeneratedScalarSearchPreservesScopeAndBindsAllSiblings(t *testing.T) {
	base := NewSelectQuery("Invoice").WithFilter(ExprEq("tenant", ValI64(7))).Limit(10).OrderAsc("id")
	before, _ := json.Marshal(base)
	var emitted []DynamicSearchWarning
	query, warnings, err := MergeGeneratedScalarSearch(base, []byte(`{"filter":{"referenceCode":{"$contains":"KEEP"},"amount":{"$gte":"100.50"},"active":true,"removed":"SECRET"},"orderBy":[{"field":"reference_code","direction":"desc"}]}`), generatedSearchFixture(), func(w DynamicSearchWarning) { emitted = append(emitted, w) })
	if err != nil {
		t.Fatal(err)
	}
	want := base.Clone().AndFilter(ExprEq("active", ValBool(true))).AndFilter(ExprGte("amount", ValDecimal(decimal.RequireFromString("100.50")))).AndFilter(ExprContain("reference_code", "KEEP"))
	if !reflect.DeepEqual(query.Filter, want.Filter) || len(query.OrderBy) != 2 || !reflect.DeepEqual(query.OrderBy[1], OrderDesc("reference_code")) || !reflect.DeepEqual(query.Slice, base.Slice) {
		t.Fatal("lost native scope, sibling clauses, ordering or pagination")
	}
	after, _ := json.Marshal(base)
	if string(before) != string(after) || len(warnings) != 1 || !reflect.DeepEqual(warnings, emitted) {
		t.Fatal("mutated original query or lost warnings")
	}
	data, _ := json.Marshal(warnings)
	if strings.Contains(string(data), "SECRET") {
		t.Fatal("warnings leaked value")
	}
}

func TestGeneratedScalarSearchRejectsInvalidInputAtomically(t *testing.T) {
	for _, source := range []string{
		`{"filter":{"reference_code":"KEEP","customer.name":"Ada","removed":"secret"}}`,
		`{"orderBy":[{"field":"customer.name","direction":"asc"}]}`,
		`{"filter":{"active":"true"}}`, `{"filter":{"amount":"NaN"}}`,
		`{"filter":{"payload":null}}`, `{"orderBy":[{"field":"payload","direction":"asc"}]}`,
		`{"filter":{"reference_code":{"$contains":null}}}`,
		`{"filter":{"created":{"$gt":null}}}`,
		`{"tenant":99}`, `{"filter":{"removed":{"$bad":1}}}`,
	} {
		t.Run(source, func(t *testing.T) {
			base := NewSelectQuery("Invoice").WithFilter(ExprEq("tenant", ValI64(7)))
			before, _ := json.Marshal(base)
			query, warnings, err := MergeGeneratedScalarSearch(base, []byte(source), generatedSearchFixture(), func(DynamicSearchWarning) { t.Fatal("warning emitted before complete binding") })
			after, _ := json.Marshal(base)
			if err == nil || query != nil || warnings != nil || string(before) != string(after) {
				t.Fatal("fatal input changed query or returned partial results")
			}
		})
	}
}

func TestGeneratedScalarSearchNativeTypesAndOperators(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-10-11")
	for _, fixture := range []struct {
		source   string
		expected *Expr
	}{
		{`{"filter":{"date":"2026-10-11"}}`, ExprEq("date", ValDate(date))},
		{`{"filter":{"created":1791676800000}}`, ExprEq("created", ValTimestamp(1791676800000))},
		{`{"filter":{"ratio":{"$lt":1.25}}}`, ExprLt("ratio", ValF64(1.25))},
		{`{"filter":{"customer":{"$in":[1,2]}}}`, ExprInList("customer_id", []Value{ValI64(1), ValI64(2)})},
		{`{"filter":{"customer":{"$notIn":[1]}}}`, ExprNotInList("customer_id", []Value{ValI64(1)})},
		{`{"filter":{"referenceCode":null}}`, ExprIsNullNode("reference_code")},
		{`{"filter":{"referenceCode":{"$ne":null}}}`, ExprIsNotNullNode("reference_code")},
		{`{"filter":{"amount":{"$lte":1e2}}}`, ExprLte("amount", ValDecimal(decimal.RequireFromString("1e2")))},
		{`{"filter":{"customer":{"$gt":1}}}`, ExprGt("customer_id", ValI64(1))},
		{`{"filter":{"referenceCode":{"$ne":"DROP"}}}`, ExprNe("reference_code", ValText("DROP"))},
	} {
		t.Run(fixture.source, func(t *testing.T) {
			query, _, err := MergeGeneratedScalarSearch(NewSelectQuery("Invoice"), []byte(fixture.source), generatedSearchFixture(), nil)
			if err != nil || !reflect.DeepEqual(query.Filter, fixture.expected) {
				t.Fatalf("wrong typed predicate: %#v; error=%v", query, err)
			}
		})
	}
}

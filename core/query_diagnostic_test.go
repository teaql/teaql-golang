package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/internal/logprivacy"
)

func TestFutureQueryDiagnosticsAreOwnedAndNotExecutable(t *testing.T) {
	root := NewSelectQuery("Payment").Comment("inspect facets").Purpose("safe diagnostics")
	future := NewSelectQuery("OrderItem").WithFilter(ExprEq("name", ValText("PRIVATE-FUTURE")))
	captured := root.WithDiagnosticQueries(future)
	future.Filter = nil
	if len(captured.ChildEnhancements) != 0 || len(captured.Relations) != 0 {
		t.Fatal("diagnostic-only capture scheduled work")
	}
	if logprivacy.ReadIntentSource(root.DiagnosticOrigin()) != nil {
		t.Fatal("caller query mutated")
	}
	origin, ok := logprivacy.ReadIntentSource(captured.DiagnosticOrigin()).(*SelectQuery)
	if !ok || len(origin.ChildEnhancements) != 1 || origin.ChildEnhancements[0].Filter == nil {
		t.Fatal("future query was not captured independently")
	}
	for _, query := range []*SelectQuery{captured, captured.ForExactCount("n"), captured.Clone()} {
		encoded, err := json.Marshal(query)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), "PRIVATE-FUTURE") || strings.Contains(fmt.Sprintf("%+v %#v", query, query), "PRIVATE-FUTURE") {
			t.Fatal("private diagnostic origin leaked through serialization or formatting")
		}
		if logprivacy.ReadIntentSource(query.DiagnosticOrigin()) == nil {
			t.Fatal("derived query lost diagnostic origin")
		}
	}
}

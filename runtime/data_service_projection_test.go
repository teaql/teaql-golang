package runtime

import (
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestEnsureProjectionPreservesSelectAllAndCompletesNarrowProjection(t *testing.T) {
	selectAll := core.NewSelectQuery("Order Line")
	ensureProjection(selectAll, "customer_order_id")
	if len(selectAll.Projection) != 0 {
		t.Fatalf("empty projection means select-all and must remain empty, got %v", selectAll.Projection)
	}

	narrow := core.NewSelectQuery("Order Line")
	narrow.Projection = []string{"id", "product_name"}
	ensureProjection(narrow, "customer_order_id")
	expected := []string{"id", "product_name", "customer_order_id"}
	if !reflect.DeepEqual(narrow.Projection, expected) {
		t.Fatalf("foreign key must be added to a narrow projection: expected %v, got %v", expected, narrow.Projection)
	}
	ensureProjection(narrow, "customer_order_id")
	if !reflect.DeepEqual(narrow.Projection, expected) {
		t.Fatalf("foreign key must not be duplicated: expected %v, got %v", expected, narrow.Projection)
	}
}

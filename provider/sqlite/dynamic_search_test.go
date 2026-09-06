package sqlite

import (
	"context"
	"database/sql"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDynamicSearchScopedSQLiteExecution(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "search.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, statement := range []string{
		"CREATE TABLE customer_data(id INTEGER PRIMARY KEY, tenant INTEGER, name TEXT, version INTEGER)",
		"CREATE TABLE order_data(id INTEGER PRIMARY KEY, tenant INTEGER, customer INTEGER, name TEXT, version INTEGER)",
		"INSERT INTO customer_data VALUES (1,1,'Ada',1),(2,2,'Ada',1)",
		"INSERT INTO order_data VALUES (1,1,1,'matched',1),(2,2,2,'matched',1)",
	} {
		if _, err = db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	customer := core.NewEntityDescriptor("Customer").TableName("customer_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("tenant", core.TypeI64))
	order := core.NewEntityDescriptor("Order").TableName("order_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).
		Property(core.NewPropertyDescriptor("tenant", core.TypeI64)).
		Property(core.NewPropertyDescriptor("customer", core.TypeU64))
	metadata.Register(customer)
	metadata.Register(order)
	service := runtime.NewRuntimeDataService(metadata, teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata))
	models := map[string]core.SearchModel{
		"Order":    {Fields: map[string]string{"name": "string", "id": "integer"}, Relations: map[string]string{"customer": "Customer"}},
		"Customer": {Fields: map[string]string{"name": "string"}},
	}
	base := core.NewSelectQuery("Order").AndFilter(core.ExprEq("tenant", core.ValI64(1))).Limit(10).OrderAsc("id").
		Comment("what: scoped dynamic search").Purpose("why: verify schema drift behavior")
	original := base.Filter
	query, warnings, err := core.MergeDynamicSearch(base, []byte(`{"filter":{"removed":"SECRET","missing.name":"SECRET","customer.removed":"SECRET","customer.name":"Ada","name":"matched"},"orderBy":[{"field":"removed","direction":"desc"}]}`), models,
		func(path string, predicate map[string]any) (*core.Expr, error) {
			value := core.ValText(predicate["$eq"].(string))
			if path == "customer.name" {
				child := core.NewSelectQuery("Customer").AndFilter(core.ExprEq("tenant", core.ValI64(1))).AndFilter(core.ExprEq("name", value))
				return core.ExprInSubQuery("customer", customer, child, "id"), nil
			}
			return core.ExprEq(path, value), nil
		}, func(path, direction string) (*core.OrderBy, error) { return core.OrderAsc(path), nil }, func(core.DynamicSearchWarning) {})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := service.FetchAll(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["tenant"].V != int64(1) {
		t.Fatalf("scope lost: %#v", rows)
	}
	if len(warnings) != 4 || base.Filter != original || !reflect.DeepEqual(base.Slice, query.Slice) || !reflect.DeepEqual(base.OrderBy, query.OrderBy) {
		t.Fatal("query controls changed")
	}
}

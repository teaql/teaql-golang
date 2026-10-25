package sqlite

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

// #236 EA-01..08: the SQLite result and relation/facet fillers must agree.
func TestEmptyAggregateNullZeroAndBuckets(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		"CREATE TABLE ea_bucket(id INTEGER PRIMARY KEY, version INTEGER)",
		"CREATE TABLE ea_entry(id INTEGER PRIMARY KEY, version INTEGER, bucket INTEGER, amount INTEGER)",
		"INSERT INTO ea_bucket VALUES(1,1),(2,1),(3,1),(4,1)",
		"INSERT INTO ea_entry VALUES(1,1,1,4),(2,1,1,-4),(3,1,2,NULL),(4,1,2,NULL),(5,1,3,NULL),(6,1,3,2),(7,1,3,-1),(8,1,3,NULL)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(core.NewEntityDescriptor("Bucket").TableName("ea_bucket").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Relation(core.NewRelationDescriptor("entries", "Entry").ForeignKey("bucket").Many()))
	metadata.Register(core.NewEntityDescriptor("Entry").TableName("ea_entry").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("bucket", core.TypeI64)).
		Property(core.NewPropertyDescriptor("amount", core.TypeI64)).
		Relation(core.NewRelationDescriptor("bucket", "Bucket").LocalKey("bucket").ForeignKey("id")))
	service := runtime.NewRuntimeDataService(metadata,
		teaqlsql.NewSqlDataServiceExecutor(&SqliteDialect{}, NewSqliteMutationExecutor(db), metadata))
	ctx := runtime.NewUserContext()
	ctx.DisableSqlLog()
	bounded := func(entity string) *core.SelectQuery {
		return core.NewSelectQuery(entity).Limit(10).Comment("inspect empty aggregates").Purpose("verify EA-01 through EA-08")
	}
	read := func(query *core.SelectQuery) []core.Record {
		t.Helper()
		rows, err := service.FetchAll(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	for _, vector := range []struct {
		bucket       int64
		rows, values int64
		total        any
	}{
		{4, 0, 0, nil}, {2, 2, 0, nil}, {3, 4, 2, int64(1)}, {1, 2, 2, int64(0)}, {0, 8, 4, int64(1)},
	} {
		query := bounded("Entry").Count("rows").CountField("amount", "values").Sum("amount", "total")
		if vector.bucket != 0 {
			query.WithFilter(core.ExprEq("bucket", core.ValI64(vector.bucket)))
		}
		rows := read(query)
		if len(rows) != 1 {
			t.Fatalf("scalar aggregate shape: %v", rows)
		}
		for key, expected := range map[string]int64{"rows": vector.rows, "values": vector.values} {
			actual, ok := rows[0][key].TryI64()
			if !ok || actual != expected {
				t.Fatalf("%s: got %v want %v", key, rows[0][key], expected)
			}
		}
		total, present := rows[0]["total"]
		if !present || !reflect.DeepEqual(total.V, vector.total) {
			t.Fatalf("SUM: got %v want %v", rows[0], vector.total)
		}
		fallback := total.V
		if fallback == nil {
			fallback = int64(0)
		}
		if fallback == nil || !reflect.DeepEqual(total.V, vector.total) {
			t.Fatal("fallback mutated raw result")
		}
	}
	grouped := bounded("Entry").Sum("amount", "total").WithGroupBy("bucket").WithFilter(core.ExprEq("bucket", core.ValI64(4)))
	if len(read(grouped)) != 0 {
		t.Fatal("fabricated an empty group")
	}
	parents := bounded("Bucket").OrderAsc("id")
	parents.RelationAggregates = []*core.RelationAggregate{
		core.NewRelationAggregate("entries", "entryCount", bounded("Entry").Count("count"), true),
		core.NewRelationAggregate("entries", "entryTotal", bounded("Entry").Sum("amount", "sum"), true),
	}
	assertBuckets := func(rows []core.Record) {
		t.Helper()
		if len(rows) != 4 {
			t.Fatalf("bucket shape: %v", rows)
		}
		for i, expected := range []any{int64(0), nil, int64(1), nil} {
			actual, present := rows[i]["entryTotal"]
			if !present || !reflect.DeepEqual(actual.V, expected) {
				t.Fatalf("bucket %d sum: %v", i, rows[i])
			}
			count, ok := rows[i]["entryCount"].TryI64()
			if !ok || count != []int64{2, 2, 4, 0}[i] {
				t.Fatalf("bucket %d count: %v", i, rows[i])
			}
		}
	}
	assertBuckets(read(parents))
	choices := bounded("Bucket").OrderAsc("id").Count("entryCount")
	choices.RelationAggregates = parents.RelationAggregates[1:]
	options := core.NewQueryOptions()
	options.Facets = append(options.Facets, core.NewFacetRequest("buckets", "bucket", core.NewQuerySelection(choices), true))
	facets, err := runtime.ExecuteFacets(ctx, service, bounded("Entry"), options)
	if err != nil {
		t.Fatal(err)
	}
	assertBuckets(facets["buckets"].Data)
	t.Log("EMPTY_AGGREGATES_PASS go sqlite EA-01..08")
}

package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
	"testing"
)

func TestNestedPopulationOverTenThousand(t *testing.T) {
	for _, wrapper := range []bool{false, true} {
		t.Run(fmt.Sprint(wrapper), func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
			for _, statement := range []string{
				"CREATE TABLE np_bucket(id INTEGER PRIMARY KEY, version INTEGER)",
				"CREATE TABLE np_entry(id INTEGER PRIMARY KEY, version INTEGER, bucket INTEGER)",
				"WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<10003) INSERT INTO np_bucket SELECT x,1 FROM n",
				"WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<10003) INSERT INTO np_entry SELECT x,1,1 FROM n",
			} {
				if _, err := db.Exec(statement); err != nil {
					t.Fatal(err)
				}
			}
			metadata := runtime.NewInMemoryMetadataStore()
			metadata.Register(core.NewEntityDescriptor("Bucket").TableName("np_bucket").
				Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
				Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
				Relation(core.NewRelationDescriptor("entries", "Entry").ForeignKey("bucket").Many()))
			metadata.Register(core.NewEntityDescriptor("Entry").TableName("np_entry").
				Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
				Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
				Property(core.NewPropertyDescriptor("bucket", core.TypeI64)))
			transport := NewSqliteMutationExecutor(db)
			var executor ds.DataServiceExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
			if wrapper {
				executor = runtime.NewSqlDataServiceExecutor(transport, &SqliteDialect{}, metadata)
			}
			service := runtime.NewRuntimeDataService(metadata, executor)
			ctx := runtime.NewUserContext()
			ctx.DisableSqlLog()
			query := func(entity string) *core.SelectQuery {
				return core.NewSelectQuery(entity).
					Comment("inspect complete nested population").Purpose("avoid silent statistical truncation")
			}
			root := query("Bucket").WithFilter(core.ExprEq("id", core.ValI64(1))).Limit(1).
				RelationQuery("entries", query("Entry"))
			rows, err := service.FetchAll(ctx, root)
			if err != nil {
				t.Fatal(err)
			}
			if got := len(rows[0]["entries"].V.([]core.Record)); got != 10003 {
				t.Fatalf("nested detail count=%d want 10003", got)
			}
			root = query("Bucket").WithFilter(core.ExprEq("id", core.ValI64(1))).Limit(1)
			root.RelationAggregates = []*core.RelationAggregate{core.NewRelationAggregate("entries", "n", query("Entry").Count("n"), true)}
			rows, err = service.FetchAll(ctx, root)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(rows[0]["n"].V) != "10003" {
				t.Fatal("nested count", rows[0]["n"])
			}
			owner, err := ds.NewQueryRequest(query("Bucket").Limit(1))
			if err != nil {
				t.Fatal(err)
			}
			grouped, err := owner.Derive(query("Entry").WithGroupBy("id").Count("n"), nil)
			if err != nil {
				t.Fatal(err)
			}
			statistics, err := executor.(ds.QueryExecutor).Query(ctx, grouped)
			if err != nil {
				t.Fatal(err)
			}
			if len(statistics.Rows) != 10003 {
				t.Fatalf("nested statistical groups=%d want 10003", len(statistics.Rows))
			}
			page, err := service.FetchAll(ctx, query("Entry"))
			if err != nil {
				t.Fatal(err)
			}
			if len(page) != 10000 {
				t.Fatalf("outer limit changed: %d", len(page))
			}
			root = query("Bucket").WithFilter(core.ExprEq("id", core.ValI64(1))).Limit(1).
				RelationQuery("entries", query("Entry").Limit(2))
			rows, err = service.FetchAll(ctx, root)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows[0]["entries"].V.([]core.Record)) != 2 {
				t.Fatal("explicit Top-N changed")
			}
		})
	}
}

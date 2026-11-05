package sqlite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/teaql/teaql-golang/core"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	compiler "github.com/teaql/teaql-golang/sql"
)

func TestBooleanGroupingAliasOwnership(t *testing.T) {
	for _, ddl := range []string{"INTEGER", "BOOL", "BOOLEAN"} {
		for _, explicit := range []bool{false, true} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/explicit=%t/stream=%t", ddl, explicit, stream), func(t *testing.T) {
					db, err := provider.OpenDatabase(":memory:")
					if err != nil {
						t.Fatal(err)
					}
					defer db.Close()
					if _, err = db.Exec("CREATE TABLE probe(id INTEGER PRIMARY KEY, flag " + ddl + ", n INTEGER); INSERT INTO probe VALUES(1,1,7)"); err != nil {
						t.Fatal(err)
					}
					// The Boolean's physical column name is another selected integer's
					// logical output name. Grouping is not ownership of that output.
					entity := core.NewEntityDescriptor("Probe").TableName("probe").Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
						Property(core.NewPropertyDescriptor("active", core.TypeBool).ColumnName("flag")).
						Property(core.NewPropertyDescriptor("flag", core.TypeI64).ColumnName("n"))
					q := core.NewSelectQuery("Probe").WithGroupBy("id").WithGroupBy("active").Limit(1).
						Comment("inspect selected output ownership").Purpose("do not coerce an integer because a group key shares its alias")
					if explicit {
						q.Project("flag")
					}
					compiled, err := (&compiler.DefaultSqlDialect{Dialect: &provider.SqliteDialect{}}).CompileSelect(entity, q)
					if err != nil {
						t.Fatal(err)
					}
					transport := provider.NewSqliteMutationExecutor(db)
					var rows []core.Record
					if stream {
						err = transport.StreamSql(context.Background(), compiled, 1, func(chunk []core.Record) error { rows = append(rows, chunk...); return nil })
					} else {
						rows, err = transport.FetchAllSql(context.Background(), compiled)
					}
					if err != nil || len(rows) != 1 {
						t.Fatalf("group key changed an unrelated integer result: %v", err)
					}
					if n, ok := rows[0]["flag"].TryI64(); !ok || n != 7 {
						t.Fatalf("integer output lost ownership: %#v", rows[0])
					}
					if explicit {
						if _, present := rows[0]["active"]; present {
							t.Fatal("unselected group key was invented")
						}
					} else {
						if b, ok := rows[0]["active"].TryBool(); !ok || !b {
							t.Fatal("selected modeled Boolean lost type")
						}
					}
				})
			}
		}
	}
}

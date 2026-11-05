package sqlite_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	compiler "github.com/teaql/teaql-golang/sql"
)

func TestModeledBooleanRenamedResultColumnsKeepPredicateAffinity(t *testing.T) {
	for _, shape := range []string{"plain", "named", "grouped", "aggregate-projection"} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", shape, stream), func(t *testing.T) {
				db, err := provider.OpenDatabase(":memory:")
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				if _, err = db.Exec("CREATE TABLE probe(active_flag BOOLEAN, n INTEGER); INSERT INTO probe VALUES(0,1),(1,2),(NULL,3)"); err != nil {
					t.Fatal(err)
				}
				entity := core.NewEntityDescriptor("Probe").TableName("probe").Property(core.NewPropertyDescriptor("is_enabled", core.TypeBool).ColumnName("active_flag")).Property(core.NewPropertyDescriptor("n", core.TypeI64))
				q := core.NewSelectQuery("Probe").Limit(10)
				field := "is_enabled"
				switch shape {
				case "plain":
					q.Project("is_enabled")
				case "named":
					q.ProjectExpr("shown", &core.Expr{Type: core.ExprTypeColumn, Column: "is_enabled"})
					field = "shown"
				case "grouped":
					q.WithGroupBy("is_enabled").CountField("n", "count")
					field = "active_flag"
				case "aggregate-projection":
					q.Project("is_enabled").CountField("n", "count")
					field = "active_flag"
				}
				d := &compiler.DefaultSqlDialect{Dialect: &provider.SqliteDialect{}}
				compiled, err := d.CompileSelect(entity, q)
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
				if err != nil || len(rows) == 0 {
					t.Fatal(err)
				}
				for _, row := range rows {
					value, present := row[field]
					if !present {
						t.Fatal("renamed field lost")
					}
					if !value.IsNull() {
						if _, ok := value.TryBool(); !ok {
							t.Fatalf("renamed Boolean lost type: %#v", value)
						}
					}
				}
				if _, err = db.Exec("UPDATE probe SET active_flag=2"); err != nil {
					t.Fatal(err)
				}
				if _, err = transport.FetchAllSql(context.Background(), compiled); err == nil {
					t.Fatal("renamed corrupt Boolean accepted")
				}
				predicate, err := d.CompileSelect(entity, core.NewSelectQuery("Probe").Project("is_enabled").WithFilter(core.ExprEq("is_enabled", core.ValBool(true))).OrderAsc("is_enabled"))
				if err != nil {
					t.Fatal(err)
				}
				column := (&provider.SqliteDialect{}).QuoteIdent("active_flag")
				if !strings.Contains(predicate.Sql, "WHERE ("+column+" = ?)") || !strings.Contains(predicate.Sql, "ORDER BY "+column+" ASC") {
					t.Fatalf("result repair changed predicate/order affinity: %s", predicate.Sql)
				}
			})
		}
	}
}

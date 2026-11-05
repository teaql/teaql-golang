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

func booleanGroupingCase(t *testing.T, explicit bool) {
	t.Helper()
	for _, ddl := range []string{"INTEGER", "BOOL", "BOOLEAN"} {
		for _, renamed := range []bool{false, true} {
			for _, stream := range []bool{false, true} {
				for _, sample := range []string{"0", "1", "NULL", "2"} {
					t.Run(fmt.Sprintf("%s/renamed=%t/stream=%t/value=%s", ddl, renamed, stream, sample), func(t *testing.T) {
						db, err := provider.OpenDatabase(":memory:")
						if err != nil {
							t.Fatal(err)
						}
						defer db.Close()
						column := "active"
						if renamed {
							column = "active_flag"
						}
						// A unique grouping key keeps the selected row functionally dependent
						// on its group: this is not ambiguous arbitrary-row aggregation.
						if _, err = db.Exec("CREATE TABLE probe(id INTEGER PRIMARY KEY, " + column + " " + ddl + "); INSERT INTO probe VALUES(1," + sample + ")"); err != nil {
							t.Fatal(err)
						}
						entity := core.NewEntityDescriptor("Probe").TableName("probe").Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).Property(core.NewPropertyDescriptor("active", core.TypeBool).ColumnName(column))
						q := core.NewSelectQuery("Probe").WithGroupBy("id").Limit(1).Comment("read a bounded grouped Boolean row").Purpose("preserve modeled types without an aggregate function")
						if explicit {
							q.Project("active")
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
						if sample == "2" {
							if err == nil || !strings.Contains(err.Error(), "Boolean") {
								t.Fatalf("group-only corrupt Boolean accepted: %#v, %v", rows, err)
							}
							return
						}
						if err != nil || len(rows) != 1 {
							t.Fatal("group-only read failed", err)
						}
						value, present := rows[0]["active"]
						if !present {
							t.Fatal("selected Boolean disappeared")
						}
						if sample == "NULL" {
							if !value.IsNull() {
								t.Fatal("group-only null lost")
							}
							return
						}
						actual, ok := value.TryBool()
						if !ok || actual != (sample == "1") {
							t.Fatalf("group-only modeled Boolean lost type: %#v", value)
						}
					})
				}
			}
		}
	}
}

func TestBooleanGroupingDefaultProjection(t *testing.T)  { booleanGroupingCase(t, false) }
func TestBooleanGroupingExplicitProjection(t *testing.T) { booleanGroupingCase(t, true) }

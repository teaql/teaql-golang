package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	provider "github.com/teaql/teaql-golang/provider/sqlite"
	compiler "github.com/teaql/teaql-golang/sql"
)

func booleanBoundaryDatabase(t *testing.T, ddl string) (*sql.DB, *provider.SqliteMutationExecutor, *core.EntityDescriptor) {
	t.Helper()
	db, err := provider.OpenDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec("CREATE TABLE boolean_probe(id INTEGER, active " + ddl + ")"); err != nil {
		t.Fatal(err)
	}
	entity := core.NewEntityDescriptor("BooleanProbe").TableName("boolean_probe").
		Property(core.NewPropertyDescriptor("id", core.TypeI64)).Property(core.NewPropertyDescriptor("active", core.TypeBool))
	return db, provider.NewSqliteMutationExecutor(db), entity
}

func booleanBoundaryQuery(t *testing.T, entity *core.EntityDescriptor, shape string) *compiler.CompiledQuery {
	t.Helper()
	q := core.NewSelectQuery("BooleanProbe").Limit(10).Comment("inspect modeled Boolean values").Purpose("prove Boolean storage admission")
	switch shape {
	case "plain":
		q.Project("active")
	case "named":
		q.ProjectExpr("selected", &core.Expr{Type: core.ExprTypeColumn, Column: "active"})
	case "grouped":
		q.WithGroupBy("active").CountField("id", "n")
	default:
		t.Fatal("unknown shape")
	}
	d := &compiler.DefaultSqlDialect{Dialect: &provider.SqliteDialect{}}
	compiled, err := d.CompileSelect(entity, q)
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func booleanBoundaryRows(transport *provider.SqliteMutationExecutor, query *compiler.CompiledQuery, stream bool) ([]core.Record, error) {
	if !stream {
		return transport.FetchAllSql(context.Background(), query)
	}
	var rows []core.Record
	err := transport.StreamSql(context.Background(), query, 1, func(chunk []core.Record) error { rows = append(rows, chunk...); return nil })
	return rows, err
}

func TestBooleanBoundaryRead(t *testing.T) {
	for _, ddl := range []string{"INTEGER", "BOOL", "BOOLEAN"} {
		for _, value := range []string{"2", "-1", "1.5", "'private-corrupt-boolean'", "X'74727565'"} {
			for _, shape := range []string{"plain", "named", "grouped"} {
				for _, stream := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/stream=%t", ddl, value, shape, stream), func(t *testing.T) {
						db, transport, entity := booleanBoundaryDatabase(t, ddl)
						if _, err := db.Exec("INSERT INTO boolean_probe VALUES(1," + value + ")"); err != nil {
							t.Fatal(err)
						}
						rows, err := booleanBoundaryRows(transport, booleanBoundaryQuery(t, entity, shape), stream)
						if err == nil || !strings.Contains(err.Error(), "Boolean") {
							t.Fatalf("invalid stored Boolean accepted: rows=%#v err=%v", rows, err)
						}
						if strings.Contains(err.Error(), "private-corrupt-boolean") {
							t.Fatal("error leaked stored value")
						}
						idQuery, err := (&compiler.DefaultSqlDialect{Dialect: &provider.SqliteDialect{}}).CompileSelect(entity, core.NewSelectQuery("BooleanProbe").Project("id"))
						if err != nil {
							t.Fatal(err)
						}
						if rows, err := transport.FetchAllSql(context.Background(), idQuery); err != nil || len(rows) != 1 {
							t.Fatal("unselected corruption prevented identity-only read", err)
						}
					})
				}
			}
		}
	}
}

func TestBooleanBoundaryReadPositive(t *testing.T) {
	for _, ddl := range []string{"INTEGER", "BOOL", "BOOLEAN"} {
		for _, shape := range []string{"plain", "named", "grouped"} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/stream=%t", ddl, shape, stream), func(t *testing.T) {
					db, transport, entity := booleanBoundaryDatabase(t, ddl)
					if _, err := db.Exec("INSERT INTO boolean_probe VALUES(1,0),(2,1),(3,NULL)"); err != nil {
						t.Fatal(err)
					}
					rows, err := booleanBoundaryRows(transport, booleanBoundaryQuery(t, entity, shape), stream)
					if err != nil || len(rows) != 3 {
						t.Fatal("valid Boolean read", err)
					}
					field := "active"
					if shape == "named" {
						field = "selected"
					}
					seen := map[string]bool{}
					for _, row := range rows {
						value, present := row[field]
						if !present {
							t.Fatal("selected field disappeared")
						}
						if value.IsNull() {
							seen["null"] = true
							continue
						}
						v, ok := value.TryBool()
						if !ok {
							t.Fatalf("modeled Boolean lost type: %#v", value)
						}
						seen[fmt.Sprint(v)] = true
					}
					if len(seen) != 3 {
						t.Fatalf("true/false/null collapsed: %#v", seen)
					}
				})
			}
		}
	}
}

func TestBooleanBoundaryNumericAliasesAndExtrema(t *testing.T) {
	for _, ddl := range []string{"INTEGER", "BOOL", "BOOLEAN"} {
		t.Run(ddl, func(t *testing.T) {
			db, transport, entity := booleanBoundaryDatabase(t, ddl)
			if _, err := db.Exec("INSERT INTO boolean_probe VALUES(1,0),(2,1),(3,NULL)"); err != nil {
				t.Fatal(err)
			}
			d := &compiler.DefaultSqlDialect{Dialect: &provider.SqliteDialect{}}
			for _, q := range []*core.SelectQuery{core.NewSelectQuery("BooleanProbe").CountField("active", "active"), core.NewSelectQuery("BooleanProbe").Min("active", "low").Max("active", "high")} {
				query, err := d.CompileSelect(entity, q)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := transport.FetchAllSql(context.Background(), query)
				if err != nil || len(rows) != 1 {
					t.Fatal(err)
				}
				if _, count := rows[0]["active"]; count {
					if n, ok := rows[0]["active"].TryI64(); !ok || n != 2 {
						t.Fatal("COUNT converted to Boolean")
					}
				} else {
					lo, lok := rows[0]["low"].TryBool()
					hi, hok := rows[0]["high"].TryBool()
					if !lok || !hok || lo || !hi {
						t.Fatal("Boolean MIN/MAX lost type")
					}
				}
			}
		})
	}
}

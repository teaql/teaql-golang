package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

// A forward relation may deliberately use the scalar FK's model name. Its
// filtered result must not destroy the child's independent list membership.
func TestRelationMembershipSurvivesForwardHydration(t *testing.T) {
	for _, wrapper := range []bool{false, true} {
		for _, nested := range []bool{false, true} {
			for _, logging := range []bool{false, true} {
				for _, shape := range []string{"scalar-control", "visible", "filtered", "filtered-sibling"} {
					t.Run(fmt.Sprintf("wrapper=%v/nested=%v/logging=%v/%s", wrapper, nested, logging, shape), func(t *testing.T) {
						t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
						db, err := sql.Open("sqlite3", ":memory:")
						if err != nil {
							t.Fatal(err)
						}
						defer db.Close()
						db.SetMaxOpenConns(1)
						for _, statement := range []string{
							"CREATE TABLE owner_data(id INTEGER PRIMARY KEY, version INTEGER)",
							"CREATE TABLE parent_data(id INTEGER PRIMARY KEY, version INTEGER, owner_id INTEGER, code TEXT, name TEXT)",
							"CREATE TABLE child_data(id INTEGER PRIMARY KEY, version INTEGER, parent_ref TEXT)",
							"INSERT INTO owner_data VALUES (1,1)",
							"INSERT INTO parent_data VALUES (1,1,1,'P-A','PRIVATE-PARENT'),(2,1,1,'P-B','PRIVATE-EMPTY')",
							"INSERT INTO child_data VALUES (1,1,'P-A'),(2,1,'P-A')",
						} {
							if _, err := db.Exec(statement); err != nil {
								t.Fatal(err)
							}
						}
						metadata := runtime.NewInMemoryMetadataStore()
						for name, fields := range map[string][]string{"Owner": {}, "Parent": {"owner_id", "code", "name"}, "Child": {"parent_ref"}} {
							entity := core.NewEntityDescriptor(name).TableName(strings.ToLower(name) + "_data").
								Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
								Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).AuditMaskFields([]string{"name"})
							for _, field := range fields {
								kind := core.TypeText
								if field == "owner_id" {
									kind = core.TypeI64
								}
								entity.Property(core.NewPropertyDescriptor(field, kind))
							}
							switch name {
							case "Owner":
								entity.Relation(core.NewRelationDescriptor("parents", "Parent").ForeignKey("owner_id").Many())
							case "Parent":
								entity.Relation(core.NewRelationDescriptor("children", "Child").LocalKey("code").ForeignKey("parent_ref").Many())
							case "Child":
								for _, relation := range []string{"parent_ref", "parent_again"} {
									entity.Relation(core.NewRelationDescriptor(relation, "Parent").LocalKey("parent_ref").ForeignKey("code"))
								}
							}
							metadata.Register(entity)
						}
						transport := &relationLogTransport{delegate: NewSqliteMutationExecutor(db)}
						var executor ds.DataServiceExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
						if wrapper {
							executor = runtime.NewSqlDataServiceExecutor(transport, &SqliteDialect{}, metadata)
						}
						service := runtime.NewRuntimeDataService(metadata, executor)
						var output bytes.Buffer
						capture := &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&output)}
						ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(capture)
						if !logging {
							ctx.DisableSqlLog()
						}
						forward := core.NewSelectQuery("Parent").WithFilter(core.ExprEq("name", core.ValText("PRIVATE-PARENT")))
						if shape != "visible" {
							forward.WithFilter(core.ExprEq("name", core.ValText("NOT-VISIBLE")))
						}
						child := core.NewSelectQuery("Child").OrderAsc("id").Limit(10).TopNProbeParentThreshold(0)
						if shape != "scalar-control" {
							child.RelationQuery("parent_ref", forward)
						}
						child.RelationAggregates = []*core.RelationAggregate{core.NewRelationAggregate("parent_ref", "parent_count", core.NewSelectQuery("Parent"), true)}
						if shape == "filtered-sibling" {
							child.RelationQuery("parent_again", core.NewSelectQuery("Parent"))
						}
						parents := core.NewSelectQuery("Parent").RelationQuery("children", child).OrderAsc("id").Limit(2)
						query := parents
						if nested {
							query = core.NewSelectQuery("Owner").RelationQuery("parents", parents).Limit(1)
						}
						query.Comment("what: load parents and children").Purpose("why: verify preserved scalar membership")
						rows, err := service.FetchAll(ctx, query)
						if err != nil {
							t.Fatal(err)
						}
						if nested {
							if len(rows) != 1 {
								t.Fatalf("owners: %v", rows)
							}
							rows = rows[0]["parents"].V.([]core.Record)
						}
						if len(rows) != 2 {
							t.Fatalf("parents: %v", rows)
						}
						children := rows[0]["children"].V.([]core.Record)
						if len(children) != 2 || len(rows[1]["children"].V.([]core.Record)) != 0 {
							t.Fatalf("membership lost: %v", rows)
						}
						for _, child := range children {
							if fmt.Sprint(child["parent_count"].V) != "1" {
								t.Fatalf("count lost: %v", child)
							}
							if shape == "scalar-control" {
								if child["parent_ref"].V != "P-A" {
									t.Fatal("scalar key changed")
								}
							} else if shape == "visible" {
								if child["parent_ref"].V.(core.Record)["code"].V != "P-A" {
									t.Fatal("wrong visible reference")
								}
							} else if child["parent_ref"].V != nil {
								t.Fatal("filtered reference must remain null")
							}
							if shape == "filtered-sibling" && child["parent_again"].V.(core.Record)["code"].V != "P-A" {
								t.Fatal("sibling lost original key")
							}
						}
						expected := 4
						if shape == "scalar-control" {
							expected--
						}
						if nested {
							expected++
						}
						if shape == "filtered-sibling" {
							expected++
						}
						if len(transport.reads) != expected {
							t.Fatalf("physical SQL count: %d != %d", len(transport.reads), expected)
						}
						if !logging {
							if len(capture.entries) != 0 || output.Len() != 0 {
								t.Fatal("logging off still writes")
							}
						} else {
							if len(capture.entries) != expected {
								t.Fatal("missing physical SQL evidence")
							}
							for _, entry := range capture.entries {
								if entry.Comment == nil || entry.Purpose == nil || entry.TraceChain[0].Name != query.Entity {
									t.Fatal("intent/root lost")
								}
								if strings.Contains(entry.ParameterizedSQL, "child_data") && !strings.Contains(fmt.Sprint(entry.TraceChain), "children") {
									t.Fatal("child ancestry lost")
								}
							}
							if strings.Contains(output.String(), "PRIVATE-PARENT") {
								t.Fatal("private intent leaked")
							}
							// The final physical query is the count on the same forward
							// edge, not a fabricated edge named after its output alias.
							path := capture.entries[len(capture.entries)-1].TraceChain
							relations := []string{"children", "parent_ref"}
							if nested {
								relations = append([]string{"parents"}, relations...)
							}
							if len(path) != len(relations)+4 {
								t.Fatalf("aggregate ancestry: %+v", path)
							}
							for i, name := range relations {
								if path[i+2].Kind != "relation" || path[i+2].Name != name {
									t.Fatalf("aggregate edge: %+v", path)
								}
							}
						}
						// Original database keys and a subsequent independent request
						// must not inherit this invocation's assembly or trace state.
						capture.entries = nil
						plain, err := service.FetchAll(ctx, core.NewSelectQuery("Child").OrderAsc("id").Limit(2).
							Comment("independent request").Purpose("verify no assembly state escapes"))
						if err != nil || len(plain) != 2 {
							t.Fatalf("independent read: %v %v", plain, err)
						}
						for _, row := range plain {
							if row["parent_ref"].V != "P-A" || len(row) != 3 {
								t.Fatalf("assembly mutated stored/wire state: %v", row)
							}
						}
						if logging && (len(capture.entries) != 1 || len(capture.entries[0].TraceChain) != 4 || capture.entries[0].TraceChain[0].Name != "Child") {
							t.Fatal("independent read inherited ancestry")
						}
					})
				}
			}
		}
	}
}

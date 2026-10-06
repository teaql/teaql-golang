package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

type likeIntentTransport struct {
	tsql.SqlTransport
	reads []*tsql.CompiledQuery
}

func (p *likeIntentTransport) FetchAllSql(ctx context.Context, query *tsql.CompiledQuery) ([]core.Record, error) {
	p.reads = append(p.reads, query)
	return p.SqlTransport.FetchAllSql(ctx, query)
}

type likeIntentPolicy struct {
	runtime.DefaultRequestPolicy
	queries []*core.SelectQuery
}

func (p *likeIntentPolicy) EnforceSelect(_ *runtime.UserContext, query *core.SelectQuery) error {
	p.queries = append(p.queries, query.Clone())
	return nil
}

type likeIntentFixture struct {
	ctx       *runtime.UserContext
	service   *runtime.RuntimeDataService
	transport *likeIntentTransport
	policy    *likeIntentPolicy
	logs      *maskingCapture
	text      bytes.Buffer
}

func newLikeIntentFixture(t *testing.T, operand string) *likeIntentFixture {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		"CREATE TABLE like_owner (id INTEGER PRIMARY KEY, version INTEGER)",
		"CREATE TABLE like_child (id INTEGER PRIMARY KEY, owner_id INTEGER, name TEXT, visible TEXT, version INTEGER)",
		"INSERT INTO like_owner VALUES (1, 1)",
		"INSERT INTO like_child VALUES (2, 1, 'UNRELATED', 'UNRELATED', 1)",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("INSERT INTO like_child VALUES (1, 1, ?, ?, 1)", operand, operand); err != nil {
		t.Fatal(err)
	}
	visible := core.NewPropertyDescriptor("visible", core.TypeText)
	visible.LogPolicy = "plain"
	owner := core.NewEntityDescriptor("Owner").TableName("like_owner").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).AuditMaskFields([]string{}).
		Relation(core.NewRelationDescriptor("children", "LikeChild").Many().LocalKey("id").ForeignKey("owner_id"))
	child := core.NewEntityDescriptor("LikeChild").TableName("like_child").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("owner_id", core.TypeI64)).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).Property(visible).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).AuditMaskFields([]string{"name"})
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(owner)
	metadata.Register(child)
	f := &likeIntentFixture{policy: &likeIntentPolicy{}, transport: &likeIntentTransport{SqlTransport: NewSqliteMutationExecutor(db)}}
	f.logs = &maskingCapture{sink: runtime.NewTextDiagnosticSQLLogSink(&f.text)}
	f.ctx = runtime.NewUserContext().WithDiagnosticSQLLogSink(f.logs).WithRequestPolicy(f.policy)
	f.ctx.Metadata = metadata
	f.service = runtime.NewRuntimeDataService(metadata, tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, f.transport, metadata))
	return f
}

func (f *likeIntentFixture) execute(t *testing.T, query *core.SelectQuery) []core.Record {
	t.Helper()
	prepared, err := f.ctx.PrepareQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := f.service.FetchAll(f.ctx, prepared)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func likeIntentExpr(kind int, field, operand string) *core.Expr {
	return []func(string, string) *core.Expr{
		core.ExprContain, core.ExprNotContain, core.ExprBeginWith,
		core.ExprNotBeginWith, core.ExprEndWith, core.ExprNotEndWith,
	}[kind](field, operand)
}

func likeIntentPattern(kind int, operand string) string {
	if kind < 2 {
		return "%" + operand + "%"
	}
	if kind < 4 {
		return operand + "%"
	}
	return "%" + operand
}

func TestTypedLikeOriginalIntentSQLite(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, future := range []bool{false, true} {
		for _, operand := range []string{"GO-LIKE-PRIVATE", "%GO_LIKE\\_%"} {
			for kind := 0; kind < 6; kind++ {
				for _, masked := range []bool{false, true} {
					for _, logging := range []bool{false, true} {
						t.Run(fmt.Sprintf("future=%v/operand=%s/kind=%d/masked=%v/log=%v", future, operand, kind, masked, logging), func(t *testing.T) {
							f := newLikeIntentFixture(t, operand)
							if !logging {
								f.ctx.DisableSqlLog()
							}
							field := "visible"
							if masked {
								field = "name"
							}
							query := core.NewSelectQuery("LikeChild").WithFilter(likeIntentExpr(kind, field, operand)).Limit(10)
							root := "LikeChild"
							if future {
								root = "Owner"
								query = core.NewSelectQuery(root).RelationQuery("children", query).Limit(1)
							}
							comment, purpose := "load "+operand+" PLAIN-CONTROL", "purpose "+operand+" PLAIN-CONTROL"
							query.Comment(comment).Purpose(purpose)
							before, _ := json.Marshal(query)
							rows := f.execute(t, query)
							if len(rows) != 1 {
								t.Fatalf("row count=%d", len(rows))
							}
							selected := rows[0]
							if future {
								children, ok := rows[0]["children"].V.([]core.Record)
								if !ok || len(children) != 1 {
									t.Fatalf("children=%#v", rows[0]["children"])
								}
								selected = children[0]
							}
							expectedID := int64(1)
							if kind%2 == 1 {
								expectedID = 2
							}
							if id, ok := selected["id"].TryI64(); !ok || id != expectedID {
								t.Fatalf("selected id=%v", selected["id"])
							}
							physicalCount := 1
							if future {
								physicalCount = 2
							}
							if len(f.transport.reads) != physicalCount {
								t.Fatalf("physical count=%d", len(f.transport.reads))
							}
							physical := f.transport.reads[physicalCount-1]
							index := -1
							for i, value := range physical.Params {
								if value.V == likeIntentPattern(kind, operand) {
									index = i
								}
							}
							if index < 0 {
								t.Fatalf("decorated bind changed: %#v", physical.Params)
							}
							policy := "plain"
							if masked {
								policy = "masked"
							}
							if physical.ParameterLogPolicies[index] != policy {
								t.Fatalf("binding policy=%s", physical.ParameterLogPolicies[index])
							}
							if future {
								for _, value := range f.transport.reads[0].Params {
									if value.V == likeIntentPattern(kind, operand) {
										t.Fatal("future source already bound at root")
									}
								}
							}
							after, _ := json.Marshal(query)
							if !bytes.Equal(before, after) {
								t.Fatal("caller query changed")
							}
							if len(f.policy.queries) == 0 {
								t.Fatal("missing policy invocation")
							}
							for _, seen := range f.policy.queries {
								if seen.CommentText == nil || *seen.CommentText != comment || seen.PurposeText == nil || *seen.PurposeText != purpose {
									t.Fatal("policy intent changed")
								}
							}
							wantLogs := physicalCount
							if !logging {
								wantLogs = 0
							}
							if len(f.logs.entries) != wantLogs {
								t.Fatalf("log count=%d", len(f.logs.entries))
							}
							for depth, entry := range f.logs.entries {
								if len(entry.TraceChain) < 4 || entry.TraceChain[0].Name != root {
									t.Fatalf("root trace=%+v", entry.TraceChain)
								}
								relations := 0
								for _, node := range entry.TraceChain {
									if node.Kind == "relation" {
										relations++
										if node.Name != "children" {
											t.Fatal("invented relation")
										}
									}
								}
								if relations != depth {
									t.Fatal("incorrect relation depth")
								}
								if entry.Comment == nil || entry.Purpose == nil || !strings.Contains(*entry.Comment, "PLAIN-CONTROL") || !strings.Contains(*entry.Purpose, "PLAIN-CONTROL") {
									t.Fatal("ordinary intent discarded")
								}
								encoded, _ := json.Marshal(entry)
								if masked && (strings.Contains(string(encoded), operand) || strings.Contains(*entry.Comment, operand) || strings.Contains(*entry.Purpose, operand)) {
									t.Errorf("safe sink leaked original operand at depth %d", depth)
								}
								if !masked && (*entry.Comment != comment || *entry.Purpose != purpose) {
									t.Fatal("plain field over-redacted")
								}
							}
							if masked && logging && strings.Contains(f.text.String(), operand) {
								t.Error("safe text sink leaked original operand")
							}
							if !logging && f.text.Len() != 0 {
								t.Fatal("logging-off emitted text")
							}
							f.logs.entries = nil
							f.text.Reset()
							f.execute(t, core.NewSelectQuery(root).Limit(1).Comment(comment).Purpose(purpose))
							if logging && (len(f.logs.entries) != 1 || *f.logs.entries[0].Comment != comment || *f.logs.entries[0].Purpose != purpose) {
								t.Fatal("privacy escaped into independent request")
							}
						})
					}
				}
			}
		}
	}
}

func TestTypedLikeRawAndRewrittenControlsSQLite(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		for _, shape := range []string{"raw", "rewritten-value", "rewritten-op"} {
			t.Run(fmt.Sprintf("%s/log=%v", shape, logging), func(t *testing.T) {
				const original = "RAW_LITERAL\\"
				f := newLikeIntentFixture(t, original)
				if !logging {
					f.ctx.DisableSqlLog()
				}
				expr := core.ExprBeginWith("name", original)
				pattern := original + "%"
				switch shape {
				case "raw":
					pattern = "%" + original + "%"
					expr = core.ExprLike("name", pattern)
				case "rewritten-value":
					pattern = "UNRELATED%"
					expr.Right.Value = core.ValText(pattern)
				case "rewritten-op":
					expr.Op = core.OpEq
				}
				comment := "ordinary " + original
				rows := f.execute(t, core.NewSelectQuery("LikeChild").WithFilter(expr).Limit(10).Comment(comment).Purpose("raw/stale provenance control"))
				want := 1
				if shape == "rewritten-op" {
					want = 0
				}
				if len(rows) != want {
					t.Fatalf("row count=%d", len(rows))
				}
				if !reflect.DeepEqual(f.transport.reads[0].Params, []core.Value{core.ValText(pattern)}) {
					t.Fatal("raw pattern changed")
				}
				if logging && (len(f.logs.entries) != 1 || *f.logs.entries[0].Comment != comment) {
					t.Fatal("inferred or stale original redaction")
				}
				if !logging && (len(f.logs.entries) != 0 || f.text.Len() != 0) {
					t.Fatal("logging-off emitted")
				}
			})
		}
	}
}

func TestTypedLikeReusableExecutorKeepsCurrentOperandOnlySQLite(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("log=%v", logging), func(t *testing.T) {
			const first, second = "FIRST-LIKE-PRIVATE", "SECOND-LIKE-PRIVATE"
			f := newLikeIntentFixture(t, first)
			if !logging {
				f.ctx.DisableSqlLog()
			}
			for _, pair := range [][2]string{{first, second}, {second, first}} {
				current, ordinary := pair[0], pair[1]
				f.logs.entries = nil
				f.text.Reset()
				f.execute(t, core.NewSelectQuery("LikeChild").WithFilter(core.ExprBeginWith("name", current)).Limit(1).
					Comment("private "+current+"; ordinary "+ordinary).Purpose("purpose "+current+"; ordinary "+ordinary))
				if logging {
					if len(f.logs.entries) != 1 {
						t.Fatal("missing repeated query safe sink")
					}
					entry := f.logs.entries[0]
					if strings.Contains(*entry.Comment, current) || strings.Contains(*entry.Purpose, current) || strings.Contains(f.text.String(), current) {
						t.Fatal("current original leaked")
					}
					if !strings.Contains(*entry.Comment, ordinary) || !strings.Contains(*entry.Purpose, ordinary) {
						t.Fatal("previous query provenance retained")
					}
				} else if len(f.logs.entries) != 0 || f.text.Len() != 0 {
					t.Fatal("logging-off emitted")
				}
			}
			if len(f.transport.reads) != 2 || f.transport.reads[0].Sql != f.transport.reads[1].Sql ||
				!reflect.DeepEqual(f.transport.reads[0].Params, []core.Value{core.ValText(first + "%")}) ||
				!reflect.DeepEqual(f.transport.reads[1].Params, []core.Value{core.ValText(second + "%")}) {
				t.Fatal("same-shape physical rebinding changed")
			}
		})
	}
}

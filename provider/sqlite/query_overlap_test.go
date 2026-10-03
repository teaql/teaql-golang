package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

type pausedQueryRoots struct {
	tsql.SqlTransport
	entered chan struct{}
	release chan struct{}
}

func (p *pausedQueryRoots) FetchAllSql(ctx context.Context, q *tsql.CompiledQuery) ([]core.Record, error) {
	rows, err := p.SqlTransport.FetchAllSql(ctx, q)
	if err != nil {
		return nil, err
	}
	if strings.Contains(strings.ToLower(strings.ReplaceAll(q.Sql, `"`, "")), "from graphroot_data") {
		p.entered <- struct{}{}
		select {
		case <-p.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return rows, nil
}

type queryStatementCapture struct {
	ds.DataServiceExecutor
	query   ds.QueryExecutor
	mu      sync.Mutex
	entries []ds.ExecutionMetadata
}

func (c *queryStatementCapture) Query(ctx context.Context, request *ds.QueryRequest) (*ds.QueryResult, error) {
	result, err := c.query.Query(ctx, request)
	if err == nil {
		c.mu.Lock()
		c.entries = append(c.entries, result.Metadata)
		c.mu.Unlock()
	}
	return result, err
}

func (c *queryStatementCapture) snapshot() []ds.ExecutionMetadata {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]ds.ExecutionMetadata(nil), c.entries...)
}

// TC-REQ-06: unlike simultaneous starts on different Contexts, both real root
// SQL results are held while two queries on the SAME Context remain in flight.
func TestSharedContextLiveQueryGraphs(t *testing.T) {
	for _, wrapper := range []bool{false, true} {
		for _, logging := range []bool{false, true} {
			t.Run(fmt.Sprintf("runtimeExecutor=%v/logging=%v", wrapper, logging), func(t *testing.T) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				db.SetMaxOpenConns(1)
				metadata := runtime.NewInMemoryMetadataStore()
				names := []string{"GraphRoot", "GraphChild", "GraphDetail", "GraphLeaf"}
				for i, name := range names {
					entity := core.NewEntityDescriptor(name).TableName(strings.ToLower(name) + "_data").
						Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
						Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
						Property(core.NewPropertyDescriptor("parent_id", core.TypeI64))
					if i < 3 {
						entity.Relation(core.NewRelationDescriptor("children", names[i+1]).LocalKey("id").ForeignKey("parent_id").Many())
					}
					metadata.Register(entity)
					if _, err := db.Exec("CREATE TABLE " + strings.ToLower(name) + "_data(id INTEGER PRIMARY KEY, version INTEGER, parent_id INTEGER)"); err != nil {
						t.Fatal(err)
					}
				}
				transport := &pausedQueryRoots{SqlTransport: NewSqliteMutationExecutor(db), entered: make(chan struct{}, 2), release: make(chan struct{})}
				var executor ds.DataServiceExecutor = tsql.NewSqlDataServiceExecutor(&SqliteDialect{}, transport, metadata)
				if wrapper {
					executor = runtime.NewSqlDataServiceExecutor(transport, &SqliteDialect{}, metadata)
				}
				ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(nil)
				ctx.Metadata = metadata
				// Native Mutation fixture, not generated-bootstrap acceptance.
				for _, name := range names {
					request, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: core.NewInsertCommand(name).
						Value("id", core.ValI64(1)).Value("version", core.ValI64(1)).Value("parent_id", core.ValI64(1))}, "seed overlap fixture")
					if err != nil {
						t.Fatal(err)
					}
					if _, err := executor.(ds.MutationExecutor).Mutate(ctx, request); err != nil {
						t.Fatal(err)
					}
				}
				capture := &queryStatementCapture{DataServiceExecutor: executor, query: executor.(ds.QueryExecutor)}
				service := runtime.NewRuntimeDataService(metadata, capture)
				logs := runtime.NewSQLExecutionEvidenceStore()
				ctx.WithDiagnosticSQLLogSink(logs)
				if !logging {
					ctx.DisableSqlLog()
				}
				deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				ctx.Context = deadline
				defer cancel()
				var group sync.WaitGroup
				var once sync.Once
				release := func() { once.Do(func() { close(transport.release) }) }
				defer func() { release(); cancel(); group.Wait() }()
				type outcome struct {
					rows []core.Record
					err  error
				}
				results := make(chan outcome, 2)
				for _, label := range []string{"alpha", "beta"} {
					query := core.NewSelectQuery("GraphRoot").Limit(1).Comment("load "+label+" graph").Purpose("render "+label+" graph").
						RelationQuery("children", core.NewSelectQuery("GraphChild").Limit(1).
							RelationQuery("children", core.NewSelectQuery("GraphDetail").Limit(1).
								RelationQuery("children", core.NewSelectQuery("GraphLeaf").Limit(1))))
					group.Add(1)
					go func(q *core.SelectQuery) {
						defer group.Done()
						rows, err := service.FetchAll(ctx, q)
						results <- outcome{rows, err}
					}(query)
				}
				for i := 0; i < 2; i++ {
					select {
					case <-transport.entered:
					case result := <-results:
						t.Fatalf("query escaped live-root barrier: %v", result.err)
					case <-deadline.Done():
						t.Fatal("root overlap timeout")
					}
				}
				if len(capture.snapshot()) != 0 || len(logs.Snapshot()) != 0 || len(results) != 0 {
					t.Fatal("query completed before root barrier release")
				}
				release()
				for i := 0; i < 2; i++ {
					select {
					case result := <-results:
						if result.err != nil {
							t.Fatal(result.err)
						}
						rows := result.rows
						for depth := 0; depth < 3; depth++ {
							if len(rows) != 1 {
								t.Fatal("missing graph record")
							}
							rows = rows[0]["children"].V.([]core.Record)
						}
						if len(rows) != 1 || fmt.Sprint(rows[0]["id"].V) != "1" {
							t.Fatal("missing real leaf")
						}
					case <-deadline.Done():
						t.Fatal("graph completion timeout")
					}
				}
				for index, entries := range [][]ds.ExecutionMetadata{capture.snapshot(), logs.Snapshot()} {
					want := 8
					if index == 1 && !logging {
						want = 0
					}
					if len(entries) != want {
						t.Fatalf("observation %d: got %d statements, want %d", index, len(entries), want)
					}
					for _, label := range []string{"alpha", "beta"} {
						depth := 0
						for _, entry := range entries {
							if entry.Comment == nil || *entry.Comment != "load "+label+" graph" {
								continue
							}
							if entry.Purpose == nil || *entry.Purpose != "render "+label+" graph" || entry.ExecutionOutcome != "success" {
								t.Fatal("intent/outcome contamination")
							}
							kinds, names := []string{}, []string{}
							for _, node := range entry.TraceChain {
								kinds = append(kinds, node.Kind)
								names = append(names, node.Name)
							}
							wantKinds, wantNames := []string{"operation", "request"}, []string{"GraphRoot", "GraphRoot"}
							for d := 0; d < depth; d++ {
								wantKinds = append(wantKinds, "relation")
								wantNames = append(wantNames, "children")
							}
							wantKinds = append(wantKinds, "provider", "sql")
							wantNames = append(wantNames, "sqlite", "select")
							if !reflect.DeepEqual(kinds, wantKinds) || !reflect.DeepEqual(names, wantNames) {
								t.Fatalf("wrong trace: %v %v", kinds, names)
							}
							for d, source := range []string{"GraphRoot", "GraphChild", "GraphDetail"} {
								if d < depth && entry.TraceChain[d+2].Comment != source+".children" {
									t.Fatal("relation trace lost its qualified source")
								}
							}
							depth++
						}
						if depth != want/2 {
							t.Fatalf("missing %s statements", label)
						}
					}
				}
			})
		}
	}
}

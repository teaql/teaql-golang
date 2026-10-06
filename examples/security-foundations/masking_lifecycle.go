package main

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

// Runtime-owned fixture: generated domain-library files remain untouched.
func verifyMaskingLifecycle() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	entity := core.NewEntityDescriptor("MaskCustomer").TableName("mask_customer_data").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("display_name", core.TypeText)).AuditMaskFields([]string{"display_name"})
	ddl, err := (&tsql.DefaultSqlDialect{Dialect: &sqlite.SqliteDialect{}}).CompileCreateTable(entity)
	if err != nil {
		panic(err)
	}
	if _, err = db.Exec(ddl); err != nil {
		panic(err)
	}
	metadata := runtime.NewInMemoryMetadataStore()
	metadata.Register(entity)
	executor := tsql.NewSqlDataServiceExecutor(&sqlite.SqliteDialect{}, sqlite.NewSqliteMutationExecutor(db), metadata)
	var output bytes.Buffer
	ctx := runtime.NewUserContext().WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&output))
	insert := func(id int64) error {
		cmd := core.NewInsertCommand("MaskCustomer").Value("id", core.ValI64(id)).Value("version", core.ValI64(1)).Value("display_name", core.ValText("Riverside"))
		cmd.TraceChain = []*core.TraceNode{core.NewTraceNode("MaskCustomer", nil, "seed lifecycle fixture")}
		request, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: cmd}, "seed lifecycle fixture")
		if err != nil {
			return err
		}
		_, err = executor.Mutate(ctx, request)
		return err
	}
	for _, id := range []int64{1, 2, 3} {
		if err = insert(id); err != nil {
			panic(err)
		}
	}
	output.Reset()
	comment, purpose := "what: inspect customers", "why: verify cursor cleanup"
	q := core.NewSelectQuery("MaskCustomer").AndFilter(core.ExprEq("display_name", core.ValText("Riverside"))).Limit(3).Comment(comment).Purpose(purpose)
	request := &ds.QueryRequest{Query: q, Comment: &comment, Purpose: &purpose}
	stop := errors.New("intentional consumer stop")
	err = executor.QueryStream(ctx, request, 1, func(chunk *ds.StreamChunk) error {
		require(chunk.Rows[0]["display_name"].V == "Riverside", "stream changed original data")
		return stop
	})
	require(errors.Is(err, stop), "stream lost callback error")
	require(db.Stats().InUse == 0, "stream leaked its database connection")
	require(insert(1) != nil, "expected real SQLite duplicate ID failure")
	result, err := executor.Query(ctx, request)
	if err != nil {
		panic(err)
	}
	require(len(result.Rows) == 3 && result.Rows[0]["display_name"].V == "Riverside", "failure changed persisted data")
	text := output.String()
	for _, expected := range []string{"outcome=cancelled", "outcome=failure", "outcome=success", "Ri*****de", "1 rows returned", comment, purpose} {
		require(strings.Contains(text, expected), "missing lifecycle diagnostic: "+expected)
	}
	require(!strings.Contains(text, "Riverside"), "SQL lifecycle leaked masked value")

	// SQLite really accepts the write, then a trigger makes the authoritative
	// snapshot unavailable. Preserve both SQL facts while rolling the write back.
	if _, err = db.Exec("CREATE TRIGGER remove_probe AFTER INSERT ON mask_customer_data WHEN NEW.id = 777 BEGIN DELETE FROM mask_customer_data WHERE id = NEW.id; END"); err != nil {
		panic(err)
	}
	output.Reset()
	evidence := runtime.NewSQLExecutionEvidenceStore()
	ctx.WithRuntimeTelemetrySink(evidence)
	makeInsert := func(id int64) *ds.InsertMutation {
		cmd := core.NewInsertCommand("MaskCustomer").Value("id", core.ValI64(id)).Value("version", core.ValI64(1)).Value("display_name", core.ValText("Riverside"))
		cmd.TraceChain = []*core.TraceNode{core.NewTraceNode("MaskCustomer", nil, "what: insert Riverside for readback verification")}
		comment := "what: insert Riverside for readback verification"
		return &ds.InsertMutation{Cmd: cmd, RootComment: &comment}
	}
	_, err = executor.Mutate(ctx, makeInsert(777))
	require(err != nil, "expected missing authoritative snapshot")
	entries := evidence.Snapshot()
	require(len(entries) == 2, "readback must retain write and read statements")
	require(entries[0].ExecutionOutcome == "success" && entries[0].AffectedRows != nil && *entries[0].AffectedRows == 1, "write fact lost")
	require(entries[1].Operation == ds.OpQuery && entries[1].ExecutionOutcome == "success" && entries[1].ResultCount != nil && *entries[1].ResultCount == 0, "readback zero rows is SQL success, not driver failure")
	require(entries[1].AuditReason != nil && strings.Contains(*entries[1].AuditReason, "what: insert"), "missing inherited intent")
	require(!strings.Contains(output.String(), "Riverside"), "readback intent leaked original masked value")

	// A later failing statement must not erase earlier SQL success or allow the
	// following child to execute. Neither success implies transaction commit.
	evidence = runtime.NewSQLExecutionEvidenceStore()
	ctx.WithRuntimeTelemetrySink(evidence)
	batch, err := ds.NewMutationRequest(&ds.BatchMutation{Mutations: []ds.MutationRequest{makeInsert(30), makeInsert(1), makeInsert(31)}}, "verify partial batch rollback")
	if err != nil {
		panic(err)
	}
	_, err = executor.Mutate(ctx, batch)
	require(err != nil, "expected partial batch duplicate failure")
	entries = evidence.Snapshot()
	require(len(entries) == 3 && entries[0].ExecutionOutcome == "success" && entries[0].Operation == ds.OpInsert &&
		entries[1].ExecutionOutcome == "success" && entries[1].Operation == ds.OpQuery &&
		entries[2].ExecutionOutcome == "failure" && entries[2].Operation == ds.OpInsert, "partial batch must retain successful write/readback before failed write")
	var remaining int
	if err = db.QueryRow("SELECT count(*) FROM mask_customer_data WHERE id IN (30,31,777)").Scan(&remaining); err != nil {
		panic(err)
	}
	require(remaining == 0 && db.Stats().InUse == 0, "failed mutation did not roll back/release connection")

	// Explicit, temporary test authorization. Keep the debug record in memory,
	// revoke authorization, and only then write its safe snapshot to a file.
	const envName = "TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS"
	previous, existed := os.LookupEnv(envName)
	defer func() {
		if existed {
			_ = os.Setenv(envName, previous)
		} else {
			_ = os.Unsetenv(envName)
		}
	}()
	if err = os.Setenv(envName, "I_UNDERSTAND_SENSITIVE_DATA_MAY_BE_WRITTEN_TO_DISK"); err != nil {
		panic(err)
	}
	debugEvidence := runtime.NewSQLExecutionEvidenceStore()
	ctx.WithSensitiveDiagnosticSQLLogSink(debugEvidence)
	_, err = executor.Mutate(ctx, makeInsert(777))
	require(err != nil, "expected debug readback rejection")
	debugEntries := debugEvidence.Snapshot()
	require(len(debugEntries) == 2 && strings.Contains(*debugEntries[1].AuditReason, "Riverside"), "debug fixture did not retain inherited intent")
	if err = os.Unsetenv(envName); err != nil {
		panic(err)
	}
	revoked := debugEvidence.Snapshot()
	require(!strings.Contains(*revoked[1].AuditReason, "Riverside") && strings.Contains(*revoked[1].AuditReason, "what: insert"), "revoked debug intent unsafe or lost")
	file, err := os.CreateTemp("", "teaql-go-revoked-log-*.log")
	if err != nil {
		panic(err)
	}
	runtime.NewTextDiagnosticSQLLogSink(file).WriteSQLLog(revoked[1])
	if err = file.Close(); err != nil {
		panic(err)
	}
	contents, err := os.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}
	require(!strings.Contains(string(contents), "Riverside") && strings.Contains(string(contents), "SELECT"), "revoked file diagnostic unsafe")

	// Exercise a real relation failure through RuntimeDataService, not just a
	// direct SQL executor call. The child binds an FK but inherits sensitive intent.
	entity.Relation(core.NewRelationDescriptor("children", "MaskChild").ForeignKey("parent_id").Many())
	child := core.NewEntityDescriptor("MaskChild").TableName("mask_child_data").
		Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("parent_id", core.TypeI64))
	metadata.Register(child)
	childDDL, err := (&tsql.DefaultSqlDialect{Dialect: &sqlite.SqliteDialect{}}).CompileCreateTable(child)
	if err != nil {
		panic(err)
	}
	if _, err = db.Exec(childDDL); err != nil {
		panic(err)
	}
	childCommand := core.NewInsertCommand("MaskChild").Value("id", core.ValI64(1)).Value("version", core.ValI64(1)).Value("parent_id", core.ValI64(1))
	childCommand.TraceChain = []*core.TraceNode{core.NewTraceNode("MaskChild", nil, "what: seed child for relation masking")}
	childRequest, err := ds.NewMutationRequest(&ds.InsertMutation{Cmd: childCommand}, "what: seed child for relation masking")
	if err != nil {
		panic(err)
	}
	if _, err = executor.Mutate(ctx, childRequest); err != nil {
		panic(err)
	}
	relationFile, err := os.CreateTemp("", "teaql-go-relation-log-*.log")
	if err != nil {
		panic(err)
	}
	defer relationFile.Close()
	relationEvidence := runtime.NewSQLExecutionEvidenceStore()
	ctx.WithSensitiveDiagnosticSQLLogSink(nil).WithRuntimeTelemetrySink(relationEvidence).
		WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(relationFile))
	service := runtime.NewRuntimeDataService(metadata, executor)
	graphQuery := core.NewSelectQuery("MaskCustomer").Project("id").
		AndFilter(core.ExprEq("id", core.ValI64(1))).AndFilter(core.ExprEq("display_name", core.ValText("Riverside"))).
		RelationQuery("children", core.NewSelectQuery("MaskChild").Project("id").Limit(2)).Limit(1).
		Comment("what: load Riverside graph").Purpose("why: verify relation masking")
	func() {
		if _, err = db.Exec("ALTER TABLE mask_child_data RENAME TO mask_child_unavailable"); err != nil {
			panic(err)
		}
		defer func() {
			if _, restoreErr := db.Exec("ALTER TABLE mask_child_unavailable RENAME TO mask_child_data"); restoreErr != nil {
				panic(restoreErr)
			}
		}()
		_, queryErr := service.FetchAll(ctx, graphQuery)
		require(queryErr != nil, "missing relation table unexpectedly succeeded")
	}()
	relationEntries := relationEvidence.Snapshot()
	require(len(relationEntries) == 2 && relationEntries[0].ExecutionOutcome == "success" && relationEntries[1].ExecutionOutcome == "failure", "incorrect relation statement evidence")
	require(!strings.Contains(*relationEntries[1].Comment, "Riverside") && strings.Contains(*relationEntries[1].Comment, "what: load"), "inherited relation intent leaked or lost")
	relationContents, err := os.ReadFile(relationFile.Name())
	if err != nil {
		panic(err)
	}
	require(!strings.Contains(string(relationContents), "Riverside") && strings.Contains(string(relationContents), "mask_child_data"), "relation file log leaked or lost SQL")
	graph, err := service.FetchAll(ctx, graphQuery)
	if err != nil {
		panic(err)
	}
	require(len(graph) == 1, "restored parent missing")
	children := graph[0]["children"].V.([]core.Record)
	require(len(children) == 1 && children[0]["id"].V == int64(1), "restored child graph changed")
}

package postgres_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/provider/mysql"
	"github.com/teaql/teaql-golang/provider/postgres"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

func TestLiveProviderMaskedQAndMutation(t *testing.T) {
	t.Setenv("TEAQL_ALLOW_SENSITIVE_PLAINTEXT_LOGS", "")
	providers := []struct {
		name      string
		env       string
		driver    string
		dialect   teaqlsql.SqlDialect
		transport func(*sql.DB) teaqlsql.SqlTransport
	}{
		{"postgres", "TEAQL_TEST_POSTGRES_DSN", "postgres", &postgres.PostgresDialect{},
			func(db *sql.DB) teaqlsql.SqlTransport { return postgres.NewPgMutationExecutor(db) }},
		{"mysql", "TEAQL_TEST_MYSQL_DSN", "mysql", &mysql.MysqlDialect{},
			func(db *sql.DB) teaqlsql.SqlTransport { return mysql.NewMysqlMutationExecutor(db) }},
	}
	for _, provider := range providers {
		t.Run(provider.name, func(t *testing.T) {
			dsn := os.Getenv(provider.env)
			if dsn == "" {
				if strings.EqualFold(os.Getenv("TEAQL_REQUIRE_LIVE_DB"), "true") {
					t.Fatalf("%s is required for live provider tests", provider.env)
				}
				t.Skip(provider.env + " is not set")
			}
			db, err := sql.Open(provider.driver, dsn)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if err := db.Ping(); err != nil {
				t.Fatal(err)
			}
			table := fmt.Sprintf("teaql_mask_%d", time.Now().UnixNano())
			plain := core.NewPropertyDescriptor("public_address", core.TypeText)
			plain.LogPolicy = "plain"
			entity := core.NewEntityDescriptor("Customer").TableName(table).
				Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
				Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
				Property(core.NewPropertyDescriptor("display_name", core.TypeText)).
				Property(plain).
				Property(core.NewPropertyDescriptor("password_hash", core.TypeText)).
				AuditMaskFields([]string{"display_name", "password_hash"})
			compiler := &teaqlsql.DefaultSqlDialect{Dialect: provider.dialect}
			ddl, err := compiler.CompileCreateTable(entity)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(ddl); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
					t.Error(err)
				}
			}()
			metadata := runtime.NewInMemoryMetadataStore()
			metadata.Register(entity)
			service := teaqlsql.NewSqlDataServiceExecutor(provider.dialect, provider.transport(db), metadata)
			var output bytes.Buffer
			context := runtime.NewUserContext().WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&output))
			insert := core.NewInsertCommand("Customer").Value("id", core.ValI64(1)).
				Value("version", core.ValI64(1)).Value("display_name", core.ValText("Riverside")).
				Value("public_address", core.ValText("1 Runtime Road")).
				Value("password_hash", core.ValText("PASSWORD-CANARY"))
			insert.TraceChain = []*core.TraceNode{core.NewTraceNode("Customer", nil, "what: create masked customer")}
			if _, err := service.Mutate(context, &ds.InsertMutation{Cmd: insert}); err != nil {
				t.Fatal(err)
			}
			comment, purpose := "what: read masked customer", "why: verify live-provider SQL masking"
			query := core.NewSelectQuery("Customer").
				AndFilter(core.ExprEq("display_name", core.ValText("Riverside"))).
				AndFilter(core.ExprEq("public_address", core.ValText("1 Runtime Road"))).
				Limit(1).Comment(comment).Purpose(purpose)
			result, err := service.Query(context, &ds.QueryRequest{Query: query, Comment: &comment, Purpose: &purpose})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Rows) != 1 || result.Rows[0]["display_name"].V != "Riverside" {
				t.Fatalf("unexpected query result count: %d", len(result.Rows))
			}
			logged := output.String()
			for _, required := range []string{"INSERT", "SELECT", "Ri*****de", "1 Runtime Road", comment, purpose} {
				if !strings.Contains(logged, required) {
					t.Fatalf("missing safe SQL fragment %q", required)
				}
			}
			for _, forbidden := range []string{"Riverside", "PASSWORD-CANARY"} {
				if strings.Contains(logged, forbidden) {
					t.Fatalf("SQL log exposed protected value")
				}
			}
		})
	}
}

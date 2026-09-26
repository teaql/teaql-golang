package main

import (
	"bytes"
	stdcontext "context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
	"github.com/teaql/teaql-golang/tfp_endpoint"
)

const goldenToken = "tqr1.AAAAAjMzMzMzMzMzMzMzM3bKiZgRSQQhfIj2cBXRDZIloUGHWLBp8QrXL_aejwIXPFtvV_E71O7wbOXy3cvYo_SwxvuS-89x572T9CO_pDAY4tbjWCNv"

type queryExecutor struct{}

func (*queryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Query: true}
}
func (*queryExecutor) Query(stdcontext.Context, *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return &data_service.QueryResult{}, nil
}

type mutationExecutor struct{ request data_service.MutationRequest }

func (*mutationExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{Mutation: true}
}
func (e *mutationExecutor) Mutate(_ stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	e.request = request
	return &data_service.MutationResult{AffectedRows: 1}, nil
}

func require(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

func verifyLogBoundary() {
	var ordinary, sensitive bytes.Buffer
	context := runtime.NewUserContext().
		WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(&ordinary)).
		WithSensitiveDiagnosticSQLLogSink(runtime.NewSensitiveDiagnosticSQLLogSink(&sensitive))
	debug := `SELECT * FROM customer_order_data WHERE card_number = '4111111111111111'`
	count := 1
	context.RecordExecutionMetadata(data_service.ExecutionMetadata{
		Operation: data_service.OpQuery, ParameterizedSQL: "SELECT * FROM customer_order_data WHERE card_number = ?",
		Parameters: []core.Value{core.ValText("4111111111111111")}, DebugQuery: &debug,
		StartedAt: time.Unix(1, 0), EndedAt: time.Unix(1, 10_000), ResultCount: &count,
	})
	require(!strings.Contains(ordinary.String(), "4111111111111111") && !strings.Contains(ordinary.String(), "Debug SQL:"), "ordinary SQL log leaked a value")
	require(strings.Contains(ordinary.String(), "parameterCount=1"), "ordinary SQL log lost parameter count")
	require(strings.Contains(sensitive.String(), debug), "explicit sensitive sink did not receive copy-paste SQL")
}

func verifyTrustedTFP() {
	mutation := &mutationExecutor{}
	trusted := tfp_endpoint.TrustedFederalContext{
		TenantField: "tenant_id", TenantID: core.ValI64(7), AuthenticatedUser: "example-user", ApprovedPurpose: "example",
		AllowedEntities: map[string]bool{"Order": true},
		ReadableFields:  map[string]map[string]string{"Order": {"id": "id", "status": "status"}},
		WritableFields:  map[string]map[string]string{"Order": {"status": "status"}},
		AllowedActions:  map[string]map[string]bool{"Order": {"Update": true}}, MaxPageSize: 100, MaxOffset: 1000,
	}
	endpoint := tfp_endpoint.NewTfpEndpoint(&queryExecutor{}, mutation).WithTrustedContext(trusted)
	for _, payload := range []string{
		`{"entity":"Order","limitValue":10,"tenantId":99,"commentText":"list","purposeText":"example"}`,
		`{"entity":"Order","limitValue":10,"rawSql":"select * from secrets","commentText":"list","purposeText":"example"}`,
	} {
		_, err := endpoint.HandleQuery(stdcontext.Background(), []byte(payload))
		require(err != nil, "untrusted TFP control was accepted")
	}
	_, err := endpoint.HandleMutation(stdcontext.Background(), []byte(`{"entity":"Order","action":"Update","id":42,"expectedVersion":3,"payload":{"status":"PAID"},"comment":"mark paid"}`))
	if err != nil {
		panic(err)
	}
	update := mutation.request.(*data_service.UpdateMutation).Cmd
	require(update.Guards["tenant_id"].V == int64(7), "trusted tenant guard was not retained")

	entity := core.NewEntityDescriptor("Order").TableName("customer_order_data").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version().NotNull()).
		Property(core.NewPropertyDescriptor("status", core.TypeText).ColumnName("status")).
		Property(core.NewPropertyDescriptor("tenant_id", core.TypeI64).ColumnName("tenant_id"))
	compiled, err := (&teaqlsql.DefaultSqlDialect{Dialect: &sqlite.SqliteDialect{}}).CompileUpdate(entity, update)
	if err != nil {
		panic(err)
	}
	parts := strings.SplitN(compiled.Sql, " WHERE ", 2)
	require(len(parts) == 2 && strings.Contains(parts[1], "id") && strings.Contains(parts[1], "version") && strings.Contains(parts[1], "tenant_id"), "ID, version, and tenant are not guarded in one SQL statement: "+compiled.Sql)
}

func verifyOpaqueReference() {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	codec, err := runtime.NewAEADEntityReferenceCodec(2, map[uint32][]byte{
		1: bytes.Repeat([]byte{0x11}, 32), 2: bytes.Repeat([]byte{0x22}, 32),
	})
	if err != nil {
		panic(err)
	}
	codec.WithClock(func() time.Time { return now }).WithNonceSource(bytes.NewReader(bytes.Repeat([]byte{0x33}, 12)))
	context := runtime.NewUserContext().WithEntityReferenceCodec(codec)
	token, err := context.EncodeEntityReference("OrderItem", 42, 7, "edit-order", time.Hour)
	if err != nil {
		panic(err)
	}
	require(token == goldenToken, "Go token does not match the portable golden vector")
	claims, err := context.DecodeEntityReference(token, "OrderItem", "edit-order")
	if err != nil || claims.ID != 42 || claims.Version != 7 {
		panic("opaque reference did not round-trip")
	}
	_, err = context.DecodeEntityReference(token, "OrderItem", "other-purpose")
	require(err != nil && errors.As(err, new(*runtime.EntityReferenceTokenError)), "wrong-purpose token did not fail closed")
}

func main() {
	verifyLogBoundary()
	verifyTrustedTFP()
	verifyOpaqueReference()
	fmt.Println("PASS Go security foundations example")
}

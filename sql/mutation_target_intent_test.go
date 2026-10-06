package sql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

func TestMutationTargetIntentSurvivesExecutionAndReadback(t *testing.T) {
	// Nonstandard ID property proves metadata is used, not a hard-coded "id".
	model := core.NewEntityDescriptor("Ticket").TableName("ticket_data").
		Property(core.NewPropertyDescriptor("record_key", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("name", core.TypeText))
	id := core.ValI64(7123)
	for _, kind := range []string{"insert", "update", "delete", "recover"} {
		for _, fail := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/failure=%t", kind, fail), func(t *testing.T) {
				var raw ds.MutationRequest
				switch kind {
				case "insert":
					raw = &ds.InsertMutation{Cmd: core.NewInsertCommand("Ticket").Value("record_key", id).Value("version", core.ValI64(1))}
				case "update":
					raw = &ds.UpdateMutation{Cmd: core.NewUpdateCommand("Ticket", id).Value("name", core.ValText("changed")).WithExpectedVersion(1)}
				case "delete":
					raw = &ds.DeleteMutation{Cmd: core.NewDeleteCommand("Ticket", id).WithExpectedVersion(1)}
				case "recover":
					raw = &ds.RecoverMutation{Cmd: core.NewRecoverCommand("Ticket", id, -1)}
				}
				request, err := ds.NewMutationRequest(raw, "change ticket 7123")
				if err != nil {
					t.Fatal(err)
				}
				driverError := errors.New("synthetic driver failure")
				transport := &mockSqlTransport{executeSql: func(context.Context, *CompiledQuery) (uint64, error) {
					if fail {
						return 0, driverError
					}
					return 1, nil
				}}
				executor := NewSqlDataServiceExecutor(&TestDialect{}, transport,
					&mockSchemaProvider{getEntity: func(string) *core.EntityDescriptor { return model }})
				observer := &traceRecorderContext{Context: context.Background()}
				_, err = executor.Mutate(observer, request)
				if (err != nil) != fail {
					t.Fatalf("error=%v expected failure=%t", err, fail)
				}
				if fail && !errors.Is(err, driverError) {
					t.Fatal(err)
				}
				if len(observer.entries) != 1 {
					t.Fatalf("statements=%d", len(observer.entries))
				}
				write := observer.entries[0]
				if !reflect.DeepEqual(logprivacy.ReadIntentSource(write.IntentTargetID), id) {
					t.Fatal("target ID provenance missing")
				}
				if *write.AuditReason != "change ticket 7123" || *request.Comment() != "change ticket 7123" {
					t.Fatal("internal request was pre-redacted")
				}
				if !fail {
					recordMutationReadback(observer, &CompiledQuery{Sql: "SELECT version FROM ticket_data WHERE record_key = ?", Params: []core.Value{id}}, write, time.Now(), 1, nil)
					if !reflect.DeepEqual(logprivacy.ReadIntentSource(observer.entries[1].IntentTargetID), id) {
						t.Fatal("readback lost target ID provenance")
					}
				}
			})
		}
	}
}

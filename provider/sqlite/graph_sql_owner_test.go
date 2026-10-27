package sqlite

import (
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

func TestOptionalModuleUsesSamePhysicalGraphSQLAndCaughtFailureRollsBack(t *testing.T) {
	e := nativeBatchEnvironment(t)
	if _, err := e.db.Exec("CREATE TABLE blob_binding_probe (id INTEGER PRIMARY KEY, key TEXT NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	comment := "save business and module rows"
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		t.Fatal(err)
	}
	writeModule := func(handle *runtime.GraphSQLHandle, id int64) error {
		return handle.WithSQL(func(transport tsql.SqlTransport) error {
			_, err := transport.ExecuteSql(e.ctx, &tsql.CompiledQuery{
				Sql:    "INSERT INTO blob_binding_probe(id, key) VALUES (?, ?)",
				Params: []core.Value{core.ValI64(id), core.ValText("blob_native_key")},
			})
			return err
		})
	}
	businessMutation := func(id int64) ds.MutationRequest {
		t.Helper()
		request, err := ds.NewMutationRequest(nativeBatchLeaf(id, "customer", ""), "save customer in graph")
		if err != nil {
			t.Fatal(err)
		}
		return request
	}
	count := func(table string) int {
		t.Helper()
		var found int
		if err := e.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&found); err != nil {
			t.Fatal(err)
		}
		return found
	}
	var stale *runtime.GraphSQLHandle
	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(handle *runtime.GraphSQLHandle) error {
		stale = handle
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(
			e.ctx, businessMutation(1)); err != nil {
			return err
		}
		return writeModule(handle, 1)
	})
	if err != nil || count("customer_data") != 1 || count("blob_binding_probe") != 1 {
		t.Fatalf("business and module rows must commit together: %v", err)
	}
	if err := writeModule(stale, 9); err == nil || err.Error() != "BLOB_TRANSACTION_REQUIRED" {
		t.Fatalf("stale owner handle accepted: %v", err)
	}

	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(handle *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(
			e.ctx, businessMutation(2)); err != nil {
			return err
		}
		if err := writeModule(handle, 2); err != nil {
			return err
		}
		return errors.New("abort business graph")
	})
	if err == nil || count("customer_data") != 1 || count("blob_binding_probe") != 1 {
		t.Fatalf("business and module rows did not roll back together: %v", err)
	}

	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(handle *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(
			e.ctx, businessMutation(3)); err != nil {
			return err
		}
		_ = handle.WithSQL(func(transport tsql.SqlTransport) error {
			_, err := transport.ExecuteSql(e.ctx, &tsql.CompiledQuery{
				Sql:    "INSERT INTO blob_binding_probe(id, key) VALUES (?, ?)",
				Params: []core.Value{core.ValI64(3), core.ValText("blob_native_key")},
			})
			if err != nil {
				return err
			}
			return errors.New("module rejected binding")
		})
		return nil // The owner still must reject commit.
	})
	if err == nil || count("customer_data") != 1 || count("blob_binding_probe") != 1 {
		t.Fatalf("caught module failure committed the graph: %v", err)
	}

	err = e.ctx.ExecuteGraphSaveWithSQL(intent, func(handle *runtime.GraphSQLHandle) error {
		if _, err := e.ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(
			e.ctx, businessMutation(4)); err != nil {
			return err
		}
		_ = handle.WithSQL(func(transport tsql.SqlTransport) error {
			query := &tsql.CompiledQuery{Sql: "INSERT INTO blob_binding_probe(id, key) VALUES (?, ?)",
				Params: []core.Value{core.ValI64(4), core.ValText("blob_native_key")}}
			if _, err := transport.ExecuteSql(e.ctx, query); err != nil {
				return err
			}
			if _, err := transport.ExecuteSql(e.ctx, query); err == nil {
				t.Error("duplicate module SQL unexpectedly succeeded")
			}
			return nil // Catching the SQL error must still poison the graph.
		})
		return nil
	})
	if err == nil || count("customer_data") != 1 || count("blob_binding_probe") != 1 {
		t.Fatalf("caught module SQL error committed the graph: %v", err)
	}
}

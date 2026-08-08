package runtime_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

type dummyExecutor struct{}

func (e *dummyExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func (e *dummyExecutor) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return &data_service.QueryResult{
		Rows: []core.Record{
			{"id": core.ValI64(1)},
		},
		Metadata: data_service.ExecutionMetadata{
			Backend:   "dummy",
			Operation: data_service.OpQuery,
			StartedAt: time.Now(),
			EndedAt:   time.Now(),
		},
	}, nil
}

type dummyFailingExecutor struct{}

func (e *dummyFailingExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func (e *dummyFailingExecutor) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return nil, errors.New("dummy query error")
}

type dummyNonQueryExecutor struct{}

func (e *dummyNonQueryExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{}
}

func TestRuntimeDataService_FetchAll(t *testing.T) {
	metadata := runtime.NewInMemoryMetadataStore()
	ctx := context.Background()

	t.Run("successful query", func(t *testing.T) {
		executor := &dummyExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(ctx, query)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}

		if rows[0]["id"] != core.ValI64(1) {
			t.Errorf("expected id 1, got %v", rows[0]["id"])
		}
	})

	t.Run("failing query", func(t *testing.T) {
		executor := &dummyFailingExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(ctx, query)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var dsErr *runtime.DataServiceError
		if !errors.As(err, &dsErr) {
			t.Fatalf("expected error to be *runtime.DataServiceError, got %T", err)
		}

		if dsErr.Type != "Executor" {
			t.Errorf("expected error type Executor, got %s", dsErr.Type)
		}

		if rows != nil {
			t.Errorf("expected nil rows, got %v", rows)
		}
	})

	t.Run("unsupported executor", func(t *testing.T) {
		executor := &dummyNonQueryExecutor{}
		svc := runtime.NewRuntimeDataService(metadata, executor)

		query := &core.SelectQuery{}

		rows, err := svc.FetchAll(ctx, query)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "executor does not support Query" {
			t.Errorf("unexpected error message: %v", err)
		}

		if rows != nil {
			t.Errorf("expected nil rows, got %v", rows)
		}
	})
}

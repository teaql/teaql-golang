package runtime

import (
	"context"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type RuntimeDataService struct {
	metadata MetadataStore
	executor data_service.DataServiceExecutor
}

func NewRuntimeDataService(metadata MetadataStore, executor data_service.DataServiceExecutor) *RuntimeDataService {
	return &RuntimeDataService{
		metadata: metadata,
		executor: executor,
	}
}

func (s *RuntimeDataService) FetchAll(ctx context.Context, query *core.SelectQuery) ([]core.Record, error) {
	qExec, ok := s.executor.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("executor does not support Query")
	}

	req := &data_service.QueryRequest{
		Query:      query,
		TraceChain: query.TraceChain,
		Comment:    query.Comment,
	}

	res, err := qExec.Query(ctx, req)
	if err != nil {
		return nil, &DataServiceError{Type: "Executor", ExecutorError: err}
	}

	return res.Rows, nil
}

// TODO: fetch entities etc


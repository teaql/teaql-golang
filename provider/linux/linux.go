package linux

import (
	"context"
	"fmt"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type Collector interface {
	EntityName() string
	CollectAll() ([]core.Record, error)
}

type LinuxDataServiceExecutor struct {
	collectors map[string]Collector
}

func NewLinuxDataServiceExecutor() *LinuxDataServiceExecutor {
	return &LinuxDataServiceExecutor{
		collectors: make(map[string]Collector),
	}
}

func (e *LinuxDataServiceExecutor) WithCollector(collector Collector) *LinuxDataServiceExecutor {
	e.collectors[collector.EntityName()] = collector
	return e
}

func (e *LinuxDataServiceExecutor) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{
		Query:        true,
		Mutation:     false,
		Transaction:  false,
		Schema:       false,
		IdGeneration: false,
	}
}

func (e *LinuxDataServiceExecutor) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	startedAt := time.Now()
	entity := request.Query.Entity

	collector, ok := e.collectors[entity]
	if !ok {
		return nil, fmt.Errorf("unknown entity: %s", entity)
	}

	rows, err := collector.CollectAll()
	if err != nil {
		return nil, err
	}

	// NOTE: filtering and projection should be handled by a memory engine, similar to teaql-rs.
	// For now, we return raw rows.

	resultCount := len(rows)
	return &data_service.QueryResult{
		Rows: rows,
		Metadata: data_service.ExecutionMetadata{
			Backend:      "linux-proc",
			Operation:    data_service.OpQuery,
			StartedAt:    startedAt,
			EndedAt:      time.Now(),
			ResultCount:  &resultCount,
			TraceChain:   request.TraceChain,
			Comment:      request.Comment,
		},
	}, nil
}

func (e *LinuxDataServiceExecutor) Mutate(ctx context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	return nil, fmt.Errorf("linux provider is read-only")
}

package data_service

import (
	"context"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type DataServiceCapabilities struct {
	Query         bool
	Mutation      bool
	Transaction   bool
	Schema        bool
	IdGeneration  bool
	BatchMutation bool
	Returning     bool
}

type QueryRequest struct {
	Query      *core.SelectQuery
	TraceChain []*core.TraceNode
	Comment    *string
}

type QueryResult struct {
	Rows     []core.Record
	Metadata ExecutionMetadata
}

type MutationRequest interface {
	TraceChain() []*core.TraceNode
	Comment() *string
}

type InsertMutation struct {
	Cmd *core.InsertCommand
}

func (m *InsertMutation) TraceChain() []*core.TraceNode { return m.Cmd.TraceChain }
func (m *InsertMutation) Comment() *string {
	if len(m.Cmd.TraceChain) > 0 {
		return &m.Cmd.TraceChain[len(m.Cmd.TraceChain)-1].Comment
	}
	return nil
}

type UpdateMutation struct {
	Cmd *core.UpdateCommand
}

func (m *UpdateMutation) TraceChain() []*core.TraceNode { return m.Cmd.TraceChain }
func (m *UpdateMutation) Comment() *string {
	if len(m.Cmd.TraceChain) > 0 {
		return &m.Cmd.TraceChain[len(m.Cmd.TraceChain)-1].Comment
	}
	return nil
}

type DeleteMutation struct {
	Cmd *core.DeleteCommand
}

func (m *DeleteMutation) TraceChain() []*core.TraceNode { return m.Cmd.TraceChain }
func (m *DeleteMutation) Comment() *string {
	if len(m.Cmd.TraceChain) > 0 {
		return &m.Cmd.TraceChain[len(m.Cmd.TraceChain)-1].Comment
	}
	return nil
}

type RecoverMutation struct {
	Cmd *core.RecoverCommand
}

func (m *RecoverMutation) TraceChain() []*core.TraceNode { return m.Cmd.TraceChain }
func (m *RecoverMutation) Comment() *string {
	if len(m.Cmd.TraceChain) > 0 {
		return &m.Cmd.TraceChain[len(m.Cmd.TraceChain)-1].Comment
	}
	return nil
}

type BatchMutation struct {
	Mutations []MutationRequest
}

func (m *BatchMutation) TraceChain() []*core.TraceNode { return nil }
func (m *BatchMutation) Comment() *string              { return nil }

type MutationResult struct {
	AffectedRows    uint64
	GeneratedValues core.Record
	PersistedRecord core.Record
	Metadata        ExecutionMetadata
}

type DataServiceOperation string

const (
	OpQuery   DataServiceOperation = "Query"
	OpInsert  DataServiceOperation = "Insert"
	OpUpdate  DataServiceOperation = "Update"
	OpDelete  DataServiceOperation = "Delete"
	OpRecover DataServiceOperation = "Recover"
	OpBatch   DataServiceOperation = "Batch"
	OpSchema  DataServiceOperation = "Schema"
)

type ExecutionMetadata struct {
	Backend          string
	Operation        DataServiceOperation
	ParameterizedSQL string
	Parameters       []core.Value
	StartedAt        time.Time
	EndedAt          time.Time
	AffectedRows     *uint64
	ResultCount      *int
	TraceChain       []*core.TraceNode
	Comment          *string
	BackendRequestId *string
	DebugQuery       *string
}

type DataServiceExecutor interface {
	Capabilities() DataServiceCapabilities
}

type QueryExecutor interface {
	DataServiceExecutor
	Query(ctx context.Context, request *QueryRequest) (*QueryResult, error)
}

type StreamChunk struct {
	Rows       []core.Record
	ChunkIndex int
	IsLast     bool
}

type StreamQueryExecutor interface {
	DataServiceExecutor
	QueryStream(ctx context.Context, request *QueryRequest, chunkSize int, yield func(*StreamChunk) error) error
}

type MutationExecutor interface {
	DataServiceExecutor
	Mutate(ctx context.Context, request MutationRequest) (*MutationResult, error)
}

type TransactionExecutor interface {
	DataServiceExecutor
	Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	QueryExecutor
	MutationExecutor
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type SchemaRequest struct {
	EntityName string
}

type SchemaResult struct {
	Changed bool
}

type SchemaExecutor interface {
	DataServiceExecutor
	EnsureSchema(ctx context.Context, request *SchemaRequest) (*SchemaResult, error)
}

type IdGeneratorExecutor interface {
	DataServiceExecutor
	NextId(ctx context.Context, entity string) (uint64, error)
}

type SchemaProvider interface {
	GetEntity(name string) *core.EntityDescriptor
}

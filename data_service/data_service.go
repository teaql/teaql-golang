package data_service

import (
	stdcontext "context"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/internal/logprivacy"
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
	Purpose    *string
	// Internal invocation provenance, never serialized into a request or log.
	InheritedIntent logprivacy.IntentSource `json:"-"`
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
	// ExecutionOutcome describes statement/cursor termination, not transaction commit.
	// Empty means the producer has not supplied an outcome.
	ExecutionOutcome     string
	Backend              string
	Operation            DataServiceOperation
	ParameterizedSQL     string
	Parameters           []core.Value
	ParameterCount       int
	StartedAt            time.Time
	EndedAt              time.Time
	AffectedRows         *uint64
	ResultCount          *int
	TraceChain           []*core.TraceNode
	Comment              *string
	Purpose              *string
	AuditReason          *string
	BackendRequestId     *string
	DebugQuery           *string
	ParameterLogPolicies []string
	GeneratedSQL         bool
	MaskedParameters     []bool
	LogMode              string
	OmissionReason       string
	// Internal runtime plumbing, not a workspace/wire contract. Removed by safe projection.
	InheritedIntent logprivacy.IntentSource `json:"-"`
	// Mutation target ID is private projection provenance for free-text SQL intent only.
	IntentTargetID logprivacy.IntentSource `json:"-"`
	// Carries only an immutable safe alternative for debug revocation; not a wire API.
	LogProjection logprivacy.ProjectionState `json:"-"`
}

type DataServiceExecutor interface {
	Capabilities() DataServiceCapabilities
}

type QueryExecutor interface {
	DataServiceExecutor
	Query(context stdcontext.Context, request *QueryRequest) (*QueryResult, error)
}

type StreamChunk struct {
	Rows       []core.Record
	ChunkIndex int
	IsLast     bool
}

type StreamQueryExecutor interface {
	DataServiceExecutor
	QueryStream(context stdcontext.Context, request *QueryRequest, chunkSize int, yield func(*StreamChunk) error) error
}

type MutationExecutor interface {
	DataServiceExecutor
	Mutate(context stdcontext.Context, request MutationRequest) (*MutationResult, error)
}

type TransactionExecutor interface {
	DataServiceExecutor
	Begin(context stdcontext.Context) (Transaction, error)
}

type Transaction interface {
	QueryExecutor
	MutationExecutor
	Commit(context stdcontext.Context) error
	Rollback(context stdcontext.Context) error
}

type IdGeneratorExecutor interface {
	DataServiceExecutor
	NextId(context stdcontext.Context, entity string) (uint64, error)
}

type SchemaProvider interface {
	GetEntity(name string) *core.EntityDescriptor
}

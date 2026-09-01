package runtime

import (
	stdcontext "context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type ContinuousPageCursor struct {
	CursorID   string
	QueryKey   string
	Entity     string
	Direction  core.SortDirection
	Boundary   core.Value
	PageSize   uint64
	NextOffset uint64
	ExpiresAt  time.Time
}

type ContinuousPageCursorStore interface {
	GetContinuousPageCursor(context stdcontext.Context, queryKey string, targetOffset uint64) (*ContinuousPageCursor, error)
	PutContinuousPageCursor(context stdcontext.Context, cursor *ContinuousPageCursor) error
	InvalidateContinuousPageCursor(context stdcontext.Context, queryKey string) error
}

type InMemoryContinuousPageCursorStore struct {
	mu         sync.Mutex
	cursors    map[string]*ContinuousPageCursor
	maxEntries int
}

type RetainedIDSet struct {
	QueryKey  string
	IDs       []uint64
	ExpiresAt time.Time
}

type IDSetStore interface {
	GetIDSet(context stdcontext.Context, queryKey string) (*RetainedIDSet, error)
	PutIDSet(context stdcontext.Context, retained *RetainedIDSet) error
	InvalidateIDSet(context stdcontext.Context, queryKey string) error
}

type InMemoryIDSetStore struct {
	mu         sync.Mutex
	sets       map[string]*RetainedIDSet
	maxEntries int
	maxBytes   uint64
}

var defaultIDSetStore IDSetStore = NewInMemoryIDSetStore()

func NewInMemoryIDSetStore() *InMemoryIDSetStore {
	return &InMemoryIDSetStore{sets: make(map[string]*RetainedIDSet), maxEntries: 64, maxBytes: 256 << 20}
}

func (s *InMemoryIDSetStore) GetIDSet(_ stdcontext.Context, queryKey string) (*RetainedIDSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	retained := s.sets[queryKey]
	if retained == nil {
		return nil, nil
	}
	if !retained.ExpiresAt.After(time.Now()) {
		delete(s.sets, queryKey)
		return nil, nil
	}
	copySet := *retained
	copySet.IDs = append([]uint64(nil), retained.IDs...)
	return &copySet, nil
}

func (s *InMemoryIDSetStore) PutIDSet(_ stdcontext.Context, retained *RetainedIDSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copySet := *retained
	copySet.IDs = append([]uint64(nil), retained.IDs...)
	if uint64(len(copySet.IDs))*8 > s.maxBytes {
		return fmt.Errorf("retained ID set exceeds store memory ceiling")
	}
	for len(s.sets) >= s.maxEntries || retainedIDSetBytes(s.sets)+uint64(len(copySet.IDs))*8 > s.maxBytes {
		var oldestKey string
		var oldest time.Time
		for key, value := range s.sets {
			if oldestKey == "" || value.ExpiresAt.Before(oldest) {
				oldestKey, oldest = key, value.ExpiresAt
			}
		}
		if oldestKey == "" {
			break
		}
		delete(s.sets, oldestKey)
	}
	s.sets[copySet.QueryKey] = &copySet
	return nil
}

func retainedIDSetBytes(sets map[string]*RetainedIDSet) uint64 {
	var total uint64
	for _, retained := range sets {
		total += uint64(len(retained.IDs)) * 8
	}
	return total
}

func (s *InMemoryIDSetStore) InvalidateIDSet(_ stdcontext.Context, queryKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sets, queryKey)
	return nil
}

func NewInMemoryContinuousPageCursorStore() *InMemoryContinuousPageCursorStore {
	return &InMemoryContinuousPageCursorStore{cursors: make(map[string]*ContinuousPageCursor), maxEntries: 4096}
}

func continuousPageCheckpointKey(queryKey string, offset uint64) string {
	return fmt.Sprintf("%s:%d", queryKey, offset)
}

func (s *InMemoryContinuousPageCursorStore) GetContinuousPageCursor(_ stdcontext.Context, queryKey string, targetOffset uint64) (*ContinuousPageCursor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := continuousPageCheckpointKey(queryKey, targetOffset)
	cursor := s.cursors[key]
	if cursor != nil && !cursor.ExpiresAt.After(time.Now()) {
		delete(s.cursors, key)
		return nil, nil
	}
	return cursor, nil
}

func (s *InMemoryContinuousPageCursorStore) PutContinuousPageCursor(_ stdcontext.Context, cursor *ContinuousPageCursor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.cursors) >= s.maxEntries {
		var oldestKey string
		var oldest time.Time
		for key, value := range s.cursors {
			if oldestKey == "" || value.ExpiresAt.Before(oldest) {
				oldestKey, oldest = key, value.ExpiresAt
			}
		}
		delete(s.cursors, oldestKey)
	}
	s.cursors[continuousPageCheckpointKey(cursor.QueryKey, cursor.NextOffset)] = cursor
	return nil
}

func (s *InMemoryContinuousPageCursorStore) InvalidateContinuousPageCursor(_ stdcontext.Context, queryKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := queryKey + ":"
	for key := range s.cursors {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(s.cursors, key)
		}
	}
	return nil
}

type localCacheEntry struct {
	value     any
	expiresAt time.Time
}

var processLocalCache sync.Map

type localLockEntry struct {
	owner     *UserContext
	expiresAt time.Time
}

var processLocalLocks = struct {
	sync.Mutex
	entries map[string]localLockEntry
}{entries: make(map[string]localLockEntry)}

type DataStore interface {
	Get(context stdcontext.Context, key string) (core.Value, bool)
	Put(context stdcontext.Context, key string, value core.Value, timeoutSeconds *uint64)
	Remove(context stdcontext.Context, key string)
}

type InMemoryDataStore struct {
	cache map[string]core.Value
}

func NewInMemoryDataStore() *InMemoryDataStore {
	return &InMemoryDataStore{
		cache: make(map[string]core.Value),
	}
}

func (s *InMemoryDataStore) Get(context stdcontext.Context, key string) (core.Value, bool) {
	val, ok := s.cache[key]
	return val, ok
}

func (s *InMemoryDataStore) Put(context stdcontext.Context, key string, value core.Value, timeoutSeconds *uint64) {
	s.cache[key] = value
}

func (s *InMemoryDataStore) Remove(context stdcontext.Context, key string) {
	delete(s.cache, key)
}

type GraphNode struct {
	Entity       string
	Values       core.Record
	RelationsMap map[string][]*GraphNode
	IsDeleted    bool
	CommentText  string
}

func (n *GraphNode) Child(rel string) *GraphNode {
	if n.RelationsMap == nil {
		n.RelationsMap = make(map[string][]*GraphNode)
	}
	child := &GraphNode{Entity: rel, Values: make(core.Record)}
	n.RelationsMap[rel] = append(n.RelationsMap[rel], child)
	return child
}

func (n *GraphNode) Relation(rel string) *GraphNode {
	return n.Child(rel)
}

func (n *GraphNode) Relations() map[string][]*GraphNode {
	return n.RelationsMap
}

func (n *GraphNode) Remove(rel string) {
	if n.RelationsMap != nil {
		delete(n.RelationsMap, rel)
	}
}

func (n *GraphNode) Value(field string) any {
	if n.Values != nil {
		if v, ok := n.Values[field]; ok {
			return v.V
		}
	}
	return nil
}

func (n *GraphNode) Comment(text string) *GraphNode {
	n.CommentText = text
	return n
}

func (n *GraphNode) SetComment(text string) {
	n.CommentText = text
}

func (n *GraphNode) Id() any {
	return n.Value("id")
}

func (n *GraphNode) Operation() core.EntityGraphOperation {
	if n.IsDeleted {
		return core.EntityGraphOpDelete
	}
	return core.EntityGraphOpSave
}

func (n *GraphNode) Reference(rel string, refId any) *GraphNode {
	child := n.Child(rel)
	child.Values["id"] = core.Value{V: refId}
	return child
}

type UserContext struct {
	stdcontext.Context
	Metadata       MetadataStore
	EntityRegistry EntityRegistry
	Behaviors      EntityDataServiceBehaviorRegistry

	initialGraphs             []*GraphNode
	rootGraphs                []*GraphNode
	resources                 map[string]interface{}
	standardAuditSink         RawAuditEventSink
	appAuditSink              AppAuditEventSink
	runtimeTelemetrySink      RuntimeTelemetrySink
	diagnosticSQLLogSink      DiagnosticSQLLogSink
	querySQLLogEnabled        bool
	mutationSQLLogEnabled     bool
	runtimeTelemetry          RuntimeTelemetry
	continuousPageCursorStore ContinuousPageCursorStore
	idSetStore                IDSetStore
	idSetMu                   sync.Mutex
	idSetPlan                 string
	idSetCount                uint64
	idSetCountAccuracy        string
	continuousPageMu          sync.Mutex
	continuousPagePlan        string
	continuousPageCursorID    string
	userIdentifier            string
	requestPolicy             RequestPolicy
	checkerRegistry           CheckerRegistry
	graphSaveGate             sync.Mutex
	graphSaveMu               sync.Mutex
	graphSaveActive           bool
	graphCommitActions        []func()
	graphRollbackActions      []func()
	graphFixTime              time.Time
	currentFixEvidence        []FixEvidence
	lastFixEvidence           []FixEvidence
}

type FixEvidenceSource string

const (
	FixEvidenceClock   FixEvidenceSource = "clock"
	FixEvidenceContext FixEvidenceSource = "context"
)

type FixEvidence struct {
	EntityType  string
	ModelPath   string
	Source      FixEvidenceSource
	SourceLabel string
}

// CheckAndFixInput is the provider-independent mutation state seen by model checkers.
// Values may be amended by declared fixes before SQL compilation.
type CheckAndFixInput struct {
	Entity    string
	Operation core.MutationKind
	Values    core.Record
	OldValues core.Record
	Now       time.Time
}

// EntityReference is trusted context-owned identity, never a raw magic number.
type EntityReference struct {
	Entity string
	ID     uint64
}

type ContextRootError struct {
	Reason       string
	ExpectedType string
	ActiveRoot   *EntityReference
}

func (e *ContextRootError) Error() string {
	return fmt.Sprintf("context root %s: expected %s", e.Reason, e.ExpectedType)
}

type CheckerRegistry interface {
	CheckAndFix(context *UserContext, input *CheckAndFixInput) []CheckResult
}

type userContextKey struct{}

// UserContextFrom preserves access to TeaQL services when an instrumentation
// adapter derives a child context for trace propagation.
func UserContextFrom(context stdcontext.Context) (*UserContext, bool) {
	if value, ok := context.(*UserContext); ok {
		return value, true
	}
	value, ok := context.Value(userContextKey{}).(*UserContext)
	return value, ok
}

type RuntimeTelemetrySink interface {
	RecordExecutionMetadata(data_service.ExecutionMetadata)
}

// DiagnosticSQLLogSink receives value-bearing operator logs. The text sink is
// installed by default and can be disabled independently for queries/mutations.
type DiagnosticSQLLogSink interface {
	WriteSQLLog(data_service.ExecutionMetadata)
}

type TextDiagnosticSQLLogSink struct {
	mu     sync.Mutex
	writer io.Writer
}

func sqlLogText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func NewTextDiagnosticSQLLogSink(writer io.Writer) *TextDiagnosticSQLLogSink {
	return &TextDiagnosticSQLLogSink{writer: writer}
}

func (s *TextDiagnosticSQLLogSink) WriteSQLLog(metadata data_service.ExecutionMetadata) {
	if s == nil || s.writer == nil || metadata.DebugQuery == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	duration := metadata.EndedAt.Sub(metadata.StartedAt).Microseconds()
	summary := ""
	if metadata.ResultCount != nil {
		summary = fmt.Sprintf("%d rows returned", *metadata.ResultCount)
	} else if metadata.AffectedRows != nil {
		summary = fmt.Sprintf("%d rows affected", *metadata.AffectedRows)
	}
	fmt.Fprintf(s.writer, "[TeaQL SQL][%s][%dus] %s comment=%q purpose=%q auditReason=%q tracePath=%v\nParameterized SQL: %s params=%v\nDebug SQL: %s\n",
		strings.ToLower(string(metadata.Operation)), duration, summary,
		sqlLogText(metadata.Comment), sqlLogText(metadata.Purpose), sqlLogText(metadata.AuditReason), metadata.TraceChain,
		metadata.ParameterizedSQL, metadata.Parameters, *metadata.DebugQuery)
}

type SQLExecutionEvidenceMode int

const (
	SQLExecutionEvidenceAll SQLExecutionEvidenceMode = iota
	SQLExecutionEvidenceQuery
	SQLExecutionEvidenceMutation
	SQLExecutionEvidenceDisabled
)

type SQLExecutionEvidenceStore struct {
	mu      sync.Mutex
	mode    SQLExecutionEvidenceMode
	entries []data_service.ExecutionMetadata
}

func NewSQLExecutionEvidenceStore() *SQLExecutionEvidenceStore {
	return &SQLExecutionEvidenceStore{mode: SQLExecutionEvidenceAll}
}

func (s *SQLExecutionEvidenceStore) RecordExecutionMetadata(metadata data_service.ExecutionMetadata) {
	s.mu.Lock()
	defer s.mu.Unlock()
	isQuery := metadata.Operation == data_service.OpQuery
	if s.mode == SQLExecutionEvidenceDisabled ||
		(s.mode == SQLExecutionEvidenceQuery && !isQuery) ||
		(s.mode == SQLExecutionEvidenceMutation && isQuery) {
		return
	}
	metadata.Parameters = append([]core.Value(nil), metadata.Parameters...)
	s.entries = append(s.entries, metadata)
}

func (s *SQLExecutionEvidenceStore) setMode(mode SQLExecutionEvidenceMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode, s.entries = mode, nil
}

func (s *SQLExecutionEvidenceStore) EnableAll()      { s.setMode(SQLExecutionEvidenceAll) }
func (s *SQLExecutionEvidenceStore) EnableQuery()    { s.setMode(SQLExecutionEvidenceQuery) }
func (s *SQLExecutionEvidenceStore) EnableMutation() { s.setMode(SQLExecutionEvidenceMutation) }
func (s *SQLExecutionEvidenceStore) Disable()        { s.setMode(SQLExecutionEvidenceDisabled) }
func (s *SQLExecutionEvidenceStore) Snapshot() []data_service.ExecutionMetadata {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]data_service.ExecutionMetadata, len(s.entries))
	copy(result, s.entries)
	for i := range result {
		result[i].Parameters = append([]core.Value(nil), result[i].Parameters...)
	}
	return result
}

func (c *UserContext) WithRuntimeTelemetrySink(sink RuntimeTelemetrySink) *UserContext {
	c.runtimeTelemetrySink = sink
	return c
}

func (c *UserContext) WithDiagnosticSQLLogSink(sink DiagnosticSQLLogSink) *UserContext {
	c.diagnosticSQLLogSink = sink
	return c
}

func (c *UserContext) RecordExecutionMetadata(metadata data_service.ExecutionMetadata) {
	isQuery := metadata.Operation == data_service.OpQuery
	if (isQuery && !c.querySQLLogEnabled) || (!isQuery && !c.mutationSQLLogEnabled) {
		return
	}
	if c.runtimeTelemetrySink != nil {
		c.runtimeTelemetrySink.RecordExecutionMetadata(metadata)
	}
	if c.diagnosticSQLLogSink != nil {
		c.diagnosticSQLLogSink.WriteSQLLog(metadata)
	}
}

func (c *UserContext) WithRuntimeTelemetry(telemetry RuntimeTelemetry) *UserContext {
	if telemetry == nil {
		telemetry = NoopRuntimeTelemetry{}
	}
	c.runtimeTelemetry = telemetry
	return c
}

func (c *UserContext) RuntimeTelemetry() RuntimeTelemetry {
	if c.runtimeTelemetry == nil {
		return NoopRuntimeTelemetry{}
	}
	return c.runtimeTelemetry
}

func NewUserContext() *UserContext {
	context := &UserContext{
		Context:                   stdcontext.Background(),
		resources:                 make(map[string]interface{}),
		continuousPageCursorStore: NewInMemoryContinuousPageCursorStore(),
		idSetStore:                defaultIDSetStore,
		idSetPlan:                 "ID_SET_DISABLED",
		idSetCountAccuracy:        "UNKNOWN",
		continuousPagePlan:        "DISABLED",
		userIdentifier:            "main",
		runtimeTelemetry:          NoopRuntimeTelemetry{},
		diagnosticSQLLogSink:      NewTextDiagnosticSQLLogSink(os.Stderr),
		querySQLLogEnabled:        true,
		mutationSQLLogEnabled:     true,
	}
	context.Context = stdcontext.WithValue(context.Context, userContextKey{}, context)
	return context
}

func (c *UserContext) SetIDSetStore(store IDSetStore) {
	if store == nil {
		panic("ID set store must not be nil")
	}
	c.idSetStore = store
}

func (c *UserContext) IDSetStore() IDSetStore { return c.idSetStore }

func (c *UserContext) ObserveIDSet(plan, accuracy string, count uint64) {
	c.idSetMu.Lock()
	defer c.idSetMu.Unlock()
	c.idSetPlan, c.idSetCountAccuracy, c.idSetCount = plan, accuracy, count
}

func (c *UserContext) IDSetPlan() string {
	c.idSetMu.Lock()
	defer c.idSetMu.Unlock()
	return c.idSetPlan
}

func (c *UserContext) IDSetCount() (uint64, string) {
	c.idSetMu.Lock()
	defer c.idSetMu.Unlock()
	return c.idSetCount, c.idSetCountAccuracy
}

func (c *UserContext) SetContinuousPageCursorStore(store ContinuousPageCursorStore) {
	if store == nil {
		panic("continuous page cursor store must not be nil")
	}
	c.continuousPageCursorStore = store
}

func (c *UserContext) ContinuousPageCursorStore() ContinuousPageCursorStore {
	return c.continuousPageCursorStore
}

func (c *UserContext) ObserveContinuousPage(plan, cursorID string) {
	c.continuousPageMu.Lock()
	defer c.continuousPageMu.Unlock()
	c.continuousPagePlan, c.continuousPageCursorID = plan, cursorID
}

func (c *UserContext) ContinuousPagePlan() string {
	c.continuousPageMu.Lock()
	defer c.continuousPageMu.Unlock()
	return c.continuousPagePlan
}

func (c *UserContext) ContinuousPageCursorID() string {
	c.continuousPageMu.Lock()
	defer c.continuousPageMu.Unlock()
	return c.continuousPageCursorID
}

func (c *UserContext) InitialGraphs() []*GraphNode {
	return c.initialGraphs
}

func (c *UserContext) RootGraphs() []*GraphNode { return c.rootGraphs }

func (c *UserContext) SetRootGraphs(graphs []*GraphNode) { c.rootGraphs = graphs }

func (c *UserContext) SetInitialGraphs(graphs []*GraphNode) {
	c.initialGraphs = graphs
}

func (c *UserContext) AllEntities() []*core.EntityDescriptor {
	if c.Metadata != nil {
		return c.Metadata.AllEntities()
	}
	return nil
}

func (c *UserContext) Entity(name string) *core.EntityDescriptor {
	if c.Metadata != nil {
		return c.Metadata.Entity(name)
	}
	return nil
}

func (c *UserContext) SetSchemaProvider(provider data_service.SchemaProvider) {
	// For compatibility with old interface, though MetadataStore replaces it
}

func (c *UserContext) InsertResource(name string, resource interface{}) {
	c.resources[name] = resource
}

func (c *UserContext) GetResource(name string) interface{} {
	return c.resources[name]
}

// ExecuteGraphSave coordinates one generated object graph through one provider
// transaction. Nested entity saves join the active graph transaction.
func (c *UserContext) ExecuteGraphSave(work func() error) error {
	// Generated child entities use their within-graph save entry point directly.
	// Every public/root save is serialized here, preventing an unrelated goroutine
	// from joining whichever transaction happens to be active on this context.
	c.graphSaveGate.Lock()
	defer c.graphSaveGate.Unlock()
	c.graphSaveMu.Lock()
	provider := c.resources["dataService"]
	originalIDGenerator, hadIDGenerator := c.resources["idGenerator"]
	executor, ok := provider.(data_service.TransactionExecutor)
	if !ok {
		c.graphSaveMu.Unlock()
		return fmt.Errorf("configured dataService does not support graph transactions")
	}
	transaction, err := executor.Begin(c)
	if err != nil {
		c.graphSaveMu.Unlock()
		return err
	}
	c.graphSaveActive = true
	c.graphCommitActions = nil
	c.graphRollbackActions = nil
	c.graphFixTime = time.Now()
	c.currentFixEvidence = nil
	c.resources["dataService"] = transaction
	if _, ok := transaction.(interface{ GenerateId(string) (uint64, error) }); ok {
		c.resources["idGenerator"] = transaction
	}
	c.graphSaveMu.Unlock()

	err = work()
	if err == nil {
		err = transaction.Commit(c)
	}
	if err != nil {
		rollbackErr := transaction.Rollback(c)
		for index := len(c.graphRollbackActions) - 1; index >= 0; index-- {
			c.graphRollbackActions[index]()
		}
		if rollbackErr != nil {
			err = fmt.Errorf("graph save failed: %w; rollback failed: %v", err, rollbackErr)
		}
	} else {
		for _, action := range c.graphCommitActions {
			action()
		}
	}

	c.graphSaveMu.Lock()
	c.resources["dataService"] = provider
	if hadIDGenerator {
		c.resources["idGenerator"] = originalIDGenerator
	} else {
		delete(c.resources, "idGenerator")
	}
	c.graphSaveActive = false
	c.graphCommitActions = nil
	c.graphRollbackActions = nil
	c.graphFixTime = time.Time{}
	c.lastFixEvidence = append([]FixEvidence(nil), c.currentFixEvidence...)
	c.currentFixEvidence = nil
	c.graphSaveMu.Unlock()
	return err
}

func (c *UserContext) RecordFixEvidence(evidence FixEvidence) error {
	normalized := strings.ToLower(evidence.SourceLabel)
	if evidence.EntityType == "" || evidence.ModelPath == "" || evidence.SourceLabel == "" ||
		(evidence.Source != FixEvidenceClock && evidence.Source != FixEvidenceContext) ||
		strings.Contains(normalized, "authorization") || strings.Contains(normalized, "cookie") || strings.Contains(normalized, "token=") {
		return fmt.Errorf("fix evidence must contain only safe framework provenance labels")
	}
	c.graphSaveMu.Lock()
	c.currentFixEvidence = append(c.currentFixEvidence, evidence)
	c.graphSaveMu.Unlock()
	return nil
}

func (c *UserContext) LastFixEvidence() []FixEvidence {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	return append([]FixEvidence(nil), c.lastFixEvidence...)
}

// FixTime is captured once for an active graph save and otherwise reflects the
// current standalone mutation time.
func (c *UserContext) FixTime() time.Time {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	if !c.graphFixTime.IsZero() {
		return c.graphFixTime
	}
	return time.Now()
}

func (c *UserContext) AfterGraphCommit(action func()) {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	if !c.graphSaveActive {
		panic("no graph save is active")
	}
	c.graphCommitActions = append(c.graphCommitActions, action)
}

func (c *UserContext) AfterGraphRollback(action func()) {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	if !c.graphSaveActive {
		panic("no graph save is active")
	}
	c.graphRollbackActions = append(c.graphRollbackActions, action)
}

func (c *UserContext) SendEvent(event *RawAuditEvent) error {
	return c.sendEvent(c, event)
}

func (c *UserContext) sendEvent(context stdcontext.Context, event *RawAuditEvent) (err error) {
	context, scope := StartRuntimeOperation(context, c.RuntimeTelemetry(), NewRuntimeOperation("audit", event.Entity+".event", map[string]RuntimeAttributeValue{
		"teaql.entity.type": event.Entity,
	}))
	_ = context
	defer func() {
		if err != nil {
			scope.Failure(RuntimeErrorType(err))
		} else {
			scope.Success(nil)
		}
	}()
	if c.standardAuditSink != nil {
		if err = c.standardAuditSink.OnEvent(c, event); err != nil {
			return err
		}
	}
	if c.appAuditSink != nil {
		descriptor := c.Entity(event.Entity)
		var maskFields []string
		var maxLen *int
		if descriptor != nil {
			maskFields = descriptor.AuditMaskFlds
			maxLen = descriptor.AuditValueMaxL
		}
		return c.appAuditSink.OnSafeEvent(c, event.BuildSafeEvent(maskFields, maxLen))
	}
	return nil
}

// EmitMutationAudit is deliberately discovered by mutation executors through a
// narrow interface. It keeps the standard sink owned by the initialized server
// context while still allowing an independent application sink.
func (c *UserContext) EmitMutationAudit(request data_service.MutationRequest, result *data_service.MutationResult) error {
	return c.emitMutationAudit(c, request, result)
}

func (c *UserContext) emitMutationAudit(context stdcontext.Context, request data_service.MutationRequest, result *data_service.MutationResult) error {
	if result == nil || result.AffectedRows == 0 {
		return nil
	}
	var event *RawAuditEvent
	switch req := request.(type) {
	case *data_service.InsertMutation:
		values := make(core.Record, len(req.Cmd.Values)+len(result.GeneratedValues))
		for k, v := range req.Cmd.Values {
			values[k] = v
		}
		for k, v := range result.GeneratedValues {
			values[k] = v
		}
		event = Created(req.Cmd.Entity, values)
	case *data_service.UpdateMutation:
		fields := make([]string, 0, len(req.Cmd.Values))
		for k := range req.Cmd.Values {
			fields = append(fields, k)
		}
		oldValues := req.Cmd.OldValues
		event = UpdatedWithOldValues(req.Cmd.Entity, req.Cmd.Values, &oldValues, req.Cmd.Values, fields)
	case *data_service.DeleteMutation:
		event = Deleted(req.Cmd.Entity, req.Cmd.Id, req.Cmd.ExpectedVersion)
	case *data_service.RecoverMutation:
		event = Recovered(req.Cmd.Entity, req.Cmd.Id, req.Cmd.ExpectedVersion)
	default:
		return nil
	}
	event.TraceChain = append([]*core.TraceNode(nil), request.TraceChain()...)
	return c.sendEvent(context, event)
}

func (c *UserContext) SetAppAuditEventSink(sink AppAuditEventSink) { c.appAuditSink = sink }
func (c *UserContext) WithAppAuditEventSink(sink AppAuditEventSink) *UserContext {
	c.SetAppAuditEventSink(sink)
	return c
}

func (c *UserContext) setStandardAuditEventSink(sink RawAuditEventSink) { c.standardAuditSink = sink }

// --- Autogenerated Parity Methods ---
func (c *UserContext) CheckAndFixRecord(args ...any) {
	if len(args) == 1 {
		if input, ok := args[0].(*CheckAndFixInput); ok {
			_ = c.checkAndFix(input)
		}
	}
}

func (c *UserContext) CheckAndFixRecordAt(args ...any) {
	// TODO: invoke checker registry
}

func (c *UserContext) ClearInStore(args ...any) any {
	return nil
}

func (c *UserContext) ClearRequestPolicy(args ...any) any {
	c.requestPolicy = nil
	return c
}

func (c *UserContext) ClearSqlLogs(args ...any) any {
	return nil
}

func (c *UserContext) DataServiceInternal(args ...any) any {
	return nil
}

func (c *UserContext) DisableSqlLog(args ...any) any {
	c.querySQLLogEnabled = false
	c.mutationSQLLogEnabled = false
	return nil
}

func (c *UserContext) DisableSelectSqlLog(args ...any) any {
	c.querySQLLogEnabled = false
	return c
}

func (c *UserContext) DisableMutationSqlLog(args ...any) any {
	c.mutationSQLLogEnabled = false
	return c
}

func (c *UserContext) QuerySqlLogEnabled() bool { return c.querySQLLogEnabled }

func (c *UserContext) MutationSqlLogEnabled() bool { return c.mutationSQLLogEnabled }

func (c *UserContext) EnableAllSqlLog(args ...any) any {
	c.querySQLLogEnabled = true
	c.mutationSQLLogEnabled = true
	return nil
}

func (c *UserContext) EnableMutationSqlLog(args ...any) any {
	c.mutationSQLLogEnabled = true
	return nil
}

func (c *UserContext) EnableSelectSqlLog(args ...any) any {
	c.querySQLLogEnabled = true
	return nil
}

func (c *UserContext) EntityDataService(args ...any) any {
	return nil
}

func (c *UserContext) EntityDataServiceBehavior(args ...any) any {
	return nil
}

func (c *UserContext) GenerateId(args ...any) any {
	return nil
}

func (c *UserContext) GetInStore(val any) any {
	return c.resources["get_in_store"]
}

func (c *UserContext) GetNamedResource(val any) any {
	return c.resources["get_named_resource"]
}

func (c *UserContext) HasChecker(args ...any) bool {
	return c.checkerRegistry != nil
}

func (c *UserContext) HasEntityDataService(args ...any) bool {
	return false
}

func (c *UserContext) InsertNamedResource(args ...any) any {
	return nil
}

func (c *UserContext) Language(args ...any) any {
	if value, ok := c.resources["locale"].(Locale); ok {
		return value
	}
	return LocaleEnglish
}

func (c *UserContext) Local(args ...any) any {
	return nil
}

func (c *UserContext) NextId(args ...any) any {
	return nil
}

func (c *UserContext) PutInStore(args ...any) any {
	return nil
}

func (c *UserContext) PutLocal(args ...any) any {
	return nil
}

func (c *UserContext) RecordMetadataLog(args ...any) {
	// TODO: forward to log buffer
}

func (c *UserContext) RecordSqlLog(args ...any) {
	// TODO: forward to log buffer
}

func (c *UserContext) RegisterExecutor(args ...any) any {
	return nil
}

func (c *UserContext) RemoveLocal(args ...any) any {
	return nil
}

func (c *UserContext) RequireEntity(args ...any) any {
	return nil
}

func (c *UserContext) RequireNamedResource(args ...any) any {
	return nil
}

func (c *UserContext) RequireResource(args ...any) any {
	return nil
}

func (c *UserContext) SetCheckerRegistry(val any) {
	if val == nil {
		c.checkerRegistry = nil
		return
	}
	registry, ok := val.(CheckerRegistry)
	if !ok {
		panic("checker registry must implement runtime.CheckerRegistry")
	}
	c.checkerRegistry = registry
}

func (c *UserContext) SetCustomEventSink(val any) {
	if sink, ok := val.(AppAuditEventSink); ok {
		c.SetAppAuditEventSink(sink)
	}
}

func (c *UserContext) SetEntityDataServiceBehaviorRegistry(val any) {
	c.resources["set_entity_data_service_behavior_registry"] = val
}

func (c *UserContext) SetEntityRegistry(val any) {
	c.resources["set_entity_registry"] = val
}

func (c *UserContext) SetEventSink(val any) {
	if sink, ok := val.(AppAuditEventSink); ok {
		c.SetAppAuditEventSink(sink)
	}
}

func (c *UserContext) SetInternalIdGenerator(val any) {
	c.resources["set_internal_id_generator"] = val
}

func (c *UserContext) SetLanguage(val any) {
	if code, ok := val.(string); ok {
		if err := c.SetLocaleCode(code); err != nil {
			panic(err)
		}
	}
}

func (c *UserContext) SetLanguageCode(val any) {
	if code, ok := val.(string); ok {
		if err := c.SetLocaleCode(code); err != nil {
			panic(err)
		}
	}
}

func (c *UserContext) SetLocaleCode(code string) error {
	locale, err := ParseLocale(code)
	if err != nil {
		return err
	}
	c.resources["locale"] = locale
	return nil
}
func (c *UserContext) InstallI18nCatalog(catalog *I18nCatalog) error {
	if catalog == nil {
		return fmt.Errorf("catalog is required")
	}
	c.resources["i18n_catalog"] = catalog
	return nil
}
func (c *UserContext) I18nCatalog() *I18nCatalog {
	if catalog, ok := c.resources["i18n_catalog"].(*I18nCatalog); ok {
		return catalog
	}
	return BuiltinI18nCatalog()
}

func (c *UserContext) SetMetadata(val any) {
	c.resources["set_metadata"] = val
}

func (c *UserContext) SetRequestPolicy(val any) {
	if val == nil {
		c.requestPolicy = nil
		return
	}
	policy, ok := val.(RequestPolicy)
	if !ok {
		panic("request policy must implement runtime.RequestPolicy")
	}
	c.requestPolicy = policy
}

func (c *UserContext) SetSqlLogOptions(val any) {
	c.resources["set_sql_log_options"] = val
}

func (c *UserContext) SetTimezone(val any) {
	c.resources["set_timezone"] = val
}

func (c *UserContext) SetTraceId(val any) {
	c.resources["set_trace_id"] = val
}

func (c *UserContext) SetUserIdentifier(val any) {
	c.resources["set_user_identifier"] = val
	if value, ok := val.(string); ok {
		c.userIdentifier = value
	} else if val == nil {
		c.userIdentifier = ""
	}
}

func (c *UserContext) SetUserIdentifierOption(val any) {
	c.resources["set_user_identifier_option"] = val
}

func (c *UserContext) SqlLogOptions(args ...any) any {
	return nil
}

func (c *UserContext) SqlLogs(args ...any) any {
	return nil
}

func (c *UserContext) Timezone(args ...any) any {
	return nil
}

func (c *UserContext) TraceId(args ...any) any {
	return nil
}

func (c *UserContext) TranslateCheckResults(args ...any) {
	if len(args) == 0 {
		return
	}
	results, ok := args[0].([]CheckResult)
	if !ok {
		return
	}
	locale := c.Language().(Locale)
	for i := range results {
		c.I18nCatalog().Translate(&results[i], locale)
	}
}

func (c *UserContext) UserIdentifier(args ...any) any {
	return c.userIdentifier
}

func (c *UserContext) WithCheckerRegistry(val any) *UserContext {
	c.SetCheckerRegistry(val)
	return c
}

func (c *UserContext) WithActiveRoot(root EntityReference) *UserContext {
	c.InsertResource("activeRoot", root)
	return c
}

func (c *UserContext) RequireActiveRoot(expectedType string) (EntityReference, error) {
	value := c.GetResource("activeRoot")
	root, ok := value.(EntityReference)
	if !ok {
		if pointer, pointerOK := value.(*EntityReference); pointerOK && pointer != nil {
			root, ok = *pointer, true
		}
	}
	if !ok {
		return EntityReference{}, &ContextRootError{Reason: "missing", ExpectedType: expectedType}
	}
	if root.Entity != expectedType {
		return EntityReference{}, &ContextRootError{Reason: "type_mismatch", ExpectedType: expectedType, ActiveRoot: &root}
	}
	return root, nil
}

// CheckAndFix runs the installed model checker against provider-independent
// mutation state. Generated save code calls this before ID allocation or I/O.
func (c *UserContext) CheckAndFix(input *CheckAndFixInput) error {
	if input == nil {
		return nil
	}
	return c.checkAndFix(input)
}

func (c *UserContext) checkAndFix(input *CheckAndFixInput) error {
	if c.checkerRegistry == nil {
		return nil
	}
	if input.Now.IsZero() {
		input.Now = c.FixTime()
	}
	results := c.checkerRegistry.CheckAndFix(c, input)
	c.TranslateCheckResults(results)
	if len(results) != 0 {
		return &RuntimeError{Type: "Check", CheckResults: results}
	}
	return nil
}

func (c *UserContext) WithCustomEventSink(val any) *UserContext {
	c.SetCustomEventSink(val)
	return c
}

func (c *UserContext) WithEntityDataServiceBehaviorRegistry(val any) *UserContext {
	c.resources["with_entity_data_service_behavior_registry"] = val
	return c
}

func (c *UserContext) WithEntityRegistry(val any) *UserContext {
	c.resources["with_entity_registry"] = val
	return c
}

func (c *UserContext) WithEventSink(val any) *UserContext {
	c.SetEventSink(val)
	return c
}

func (c *UserContext) WithInternalIdGenerator(val any) *UserContext {
	c.resources["with_internal_id_generator"] = val
	return c
}

func (c *UserContext) WithLanguage(val any) *UserContext {
	c.SetLanguage(val)
	return c
}

func (c *UserContext) WithMetadata(val any) *UserContext {
	c.resources["with_metadata"] = val
	return c
}

func (c *UserContext) WithModule(val any) *UserContext {
	c.resources["with_module"] = val
	return c
}

func (c *UserContext) WithRequestPolicy(val any) *UserContext {
	c.SetRequestPolicy(val)
	return c
}

// WithTrustedTenant installs application-owned tenancy outside generated and
// federated request payloads.
func (c *UserContext) WithTrustedTenant(tenant string) *UserContext {
	if strings.TrimSpace(tenant) == "" {
		panic("trusted tenant is required")
	}
	c.InsertResource("trustedTenant", tenant)
	return c
}

func (c *UserContext) TrustedTenant() (string, error) {
	tenant, ok := c.GetResource("trustedTenant").(string)
	if !ok || strings.TrimSpace(tenant) == "" {
		return "", fmt.Errorf("trusted tenant is required")
	}
	return tenant, nil
}

// RuntimeReadiness verifies the same provider installed by generated startup.
// Generated startup owns schema creation; readiness then performs a real ping
// when the provider exposes PingContext.
func (c *UserContext) RuntimeReadiness() error {
	if c.Metadata == nil {
		return fmt.Errorf("metadata is required")
	}
	if c.requestPolicy == nil {
		return fmt.Errorf("request policy is required")
	}
	if _, err := c.TrustedTenant(); err != nil {
		return err
	}
	if c.GetResource("dataService") == nil {
		return fmt.Errorf("dataService is required")
	}
	if provider, ok := c.GetResource("db").(interface {
		PingContext(stdcontext.Context) error
	}); ok {
		if err := provider.PingContext(c); err != nil {
			return fmt.Errorf("provider readiness: %w", err)
		}
	}
	return nil
}

// PrepareQuery snapshots a request and applies trusted authorization exactly
// once before callers derive row and aggregate executions from it.
func (c *UserContext) PrepareQuery(query *core.SelectQuery) (*core.SelectQuery, error) {
	prepared := query.Clone()
	if c.requestPolicy != nil {
		if err := c.requestPolicy.EnforceSelect(c, prepared); err != nil {
			return nil, err
		}
	}
	return prepared, nil
}

func (c *UserContext) WithSchemaProvider(val any) *UserContext {
	c.resources["with_schema_provider"] = val
	return c
}

func (c *UserContext) WithSqlLogOptions(val any) *UserContext {
	c.resources["with_sql_log_options"] = val
	return c
}

func (c *UserContext) WithTimezone(val any) *UserContext {
	c.resources["with_timezone"] = val
	return c
}

func (c *UserContext) WithTraceId(val any) *UserContext {
	c.resources["with_trace_id"] = val
	return c
}

func (c *UserContext) WithUserIdentifier(val any) *UserContext {
	c.SetUserIdentifier(val)
	return c
}

func (c *UserContext) WithUserIdentifierOption(val any) *UserContext {
	c.resources["with_user_identifier_option"] = val
	return c
}

// ==========================================
// Context Attribute
// ==========================================
func (c *UserContext) PutAttribute(key string, value any) {
}

func (c *UserContext) GetAttribute(key string) any {
	return nil
}

// ==========================================
// Local Cache
// ==========================================
func (c *UserContext) PutToLocalCache(key string, value any, timeToLiveInSeconds ...int) {
	entry := localCacheEntry{value: value}
	if len(timeToLiveInSeconds) > 0 && timeToLiveInSeconds[0] > 0 {
		entry.expiresAt = time.Now().Add(time.Duration(timeToLiveInSeconds[0]) * time.Second)
	}
	processLocalCache.Store(key, entry)
}

func (c *UserContext) GetFromLocalCache(key string) any {
	value, ok := processLocalCache.Load(key)
	if !ok {
		return nil
	}
	entry := value.(localCacheEntry)
	if !entry.expiresAt.IsZero() && !time.Now().Before(entry.expiresAt) {
		processLocalCache.Delete(key)
		return nil
	}
	return entry.value
}

func (c *UserContext) RemoveFromLocalCache(key string) {
	processLocalCache.Delete(key)
}

// ==========================================
// Remote Cache
// ==========================================

type RemoteCacheProvider interface {
	PutToRemoteCache(context stdcontext.Context, key string, value any, timeToLiveInSeconds ...int)
	GetFromRemoteCache(context stdcontext.Context, key string) any
	RemoveFromRemoteCache(context stdcontext.Context, key string)
}

func (c *UserContext) PutToRemoteCache(key string, value any, timeToLiveInSeconds ...int) {
	if provider, ok := c.GetResource("RemoteCacheProvider").(RemoteCacheProvider); ok {
		provider.PutToRemoteCache(c.Context, key, value, timeToLiveInSeconds...)
	}
}

func (c *UserContext) GetFromRemoteCache(key string) any {
	if provider, ok := c.GetResource("RemoteCacheProvider").(RemoteCacheProvider); ok {
		return provider.GetFromRemoteCache(c.Context, key)
	}
	return nil
}

func (c *UserContext) RemoveFromRemoteCache(key string) {
	if provider, ok := c.GetResource("RemoteCacheProvider").(RemoteCacheProvider); ok {
		provider.RemoveFromRemoteCache(c.Context, key)
	}
}

// ==========================================
// Local Lock
// ==========================================
func (c *UserContext) TryLocalLock(key string, timeoutMillis int64, expireMillis int64) bool {
	if timeoutMillis < 0 {
		timeoutMillis = 0
	}
	deadline := time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for {
		now := time.Now()
		processLocalLocks.Lock()
		entry, exists := processLocalLocks.entries[key]
		if !exists || (!entry.expiresAt.IsZero() && !now.Before(entry.expiresAt)) || entry.owner == c {
			expiresAt := time.Time{}
			if expireMillis > 0 {
				expiresAt = now.Add(time.Duration(expireMillis) * time.Millisecond)
			}
			processLocalLocks.entries[key] = localLockEntry{owner: c, expiresAt: expiresAt}
			processLocalLocks.Unlock()
			return true
		}
		processLocalLocks.Unlock()
		if timeoutMillis <= 0 || !time.Now().Before(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}

func (c *UserContext) UnlockLocal(key string) {
	processLocalLocks.Lock()
	if entry, exists := processLocalLocks.entries[key]; exists && entry.owner == c {
		delete(processLocalLocks.entries, key)
	}
	processLocalLocks.Unlock()
}

// ==========================================
// Remote Lock
// ==========================================

type RemoteLockProvider interface {
	TryRemoteLock(context stdcontext.Context, key string, timeoutMillis int64, expireMillis int64) bool
	UnlockRemote(context stdcontext.Context, key string)
}

func (c *UserContext) TryRemoteLock(key string, timeoutMillis int64, expireMillis int64) bool {
	if provider, ok := c.GetResource("RemoteLockProvider").(RemoteLockProvider); ok {
		return provider.TryRemoteLock(c.Context, key, timeoutMillis, expireMillis)
	}
	return true
}

func (c *UserContext) UnlockRemote(key string) {
	if provider, ok := c.GetResource("RemoteLockProvider").(RemoteLockProvider); ok {
		provider.UnlockRemote(c.Context, key)
	}
}

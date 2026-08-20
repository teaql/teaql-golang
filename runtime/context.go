package runtime

import (
	stdcontext "context"
	"fmt"
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
	resources                 map[string]interface{}
	standardAuditSink         RawAuditEventSink
	appAuditSink              AppAuditEventSink
	runtimeTelemetrySink      RuntimeTelemetrySink
	runtimeTelemetry          RuntimeTelemetry
	continuousPageCursorStore ContinuousPageCursorStore
	continuousPageMu          sync.Mutex
	continuousPagePlan        string
	continuousPageCursorID    string
	userIdentifier            string
	requestPolicy             RequestPolicy
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

func (c *UserContext) RecordExecutionMetadata(metadata data_service.ExecutionMetadata) {
	if c.runtimeTelemetrySink != nil {
		c.runtimeTelemetrySink.RecordExecutionMetadata(metadata)
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
		continuousPagePlan:        "DISABLED",
		userIdentifier:            "main",
		runtimeTelemetry:          NoopRuntimeTelemetry{},
	}
	context.Context = stdcontext.WithValue(context.Context, userContextKey{}, context)
	return context
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
	// TODO: invoke checker registry
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

func (c *UserContext) CommitChangesInternal(args ...any) error {
	// TODO: iterate change set and dispatch to executor
	return nil
}

func (c *UserContext) DataServiceInternal(args ...any) any {
	return nil
}

func (c *UserContext) DisableSqlLog(args ...any) any {
	return nil
}

func (c *UserContext) EnableAllSqlLog(args ...any) any {
	return nil
}

func (c *UserContext) EnableMutationSqlLog(args ...any) any {
	return nil
}

func (c *UserContext) EnableSelectSqlLog(args ...any) any {
	return nil
}

func (c *UserContext) EnsureSchema(args ...any) any {
	return nil
}

func (c *UserContext) EntityDataService(args ...any) any {
	return nil
}

func (c *UserContext) EntityDataServiceBehavior(args ...any) any {
	return nil
}

func (c *UserContext) EntityRoot(args ...any) any {
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
	return false
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
	c.resources["set_checker_registry"] = val
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
	c.resources["with_checker_registry"] = val
	return c
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

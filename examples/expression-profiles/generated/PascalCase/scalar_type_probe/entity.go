package scalar_type_probe

import (
	stdcontext "context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"

	"time"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	"scalar-type-conformance-service-core-workspace/lib/internal/generatedjson"
)

var (
	_ = time.Time{}
	_ = decimal.Decimal{}
	_ = fmt.Sprint
	_ = strings.Join
	_ = errors.As
)

var teaqlTemporaryEntityID int64


type ScalarTypeProbe struct {
	base        *core.BaseEntityData
	loadedSnapshot core.LoadedScalarSnapshot
	dirtyFields map[string]bool
	isNew       bool
	markedAsDelete bool
	comment     *string
	purpose     *string
	loadState   map[string]bool
	restrictLoadState bool
	root        *core.EntityRoot
	ledgerID    core.Value
	relations   map[string]core.Entity
	loadedRelations map[string]bool
}

// TypePlatformFacet reads metadata of the selected relation; it never queries the database.
func (e *ScalarTypeProbe) TypePlatformFacet(name string) (*core.SmartList[core.Record], bool) {
	return e.base.RelationFacet("typePlatformEntity", name)
}

func NewScalarTypeProbe() *ScalarTypeProbe {
	temporaryID := -atomic.AddInt64(&teaqlTemporaryEntityID, 1)
	entity := &ScalarTypeProbe{
		base:        core.NewBaseEntityData(),
		dirtyFields: make(map[string]bool),
		isNew:       true,
		loadState:   make(map[string]bool),
		root:        core.NewEntityRoot(),
		ledgerID:    core.ValI64(temporaryID),
		relations:   make(map[string]core.Entity),
		loadedRelations: make(map[string]bool),
	}
	entity.root.MarkAsNew(entity.EntityKey())
	return entity
}

// Hydration is not a create request. Keep this constructor private so public
// NewEntity still records new-object intent, while a loaded snapshot starts
// with independent mutation ownership and no pending insert.
func newLoadedScalarTypeProbe() *ScalarTypeProbe {
	entity := NewScalarTypeProbe()
	entity.root.ClearEntity(entity.EntityKey())
	return entity
}

func (e *ScalarTypeProbe) EntityKey() core.EntityKey {
	if e.base.Id != 0 { return core.NewEntityKey(e.EntityName(), core.ValU64(e.base.Id)) }
	return core.NewEntityKey(e.EntityName(), e.ledgerID)
}

func (e *ScalarTypeProbe) EntityRoot() *core.EntityRoot { return e.root }

func (e *ScalarTypeProbe) AttachEntityRoot(root *core.EntityRoot) {
	if root == nil || root == e.root { return }
	root.MergeFrom(e.root)
	if root.CompositionError() != nil { return }
	e.root = root
}

func (e *ScalarTypeProbe) RelationEntity(name string) (core.Entity, bool) {
	value, ok := e.relations[name]
	return value, ok
}

func (e *ScalarTypeProbe) setRelationEntity(name string, value core.Entity) {
	e.relations[name] = value
}

func (e *ScalarTypeProbe) markRelationLoaded(name string) {
	e.loadedRelations[name] = true
}

func (e *ScalarTypeProbe) isRelationLoaded(name string) bool {
	return e.loadedRelations[name]
}

func (e *ScalarTypeProbe) MarkLoadedOnly(fields ...string) *ScalarTypeProbe {
	e.restrictLoadState = true
	e.loadState = make(map[string]bool, len(fields))
	for _, field := range fields { e.loadState[field] = true }
	return e
}

func (e *ScalarTypeProbe) IsLoaded(field string) bool {
	if e.isNew && !e.restrictLoadState { return true }
	return e.loadState[field]
}

func (e *ScalarTypeProbe) EntityName() string {
	return "Scalar Type Probe"
}

func (e *ScalarTypeProbe) EntityDescriptor() *core.EntityDescriptor {
	return nil // Handled by runtime context in Go
}

func (e *ScalarTypeProbe) Base() *core.BaseEntityData {
	return e.base
}

func (e *ScalarTypeProbe) IdValue() core.Value {
	return core.ValU64(e.base.Id)
}



func (e *ScalarTypeProbe) FromRecord(record core.Record) error {
	return e.teaqlFromRecord(record, true)
}

func (e *ScalarTypeProbe) teaqlFromRecord(record core.Record, captureLoaded bool) error {
	oldKey := e.EntityKey()
	base, err := core.BaseEntityDataFromRecord(record)
	if err != nil {
		return err
	}
	if value, present := base.Dynamic["required_time"]; present && !value.IsNull() {
		clock, valid := value.TryTimeOfDay()
		if !valid { return core.NewEntityError("Scalar Type Probe", "invalid field required_time: expected TimeOfDay") }
		base.Dynamic["required_time"] = core.ValTimeOfDay(clock)
	}

	if value, present := base.Dynamic["optional_time"]; present && !value.IsNull() {
		clock, valid := value.TryTimeOfDay()
		if !valid { return core.NewEntityError("Scalar Type Probe", "invalid field optional_time: expected TimeOfDay") }
		base.Dynamic["optional_time"] = core.ValTimeOfDay(clock)
	}






	if value, present := base.Dynamic["required_integer"]; present && !value.IsNull() {
		if err := core.ValidateInteger32("Scalar Type Probe", "required_integer", value); err != nil { return err }
	}

	if value, present := base.Dynamic["optional_integer"]; present && !value.IsNull() {
		if err := core.ValidateInteger32("Scalar Type Probe", "optional_integer", value); err != nil { return err }
	}


















	e.base = base
	e.root.Rekey(oldKey, e.EntityKey())
	e.root.SetOriginalVersion(e.EntityKey(), e.base.Version)
	e.isNew = false
	e.dirtyFields = make(map[string]bool)
	e.loadState = make(map[string]bool, len(record))
	e.restrictLoadState = true
	for field := range record { e.loadState[field] = true }
	if captureLoaded { e.loadedSnapshot = e.teaqlLoadedScalarSnapshot(e.IntoRecord()) }
	return nil
}

func (e *ScalarTypeProbe) teaqlScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewLoadedScalarSnapshot(record, "id", "type_platform_id", "required_string", "optional_string", "large_text", "required_boolean", "optional_boolean", "required_integer", "optional_integer", "required_long", "optional_long", "required_decimal", "optional_decimal", "required_float", "optional_float", "required_double", "optional_double", "required_date", "optional_date", "required_time", "optional_time", "required_datetime", "optional_datetime", "create_time", "update_time", "version")
}

func (e *ScalarTypeProbe) teaqlLoadedScalarSnapshot(record core.Record) core.LoadedScalarSnapshot {
	return core.NewOwnedLoadedScalarSnapshot(e.EntityKey(), e.base.Version, record, "id", "type_platform_id", "required_string", "optional_string", "large_text", "required_boolean", "optional_boolean", "required_integer", "optional_integer", "required_long", "optional_long", "required_decimal", "optional_decimal", "required_float", "optional_float", "required_double", "optional_double", "required_date", "optional_date", "required_time", "optional_time", "required_datetime", "optional_datetime", "create_time", "update_time", "version")
}

func (e *ScalarTypeProbe) teaqlAcceptPersisted(context *runtime.UserContext, record core.Record) error {
	if err := e.teaqlFromRecord(record, false); err != nil { return err }
	committed := e.teaqlLoadedScalarSnapshot(record)
	context.AfterGraphCommit(func() { e.loadedSnapshot = committed })
	return nil
}

// Capture only the explicitly composed owned graph; no shared Context state.
func (e *ScalarTypeProbe) TeaqlPrivacyEntries() []data_service.MutationPrivacyEntry {
	entries := []data_service.MutationPrivacyEntry{
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.loadedSnapshot.Values()),
		data_service.NewMutationPrivacyEntry(e.EntityName(), e.teaqlScalarSnapshot(e.IntoRecord()).Values()),
	}
	return entries
}

func (e *ScalarTypeProbe) IntoRecord() core.Record {
	rec := e.base.ToRecord()
	if e.isNew && e.base.Id == 0 {
		delete(rec, "id")
	}
	return rec
}

func (e *ScalarTypeProbe) DirtyFields() []string {
	var fields []string
	for k, v := range e.dirtyFields {
		if v {
			fields = append(fields, k)
		}
	}
	return fields
}

func (e *ScalarTypeProbe) IsMarkedAsDelete() bool {
	return e.markedAsDelete
}

func (e *ScalarTypeProbe) MarkForDeletion() *ScalarTypeProbe {
	e.markedAsDelete = true
	e.root.MarkAsDeleted(e.EntityKey())
	return e
}

func (e *ScalarTypeProbe) IsNew() bool {
	return e.isNew
}

func (e *ScalarTypeProbe) MarkAsNew() {
	e.isNew = true
}

func (e *ScalarTypeProbe) GetComment() *string {
	return e.comment
}

func (e *ScalarTypeProbe) SetComment(comment string) {
	e.comment = &comment
}

func (e *ScalarTypeProbe) AuditAs(comment string) *ScalarTypeProbe {
	if _, err := core.NewMutationIntent(&comment); err != nil { panic(err) }
	e.comment = &comment
	return e
}

func (e *ScalarTypeProbe) Comment(comment string) *ScalarTypeProbe {
	e.comment = &comment
	return e
}

func (e *ScalarTypeProbe) Purpose(purpose string) *ScalarTypeProbe {
	e.purpose = &purpose
	return e
}

func (e *ScalarTypeProbe) OriginalValues() core.Record {
	return make(core.Record) // Basic implementation
}

func (e *ScalarTypeProbe) OnLoaded(context any) {
}

func (e *ScalarTypeProbe) IntoJson() any {
	return e // encoding/json dispatches MarshalJSON and preserves any error.
}

func (e *ScalarTypeProbe) ToJSON(profile ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	return generatedjson.Model.Serialize(e.EntityName(), e, profile...)
}

func (e *ScalarTypeProbe) MarshalJSON() ([]byte, error) { return e.ToJSON() }

// Only aliases captured from an executed query are public query properties.
// Never dump Base().Dynamic: it also contains modeled and private runtime data.
func (e *ScalarTypeProbe) JSONQueryProperties() core.Record { return e.base.QueryProperties() }

// Only shallow loaded values escape. Related documents are visited lazily by
// the runtime codec under its depth, byte and cycle guards.
func (e *ScalarTypeProbe) LoadedJSONFields() core.Record {
	all := e.IntoRecord()
	selected := make(core.Record)
	if e.IsLoaded("id") { if value, present := all["id"]; present { selected["id"] = value } }
	if e.IsLoaded("type_platform_id") { if value, present := all["type_platform_id"]; present { selected["type_platform_id"] = value } }
	if e.IsLoaded("required_string") { if value, present := all["required_string"]; present { selected["required_string"] = value } }
	if e.IsLoaded("optional_string") { if value, present := all["optional_string"]; present { selected["optional_string"] = value } }
	if e.IsLoaded("large_text") { if value, present := all["large_text"]; present { selected["large_text"] = value } }
	if e.IsLoaded("required_boolean") { if value, present := all["required_boolean"]; present { selected["required_boolean"] = value } }
	if e.IsLoaded("optional_boolean") { if value, present := all["optional_boolean"]; present { selected["optional_boolean"] = value } }
	if e.IsLoaded("required_integer") { if value, present := all["required_integer"]; present { selected["required_integer"] = value } }
	if e.IsLoaded("optional_integer") { if value, present := all["optional_integer"]; present { selected["optional_integer"] = value } }
	if e.IsLoaded("required_long") { if value, present := all["required_long"]; present { selected["required_long"] = value } }
	if e.IsLoaded("optional_long") { if value, present := all["optional_long"]; present { selected["optional_long"] = value } }
	if e.IsLoaded("required_decimal") { if value, present := all["required_decimal"]; present { selected["required_decimal"] = value } }
	if e.IsLoaded("optional_decimal") { if value, present := all["optional_decimal"]; present { selected["optional_decimal"] = value } }
	if e.IsLoaded("required_float") { if value, present := all["required_float"]; present { selected["required_float"] = value } }
	if e.IsLoaded("optional_float") { if value, present := all["optional_float"]; present { selected["optional_float"] = value } }
	if e.IsLoaded("required_double") { if value, present := all["required_double"]; present { selected["required_double"] = value } }
	if e.IsLoaded("optional_double") { if value, present := all["optional_double"]; present { selected["optional_double"] = value } }
	if e.IsLoaded("required_date") { if value, present := all["required_date"]; present { selected["required_date"] = value } }
	if e.IsLoaded("optional_date") { if value, present := all["optional_date"]; present { selected["optional_date"] = value } }
	if e.IsLoaded("required_time") { if value, present := all["required_time"]; present { selected["required_time"] = value } }
	if e.IsLoaded("optional_time") { if value, present := all["optional_time"]; present { selected["optional_time"] = value } }
	if e.IsLoaded("required_datetime") { if value, present := all["required_datetime"]; present { selected["required_datetime"] = value } }
	if e.IsLoaded("optional_datetime") { if value, present := all["optional_datetime"]; present { selected["optional_datetime"] = value } }
	if e.IsLoaded("create_time") { if value, present := all["create_time"]; present { selected["create_time"] = value } }
	if e.IsLoaded("update_time") { if value, present := all["update_time"]; present { selected["update_time"] = value } }
	if e.IsLoaded("version") { if value, present := all["version"]; present { selected["version"] = value } }
	if e.isRelationLoaded("typePlatformEntity") { if target, present := e.relations["typePlatformEntity"]; present && target != nil { selected["typePlatformEntity"] = core.Value{V: target} } }
	return selected
}

// Generated infrastructure used by the facade's typed factories. It never
// replaces an existing caller-owned entity or imports a parent entity package.
func TeaqlScalarTypeProbeFromJSONRecord(record core.Record, factory func(string, core.Record) (core.Entity, error)) (*ScalarTypeProbe, error) {
	entity := newLoadedScalarTypeProbe()
	if err := entity.FromRecord(record); err != nil { return nil, err }
	if value, selected := record["typePlatformEntity"]; selected {
		entity.markRelationLoaded("typePlatformEntity")
		if !value.IsNull() {
			targetRecord, ok := value.TryObject(); if !ok { return nil, fmt.Errorf("invalid generated JSON relation record") }
			target, err := factory("Type Platform", targetRecord); if err != nil { return nil, err }
			entity.setRelationEntity("typePlatformEntity", target)
		}
	}
	return entity, nil
}

func (e *ScalarTypeProbe) Save(context *runtime.UserContext) (*ScalarTypeProbe, error) {
	intent, intentErr := core.NewMutationIntent(e.comment)
	if intentErr != nil { return nil, intentErr }
	var saved *ScalarTypeProbe
	var privacy *data_service.MutationPrivacy
	err := context.ExecutePreparedGraphSave(intent, func() (*runtime.MutationPlan, error) {
		if preflightErr := e.TeaqlPreflightGraph(context, intent); preflightErr != nil { return nil, preflightErr }
		privacy = data_service.NewMutationPrivacy(e.TeaqlPrivacyEntries()...)
		auditReason := intent.AuditReason()
		return runtime.MutationPlanFromEntityRoot(e.root, e.EntityName(), auditReason), nil
	}, func() error {
		var innerErr error
		saved, innerErr = e.TeaqlSaveWithinGraph(context, intent, nil, privacy)
		return innerErr
	})
	return saved, err
}

// TeaqlPreflightGraph runs Checker/Fix for the complete aggregate before the
// first provider mutation. It is generated infrastructure, not application API.
func (e *ScalarTypeProbe) TeaqlPreflightGraph(context *runtime.UserContext, intent core.MutationIntent) error {
	if err := intent.Validate(); err != nil { return err }
	if err := e.TeaqlNormalizeMutationGraph(context); err != nil { return err }
	if err := context.ValidateGraphMutationVersions(e.root); err != nil { return err }
	return e.TeaqlPreflightWithinGraph(context, intent, nil)
}

// Cancel known-baseline equal assignments and attach the complete pending graph
// before the first Checker, Fix, business-ID allocation or version admission.
func (e *ScalarTypeProbe) TeaqlNormalizeMutationGraph(context *runtime.UserContext) error {
	if e.isNew && e.markedAsDelete { return nil }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	return nil
}

func (e *ScalarTypeProbe) TeaqlPreflightWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parent *core.EntityKey) error {
	if err := intent.Validate(); err != nil { return err }
	if e.isNew && e.markedAsDelete {
		if parent == nil { return fmt.Errorf("cancelled new root %s: create-then-delete has no persisted entity to return", e.EntityName()) }
		return nil
	}
	if err := context.PrepareGraphMutationLineage(e.root, e.EntityKey(), parent, intent, e.comment); err != nil { return err }
	restoredFields, err := context.DiscardRestoredScalarChanges(e.root, e.EntityKey(), e.loadedSnapshot)
	if err != nil { return err }
	for _, field := range restoredFields { delete(e.dirtyFields, field) }
	if !e.markedAsDelete && (e.isNew || len(e.root.Change(e.EntityKey())) > 0) {
		if e.isNew {
		}
		operation := core.MutationUpdate
		if e.isNew { operation = core.MutationInsert }
		if operation == core.MutationUpdate {
			if !e.IsLoaded("id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("id"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("type_platform_id") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("type_platform"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_string") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_string"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_string") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_string"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("large_text") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("large_text"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_boolean") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_boolean"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_boolean") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_boolean"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_integer") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_integer"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_integer") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_integer"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_long") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_long"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_long") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_long"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_decimal") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_decimal"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_decimal") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_decimal"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_float") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_float"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_float") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_float"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_double") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_double"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_double") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_double"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_date") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_date"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_date") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_date"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("required_datetime") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("required_datetime"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("optional_datetime") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("optional_datetime"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("create_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("create_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("update_time") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("update_time"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
			if !e.IsLoaded("version") {
				result := runtime.CheckResult{RuleID: "invalid_type", EntityType: "ScalarTypeProbe", CanonicalLocation: runtime.Location().Property("version"), Message: "Mutation requires a fully loaded entity"}
				return &runtime.RuntimeError{Type: "Check", CheckResults: []runtime.CheckResult{result}}
			}
		}
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Scalar Type Probe", Operation: operation, Values: checkedValues}, e.root, e.EntityKey())
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				e.base.PutDynamic(field, value)
				canonical, mappingErr := context.CanonicalLedgerField(e.EntityName(), field)
				if mappingErr != nil { return mappingErr }
				e.root.Set(e.EntityKey(), canonical, value)
			}
		}
		if checkErr != nil { return checkErr }
	}
	return nil
}

type teaqlScalarTypeProbeSaveSnapshot struct {
	loadedSnapshot core.LoadedScalarSnapshot
	record core.Record
	dirtyFields map[string]bool
	isNew bool
	markedAsDelete bool
	loadState map[string]bool
	restrictLoadState bool
	ledgerID core.Value
}

func (e *ScalarTypeProbe) teaqlSaveSnapshot() teaqlScalarTypeProbeSaveSnapshot {
	dirty := make(map[string]bool, len(e.dirtyFields))
	for field, value := range e.dirtyFields { dirty[field] = value }
	loaded := make(map[string]bool, len(e.loadState))
	for field, value := range e.loadState { loaded[field] = value }
	return teaqlScalarTypeProbeSaveSnapshot{
		loadedSnapshot: e.loadedSnapshot,
		record: e.IntoRecord(), dirtyFields: dirty, isNew: e.isNew,
		markedAsDelete: e.markedAsDelete, loadState: loaded,
		restrictLoadState: e.restrictLoadState, ledgerID: e.ledgerID,
	}
}

func (e *ScalarTypeProbe) teaqlRegisterGraphOutcome(context *runtime.UserContext, snapshot teaqlScalarTypeProbeSaveSnapshot) {
	context.AfterGraphRollback(func() {
		if err := e.FromRecord(snapshot.record); err != nil { panic(err) }
		e.loadedSnapshot = snapshot.loadedSnapshot
		e.dirtyFields = snapshot.dirtyFields
		e.isNew = snapshot.isNew
		e.markedAsDelete = snapshot.markedAsDelete
		e.loadState = snapshot.loadState
		e.restrictLoadState = snapshot.restrictLoadState
		e.ledgerID = snapshot.ledgerID
	})
	context.AfterGraphCommit(func() { e.root.ClearEntity(e.EntityKey()) })
}

// TeaqlSaveWithinGraph is generated infrastructure used by related entity
// packages after the public root Save has opened the graph transaction.
func (e *ScalarTypeProbe) TeaqlSaveWithinGraph(context *runtime.UserContext, intent core.MutationIntent, parentScope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) (*ScalarTypeProbe, error) {
	if err := intent.Validate(); err != nil { return nil, err }
	if e.isNew && e.markedAsDelete {
		if parentScope == nil { return nil, fmt.Errorf("cancelled new root %s: create-then-delete has no persisted entity to return", e.EntityName()) }
		return e, nil
	}
	// Reached but unchanged entities are ancestry, not writes. Their changed
	// descendants still participate without requiring a complete parent projection.
	if !e.isNew && !e.markedAsDelete && len(e.root.Change(e.EntityKey())) == 0 {
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
		return e, nil
	}
	snapshot := e.teaqlSaveSnapshot()
	e.teaqlRegisterGraphOutcome(context, snapshot)
	dsRaw := context.GetResource("dataService")
	if dsRaw == nil {
		return nil, fmt.Errorf("dataService not found in UserContext")
	}
	// Dynamic assert
	type mutator interface {
		Mutate(stdcontext.Context, data_service.MutationRequest) (*data_service.MutationResult, error)
	}
	ds, ok := dsRaw.(mutator)
	if !ok {
		return nil, fmt.Errorf("dataService does not implement Mutator")
	}


	if e.isNew {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Scalar Type Probe", Operation: core.MutationInsert, Values: checkedValues}, e.root, e.EntityKey())
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				canonical, mappingErr := context.CanonicalLedgerField(e.EntityName(), field)
				if mappingErr != nil { return nil, mappingErr }
				e.root.Set(e.EntityKey(), canonical, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.teaqlFromRecord(checkedValues, false); err != nil { return nil, err }
		if e.base.Id == 0 {
			id, err := context.GenerateInternalId(e.EntityName())
			if err != nil {
				return nil, fmt.Errorf("generate id for %s: %w", e.EntityName(), err)
			}
			e.base.Id = id
			if err := context.RekeyPreparedGraphEntity(e.root, core.NewEntityKey(e.EntityName(), e.ledgerID), e.EntityKey()); err != nil { return nil, err }
		} else {
			if err := context.EnsureInternalIdFloor(context.Context, e.EntityName(), e.base.Id); err != nil {
				return nil, fmt.Errorf("synchronize id floor for %s: %w", e.EntityName(), err)
			}
		}
		if e.base.Version == 0 {
			e.base.Version = 1
		}
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewInsertCommand("Scalar Type Probe")
		cmd.Values = e.IntoRecord()
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.InsertMutation{Cmd: cmd}, intent.Comment())
		if err != nil { return nil, err }
		request, err = data_service.WithMutationPrivacy(request, privacy)
		if err != nil { return nil, err }
		res, err := ds.Mutate(context, request)
		if err == nil {
			e.isNew = false
			e.dirtyFields = make(map[string]bool)
			if res.GeneratedValues != nil {
				if idVal, ok := res.GeneratedValues["id"]; ok {
					if idU64, ok := idVal.TryU64(); ok {
						e.base.Id = idU64
					} else if idI64, ok := idVal.TryI64(); ok {
						e.base.Id = uint64(idI64)
					}
				}
			}
		}
		if err != nil {
			return nil, err
		}
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.teaqlAcceptPersisted(context, res.PersistedRecord); err != nil {
			return nil, err
		}
		if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
		return e, nil
	} else if e.markedAsDelete {
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		expectedVersion := e.base.Version
		cmd := core.NewDeleteCommand("Scalar Type Probe", core.ValU64(e.base.Id)).
			WithExpectedVersion(expectedVersion)
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.DeleteMutation{Cmd: cmd}, intent.Comment())
		if err != nil { return nil, err }
		request, err = data_service.WithMutationPrivacy(request, privacy)
		if err != nil { return nil, err }
		res, err := ds.Mutate(context, request)
		if err != nil { return nil, err }
		if res.AffectedRows == 0 {
			return nil, fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
		}
		e.base.Version = -(expectedVersion + 1)
		e.markedAsDelete = false
		e.dirtyFields = make(map[string]bool)
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.teaqlAcceptPersisted(context, res.PersistedRecord); err != nil { return nil, err }
		return e, nil
	} else {
		checkedValues := e.IntoRecord()
		valuesBeforeCheck := e.IntoRecord()
		checkErr := runtime.CheckAndFixGraphEntity(context, &runtime.CheckAndFixInput{Entity: "Scalar Type Probe", Operation: core.MutationUpdate, Values: checkedValues}, e.root, e.EntityKey())
		for field, value := range checkedValues {
			if before, exists := valuesBeforeCheck[field]; !exists || !reflect.DeepEqual(before, value) {
				canonical, mappingErr := context.CanonicalLedgerField(e.EntityName(), field)
				if mappingErr != nil { return nil, mappingErr }
				e.root.Set(e.EntityKey(), canonical, value)
			}
		}
		if checkErr != nil { return nil, checkErr }
		if err := e.teaqlFromRecord(checkedValues, false); err != nil { return nil, err }
		scope, err := core.MutationScopeForEntity(parentScope, e.EntityKey(), intent, e.comment)
		if err != nil { return nil, err }
		cmd := core.NewUpdateCommand("Scalar Type Probe", core.ValU64(e.base.Id))
		cmd.Values, err = context.StorageLedgerRecord(e.EntityName(), e.root.Change(e.EntityKey()))
		if err != nil { return nil, err }
		// A clean parent still carries the scope for changed descendants, but
		// must not emit an empty UPDATE or bump its optimistic version.
		if len(cmd.Values) == 0 {
			if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
			return e, nil
		}
		expectedVersion := e.base.Version
		cmd.ExpectedVersion = &expectedVersion
		cmd.TraceChain = core.MutationTraceForEntity(e.root, e.EntityKey(), scope)
		request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: cmd}, intent.Comment())
		if err != nil { return nil, err }
		request, err = data_service.WithMutationPrivacy(request, privacy)
		if err != nil { return nil, err }
		res, err := ds.Mutate(context, request)
		if err == nil {
			if res.AffectedRows == 0 {
				return nil, fmt.Errorf("optimistic lock failed for %s(%d) at version %d", e.EntityName(), e.base.Id, expectedVersion)
			}
			e.base.Version = expectedVersion + 1
			e.dirtyFields = make(map[string]bool)
		}
		if err != nil {
			return nil, err
		}
		if res.PersistedRecord == nil {
			return nil, fmt.Errorf("mutation did not return the authoritative persisted record")
		}
		if err := e.teaqlAcceptPersisted(context, res.PersistedRecord); err != nil { return nil, err }
		if err := e.saveCascade(context, intent, scope, privacy); err != nil { return nil, err }
		return e, nil
	}
}

func (e *ScalarTypeProbe) saveCascade(context *runtime.UserContext, intent core.MutationIntent, scope *core.MutationTraceScope, privacy *data_service.MutationPrivacy) error {
	return nil
}

func (e *ScalarTypeProbe) Id() uint64 {
	return e.base.Id
}

func (e *ScalarTypeProbe) UpdateId(value uint64) *ScalarTypeProbe {
	oldKey := e.EntityKey()
	e.base.Id = value
	e.root.Rekey(oldKey, e.EntityKey())
	e.loadState["id"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredString() string {
	val, _ := e.base.GetDynamic("required_string")
	res, _ := val.TryText()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredString(value string) *ScalarTypeProbe {
	e.base.PutDynamic("required_string", core.ValText(value))
	e.dirtyFields["required_string"] = true
	e.root.Set(e.EntityKey(), "required_string", core.ValText(value))
	e.loadState["required_string"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalString() *string {
	val, _ := e.base.GetDynamic("optional_string")
	if val.IsNull() { return nil }
	res, _ := val.TryText()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalString(value *string) *ScalarTypeProbe {
	e.base.PutDynamic("optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.dirtyFields["optional_string"] = true
	e.root.Set(e.EntityKey(), "optional_string", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.loadState["optional_string"] = true
	return e
}

func (e *ScalarTypeProbe) LargeText() *string {
	val, _ := e.base.GetDynamic("large_text")
	if val.IsNull() { return nil }
	res, _ := val.TryText()
	return &res}

func (e *ScalarTypeProbe) UpdateLargeText(value *string) *ScalarTypeProbe {
	e.base.PutDynamic("large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.dirtyFields["large_text"] = true
	e.root.Set(e.EntityKey(), "large_text", func() core.Value { if value == nil { return core.ValNull() }; return core.ValText((*value)) }())
	e.loadState["large_text"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredBoolean() bool {
	val, _ := e.base.GetDynamic("required_boolean")
	res, _ := val.TryBool()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredBoolean(value bool) *ScalarTypeProbe {
	e.base.PutDynamic("required_boolean", core.ValBool(value))
	e.dirtyFields["required_boolean"] = true
	e.root.Set(e.EntityKey(), "required_boolean", core.ValBool(value))
	e.loadState["required_boolean"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalBoolean() *bool {
	val, _ := e.base.GetDynamic("optional_boolean")
	if val.IsNull() { return nil }
	res, _ := val.TryBool()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalBoolean(value *bool) *ScalarTypeProbe {
	e.base.PutDynamic("optional_boolean", func() core.Value { if value == nil { return core.ValNull() }; return core.ValBool((*value)) }())
	e.dirtyFields["optional_boolean"] = true
	e.root.Set(e.EntityKey(), "optional_boolean", func() core.Value { if value == nil { return core.ValNull() }; return core.ValBool((*value)) }())
	e.loadState["optional_boolean"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredInteger() int64 {
	val, _ := e.base.GetDynamic("required_integer")
	res, _ := val.TryI64()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredInteger(value int64) *ScalarTypeProbe {
	e.base.PutDynamic("required_integer", core.ValI64(value))
	e.dirtyFields["required_integer"] = true
	e.root.Set(e.EntityKey(), "required_integer", core.ValI64(value))
	e.loadState["required_integer"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalInteger() *int64 {
	val, _ := e.base.GetDynamic("optional_integer")
	if val.IsNull() { return nil }
	res, _ := val.TryI64()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalInteger(value *int64) *ScalarTypeProbe {
	e.base.PutDynamic("optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	e.dirtyFields["optional_integer"] = true
	e.root.Set(e.EntityKey(), "optional_integer", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	e.loadState["optional_integer"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredLong() int64 {
	val, _ := e.base.GetDynamic("required_long")
	res, _ := val.TryI64()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredLong(value int64) *ScalarTypeProbe {
	e.base.PutDynamic("required_long", core.ValI64(value))
	e.dirtyFields["required_long"] = true
	e.root.Set(e.EntityKey(), "required_long", core.ValI64(value))
	e.loadState["required_long"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalLong() *int64 {
	val, _ := e.base.GetDynamic("optional_long")
	if val.IsNull() { return nil }
	res, _ := val.TryI64()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalLong(value *int64) *ScalarTypeProbe {
	e.base.PutDynamic("optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	e.dirtyFields["optional_long"] = true
	e.root.Set(e.EntityKey(), "optional_long", func() core.Value { if value == nil { return core.ValNull() }; return core.ValI64((*value)) }())
	e.loadState["optional_long"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredDecimal() decimal.Decimal {
	val, _ := e.base.GetDynamic("required_decimal")
	res, _ := val.TryDecimal()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredDecimal(value decimal.Decimal) *ScalarTypeProbe {
	e.base.PutDynamic("required_decimal", core.ValDecimal(value))
	e.dirtyFields["required_decimal"] = true
	e.root.Set(e.EntityKey(), "required_decimal", core.ValDecimal(value))
	e.loadState["required_decimal"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalDecimal() *decimal.Decimal {
	val, _ := e.base.GetDynamic("optional_decimal")
	if val.IsNull() { return nil }
	res, _ := val.TryDecimal()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalDecimal(value *decimal.Decimal) *ScalarTypeProbe {
	e.base.PutDynamic("optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }())
	e.dirtyFields["optional_decimal"] = true
	e.root.Set(e.EntityKey(), "optional_decimal", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDecimal((*value)) }())
	e.loadState["optional_decimal"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredFloat() float32 {
	val, _ := e.base.GetDynamic("required_float")
	res, _ := val.TryF64()
	converted := float32(res)
		return converted
}

func (e *ScalarTypeProbe) UpdateRequiredFloat(value float32) *ScalarTypeProbe {
	e.base.PutDynamic("required_float", core.ValF64(float64(value)))
	e.dirtyFields["required_float"] = true
	e.root.Set(e.EntityKey(), "required_float", core.ValF64(float64(value)))
	e.loadState["required_float"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalFloat() *float32 {
	val, _ := e.base.GetDynamic("optional_float")
	if val.IsNull() { return nil }
	res, _ := val.TryF64()
	converted := float32(res)
		return &converted
}

func (e *ScalarTypeProbe) UpdateOptionalFloat(value *float32) *ScalarTypeProbe {
	e.base.PutDynamic("optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }())
	e.dirtyFields["optional_float"] = true
	e.root.Set(e.EntityKey(), "optional_float", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64(float64((*value))) }())
	e.loadState["optional_float"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredDouble() float64 {
	val, _ := e.base.GetDynamic("required_double")
	res, _ := val.TryF64()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredDouble(value float64) *ScalarTypeProbe {
	e.base.PutDynamic("required_double", core.ValF64(value))
	e.dirtyFields["required_double"] = true
	e.root.Set(e.EntityKey(), "required_double", core.ValF64(value))
	e.loadState["required_double"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalDouble() *float64 {
	val, _ := e.base.GetDynamic("optional_double")
	if val.IsNull() { return nil }
	res, _ := val.TryF64()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalDouble(value *float64) *ScalarTypeProbe {
	e.base.PutDynamic("optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }())
	e.dirtyFields["optional_double"] = true
	e.root.Set(e.EntityKey(), "optional_double", func() core.Value { if value == nil { return core.ValNull() }; return core.ValF64((*value)) }())
	e.loadState["optional_double"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredDate() time.Time {
	val, _ := e.base.GetDynamic("required_date")
	res, _ := val.TryDate()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredDate(value time.Time) *ScalarTypeProbe {
	e.base.PutDynamic("required_date", core.ValDate(value))
	e.dirtyFields["required_date"] = true
	e.root.Set(e.EntityKey(), "required_date", core.ValDate(value))
	e.loadState["required_date"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalDate() *time.Time {
	val, _ := e.base.GetDynamic("optional_date")
	if val.IsNull() { return nil }
	res, _ := val.TryDate()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalDate(value *time.Time) *ScalarTypeProbe {
	e.base.PutDynamic("optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }())
	e.dirtyFields["optional_date"] = true
	e.root.Set(e.EntityKey(), "optional_date", func() core.Value { if value == nil { return core.ValNull() }; return core.ValDate((*value)) }())
	e.loadState["optional_date"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredTime() core.TimeOfDay {
	val, _ := e.base.GetDynamic("required_time")
	res, _ := val.TryTimeOfDay()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredTime(value core.TimeOfDay) *ScalarTypeProbe {
	e.base.PutDynamic("required_time", core.ValTimeOfDay(value))
	e.dirtyFields["required_time"] = true
	e.root.Set(e.EntityKey(), "required_time", core.ValTimeOfDay(value))
	e.loadState["required_time"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalTime() *core.TimeOfDay {
	val, _ := e.base.GetDynamic("optional_time")
	if val.IsNull() { return nil }
	res, _ := val.TryTimeOfDay()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalTime(value *core.TimeOfDay) *ScalarTypeProbe {
	e.base.PutDynamic("optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }())
	e.dirtyFields["optional_time"] = true
	e.root.Set(e.EntityKey(), "optional_time", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimeOfDay((*value)) }())
	e.loadState["optional_time"] = true
	return e
}

func (e *ScalarTypeProbe) RequiredDatetime() time.Time {
	val, _ := e.base.GetDynamic("required_datetime")
	res, _ := val.TryTime()
	return res}

func (e *ScalarTypeProbe) UpdateRequiredDatetime(value time.Time) *ScalarTypeProbe {
	e.base.PutDynamic("required_datetime", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["required_datetime"] = true
	e.root.Set(e.EntityKey(), "required_datetime", core.ValTimestamp(value.UnixMilli()))
	e.loadState["required_datetime"] = true
	return e
}

func (e *ScalarTypeProbe) OptionalDatetime() *time.Time {
	val, _ := e.base.GetDynamic("optional_datetime")
	if val.IsNull() { return nil }
	res, _ := val.TryTime()
	return &res}

func (e *ScalarTypeProbe) UpdateOptionalDatetime(value *time.Time) *ScalarTypeProbe {
	e.base.PutDynamic("optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }())
	e.dirtyFields["optional_datetime"] = true
	e.root.Set(e.EntityKey(), "optional_datetime", func() core.Value { if value == nil { return core.ValNull() }; return core.ValTimestamp((*value).UnixMilli()) }())
	e.loadState["optional_datetime"] = true
	return e
}

func (e *ScalarTypeProbe) CreateTime() time.Time {
	val, _ := e.base.GetDynamic("create_time")
	res, _ := val.TryTime()
	return res}

func (e *ScalarTypeProbe) UpdateCreateTime(value time.Time) *ScalarTypeProbe {
	e.base.PutDynamic("create_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["create_time"] = true
	e.root.Set(e.EntityKey(), "create_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["create_time"] = true
	return e
}

func (e *ScalarTypeProbe) UpdateTime() time.Time {
	val, _ := e.base.GetDynamic("update_time")
	res, _ := val.TryTime()
	return res}

func (e *ScalarTypeProbe) UpdateUpdateTime(value time.Time) *ScalarTypeProbe {
	e.base.PutDynamic("update_time", core.ValTimestamp(value.UnixMilli()))
	e.dirtyFields["update_time"] = true
	e.root.Set(e.EntityKey(), "update_time", core.ValTimestamp(value.UnixMilli()))
	e.loadState["update_time"] = true
	return e
}

func (e *ScalarTypeProbe) Version() int64 {
	return e.base.Version
}

func (e *ScalarTypeProbe) UpdateVersion(value int64) *ScalarTypeProbe {
	e.base.Version = value
	e.loadState["version"] = true
	return e
}
func (e *ScalarTypeProbe) TypePlatformId() uint64 {
	val, _ := e.base.GetDynamic("type_platform_id")
	res, _ := val.TryU64()
	return res
}

func (e *ScalarTypeProbe) UpdateTypePlatformId(value uint64) *ScalarTypeProbe {
	previous, present := e.base.GetDynamic("type_platform_id")
	e.base.PutDynamic("type_platform_id", core.ValU64(value))
	e.dirtyFields["type_platform_id"] = true
	current, _ := e.base.GetDynamic("type_platform_id")
	if !present || !reflect.DeepEqual(previous, current) {
		delete(e.relations, "typePlatformEntity")
		delete(e.loadedRelations, "typePlatformEntity")
	}
	e.root.Set(e.EntityKey(), "type_platform", core.ValU64(value))
	e.loadState["type_platform_id"] = true
	return e
}
// DEBUG: constantObjectField is false

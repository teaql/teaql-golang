package runtime

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/teaql/teaql-golang/core"
)

type RawAuditEventKind int

const (
	RawAuditEventKindCreated RawAuditEventKind = iota
	RawAuditEventKindUpdated
	RawAuditEventKindDeleted
	RawAuditEventKindRecovered
	RawAuditEventKindSchemaCreated
	RawAuditEventKindSchemaVerified
	RawAuditEventKindFieldAdded
	RawAuditEventKindDataSeeded
)

type EntityPropertyChange struct {
	Field    string
	OldValue *core.Value
	NewValue *core.Value
}

type RawAuditEvent struct {
	Kind          RawAuditEventKind
	Entity        string
	Values        core.Record
	UpdatedFields []string
	OldValues     *core.Record
	NewValues     *core.Record
	Changes       []*EntityPropertyChange
	TraceChain    []*core.TraceNode
}

func Created(entity string, values core.Record) *RawAuditEvent {
	changes := make([]*EntityPropertyChange, 0, len(values))
	for k, v := range values {
		val := v
		changes = append(changes, &EntityPropertyChange{
			Field:    k,
			OldValue: nil,
			NewValue: &val,
		})
	}
	newValues := values
	return &RawAuditEvent{
		Kind:          RawAuditEventKindCreated,
		Entity:        entity,
		Values:        values,
		UpdatedFields: []string{},
		OldValues:     nil,
		NewValues:     &newValues,
		Changes:       changes,
		TraceChain:    []*core.TraceNode{},
	}
}

func Updated(entity string, values core.Record) *RawAuditEvent {
	updatedFields := make([]string, 0, len(values))
	for k := range values {
		updatedFields = append(updatedFields, k)
	}
	changes := changesForFields(nil, &values, updatedFields)
	newValues := values
	return &RawAuditEvent{
		Kind:          RawAuditEventKindUpdated,
		Entity:        entity,
		Values:        values,
		UpdatedFields: updatedFields,
		OldValues:     nil,
		NewValues:     &newValues,
		Changes:       changes,
		TraceChain:    []*core.TraceNode{},
	}
}

func UpdatedWithOldValues(entity string, values core.Record, oldValues *core.Record, newValues core.Record, updatedFields []string) *RawAuditEvent {
	changes := changesForFields(oldValues, &newValues, updatedFields)
	return &RawAuditEvent{
		Kind:          RawAuditEventKindUpdated,
		Entity:        entity,
		Values:        values,
		UpdatedFields: updatedFields,
		OldValues:     oldValues,
		NewValues:     &newValues,
		Changes:       changes,
		TraceChain:    []*core.TraceNode{},
	}
}

func Deleted(entity string, id core.Value, expectedVersion *int64) *RawAuditEvent {
	values := core.Record{"id": id}
	if expectedVersion != nil {
		values["version"] = core.ValI64(*expectedVersion)
	}
	return &RawAuditEvent{
		Kind:          RawAuditEventKindDeleted,
		Entity:        entity,
		Values:        values,
		UpdatedFields: []string{},
		OldValues:     nil,
		NewValues:     nil,
		Changes:       []*EntityPropertyChange{},
		TraceChain:    []*core.TraceNode{},
	}
}

func DeletedWithOldValues(entity string, id core.Value, expectedVersion *int64, oldValues *core.Record) *RawAuditEvent {
	event := Deleted(entity, id, expectedVersion)
	if oldValues != nil {
		changes := make([]*EntityPropertyChange, 0, len(*oldValues))
		for k, v := range *oldValues {
			val := v
			changes = append(changes, &EntityPropertyChange{
				Field:    k,
				OldValue: &val,
				NewValue: nil,
			})
		}
		event.Changes = changes
	}
	event.OldValues = oldValues
	return event
}

func Recovered(entity string, id core.Value, expectedVersion int64) *RawAuditEvent {
	values := core.Record{
		"id":      id,
		"version": core.ValI64(expectedVersion),
	}
	return &RawAuditEvent{
		Kind:          RawAuditEventKindRecovered,
		Entity:        entity,
		Values:        values,
		UpdatedFields: []string{},
		OldValues:     nil,
		NewValues:     nil,
		Changes:       []*EntityPropertyChange{},
		TraceChain:    []*core.TraceNode{},
	}
}

func RecoveredWithOldValues(entity string, id core.Value, expectedVersion int64, oldValues *core.Record) *RawAuditEvent {
	recoveredVersion := -expectedVersion + 1
	var newValues core.Record
	if oldValues != nil {
		newValues = make(core.Record)
		for k, v := range *oldValues {
			newValues[k] = v
		}
	} else {
		newValues = make(core.Record)
	}
	newValues["id"] = id
	newValues["version"] = core.ValI64(recoveredVersion)

	event := Recovered(entity, id, expectedVersion)
	event.OldValues = oldValues
	event.NewValues = &newValues
	event.Changes = changesForFields(oldValues, &newValues, []string{"version"})
	return event
}

func SchemaCreated(entity, tableName string, fieldCount int) *RawAuditEvent {
	values := core.Record{
		"table_name":  core.ValText(tableName),
		"field_count": core.ValI64(int64(fieldCount)),
	}
	changes := make([]*EntityPropertyChange, 0, len(values))
	for k, v := range values {
		val := v
		changes = append(changes, &EntityPropertyChange{
			Field:    k,
			OldValue: nil,
			NewValue: &val,
		})
	}
	return &RawAuditEvent{
		Kind:          RawAuditEventKindSchemaCreated,
		Entity:        entity,
		Values:        values,
		UpdatedFields: []string{},
		OldValues:     nil,
		NewValues:     nil,
		Changes:       changes,
		TraceChain:    []*core.TraceNode{},
	}
}

func SchemaVerified(entity, tableName string, fieldCount int) *RawAuditEvent {
	event := SchemaCreated(entity, tableName, fieldCount)
	event.Kind = RawAuditEventKindSchemaVerified
	return event
}

func FieldAdded(entity, tableName, fieldName string) *RawAuditEvent {
	values := core.Record{
		"table_name": core.ValText(tableName),
		"field_name": core.ValText(fieldName),
	}
	changes := make([]*EntityPropertyChange, 0, len(values))
	for k, v := range values {
		val := v
		changes = append(changes, &EntityPropertyChange{
			Field:    k,
			OldValue: nil,
			NewValue: &val,
		})
	}
	return &RawAuditEvent{
		Kind:          RawAuditEventKindFieldAdded,
		Entity:        entity,
		Values:        values,
		UpdatedFields: []string{},
		OldValues:     nil,
		NewValues:     nil,
		Changes:       changes,
		TraceChain:    []*core.TraceNode{},
	}
}

func DataSeeded(entity, tableName string, inserted, updated int) *RawAuditEvent {
	values := core.Record{
		"table_name": core.ValText(tableName),
		"inserted":   core.ValI64(int64(inserted)),
		"updated":    core.ValI64(int64(updated)),
	}
	changes := make([]*EntityPropertyChange, 0, len(values))
	for k, v := range values {
		val := v
		changes = append(changes, &EntityPropertyChange{
			Field:    k,
			OldValue: nil,
			NewValue: &val,
		})
	}
	return &RawAuditEvent{
		Kind:          RawAuditEventKindDataSeeded,
		Entity:        entity,
		Values:        values,
		UpdatedFields: []string{},
		OldValues:     nil,
		NewValues:     nil,
		Changes:       changes,
		TraceChain:    []*core.TraceNode{},
	}
}

func changesForFields(oldValues *core.Record, newValues *core.Record, fields []string) []*EntityPropertyChange {
	changes := make([]*EntityPropertyChange, 0, len(fields))
	for _, field := range fields {
		var oldVal *core.Value
		var newVal *core.Value
		if oldValues != nil {
			if v, ok := (*oldValues)[field]; ok {
				oldVal = &v
			}
		}
		if newValues != nil {
			if v, ok := (*newValues)[field]; ok {
				newVal = &v
			}
		}
		changes = append(changes, &EntityPropertyChange{
			Field:    field,
			OldValue: oldVal,
			NewValue: newVal,
		})
	}
	return changes
}

func MaskAuditValue(value string) string {
	chars := []rune(value)
	length := len(chars)

	if length == 0 {
		return ""
	}

	allDigits := true
	for _, c := range chars {
		if c < '0' || c > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return strings.Repeat("*", length)
	}

	if length < 8 {
		return strings.Repeat("*", length)
	}

	prefix := string(chars[0:2])
	suffix := string(chars[length-2 : length])
	middle := strings.Repeat("*", length-4)
	return prefix + middle + suffix
}

func LimitAuditValue(value string, maxLen int) (string, bool) {
	chars := []rune(value)
	length := len(chars)

	if length <= maxLen {
		return value, false
	}

	if maxLen <= 3 {
		return strings.Repeat("*", maxLen), true
	}

	marker := "..."
	keepLen := maxLen - len(marker)
	headLen := keepLen / 2
	tailLen := keepLen - headLen

	head := string(chars[0:headLen])
	tail := string(chars[length-tailLen : length])

	return head + marker + tail, true
}

func BuildSafeAuditField(fieldName string, rawValue *string, auditMaskFields []string, auditValueMaxLen *int) *SafeAuditField {
	if rawValue == nil {
		return &SafeAuditField{
			Name:           fieldName,
			Value:          nil,
			Masked:         false,
			Truncated:      false,
			RawLength:      nil,
			OutputLength:   nil,
			MaskReason:     nil,
			TruncateReason: nil,
		}
	}

	raw := *rawValue
	rawLength := utf8.RuneCountInString(raw)

	shouldMask := false
	for _, f := range auditMaskFields {
		if f == fieldName {
			shouldMask = true
			break
		}
	}

	value := raw
	if shouldMask {
		value = MaskAuditValue(raw)
	}

	truncated := false
	if auditValueMaxLen != nil {
		value, truncated = LimitAuditValue(value, *auditValueMaxLen)
	}

	outputLength := utf8.RuneCountInString(value)

	var maskReason *string
	if shouldMask {
		reason := "_audit_mask_fields"
		maskReason = &reason
	}

	var truncateReason *string
	if truncated {
		reason := "_audit_value_max_len"
		truncateReason = &reason
	}

	return &SafeAuditField{
		Name:           fieldName,
		Value:          &value,
		Masked:         shouldMask,
		Truncated:      truncated,
		RawLength:      &rawLength,
		OutputLength:   &outputLength,
		MaskReason:     maskReason,
		TruncateReason: truncateReason,
	}
}

func (e *RawAuditEvent) BuildSafeEvent(auditMaskFields []string, auditValueMaxLen *int) *SafeAuditEvent {
	safeFields := make([]*SafeAuditField, 0, len(e.Changes))
	for _, change := range e.Changes {
		if strings.HasPrefix(change.Field, "_") {
			continue
		}
		var rawValStr *string
		if change.NewValue != nil {
			str := fmt.Sprintf("%v", change.NewValue.V)
			rawValStr = &str
		}
		safeFields = append(safeFields, BuildSafeAuditField(change.Field, rawValStr, auditMaskFields, auditValueMaxLen))
	}

	return &SafeAuditEvent{
		Kind:       e.Kind,
		Entity:     e.Entity,
		Fields:     safeFields,
		TraceChain: e.TraceChain,
	}
}

type SafeAuditField struct {
	Name           string
	Value          *string
	Masked         bool
	Truncated      bool
	RawLength      *int
	OutputLength   *int
	MaskReason     *string
	TruncateReason *string
}

type SafeAuditEvent struct {
	Kind       RawAuditEventKind
	Entity     string
	Fields     []*SafeAuditField
	TraceChain []*core.TraceNode
}

type RawAuditEventSink interface {
	OnEvent(context *UserContext, event *RawAuditEvent) error
}

type InMemoryRawAuditEventSink struct {
	Sinks []RawAuditEventSink
}

func NewInMemoryRawAuditEventSink() *InMemoryRawAuditEventSink {
	return &InMemoryRawAuditEventSink{
		Sinks: []RawAuditEventSink{},
	}
}

func (s *InMemoryRawAuditEventSink) Register(sink RawAuditEventSink) {
	s.Sinks = append(s.Sinks, sink)
}

func (s *InMemoryRawAuditEventSink) WithSink(sink RawAuditEventSink) *InMemoryRawAuditEventSink {
	s.Register(sink)
	return s
}

func (s *InMemoryRawAuditEventSink) OnEvent(context *UserContext, event *RawAuditEvent) error {
	for _, sink := range s.Sinks {
		if err := sink.OnEvent(context, event); err != nil {
			return err
		}
	}
	return nil
}

type AppAuditEventSink interface {
	OnSafeEvent(context *UserContext, event *SafeAuditEvent) error
}

// Package entityjson implements model-owned, bounded JSON without persistence,
// reflection over domain members, context lookup or transport authorization.
package entityjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

type Kind string

const (
	ID        Kind = "id"
	Version   Kind = "version"
	Integer   Kind = "integer"
	Long      Kind = "long"
	Float     Kind = "float"
	Double    Kind = "double"
	Decimal   Kind = "decimal"
	Text      Kind = "text"
	Boolean   Kind = "boolean"
	Date      Kind = "date"
	Time      Kind = "time"
	DateTime  Kind = "datetime"
	JSON      Kind = "json"
	Reference Kind = "reference"
	Many      Kind = "many"
)

// RecordName is the provider/runtime key, not the canonical or wire spelling.
// Reference fields keep the selected FK at RecordName and optional target detail
// at RelationName. The codec never substitutes missing detail for a known FK.
type Field struct {
	CanonicalName string
	RecordName    string
	Kind          Kind
	Target        string
	RelationName  string
}

type Schema struct {
	EntityType string
	Fields     []Field
}

// Document exposes only a shallow snapshot of loaded generated fields. Related
// values remain Documents until visited under the codec's depth/cycle guards.
type Document interface {
	EntityName() string
	LoadedJSONFields() core.Record
}

// QueryPropertyDocument exposes an explicitly captured, output-only projection.
// Unknown loaded fields and arbitrary private dynamic state remain excluded.
// Parsing a JSON document never populates these properties or mutation intent.
type QueryPropertyDocument interface{ JSONQueryProperties() core.Record }

type Limits struct{ MaxBytes, MaxDepth int }

type Error struct {
	Code         string
	Location     runtime.ObjectLocation
	InstancePath string
	Message      string
}

func (e *Error) Error() string { return e.Code + " at " + e.InstancePath + ": " + e.Message }

type WireError struct {
	Code         string                        `json:"code"`
	Location     []runtime.WireLocationSegment `json:"location"`
	InstancePath string                        `json:"instancePath"`
	Message      string                        `json:"message"`
}

func (e *Error) ToWire() WireError {
	segments := make([]runtime.WireLocationSegment, 0, len(e.Location.Segments))
	for _, segment := range e.Location.Segments {
		if segment.Property != nil {
			segments = append(segments, runtime.WireLocationSegment{Kind: "property", Name: *segment.Property})
		} else {
			segments = append(segments, runtime.WireLocationSegment{Kind: "index", Index: segment.Index})
		}
	}
	return WireError{e.Code, segments, e.InstancePath, e.Message}
}

func (e *Error) MarshalJSON() ([]byte, error) { return json.Marshal(e.ToWire()) }

type schema struct {
	fields []Field
	byKey  map[string]Field
	names  map[runtime.JsonFieldNamingProfile]map[string]Field
}

// Model is an immutable, model-local registry. It does not register entity names
// globally, and its copied definitions cannot be changed through caller slices.
type Model struct {
	schemas map[string]*schema
	profile runtime.JsonFieldNamingProfile
	limits  Limits
}

var profiles = []runtime.JsonFieldNamingProfile{runtime.JsonFieldCamelCase, runtime.JsonFieldSnakeCase, runtime.JsonFieldPascalCase}
var integral = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
var jsonNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
var plainDecimal = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)
var instantText = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?(Z|[+-]([01][0-9]|2[0-3]):[0-5][0-9])$`)

func validProfile(profile runtime.JsonFieldNamingProfile) bool {
	return profile == runtime.JsonFieldCamelCase || profile == runtime.JsonFieldSnakeCase || profile == runtime.JsonFieldPascalCase
}

func NewModel(definitions []Schema, profile runtime.JsonFieldNamingProfile, limits Limits) (*Model, error) {
	if !validProfile(profile) || len(definitions) == 0 || limits.MaxBytes < 0 || limits.MaxDepth < 0 {
		return nil, fmt.Errorf("invalid entity JSON model settings")
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1024 * 1024
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = 64
	}
	if limits.MaxDepth > 64 {
		return nil, fmt.Errorf("entity JSON depth must not exceed 64")
	}
	m := &Model{schemas: make(map[string]*schema), profile: profile, limits: limits}
	for _, definition := range definitions {
		if strings.TrimSpace(definition.EntityType) == "" || m.schemas[definition.EntityType] != nil {
			return nil, fmt.Errorf("invalid or duplicate entity JSON type")
		}
		s := &schema{fields: append([]Field(nil), definition.Fields...), byKey: make(map[string]Field), names: make(map[runtime.JsonFieldNamingProfile]map[string]Field)}
		canonical, keys := make(map[string]bool), make(map[string]bool)
		for i, field := range s.fields {
			if field.RecordName == "" {
				field.RecordName = field.CanonicalName
				s.fields[i] = field
			}
			if strings.TrimSpace(field.CanonicalName) == "" || canonical[field.CanonicalName] || keys[field.RecordName] {
				return nil, fmt.Errorf("invalid or duplicate entity JSON field")
			}
			switch field.Kind {
			case ID, Version, Integer, Long, Float, Double, Decimal, Text, Boolean, Date, Time, DateTime, JSON:
				if field.Target != "" || field.RelationName != "" {
					return nil, fmt.Errorf("scalar JSON field has relation metadata")
				}
			case Reference:
				if field.Target == "" || field.RelationName == "" || field.RelationName == field.RecordName || keys[field.RelationName] {
					return nil, fmt.Errorf("invalid JSON reference metadata")
				}
				keys[field.RelationName] = true
			case Many:
				if field.Target == "" || field.RelationName != "" {
					return nil, fmt.Errorf("invalid JSON collection metadata")
				}
			default:
				return nil, fmt.Errorf("unknown entity JSON kind")
			}
			canonical[field.CanonicalName], keys[field.RecordName] = true, true
			s.byKey[field.RecordName] = field
		}
		for _, p := range profiles {
			names := make(map[string]Field)
			for _, field := range s.fields {
				for _, spelling := range []string{p.Render(field.CanonicalName), field.CanonicalName} {
					if previous, ok := names[spelling]; ok && previous.CanonicalName != field.CanonicalName {
						return nil, fmt.Errorf("ambiguous entity JSON field spelling")
					}
					names[spelling] = field
				}
			}
			s.names[p] = names
		}
		m.schemas[definition.EntityType] = s
	}
	for _, s := range m.schemas {
		for _, field := range s.fields {
			if field.Target != "" && m.schemas[field.Target] == nil {
				return nil, fmt.Errorf("unknown entity JSON relation target")
			}
		}
	}
	return m, nil
}

func (m *Model) checkedProfile(options []runtime.JsonFieldNamingProfile) (runtime.JsonFieldNamingProfile, error) {
	if len(options) > 1 {
		return "", fmt.Errorf("at most one entity JSON naming profile is allowed")
	}
	profile := m.profile
	if len(options) == 1 {
		profile = options[0]
	}
	if !validProfile(profile) {
		return "", fmt.Errorf("unsupported entity JSON naming profile")
	}
	return profile, nil
}

func failure(code string, location runtime.ObjectLocation, pointer string) *Error {
	return &Error{Code: code, Location: location, InstancePath: pointer, Message: "Expected the declared bounded model JSON type"}
}
func pointer(parent, child string) string {
	return parent + "/" + strings.ReplaceAll(strings.ReplaceAll(child, "~", "~0"), "/", "~1")
}

type member struct {
	name  string
	value node
}
type node struct{ value any }
type object []member
type array []node

// Token-based reading retains repeated keys; decoding directly into a map would
// silently discard the evidence needed to reject aliases and duplicate fields.
func (m *Model) readNode(decoder *json.Decoder, depth int) (node, error) {
	token, err := decoder.Token()
	if err != nil {
		return node{}, failure("WIRE_INVALID_JSON", runtime.Location(), "")
	}
	if delim, ok := token.(json.Delim); ok {
		if depth >= m.limits.MaxDepth {
			return node{}, failure("WIRE_JSON_LIMIT", runtime.Location(), "")
		}
		switch delim {
		case '{':
			values := object{}
			for decoder.More() {
				key, err := decoder.Token()
				name, ok := key.(string)
				if err != nil || !ok {
					return node{}, failure("WIRE_INVALID_JSON", runtime.Location(), "")
				}
				value, err := m.readNode(decoder, depth+1)
				if err != nil {
					return node{}, err
				}
				values = append(values, member{name, value})
			}
			end, err := decoder.Token()
			if err != nil || end != json.Delim('}') {
				return node{}, failure("WIRE_INVALID_JSON", runtime.Location(), "")
			}
			return node{values}, nil
		case '[':
			values := array{}
			for decoder.More() {
				value, err := m.readNode(decoder, depth+1)
				if err != nil {
					return node{}, err
				}
				values = append(values, value)
			}
			end, err := decoder.Token()
			if err != nil || end != json.Delim(']') {
				return node{}, failure("WIRE_INVALID_JSON", runtime.Location(), "")
			}
			return node{values}, nil
		default:
			return node{}, failure("WIRE_INVALID_JSON", runtime.Location(), "")
		}
	}
	return node{token}, nil
}

// encoding/json replaces lone UTF-16 surrogates with U+FFFD. Reject them before
// tokenization so the boundary never silently repairs caller text.
func validJSONText(payload []byte) bool {
	if !utf8.Valid(payload) {
		return false
	}
	quoted := false
	for i := 0; i < len(payload); i++ {
		if payload[i] == '"' {
			quoted = !quoted
			continue
		}
		if !quoted || payload[i] != '\\' {
			continue
		}
		i++
		if i >= len(payload) {
			return false
		}
		if payload[i] != 'u' {
			continue
		}
		if i+4 >= len(payload) {
			return false
		}
		code, err := strconv.ParseUint(string(payload[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if code >= 0xdc00 && code <= 0xdfff {
			return false
		}
		if code >= 0xd800 && code <= 0xdbff {
			if i+6 >= len(payload) || payload[i+1] != '\\' || payload[i+2] != 'u' {
				return false
			}
			low, err := strconv.ParseUint(string(payload[i+3:i+7]), 16, 16)
			if err != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}

func (m *Model) Parse(entityType string, payload []byte, options ...runtime.JsonFieldNamingProfile) (core.Record, error) {
	profile, err := m.checkedProfile(options)
	if err != nil {
		return nil, err
	}
	if len(payload) > m.limits.MaxBytes {
		return nil, failure("WIRE_JSON_LIMIT", runtime.Location(), "")
	}
	if !validJSONText(payload) {
		return nil, failure("WIRE_INVALID_TEXT", runtime.Location(), "")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	root, err := m.readNode(decoder, 0)
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, failure("WIRE_INVALID_JSON", runtime.Location(), "")
	}
	return m.readEntity(entityType, root, profile, runtime.Location(), "")
}

func (m *Model) readEntity(entityType string, value node, profile runtime.JsonFieldNamingProfile, location runtime.ObjectLocation, path string) (core.Record, error) {
	s := m.schemas[entityType]
	if s == nil {
		return nil, failure("WIRE_UNKNOWN_ENTITY", location, path)
	}
	properties, ok := value.value.(object)
	if !ok {
		return nil, failure("WIRE_ENTITY_OBJECT_REQUIRED", location, path)
	}
	record := make(core.Record)
	for _, property := range properties {
		childPath := pointer(path, property.name)
		field, ok := s.names[profile][property.name]
		if !ok {
			return nil, failure("WIRE_UNKNOWN_FIELD", location.Property(property.name), childPath)
		}
		childLocation := location.Property(field.CanonicalName)
		if _, duplicate := record[field.RecordName]; duplicate {
			return nil, failure("WIRE_FIELD_COLLISION", childLocation, childPath)
		}
		value, err := m.readValue(field, property.value, profile, childLocation, childPath)
		if err != nil {
			return nil, err
		}
		if field.Kind == Reference {
			if target, object := value.TryObject(); object {
				id, known := target["id"]
				if _, valid := id.V.(uint64); !known || !valid {
					return nil, failure("WIRE_REFERENCE_ID_REQUIRED", childLocation, childPath)
				}
				record[field.RecordName], record[field.RelationName] = id, value
			} else {
				record[field.RecordName] = value
				if value.IsNull() {
					record[field.RelationName] = core.ValNull()
				} else {
					record[field.RelationName] = core.ValObject(core.Record{"id": value})
				}
			}
		} else {
			record[field.RecordName] = value
		}
	}
	return record, nil
}

func (m *Model) readValue(field Field, value node, profile runtime.JsonFieldNamingProfile, location runtime.ObjectLocation, path string) (core.Value, error) {
	bad := func() (core.Value, error) { return core.Value{}, failure("WIRE_INVALID_TYPE", location, path) }
	if value.value == nil {
		if field.Kind == Many || field.Kind == ID || field.Kind == Version {
			return bad()
		}
		return core.ValNull(), nil
	}
	text, stringValue := value.value.(string)
	switch field.Kind {
	case ID, Long, Version:
		if !stringValue || len(text) > 20 || !integral.MatchString(text) {
			return bad()
		}
		if field.Kind == ID {
			number, err := strconv.ParseUint(text, 10, 64)
			if err != nil {
				return bad()
			}
			return core.ValU64(number), nil
		}
		number, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return bad()
		}
		return core.ValI64(number), nil
	case Integer:
		text, ok := value.value.(json.Number)
		if !ok {
			return bad()
		}
		number, err := strconv.ParseInt(string(text), 10, 32)
		if err != nil {
			return bad()
		}
		return core.ValI64(number), nil
	case Text:
		if !stringValue {
			return bad()
		}
		return core.ValText(text), nil
	case Boolean:
		boolean, ok := value.value.(bool)
		if !ok {
			return bad()
		}
		return core.ValBool(boolean), nil
	case Decimal:
		if !stringValue || len(text) > 4096 || !plainDecimal.MatchString(text) {
			return bad()
		}
		number, err := decimal.NewFromString(text)
		if err != nil {
			return bad()
		}
		return core.ValDecimal(number), nil
	case Float, Double:
		text, ok := value.value.(json.Number)
		if !ok {
			return bad()
		}
		number, err := strconv.ParseFloat(string(text), 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return bad()
		}
		if field.Kind == Float {
			number = float64(float32(number))
			if math.IsInf(number, 0) {
				return bad()
			}
		}
		return core.ValF64(number), nil
	case Date:
		if !stringValue || len(text) != 10 {
			return bad()
		}
		date, err := time.Parse("2006-01-02", text)
		if err != nil || date.Year() < 1 {
			return bad()
		}
		return core.ValDate(date), nil
	case Time:
		if !stringValue {
			return bad()
		}
		clock, err := core.ParseTimeOfDay(text)
		if err != nil {
			return bad()
		}
		return core.ValTimeOfDay(clock), nil
	case DateTime:
		if !stringValue || !instantText.MatchString(text) {
			return bad()
		}
		instant, err := time.Parse(time.RFC3339Nano, text)
		if err != nil || instant.UTC().Year() < 1 || instant.UTC().Year() > 9999 {
			return bad()
		}
		if instant.Nanosecond()%int(time.Millisecond) != 0 {
			return core.Value{}, failure("WIRE_TEMPORAL_PRECISION", location, path)
		}
		return core.ValTimestamp(instant.UnixMilli()), nil
	case Reference:
		if _, objectValue := value.value.(object); objectValue {
			record, err := m.readEntity(field.Target, value, profile, location, path)
			return core.ValObject(record), err
		}
		return m.readValue(Field{Kind: ID}, value, profile, location, path)
	case Many:
		items, ok := value.value.(array)
		if !ok {
			return bad()
		}
		records := make([]core.Value, 0, len(items))
		for index, item := range items {
			record, err := m.readEntity(field.Target, item, profile, location.At(index), pointer(path, strconv.Itoa(index)))
			if err != nil {
				return core.Value{}, err
			}
			records = append(records, core.ValObject(record))
		}
		return core.ValList(records), nil
	case JSON:
		raw, err := m.readJSON(value, location, path)
		return core.ValJson(raw), err
	}
	return bad()
}

func (m *Model) readJSON(value node, location runtime.ObjectLocation, path string) (any, error) {
	switch raw := value.value.(type) {
	case object:
		result := make(map[string]any, len(raw))
		for _, property := range raw {
			childPath := pointer(path, property.name)
			if _, found := result[property.name]; found {
				return nil, failure("WIRE_FIELD_COLLISION", location.Property(property.name), childPath)
			}
			child, err := m.readJSON(property.value, location.Property(property.name), childPath)
			if err != nil {
				return nil, err
			}
			result[property.name] = child
		}
		return result, nil
	case array:
		result := make([]any, len(raw))
		for i, child := range raw {
			item, err := m.readJSON(child, location.At(i), pointer(path, strconv.Itoa(i)))
			if err != nil {
				return nil, err
			}
			result[i] = item
		}
		return result, nil
	default:
		return raw, nil
	}
}

type writer struct {
	bytes.Buffer
	max int
}

func (w *writer) add(text string, location runtime.ObjectLocation, path string) error {
	if len(text) > w.max-w.Len() {
		return failure("WIRE_JSON_LIMIT", location, path)
	}
	w.WriteString(text)
	return nil
}
func (w *writer) quoted(text string, location runtime.ObjectLocation, path string) error {
	if !utf8.ValidString(text) {
		return failure("WIRE_INVALID_TEXT", location, path)
	}
	if len(text) > w.max-w.Len() {
		return failure("WIRE_JSON_LIMIT", location, path)
	}
	encoded, err := json.Marshal(text)
	if err != nil {
		return failure("WIRE_INVALID_TEXT", location, path)
	}
	return w.add(string(encoded), location, path)
}

// Serialize accepts a generated Document or a selected runtime Record. It does
// not mutate the record, fill omitted fields or expand a scalar FK via I/O.
func (m *Model) Serialize(entityType string, value any, options ...runtime.JsonFieldNamingProfile) ([]byte, error) {
	profile, err := m.checkedProfile(options)
	if err != nil {
		return nil, err
	}
	w := &writer{max: m.limits.MaxBytes}
	if err := m.writeEntity(w, entityType, value, profile, runtime.Location(), "", 0, make(map[uintptr]bool)); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *Model) writeEntity(w *writer, entityType string, value any, profile runtime.JsonFieldNamingProfile, location runtime.ObjectLocation, path string, depth int, active map[uintptr]bool) error {
	if depth >= m.limits.MaxDepth {
		return failure("WIRE_JSON_LIMIT", location, path)
	}
	s := m.schemas[entityType]
	if s == nil {
		return failure("WIRE_UNKNOWN_ENTITY", location, path)
	}
	var record core.Record
	var identity uintptr
	switch source := value.(type) {
	case core.Record:
		if source == nil {
			return failure("WIRE_ENTITY_OBJECT_REQUIRED", location, path)
		}
		record, identity = source, uintptr(reflect.ValueOf(source).UnsafePointer())
	case Document:
		if reflect.ValueOf(source).Kind() != reflect.Ptr || reflect.ValueOf(source).IsNil() || source.EntityName() != entityType {
			return failure("WIRE_INVALID_TYPE", location, path)
		}
		identity = reflect.ValueOf(source).Pointer()
		if active[identity] {
			return failure("WIRE_ENTITY_CYCLE", location, path)
		}
		record = source.LoadedJSONFields()
	default:
		return failure("WIRE_ENTITY_OBJECT_REQUIRED", location, path)
	}
	if active[identity] {
		return failure("WIRE_ENTITY_CYCLE", location, path)
	}
	active[identity] = true
	defer delete(active, identity)
	if err := w.add("{", location, path); err != nil {
		return err
	}
	first := true
	for _, field := range s.fields {
		value, present := record[field.RecordName]
		childLocation, childPath := location.Property(field.CanonicalName), pointer(path, profile.Render(field.CanonicalName))
		if field.Kind == Reference {
			if detail, selected := record[field.RelationName]; selected && !detail.IsNull() {
				var id core.Value
				known := false
				switch target := detail.V.(type) {
				case core.Record:
					id, known = target["id"]
				case Document:
					if reflect.ValueOf(target).Kind() == reflect.Ptr && !reflect.ValueOf(target).IsNil() {
						id, known = target.LoadedJSONFields()["id"]
					}
				}
				if !known || id.IsNull() {
					return failure("WIRE_REFERENCE_ID_REQUIRED", childLocation, childPath)
				}
				if present && (value.IsNull() || known && !sameID(value, id)) {
					return failure("WIRE_REFERENCE_ID_MISMATCH", childLocation, childPath)
				}
				value, present = detail, true
			}
		}
		if !present {
			continue
		}
		if !first {
			if err := w.add(",", childLocation, childPath); err != nil {
				return err
			}
		}
		first = false
		if err := w.quoted(profile.Render(field.CanonicalName), childLocation, childPath); err != nil {
			return err
		}
		if err := w.add(":", childLocation, childPath); err != nil {
			return err
		}
		if err := m.writeValue(w, field, value, profile, childLocation, childPath, depth+1, active); err != nil {
			return err
		}
	}
	if projected, ok := value.(QueryPropertyDocument); ok {
		properties := projected.JSONQueryProperties()
		keys := make([]string, 0, len(properties))
		for name := range properties {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		for _, name := range keys {
			childLocation, childPath := location.Property(name), pointer(path, name)
			_, collision := s.byKey[name]
			for _, names := range s.names {
				if _, known := names[name]; known {
					collision = true
				}
			}
			for _, field := range s.fields {
				if field.RelationName != "" && name == field.RelationName {
					collision = true
				}
			}
			if collision {
				return failure("WIRE_QUERY_PROPERTY_COLLISION", childLocation, childPath)
			}
			if !first {
				if err := w.add(",", childLocation, childPath); err != nil {
					return err
				}
			}
			first = false
			if err := w.quoted(name, childLocation, childPath); err != nil {
				return err
			}
			if err := w.add(":", childLocation, childPath); err != nil {
				return err
			}
			var err error
			switch raw := properties[name].V.(type) {
			case decimal.Decimal:
				// Reuse bounded Decimal validation before formatting an exponent.
				err = m.writeValue(w, Field{Kind: Decimal}, properties[name], profile, childLocation, childPath, depth+1, active)
			case float32:
				if math.IsNaN(float64(raw)) || math.IsInf(float64(raw), 0) {
					return failure("WIRE_INVALID_TYPE", childLocation, childPath)
				}
				err = w.add(strconv.FormatFloat(float64(raw), 'g', -1, 32), childLocation, childPath)
			default:
				err = m.writeJSON(w, raw, childLocation, childPath, depth+1)
			}
			if err != nil {
				return err
			}
		}
	}
	return w.add("}", location, path)
}

func sameID(left, right core.Value) bool {
	unsigned := func(value core.Value) (uint64, bool) {
		switch v := value.V.(type) {
		case uint64:
			return v, true
		case int64:
			return uint64(v), v >= 0
		}
		return 0, false
	}
	a, ok := unsigned(left)
	b, valid := unsigned(right)
	return ok && valid && a == b
}

func (m *Model) writeValue(w *writer, field Field, value core.Value, profile runtime.JsonFieldNamingProfile, location runtime.ObjectLocation, path string, depth int, active map[uintptr]bool) error {
	bad := func() error { return failure("WIRE_INVALID_TYPE", location, path) }
	if value.IsNull() {
		if field.Kind == Many || field.Kind == ID || field.Kind == Version {
			return bad()
		}
		return w.add("null", location, path)
	}
	switch field.Kind {
	case ID:
		var number uint64
		switch raw := value.V.(type) {
		case uint64:
			number = raw
		case int64:
			if raw < 0 {
				return bad()
			}
			number = uint64(raw)
		default:
			return bad()
		}
		return w.quoted(strconv.FormatUint(number, 10), location, path)
	case Version, Long, Integer:
		number, ok := value.V.(int64)
		if !ok {
			return bad()
		}
		if field.Kind == Integer {
			if number < math.MinInt32 || number > math.MaxInt32 {
				return bad()
			}
			return w.add(strconv.FormatInt(number, 10), location, path)
		}
		return w.quoted(strconv.FormatInt(number, 10), location, path)
	case Text:
		text, ok := value.V.(string)
		if !ok {
			return bad()
		}
		return w.quoted(text, location, path)
	case Boolean:
		boolean, ok := value.V.(bool)
		if !ok {
			return bad()
		}
		return w.add(strconv.FormatBool(boolean), location, path)
	case Decimal:
		number, ok := value.V.(decimal.Decimal)
		if !ok {
			return bad()
		}
		if number.Exponent() < -4096 || number.Exponent() > 4096 || number.Coefficient().BitLen() > 13607 || len(number.Coefficient().String()) > 4096 {
			return failure("WIRE_DECIMAL_PRECISION", location, path)
		}
		text := number.String()
		if number.Exponent() < 0 {
			text = number.StringFixed(-number.Exponent())
		}
		if len(text) > 4096 {
			return failure("WIRE_DECIMAL_PRECISION", location, path)
		}
		return w.quoted(text, location, path)
	case Float, Double:
		number, ok := value.V.(float64)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) {
			return bad()
		}
		bits := 64
		if field.Kind == Float {
			if float64(float32(number)) != number {
				return bad()
			}
			bits = 32
		}
		return w.add(strconv.FormatFloat(number, 'g', -1, bits), location, path)
	case Date:
		date, ok := value.V.(time.Time)
		if !ok || date.Year() < 1 || date.Year() > 9999 || date.Hour() != 0 || date.Minute() != 0 || date.Second() != 0 || date.Nanosecond() != 0 {
			return bad()
		}
		return w.quoted(date.Format("2006-01-02"), location, path)
	case Time:
		clock, ok := value.V.(core.TimeOfDay)
		if !ok {
			return bad()
		}
		return w.quoted(clock.String(), location, path)
	case DateTime:
		var instant time.Time
		switch raw := value.V.(type) {
		case core.Timestamp:
			instant = raw.ToTime()
		case time.Time:
			instant = raw.UTC()
		default:
			return bad()
		}
		if instant.Year() < 1 || instant.Year() > 9999 {
			return bad()
		}
		if instant.Nanosecond()%int(time.Millisecond) != 0 {
			return failure("WIRE_TEMPORAL_PRECISION", location, path)
		}
		return w.quoted(instant.Format("2006-01-02T15:04:05.000Z"), location, path)
	case Reference:
		switch value.V.(type) {
		case core.Record, Document:
			return m.writeEntity(w, field.Target, value.V, profile, location, path, depth, active)
		}
		return m.writeValue(w, Field{Kind: ID}, value, profile, location, path, depth, active)
	case Many:
		if depth >= m.limits.MaxDepth {
			return failure("WIRE_JSON_LIMIT", location, path)
		}
		items, ok := value.V.([]core.Value)
		if !ok {
			return bad()
		}
		if err := w.add("[", location, path); err != nil {
			return err
		}
		for index, item := range items {
			childLocation, childPath := location.At(index), pointer(path, strconv.Itoa(index))
			if index > 0 {
				if err := w.add(",", childLocation, childPath); err != nil {
					return err
				}
			}
			if err := m.writeEntity(w, field.Target, item.V, profile, childLocation, childPath, depth+1, active); err != nil {
				return err
			}
		}
		return w.add("]", location, path)
	case JSON:
		return m.writeJSON(w, value.V, location, path, depth)
	}
	return bad()
}

func (m *Model) writeJSON(w *writer, value any, location runtime.ObjectLocation, path string, depth int) error {
	bad := func() error { return failure("WIRE_INVALID_TYPE", location, path) }
	switch raw := value.(type) {
	case nil:
		return w.add("null", location, path)
	case string:
		return w.quoted(raw, location, path)
	case bool:
		return w.add(strconv.FormatBool(raw), location, path)
	case json.Number:
		if !jsonNumber.MatchString(string(raw)) {
			return bad()
		}
		return w.add(string(raw), location, path)
	case int:
		return w.add(strconv.FormatInt(int64(raw), 10), location, path)
	case int64:
		return w.add(strconv.FormatInt(raw, 10), location, path)
	case uint64:
		return w.add(strconv.FormatUint(raw, 10), location, path)
	case float64:
		if math.IsNaN(raw) || math.IsInf(raw, 0) {
			return bad()
		}
		return w.add(strconv.FormatFloat(raw, 'g', -1, 64), location, path)
	case map[string]any:
		if depth >= m.limits.MaxDepth {
			return failure("WIRE_JSON_LIMIT", location, path)
		}
		if err := w.add("{", location, path); err != nil {
			return err
		}
		keys := make([]string, 0, len(raw))
		for key := range raw {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for i, key := range keys {
			childLocation, childPath := location.Property(key), pointer(path, key)
			if i > 0 {
				if err := w.add(",", childLocation, childPath); err != nil {
					return err
				}
			}
			if err := w.quoted(key, childLocation, childPath); err != nil {
				return err
			}
			if err := w.add(":", childLocation, childPath); err != nil {
				return err
			}
			if err := m.writeJSON(w, raw[key], childLocation, childPath, depth+1); err != nil {
				return err
			}
		}
		return w.add("}", location, path)
	case []any:
		if depth >= m.limits.MaxDepth {
			return failure("WIRE_JSON_LIMIT", location, path)
		}
		if err := w.add("[", location, path); err != nil {
			return err
		}
		for i, item := range raw {
			childLocation, childPath := location.At(i), pointer(path, strconv.Itoa(i))
			if i > 0 {
				if err := w.add(",", childLocation, childPath); err != nil {
					return err
				}
			}
			if err := m.writeJSON(w, item, childLocation, childPath, depth+1); err != nil {
				return err
			}
		}
		return w.add("]", location, path)
	default:
		return bad()
	}
}

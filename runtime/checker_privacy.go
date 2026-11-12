package runtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

// Carried by value-copy/relocation, but hidden from JSON and debug formatting.
// This contains only projected values and lengths, never a raw input or Context.
type checkerPresentation struct {
	value      any
	length     int
	credential bool
}

func checkerPresented(result CheckResult) (checkerPresentation, bool) {
	state, ok := logprivacy.ReadProjectionState(result.presentation).(checkerPresentation)
	return state, ok && reflect.DeepEqual(state.value, result.InputValue)
}

func checkerInputLength(result CheckResult) int {
	if state, ok := checkerPresented(result); ok {
		return state.length
	}
	value := result.InputValue
	if typed, ok := value.(core.Value); ok {
		value = typed.V
	}
	if text, ok := value.(string); ok {
		return len([]rune(text))
	}
	return 0
}

func cloneCheckerValue(value any) any {
	if typed, ok := value.(core.Value); ok {
		return core.CloneValue(typed)
	}
	return core.CloneValue(core.Value{V: value}).V
}

// The audit vocabulary is shared; canonical locations remain model identities.
func (c *UserContext) checkerFieldPolicy(entity, field string) string {
	if credentialLogName(field) {
		return "credential"
	}
	descriptor := c.Entity(entity)
	if descriptor == nil {
		return ""
	}
	storage, err := core.StorageLedgerField(descriptor, field)
	if err != nil {
		storage = field
	}
	property := descriptor.PropertyByName(storage)
	if property != nil && (credentialLogName(property.Name) || credentialLogName(property.ColName) || property.LogPolicy == "credential") {
		return "credential"
	}
	for _, masked := range descriptor.AuditMaskFlds {
		canonical, mapErr := core.CanonicalLedgerField(descriptor, masked)
		if masked == field || masked == storage || mapErr == nil && canonical == field {
			return "masked"
		}
	}
	if property != nil && property.LogPolicy == "masked" {
		return "masked"
	}
	return ""
}

type checkerValueInfo struct {
	strings    []string
	credential bool
}

// Inspect supported JSON/value containers without calling fmt on a cyclic map.
// Path-local identities allow repeated shared containers without losing values.
func inspectCheckerValue(value any) checkerValueInfo {
	var info checkerValueInfo
	active := make(map[string]bool)
	var visit func(any)
	visit = func(value any) {
		if typed, ok := value.(core.Value); ok {
			visit(typed.V)
			return
		}
		if value == nil {
			return
		}
		if _, null := value.(core.DataType); null {
			return
		}
		switch value.(type) {
		case time.Time, decimal.Decimal:
			info.strings = append(info.strings, fmt.Sprint(value))
			return
		}
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return
			}
			key := fmt.Sprintf("%T:%p", value, value)
			if active[key] {
				info.credential = true
				return
			} // unsupported cycle: fail closed
			active[key] = true
			defer delete(active, key)
		}
		switch rv.Kind() {
		case reflect.Map:
			if rv.Type().Key().Kind() == reflect.String {
				for _, key := range rv.MapKeys() {
					info.credential = info.credential || credentialLogName(key.String())
					visit(rv.MapIndex(key).Interface())
				}
				return
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < rv.Len(); i++ {
				visit(rv.Index(i).Interface())
			}
			return
		case reflect.Ptr:
			visit(rv.Elem().Interface())
			return
		case reflect.Struct:
			for i := 0; i < rv.NumField(); i++ {
				field := rv.Type().Field(i)
				info.credential = info.credential || credentialLogName(field.Name) || credentialLogName(strings.Split(field.Tag.Get("json"), ",")[0])
				if rv.Field(i).CanInterface() {
					visit(rv.Field(i).Interface())
				} else {
					info.credential = true // Opaque application internals are not safe display values.
				}
			}
			return
		}
		if text := fmt.Sprint(value); text != "" {
			info.strings = append(info.strings, text)
		}
	}
	visit(value)
	return info
}

func checkerRedactions(values []string) []string {
	unique := make(map[string]bool)
	for _, value := range values {
		if value == "" {
			continue
		}
		unique[value] = true
		quoted := strconv.Quote(value)
		unique[quoted[1:len(quoted)-1]] = true
		if encoded, err := json.Marshal(value); err == nil {
			unique[string(encoded[1:len(encoded)-1])] = true
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		if len(result[i]) == len(result[j]) {
			return result[i] < result[j]
		}
		return len(result[i]) > len(result[j])
	})
	return result
}

func scrubCheckerText(text string, values []string) string {
	for _, value := range values {
		text = strings.ReplaceAll(text, value, "[REDACTED]")
	}
	return text
}

func (c *UserContext) checkerKnownValues(entity string, allow bool, records ...core.Record) []string {
	var values []string
	for _, record := range records {
		for field, value := range record {
			info := inspectCheckerValue(value)
			policy := c.checkerFieldPolicy(entity, field)
			if info.credential || policy == "credential" || !allow && (policy == "masked" || strings.HasPrefix(field, "#")) {
				values = append(values, info.strings...)
			}
		}
	}
	return values
}

func checkerNull(value any) bool {
	if typed, ok := value.(core.Value); ok {
		return checkerNull(typed.V)
	}
	if value == nil {
		return true
	}
	_, typedNull := value.(core.DataType)
	return typedNull
}

func checkerNumericConstraint(value any) bool {
	if typed, ok := value.(core.Value); ok {
		return checkerNumericConstraint(typed.V)
	}
	switch value.(type) {
	case nil, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, decimal.Decimal, core.DataType:
		return true
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func (c *UserContext) translateCheckerResults(results []CheckResult, declaring string, known []string, allow bool, configuration ...checkerMessageConfig) {
	if len(results) == 0 {
		return
	}
	for i := range results {
		r := &results[i]
		if declaring != "" && strings.TrimSpace(r.EntityType) == "" {
			r.EntityType = declaring
		}
		field, credential := "", false
		for _, segment := range r.ObjectLocation().Segments {
			if segment.Property != nil {
				field = *segment.Property
				credential = credential || credentialLogName(field)
			}
		}
		policy := c.checkerFieldPolicy(r.EntityType, field)
		info := inspectCheckerValue(r.InputValue)
		credential = credential || policy == "credential" || info.credential
		masked := credential || !allow && (policy == "masked" || strings.HasPrefix(field, "#"))
		old, already := checkerPresented(*r)
		if masked && !(already && (old.credential || !credential)) && !checkerNull(r.InputValue) {
			length := checkerInputLength(*r)
			var value any = "[REDACTED]"
			if !credential {
				if typed, ok := r.InputValue.(core.Value); ok {
					value = businessMaskValue(typed)
				} else {
					value = businessMaskValue(core.Value{V: r.InputValue}).V
				}
			}
			known = append(known, info.strings...)
			r.InputValue = value
			r.presentation = logprivacy.NewProjectionState(checkerPresentation{value: cloneCheckerValue(value), length: length, credential: credential})
		}
		expected := inspectCheckerValue(r.SystemValue)
		if !checkerNumericConstraint(r.SystemValue) && (credential || expected.credential) {
			known = append(known, expected.strings...)
			r.SystemValue = "[REDACTED]"
		}
	}
	known = checkerRedactions(known)
	config := c.captureCheckerMessages()
	if len(configuration) > 0 {
		config = configuration[0]
	}
	seen := map[CheckerMessageWarning]bool{}
	for i := range results {
		automatic := naturalCheckerAutomatic(results[i])
		results[i].Message = scrubCheckerText(results[i].Message, known)
		renderNaturalChecker(config, &results[i], known, automatic, seen, nil)
	}
}

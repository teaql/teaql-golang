package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SearchModel is trusted model metadata, never supplied by the search payload.
type SearchModel struct {
	Fields    map[string]string
	Relations map[string]string
}
type DynamicSearchWarning struct {
	Code      string `json:"code"`
	Entity    string `json:"entity"`
	Clause    string `json:"clause"`
	FieldPath string `json:"fieldPath"`
}
type DynamicSearchOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}
type DynamicSearchResult struct {
	Filter   map[string]map[string]any
	OrderBy  []DynamicSearchOrder
	Warnings []DynamicSearchWarning
}

var decimalSearchValue = regexp.MustCompile(`^[+-]?[0-9]+(?:\.[0-9]+)?$`)

func searchError(message string) error { return errors.New("Dynamic search: " + message) }
func emitSearchWarnings(warnings []DynamicSearchWarning, sink func(DynamicSearchWarning)) {
	for _, warning := range warnings {
		if sink != nil {
			sink(warning)
		} else {
			log.Printf("%s entity=%s clause=%s fieldPath=%s", warning.Code, warning.Entity, warning.Clause, warning.FieldPath)
		}
	}
}

// NormalizeDynamicSearch is a local UI boundary. Federation decoding remains strict.
func NormalizeDynamicSearch(source []byte, entity string, models map[string]SearchModel, maxClauses int, sink func(DynamicSearchWarning)) (*DynamicSearchResult, error) {
	if maxClauses < 1 {
		return nil, searchError("invalid clause limit")
	}
	if _, ok := models[entity]; !ok {
		return nil, searchError("unknown entity")
	}
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil || root == nil {
		return nil, searchError("expected JSON object")
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return nil, searchError("trailing JSON content")
	}
	for key := range root {
		if key != "filter" && key != "orderBy" {
			return nil, searchError("unsupported control")
		}
	}
	filters := map[string]any{}
	orders := []any{}
	if value, exists := root["filter"]; exists {
		var ok bool
		filters, ok = value.(map[string]any)
		if !ok {
			return nil, searchError("invalid filter")
		}
	}
	if value, exists := root["orderBy"]; exists {
		var ok bool
		orders, ok = value.([]any)
		if !ok {
			return nil, searchError("invalid ordering")
		}
	}
	if len(filters)+len(orders) > maxClauses {
		return nil, searchError("clause limit exceeded")
	}
	result := &DynamicSearchResult{Filter: map[string]map[string]any{}}
	missing := func(path, clause string) {
		result.Warnings = append(result.Warnings, DynamicSearchWarning{"DYNAMIC_SEARCH_UNKNOWN_FIELD", entity, clause, path})
	}
	keys := make([]string, 0, len(filters))
	for path := range filters {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	for _, path := range keys {
		predicate, ok := filters[path].(map[string]any)
		if !ok {
			predicate = map[string]any{"$eq": filters[path]}
		}
		if len(predicate) != 1 {
			return nil, searchError("malformed operator")
		}
		for operator, value := range predicate {
			switch operator {
			case "$eq", "$ne", "$gt", "$gte", "$lt", "$lte", "$in", "$notIn", "$contains":
			default:
				return nil, searchError("unsupported operator")
			}
			if operator == "$in" || operator == "$notIn" {
				list, ok := value.([]any)
				if !ok || len(list) > 1000 {
					return nil, searchError("invalid value list")
				}
			}
			kind, err := resolveSearchField(path, entity, models)
			if err != nil {
				return nil, err
			}
			if kind == "" {
				missing(path, "FILTER")
				continue
			}
			if operator == "$contains" && kind != "string" {
				return nil, searchError("string operator on non-string field")
			}
			if list, ok := value.([]any); ok {
				if operator != "$in" && operator != "$notIn" {
					return nil, searchError("unexpected value list")
				}
				for _, item := range list {
					if !validSearchScalar(item, kind) {
						return nil, searchError("invalid known-field value")
					}
				}
			} else if !validSearchScalar(value, kind) {
				return nil, searchError("invalid known-field value")
			}
			result.Filter[path] = predicate
		}
	}
	for _, item := range orders {
		order, ok := item.(map[string]any)
		if !ok || len(order) != 2 {
			return nil, searchError("invalid ordering")
		}
		path, ok := order["field"].(string)
		if !ok {
			return nil, searchError("invalid order field")
		}
		direction, ok := order["direction"].(string)
		if !ok || direction != "asc" && direction != "desc" {
			return nil, searchError("invalid order direction")
		}
		kind, err := resolveSearchField(path, entity, models)
		if err != nil {
			return nil, err
		}
		if kind == "" {
			missing(path, "ORDER_BY")
			continue
		}
		result.OrderBy = append(result.OrderBy, DynamicSearchOrder{path, direction})
	}
	emitSearchWarnings(result.Warnings, sink)
	return result, nil
}

func resolveSearchField(path, entity string, models map[string]SearchModel) (string, error) {
	parts := strings.Split(path, ".")
	if len(parts) > 16 {
		return "", searchError("path limit exceeded")
	}
	for _, part := range parts {
		if part == "" || strings.HasPrefix(part, "$") || part == "__proto__" || part == "prototype" || part == "constructor" {
			return "", searchError("invalid field path")
		}
	}
	model := models[entity]
	for _, part := range parts[:len(parts)-1] {
		target, ok := model.Relations[part]
		if !ok {
			return "", nil
		}
		model, ok = models[target]
		if !ok {
			return "", searchError("invalid trusted relation metadata")
		}
	}
	return model.Fields[parts[len(parts)-1]], nil
}

func validSearchScalar(value any, kind string) bool {
	if value == nil {
		return true
	}
	switch kind {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "integer", "timestamp":
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		v, err := number.Float64()
		return err == nil && math.Trunc(v) == v && v >= -9007199254740991 && v <= 9007199254740991
	case "number":
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		_, err := strconv.ParseFloat(string(number), 64)
		return err == nil
	case "decimal":
		switch v := value.(type) {
		case string:
			return decimalSearchValue.MatchString(v)
		case json.Number:
			_, err := strconv.ParseFloat(string(v), 64)
			return err == nil
		}
	case "date":
		text, ok := value.(string)
		if !ok {
			return false
		}
		parsed, err := time.Parse("2006-01-02", text)
		return err == nil && parsed.Year() >= 1 && parsed.Format("2006-01-02") == text
	}
	return false
}

// MergeDynamicSearch composes trusted native bindings with the existing scoped query.
func MergeDynamicSearch(base *SelectQuery, source []byte, models map[string]SearchModel,
	filter func(string, map[string]any) (*Expr, error), order func(string, string) (*OrderBy, error), sink func(DynamicSearchWarning)) (*SelectQuery, []DynamicSearchWarning, error) {
	if base == nil || filter == nil || order == nil {
		return nil, nil, searchError("missing trusted query binding")
	}
	normalized, err := NormalizeDynamicSearch(source, base.Entity, models, 100, func(DynamicSearchWarning) {})
	if err != nil {
		return nil, nil, err
	}
	query := base.Clone()
	keys := make([]string, 0, len(normalized.Filter))
	for path := range normalized.Filter {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	for _, path := range keys {
		expr, err := filter(path, normalized.Filter[path])
		if err != nil {
			return nil, nil, err
		}
		if expr == nil {
			return nil, nil, searchError("invalid trusted filter binding")
		}
		query.AndFilter(expr)
	}
	for _, item := range normalized.OrderBy {
		expr, err := order(item.Field, item.Direction)
		if err != nil {
			return nil, nil, err
		}
		if expr == nil {
			return nil, nil, searchError("invalid trusted order binding")
		}
		query.OrderBy = append(query.OrderBy, expr)
	}
	emitSearchWarnings(normalized.Warnings, sink)
	return query, normalized.Warnings, nil
}

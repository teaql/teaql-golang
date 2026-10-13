package core

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// GeneratedSearchModel supplies trusted aliases, scalar types and native query keys.
// Search payloads cannot supply or replace this metadata.
type GeneratedSearchModel struct {
	SearchModel
	QueryFields map[string]string
}

// MergeGeneratedScalarSearch reuses the local-search boundary and preserves the
// original query's scope. Related paths require application-owned trusted bindings
// through MergeDynamicSearch; this default never invents relation authorization.
func MergeGeneratedScalarSearch(base *SelectQuery, source []byte, models map[string]GeneratedSearchModel, sink func(DynamicSearchWarning)) (*SelectQuery, []DynamicSearchWarning, error) {
	if base == nil {
		return nil, nil, searchError("missing query")
	}
	searchModels := make(map[string]SearchModel, len(models))
	for entity, model := range models {
		searchModels[entity] = model.SearchModel
	}
	model := models[base.Entity]
	column := func(path string) (string, error) {
		if strings.Contains(path, ".") {
			return "", searchError("related paths require a trusted relation binding")
		}
		key, ok := model.QueryFields[path]
		if !ok || key == "" {
			return "", searchError("missing trusted field binding")
		}
		return key, nil
	}
	filter := func(path string, predicate map[string]any) (*Expr, error) {
		key, err := column(path)
		if err != nil {
			return nil, err
		}
		for operator, raw := range predicate {
			if operator == "$contains" {
				text, ok := raw.(string)
				if !ok {
					return nil, searchError("contains requires a string")
				}
				return ExprContain(key, text), nil
			}
			if operator == "$in" || operator == "$notIn" {
				list := raw.([]any) // NormalizeDynamicSearch has validated the list.
				values := make([]Value, len(list))
				for index, item := range list {
					values[index], err = generatedSearchValue(item, model.Fields[path])
					if err != nil {
						return nil, err
					}
				}
				if operator == "$in" {
					return ExprInList(key, values), nil
				}
				return ExprNotInList(key, values), nil
			}
			value, err := generatedSearchValue(raw, model.Fields[path])
			if err != nil {
				return nil, err
			}
			if raw == nil {
				switch operator {
				case "$eq":
					return ExprIsNullNode(key), nil
				case "$ne":
					return ExprIsNotNullNode(key), nil
				default:
					return nil, searchError("null supports only equality predicates")
				}
			}
			switch operator {
			case "$eq":
				return ExprEq(key, value), nil
			case "$ne":
				return ExprNe(key, value), nil
			case "$gt":
				return ExprGt(key, value), nil
			case "$gte":
				return ExprGte(key, value), nil
			case "$lt":
				return ExprLt(key, value), nil
			case "$lte":
				return ExprLte(key, value), nil
			}
		}
		return nil, searchError("unsupported operator")
	}
	order := func(path, direction string) (*OrderBy, error) {
		key, err := column(path)
		if err != nil {
			return nil, err
		}
		if direction == "asc" {
			return OrderAsc(key), nil
		}
		return OrderDesc(key), nil
	}
	return MergeDynamicSearch(base, source, searchModels, filter, order, sink)
}

// All callers pass values already validated by NormalizeDynamicSearch.
func generatedSearchValue(raw any, kind string) (Value, error) {
	if raw == nil {
		return ValNull(), nil
	}
	switch kind {
	case "string":
		return ValText(raw.(string)), nil
	case "boolean":
		return ValBool(raw.(bool)), nil
	case "integer", "timestamp":
		value, err := raw.(json.Number).Float64()
		if err != nil {
			return Value{}, searchError("invalid integer")
		}
		if kind == "timestamp" {
			return ValTimestamp(int64(value)), nil
		}
		return ValI64(int64(value)), nil
	case "number":
		value, err := raw.(json.Number).Float64()
		if err != nil {
			return Value{}, searchError("invalid number")
		}
		return ValF64(value), nil
	case "decimal":
		var text string
		switch value := raw.(type) {
		case string:
			text = value
		case json.Number:
			text = value.String()
		}
		value, err := decimal.NewFromString(text)
		if err != nil {
			return Value{}, searchError("invalid decimal")
		}
		return ValDecimal(value), nil
	case "date":
		value, err := time.Parse("2006-01-02", raw.(string))
		if err != nil {
			return Value{}, searchError("invalid date")
		}
		return ValDate(value), nil
	}
	return Value{}, searchError("unsupported known-field type")
}

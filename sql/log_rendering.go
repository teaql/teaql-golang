package sql

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"strings"
)

// RenderSQLLog reuses the dialect's diagnostic scanner after safety projection.
// Its result is diagnostic only; masked literals are not executable replacements.
func RenderSQLLog(query string, params []core.Value, masked []bool, kind DatabaseKind) (string, error) {
	if strings.TrimSpace(query) == "" || len(params) != len(masked) {
		return "", fmt.Errorf("missing SQL or mismatched mask policies")
	}
	used := make(map[int]bool)
	var failure error
	literal := func(index int) string {
		if index < 0 || index >= len(params) {
			failure = fmt.Errorf("SQL binding or token mismatch")
			return ""
		}
		used[index] = true
		value := sqlLiteral(params[index], kind)
		if masked[index] {
			value += " /* masked */"
		}
		return value
	}
	var result string
	if kind == DatabaseKindPostgreSQL {
		result = replacePostgresPlaceholders(query, params, literal)
	} else {
		result = replacePositionalPlaceholders(query, params, kind, literal)
	}
	if failure != nil {
		return "", failure
	}
	if len(used) != len(params) {
		return "", fmt.Errorf("unused SQL bindings")
	}
	return result, nil
}

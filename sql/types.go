package sql

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"strings"
	"time"
)

type DatabaseKind int

const (
	DatabaseKindPostgreSQL DatabaseKind = iota
	DatabaseKindSQLite
	DatabaseKindMySQL
)

type CompiledQuery struct {
	Sql     string
	Params  []core.Value
	Comment *string
}

func (q *CompiledQuery) SqlWithComment() string {
	if q.Comment != nil && *q.Comment != "" {
		escaped := strings.ReplaceAll(*q.Comment, "*/", "* /")
		return fmt.Sprintf("/* %s */ %s", escaped, q.Sql)
	}
	return q.Sql
}

func (q *CompiledQuery) DebugSql(kind DatabaseKind) string {
	sql := q.SqlWithComment()
	switch kind {
	case DatabaseKindPostgreSQL:
		return replacePostgresPlaceholders(sql, q.Params)
	case DatabaseKindSQLite:
		return replacePositionalPlaceholders(sql, q.Params, kind)
	case DatabaseKindMySQL:
		return replacePositionalPlaceholders(sql, q.Params, kind)
	default:
		return sql
	}
}

func replacePostgresPlaceholders(sql string, params []core.Value) string {
	var output strings.Builder
	output.Grow(len(sql))

	inString := false
	chars := []rune(sql)

	for i := 0; i < len(chars); i++ {
		ch := chars[i]

		if ch == '\'' {
			output.WriteRune('\'')
			if inString && i+1 < len(chars) && chars[i+1] == '\'' {
				output.WriteRune('\'')
				i++
			} else {
				inString = !inString
			}
			continue
		}

		if !inString && ch == '$' && i+1 < len(chars) && chars[i+1] >= '0' && chars[i+1] <= '9' {
			idxStr := ""
			j := i + 1
			for j < len(chars) && chars[j] >= '0' && chars[j] <= '9' {
				idxStr += string(chars[j])
				j++
			}
			var index int
			fmt.Sscanf(idxStr, "%d", &index)
			if index > 0 && index-1 < len(params) {
				output.WriteString(sqlLiteral(params[index-1], DatabaseKindPostgreSQL))
				i = j - 1
				continue
			}
			output.WriteRune('$')
			output.WriteString(idxStr)
			i = j - 1
			continue
		}
		output.WriteRune(ch)
	}
	return output.String()
}

func replacePositionalPlaceholders(sql string, params []core.Value, kind DatabaseKind) string {
	var output strings.Builder
	output.Grow(len(sql))

	state := scanSQL
	chars := []rune(sql)
	paramIdx := 0

	for i := 0; i < len(chars); i++ {
		ch := chars[i]

		if state == scanSQL && ch == '\'' {
			output.WriteRune(ch)
			state = scanSingleQuote
			continue
		}
		if state == scanSQL && ch == '"' {
			output.WriteRune(ch)
			state = scanDoubleQuote
			continue
		}
		if state == scanSQL && ch == '-' && i+1 < len(chars) && chars[i+1] == '-' {
			output.WriteString("--")
			i++
			state = scanLineComment
			continue
		}
		if state == scanSQL && ch == '/' && i+1 < len(chars) && chars[i+1] == '*' {
			output.WriteString("/*")
			i++
			state = scanBlockComment
			continue
		}
		if state == scanSingleQuote {
			output.WriteRune(ch)
			if ch == '\'' && i+1 < len(chars) && chars[i+1] == '\'' {
				output.WriteRune('\'')
				i++
			} else if ch == '\'' {
				state = scanSQL
			}
			continue
		}
		if state == scanDoubleQuote {
			output.WriteRune(ch)
			if ch == '"' && i+1 < len(chars) && chars[i+1] == '"' {
				output.WriteRune('"')
				i++
			} else if ch == '"' {
				state = scanSQL
			}
			continue
		}
		if state == scanLineComment {
			output.WriteRune(ch)
			if ch == '\n' || ch == '\r' {
				state = scanSQL
			}
			continue
		}
		if state == scanBlockComment {
			output.WriteRune(ch)
			if ch == '*' && i+1 < len(chars) && chars[i+1] == '/' {
				output.WriteRune('/')
				i++
				state = scanSQL
			}
			continue
		}
		if ch == '?' {
			if paramIdx < len(params) {
				output.WriteString(sqlLiteral(params[paramIdx], kind))
				paramIdx++
			} else {
				output.WriteRune(ch)
			}
			continue
		}
		output.WriteRune(ch)
	}
	return output.String()
}

type sqlScanState uint8

const (
	scanSQL sqlScanState = iota
	scanSingleQuote
	scanDoubleQuote
	scanLineComment
	scanBlockComment
)

func sqlLiteral(value core.Value, kind DatabaseKind) string {
	if value.V == nil {
		return "NULL"
	}
	switch v := value.V.(type) {
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int64:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case decimal.Decimal:
		return v.String()
	case string:
		return quotedSqlString(v)
	case time.Time:
		return quotedSqlString(v.UTC().Format("2006-01-02"))
	case core.Timestamp:
		return fmt.Sprintf("%d", v)
	case []core.Value:
		var strs []string
		for _, item := range v {
			strs = append(strs, sqlLiteral(item, kind))
		}
		joined := strings.Join(strs, ", ")
		if kind == DatabaseKindPostgreSQL {
			return fmt.Sprintf("ARRAY[%s]", joined)
		}
		return fmt.Sprintf("(%s)", joined)
	case core.Record:
		b := core.ValueToJson(value)
		return quotedSqlString(fmt.Sprintf("%v", b)) // Just string representation for now, or json encoding
	default:
		// Attempt Date/Timestamp/Json mapping based on how core implements them
		// This requires some introspection, but since value.V is 'any', we can try a json marshal fallback
		return quotedSqlString(fmt.Sprintf("%v", v))
	}
}

func quotedSqlString(value string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
}

type SqlCompileError struct {
	Message string
}

func (e *SqlCompileError) Error() string {
	return e.Message
}

func ErrUnknownEntity(entity string) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("unknown entity: %s", entity)}
}

func ErrUnknownField(field string) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("unknown field: %s", field)}
}

func ErrEmptyInList() *SqlCompileError {
	return &SqlCompileError{Message: "IN requires at least one value"}
}

func ErrMissingIdProperty(entity string) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("entity %s has no id property", entity)}
}

func ErrMissingVersionProperty(entity string) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("entity %s has no version property", entity)}
}

func ErrEmptyMutation(kind string) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("%s requires at least one writable field", kind)}
}

func ErrInvalidRecoverVersion(version int64) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("recover requires a negative version, got %d", version)}
}

func ErrUnsupportedSchemaType(dataType core.DataType) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("unsupported schema type: %v", dataType)}
}

func ErrInvalidFunctionArguments(message string) *SqlCompileError {
	return &SqlCompileError{Message: message}
}

func ErrInvalidSubQueryOperator(operator string) *SqlCompileError {
	return &SqlCompileError{Message: fmt.Sprintf("subquery does not support operator: %s", operator)}
}

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"

	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

const DefaultIdSpaceTable = "teaql_id_space"
const SqliteDecimalPrefix = "__teaql_decimal__:"

type SqliteDialect struct{}

func (d *SqliteDialect) Kind() teaql_sql.DatabaseKind {
	return teaql_sql.DatabaseKindSQLite
}

func (d *SqliteDialect) QuoteIdent(ident string) string {
	return quoteIdent(ident)
}

func (d *SqliteDialect) Placeholder(index int) string {
	return "?"
}

func (d *SqliteDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	return "", fmt.Errorf("GBK function is not supported in sqlite")
}

func (d *SqliteDialect) SchemaSetupSqls() []string {
	return nil
}

func (d *SqliteDialect) SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error) {
	switch dataType {
	case core.TypeBool:
		return "INTEGER", nil
	case core.TypeI64, core.TypeU64:
		return "INTEGER", nil
	case core.TypeF64:
		return "REAL", nil
	case core.TypeDecimal:
		return "NUMERIC", nil
	case core.TypeText:
		return "VARCHAR(255)", nil
	case core.TypeLargeText:
		return "TEXT", nil
	case core.TypeJson:
		return "JSON", nil
	case core.TypeDate:
		return "DATE", nil
	case core.TypeTimestamp:
		return "TIMESTAMP", nil
	}
	return "", fmt.Errorf("unsupported schema type")
}

func (d *SqliteDialect) CompileAddColumn(entity *core.EntityDescriptor, property *core.PropertyDescriptor) (string, error) {
	defDialect := &teaql_sql.DefaultSqlDialect{Dialect: d}
	def, err := defDialect.ColumnDefinitionSql(property)
	if err != nil {
		return "", err
	}
	defWithoutNotNull := strings.Replace(def, " NOT NULL", "", -1)

	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", d.QuoteIdent(entity.TabName), defWithoutNotNull), nil
}

type SqliteMutationExecutor struct {
	db *sql.DB
}

func NewSqliteMutationExecutor(db *sql.DB) *SqliteMutationExecutor {
	return &SqliteMutationExecutor{db: db}
}

func (e *SqliteMutationExecutor) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
	log.Printf("SQL: %s | Params: %v\n", query.SqlWithComment(), params)
	rows, err := e.db.QueryContext(ctx, query.SqlWithComment(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var records []core.Record
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		record, err := decodeSqliteRow(cols, colTypes, vals)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *SqliteMutationExecutor) StreamSql(ctx context.Context, query *teaql_sql.CompiledQuery, chunkSize int, yield func([]core.Record) error) error {
	params, err := bindValues(query.Params)
	if err != nil {
		return err
	}
	rows, err := e.db.QueryContext(ctx, query.SqlWithComment(), params...)
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	chunk := make([]core.Record, 0, chunkSize)
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		record, err := decodeSqliteRow(cols, colTypes, vals)
		if err != nil {
			return err
		}
		chunk = append(chunk, record)
		if len(chunk) == chunkSize {
			if err := yield(chunk); err != nil {
				return err
			}
			chunk = make([]core.Record, 0, chunkSize)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(chunk) > 0 {
		return yield(chunk)
	}
	return nil
}

func (e *SqliteMutationExecutor) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
	log.Printf("SQL: %s | Params: %v\n", query.SqlWithComment(), params)
	res, err := e.db.ExecContext(ctx, query.SqlWithComment(), params...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(affected), nil
}

func (e *SqliteMutationExecutor) BeginSql(ctx context.Context) (teaql_sql.SqlTransactionTransportTx, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &SqliteTransactionExecutor{tx: tx}, nil
}

type SqliteTransactionExecutor struct {
	tx *sql.Tx
}

func (e *SqliteTransactionExecutor) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
	log.Printf("SQL (Tx): %s | Params: %v\n", query.SqlWithComment(), params)
	rows, err := e.tx.QueryContext(ctx, query.SqlWithComment(), params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var records []core.Record
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		record, err := decodeSqliteRow(cols, colTypes, vals)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *SqliteTransactionExecutor) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
	log.Printf("SQL (Tx): %s | Params: %v\n", query.SqlWithComment(), params)
	res, err := e.tx.ExecContext(ctx, query.SqlWithComment(), params...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(affected), nil
}

func (e *SqliteTransactionExecutor) CommitSql(ctx context.Context) error {
	return e.tx.Commit()
}

func (e *SqliteTransactionExecutor) RollbackSql(ctx context.Context) error {
	return e.tx.Rollback()
}

func quoteIdent(ident string) string {
	return teaql_sql.QuoteIdentifierIfNeeded(ident, '"')
}

func bindValues(values []core.Value) ([]interface{}, error) {
	var result []interface{}
	for _, v := range values {
		bound, err := bindSqliteValue(v)
		if err != nil {
			return nil, err
		}
		result = append(result, bound)
	}
	return result, nil
}

func bindSqliteValue(value core.Value) (interface{}, error) {
	if value.V == nil {
		return nil, nil
	}
	switch v := value.V.(type) {
	case core.DataType:
		return nil, nil
	case bool:
		if v {
			return int64(1), nil
		}
		return int64(0), nil
	case int64:
		return v, nil
	case uint64:
		return int64(v), nil
	case float64:
		return v, nil
	case decimal.Decimal:
		// SQLite NUMERIC affinity coerces a canonical numeric string while the
		// driver cannot bind Decimal structs directly.
		return v.String(), nil
	case time.Time:
		// TeaQL time.Time values are dates (timestamps use epoch milliseconds).
		// Match SQLite DATE text storage so inclusive bounds include the day.
		return v.Format("2006-01-02"), nil
	case string:
		return v, nil
	}
	return nil, fmt.Errorf("unsupported sqlite bind value type %T", value.V)
}

func decodeSqliteRow(cols []string, colTypes []*sql.ColumnType, vals []interface{}) (core.Record, error) {
	record := make(core.Record)
	for i, colName := range cols {
		val := vals[i]
		if val == nil {
			record[colName] = core.ValNull()
			continue
		}
		switch v := val.(type) {
		case int64:
			dbType := ""
			if i < len(colTypes) {
				dbType = colTypes[i].DatabaseTypeName()
			}
			if dbType == "BOOLEAN" || dbType == "BOOL" {
				record[colName] = core.ValBool(v != 0)
			} else {
				record[colName] = core.ValI64(v)
			}
		case float64:
			record[colName] = core.ValF64(v)
		case []byte:
			record[colName] = core.ValText(string(v))
		case string:
			record[colName] = core.ValText(v)
		case time.Time:
			// go-sqlite3 decodes DATE/TIMESTAMP columns as time.Time. Preserve the
			// value so generated date and timestamp accessors can convert it.
			record[colName] = core.Value{V: v}
		}
	}
	return record, nil
}

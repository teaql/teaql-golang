package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"

	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

const DefaultIdSpaceTable = "teaql_id_space"

type PostgresDialect struct{}

func (d *PostgresDialect) Kind() teaql_sql.DatabaseKind {
	return teaql_sql.DatabaseKindPostgreSQL
}

func (d *PostgresDialect) QuoteIdent(ident string) string {
	return teaql_sql.QuoteIdentifierIfNeeded(ident, '"')
}

func (d *PostgresDialect) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (d *PostgresDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	return "", fmt.Errorf("GBK function is not supported in postgres")
}

func (d *PostgresDialect) SchemaSetupSqls() []string {
	return []string{
		`CREATE OR REPLACE FUNCTION soundex(input text)
RETURNS text
LANGUAGE plpgsql
IMMUTABLE
STRICT
AS $$
DECLARE
    normalized text := upper(regexp_replace(input, '[^A-Za-z]', '', 'g'));
    first_char text;
    output text;
    previous_code text;
    code text;
    ch text;
    i integer;
BEGIN
    IF normalized = '' THEN
        RETURN '0000';
    END IF;

    first_char := substr(normalized, 1, 1);
    output := first_char;
    previous_code := CASE
        WHEN first_char IN ('B', 'F', 'P', 'V') THEN '1'
        WHEN first_char IN ('C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z') THEN '2'
        WHEN first_char IN ('D', 'T') THEN '3'
        WHEN first_char = 'L' THEN '4'
        WHEN first_char IN ('M', 'N') THEN '5'
        WHEN first_char = 'R' THEN '6'
        ELSE '0'
    END;

    FOR i IN 2..char_length(normalized) LOOP
        ch := substr(normalized, i, 1);
        code := CASE
            WHEN ch IN ('B', 'F', 'P', 'V') THEN '1'
            WHEN ch IN ('C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z') THEN '2'
            WHEN ch IN ('D', 'T') THEN '3'
            WHEN ch = 'L' THEN '4'
            WHEN ch IN ('M', 'N') THEN '5'
            WHEN ch = 'R' THEN '6'
            ELSE '0'
        END;

        IF code <> '0' AND code <> previous_code THEN
            output := output || code;
            IF char_length(output) = 4 THEN
                RETURN output;
            END IF;
        END IF;
        previous_code := code;
    END LOOP;

    RETURN rpad(output, 4, '0');
END;
$$`,
	}
}

func (d *PostgresDialect) SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error) {
	switch dataType {
	case core.TypeBool:
		return "BOOLEAN", nil
	case core.TypeI64, core.TypeU64:
		return "BIGINT", nil
	case core.TypeF64:
		return "DOUBLE PRECISION", nil
	case core.TypeDecimal:
		return "NUMERIC", nil
	case core.TypeText:
		return "VARCHAR(255)", nil
	case core.TypeLargeText:
		return "TEXT", nil
	case core.TypeJson:
		return "JSONB", nil
	case core.TypeDate:
		return "DATE", nil
	case core.TypeTimestamp:
		return "TIMESTAMPTZ", nil
	}
	return "", fmt.Errorf("unsupported schema type")
}

func (d *PostgresDialect) CompileAddColumn(entity *core.EntityDescriptor, property *core.PropertyDescriptor) (string, error) {
	defDialect := &teaql_sql.DefaultSqlDialect{Dialect: d}
	def, err := defDialect.ColumnDefinitionSql(property)
	if err != nil {
		return "", err
	}
	defWithoutNotNull := strings.Replace(def, " NOT NULL", "", -1)

	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", d.QuoteIdent(entity.TabName), defWithoutNotNull), nil
}

type PgMutationExecutor struct {
	db *sql.DB
}

func NewPgMutationExecutor(db *sql.DB) *PgMutationExecutor {
	return &PgMutationExecutor{db: db}
}

func (e *PgMutationExecutor) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
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
		record, err := decodePgRow(cols, colTypes, vals)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *PgMutationExecutor) StreamSql(ctx context.Context, query *teaql_sql.CompiledQuery, chunkSize int, yield func([]core.Record) error) error {
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
		record, err := decodePgRow(cols, colTypes, vals)
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

func (e *PgMutationExecutor) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
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

func (e *PgMutationExecutor) BeginSql(ctx context.Context) (teaql_sql.SqlTransactionTransportTx, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &PgTransactionExecutor{tx: tx}, nil
}

type PgTransactionExecutor struct {
	tx *sql.Tx
}

func (e *PgTransactionExecutor) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
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
		record, err := decodePgRow(cols, colTypes, vals)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *PgTransactionExecutor) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
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

func (e *PgTransactionExecutor) CommitSql(ctx context.Context) error {
	return e.tx.Commit()
}

func (e *PgTransactionExecutor) RollbackSql(ctx context.Context) error {
	return e.tx.Rollback()
}

func bindValues(values []core.Value) ([]interface{}, error) {
	var result []interface{}
	for _, v := range values {
		bound, err := bindPgValue(v)
		if err != nil {
			return nil, err
		}
		result = append(result, bound)
	}
	return result, nil
}

func bindPgValue(value core.Value) (interface{}, error) {
	if value.V == nil {
		return nil, nil
	}
	switch v := value.V.(type) {
	case core.DataType:
		// ValTypedNull keeps the intended database type in Value.V. Drivers
		// still need an actual nil parameter; SQL metadata carries the type.
		return nil, nil
	case bool:
		return v, nil
	case int64:
		return v, nil
	case uint64:
		return int64(v), nil
	case float64:
		return v, nil
	case decimal.Decimal:
		// lib/pq does not accept arbitrary structs as bind parameters. Decimal's
		// canonical string preserves precision and PostgreSQL NUMERIC coerces it.
		return v.String(), nil
	case time.Time:
		return v, nil
	case string:
		return v, nil
	}
	return nil, fmt.Errorf("unsupported postgres bind value type %T", value.V)
}

func decodePgRow(cols []string, colTypes []*sql.ColumnType, vals []interface{}) (core.Record, error) {
	record := make(core.Record)
	for i, colName := range cols {
		val := vals[i]
		if val == nil {
			record[colName] = core.ValNull()
			continue
		}
		switch v := val.(type) {
		case bool:
			record[colName] = core.ValBool(v)
		case int64:
			record[colName] = core.ValI64(v)
		case float64:
			record[colName] = core.ValF64(v)
		case []byte:
			record[colName] = core.ValText(string(v))
		case string:
			record[colName] = core.ValText(v)
		default:
			// Just a fallback
			record[colName] = core.ValText(fmt.Sprintf("%v", v))
		}
	}
	return record, nil
}

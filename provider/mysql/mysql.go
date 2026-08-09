package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

const DefaultIdSpaceTable = "teaql_id_space"

type MysqlDialect struct{}

func (d *MysqlDialect) Kind() teaql_sql.DatabaseKind {
	return teaql_sql.DatabaseKindMySQL
}

func (d *MysqlDialect) QuoteIdent(ident string) string {
	return teaql_sql.QuoteIdentifierIfNeeded(ident, '`')
}

func (d *MysqlDialect) Placeholder(index int) string {
	return "?"
}

func (d *MysqlDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	return "", fmt.Errorf("GBK function is not supported natively, maybe need to check")
}

func (d *MysqlDialect) SchemaSetupSqls() []string {
	return nil
}

func (d *MysqlDialect) SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error) {
	switch dataType {
	case core.TypeBool:
		return "BOOLEAN", nil
	case core.TypeI64, core.TypeU64:
		return "BIGINT", nil
	case core.TypeF64:
		return "DOUBLE", nil
	case core.TypeDecimal:
		return "DECIMAL(38, 10)", nil
	case core.TypeText:
		return "VARCHAR(255)", nil
	case core.TypeLargeText:
		return "LONGTEXT", nil
	case core.TypeJson:
		return "JSON", nil
	case core.TypeDate:
		return "DATE", nil
	case core.TypeTimestamp:
		return "DATETIME(6)", nil
	}
	return "", fmt.Errorf("unsupported schema type")
}

func (d *MysqlDialect) CompileAddColumn(entity *core.EntityDescriptor, property *core.PropertyDescriptor) (string, error) {
	defDialect := &teaql_sql.DefaultSqlDialect{Dialect: d}
	def, err := defDialect.ColumnDefinitionSql(property)
	if err != nil {
		return "", err
	}
	defWithoutNotNull := strings.Replace(def, " NOT NULL", "", -1)

	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", d.QuoteIdent(entity.TabName), defWithoutNotNull), nil
}

type MysqlMutationExecutor struct {
	db *sql.DB
}

func NewMysqlMutationExecutor(db *sql.DB) *MysqlMutationExecutor {
	return &MysqlMutationExecutor{db: db}
}

func (e *MysqlMutationExecutor) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
	rows, err := e.db.QueryContext(ctx, query.SqlComment(), params...)
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
		record, err := decodeMysqlRow(cols, colTypes, vals)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *MysqlMutationExecutor) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
	res, err := e.db.ExecContext(ctx, query.SqlComment(), params...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(affected), nil
}

func (e *MysqlMutationExecutor) BeginSql(ctx context.Context) (teaql_sql.SqlTransactionTransportTx, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &MysqlTransactionExecutor{tx: tx}, nil
}

type MysqlTransactionExecutor struct {
	tx *sql.Tx
}

func (e *MysqlTransactionExecutor) FetchAllSql(ctx context.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
	rows, err := e.tx.QueryContext(ctx, query.SqlComment(), params...)
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
		record, err := decodeMysqlRow(cols, colTypes, vals)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *MysqlTransactionExecutor) ExecuteSql(ctx context.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
	res, err := e.tx.ExecContext(ctx, query.SqlComment(), params...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(affected), nil
}

func (e *MysqlTransactionExecutor) CommitSql(ctx context.Context) error {
	return e.tx.Commit()
}

func (e *MysqlTransactionExecutor) RollbackSql(ctx context.Context) error {
	return e.tx.Rollback()
}

func bindValues(values []core.Value) ([]interface{}, error) {
	var result []interface{}
	for _, v := range values {
		bound, err := bindMysqlValue(v)
		if err != nil {
			return nil, err
		}
		result = append(result, bound)
	}
	return result, nil
}

func bindMysqlValue(value core.Value) (interface{}, error) {
	if value.V == nil {
		return nil, nil
	}
	switch v := value.V.(type) {
	case bool:
		return v, nil
	case int64:
		return v, nil
	case uint64:
		return int64(v), nil
	case float64:
		return v, nil
	case string:
		return v, nil
	}
	return nil, fmt.Errorf("unsupported mysql bind value type")
}

func decodeMysqlRow(cols []string, colTypes []*sql.ColumnType, vals []interface{}) (core.Record, error) {
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

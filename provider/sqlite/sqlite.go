package sqlite

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"

	"github.com/teaql/teaql-golang/core"
	teaql_sql "github.com/teaql/teaql-golang/sql"
)

const DefaultIdSpaceTable = "teaql_id_space"

// EnsureSoundex installs the SQLite functions and collations required by the
// generated schema, including soundex and exact Decimal ordering. The name is
// retained for generated workspaces that already call it from EnsureSchema.
func EnsureSoundex(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("sqlite database is nil")
	}
	// A user-defined SQLite function belongs to a physical connection. Keeping
	// one connection makes the registration deterministic for this embedded DB.
	db.SetMaxOpenConns(1)
	connection, err := db.Conn(stdcontext.Background())
	if err != nil {
		return err
	}
	defer connection.Close()
	return connection.Raw(func(driverConnection any) error {
		sqliteConnection, ok := driverConnection.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("unexpected SQLite driver connection %T", driverConnection)
		}
		if err := sqliteConnection.RegisterFunc("soundex", sqliteCompatibleSoundex, true); err != nil {
			return err
		}
		if err := sqliteConnection.RegisterCollation("TEAQL_DECIMAL", compareExactDecimal); err != nil {
			return err
		}
		if err := sqliteConnection.RegisterFunc("teaql_decimal_valid", func(text string) int {
			if _, err := decimal.NewFromString(text); err != nil {
				return 0
			}
			return 1
		}, true); err != nil {
			return err
		}
		return sqliteConnection.RegisterAggregator("TEAQL_DECIMAL_SUM", func() *exactDecimalSum { return &exactDecimalSum{} }, true)
	})
}

type exactDecimalSum struct {
	total decimal.Decimal
	seen  bool
}

func (aggregate *exactDecimalSum) Step(value any) error {
	if value == nil {
		return nil
	}
	var text string
	switch typed := value.(type) {
	case string:
		text = typed
	case []byte:
		text = string(typed)
	default:
		return fmt.Errorf("exact Decimal SUM expected text, got %T", value)
	}
	parsed, err := decimal.NewFromString(text)
	if err != nil {
		return fmt.Errorf("invalid Decimal in SUM: %w", err)
	}
	aggregate.total = aggregate.total.Add(parsed)
	aggregate.seen = true
	return nil
}

func (aggregate *exactDecimalSum) Done() (any, error) {
	if !aggregate.seen {
		return nil, nil
	}
	return exactDecimalText(aggregate.total), nil
}

func compareExactDecimal(left, right string) int {
	leftValue, leftErr := decimal.NewFromString(left)
	rightValue, rightErr := decimal.NewFromString(right)
	if leftErr != nil || rightErr != nil {
		// Malformed externally written values fail when decoded as Decimal. The
		// SQLite callback cannot return an error, so keep its ordering stable.
		return strings.Compare(left, right)
	}
	return leftValue.Cmp(rightValue)
}

func exactDecimalText(value decimal.Decimal) string {
	if value.Exponent() < 0 {
		return value.StringFixed(-value.Exponent())
	}
	return value.String()
}

func sqliteCompatibleSoundex(value any) string {
	if value == nil {
		return "?000"
	}
	var input string
	switch typed := value.(type) {
	case string:
		input = typed
	case []byte:
		input = string(typed)
	default:
		input = fmt.Sprint(typed)
	}
	letters := make([]byte, 0, len(input))
	for _, value := range []byte(strings.ToUpper(input)) {
		if value >= 'A' && value <= 'Z' {
			letters = append(letters, value)
		}
	}
	if len(letters) == 0 {
		return "?000"
	}
	code := func(value byte) byte {
		switch value {
		case 'B', 'F', 'P', 'V':
			return '1'
		case 'C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z':
			return '2'
		case 'D', 'T':
			return '3'
		case 'L':
			return '4'
		case 'M', 'N':
			return '5'
		case 'R':
			return '6'
		default:
			return '0'
		}
	}
	result := []byte{letters[0]}
	previous := code(letters[0])
	for _, letter := range letters[1:] {
		current := code(letter)
		if current != '0' && current != previous {
			result = append(result, current)
		}
		if len(result) == 4 {
			break
		}
		previous = current
	}
	for len(result) < 4 {
		result = append(result, '0')
	}
	return string(result)
}

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

func (d *SqliteDialect) CompileExactDecimalAggregate(function core.AggregateFunction, field string) (string, error) {
	switch function {
	case core.AggSum:
		return "TEAQL_DECIMAL_SUM(" + field + ")", nil
	case core.AggMin:
		return "MIN(" + field + ")", nil
	case core.AggMax:
		return "MAX(" + field + ")", nil
	default:
		return "", fmt.Errorf("SQLite exact Decimal aggregate %v requires an explicit precision policy or is unsupported", function)
	}
}

func (d *SqliteDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	return "", fmt.Errorf("GBK function is not supported in sqlite")
}

func (d *SqliteDialect) SchemaSetupSqls() []string {
	return []string{teaql_sql.BusinessIDSchemaSQL}
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
		// TEXT affinity preserves all base-10 digits; the collation keeps range
		// predicates and ORDER BY numeric instead of lexical. Generated columns
		// also reject malformed or non-TEXT values written outside TeaQL.
		definition := "TEAQL_DECIMAL_TEXT COLLATE TEAQL_DECIMAL"
		if property != nil {
			column := d.QuoteIdent(property.ColName)
			definition += " CHECK (" + column + " IS NULL OR (typeof(" + column + ") = 'text' AND teaql_decimal_valid(" + column + ")))"
		}
		return definition, nil
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

// VerifyExactDecimalSchema rejects an existing NUMERIC table before generated
// bootstrap or application writes can silently round modeled Decimal values.
// Call this after CREATE TABLE IF NOT EXISTS in the generated EnsureSchema path.
func VerifyExactDecimalSchema(db *sql.DB, entity *core.EntityDescriptor) error {
	if db == nil || entity == nil {
		return fmt.Errorf("SQLite Decimal schema verification requires a database and entity")
	}
	decimals := make([]*core.PropertyDescriptor, 0)
	for _, property := range entity.Properties {
		if property.DataType == core.TypeDecimal {
			decimals = append(decimals, property)
		}
	}
	if len(decimals) == 0 {
		return nil
	}
	rows, err := db.Query("PRAGMA table_info(" + quoteIdent(entity.TabName) + ")")
	if err != nil {
		return err
	}
	defer rows.Close()
	types := make(map[string]string)
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, declaredType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &declaredType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		types[strings.ToLower(name)] = declaredType
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	var createSQL string
	if err := db.QueryRow("SELECT sql FROM sqlite_schema WHERE type='table' AND name=?", entity.TabName).Scan(&createSQL); err != nil {
		return fmt.Errorf("read SQLite schema for %s: %w", entity.TabName, err)
	}
	for _, property := range decimals {
		actual, exists := types[strings.ToLower(property.ColName)]
		if !exists || !strings.EqualFold(actual, "TEAQL_DECIMAL_TEXT") {
			if !exists {
				actual = "<missing>"
			}
			return fmt.Errorf("SQLite Decimal column %s.%s has type %s; migrate to TEAQL_DECIMAL_TEXT before continuing (previously rounded NUMERIC/REAL values cannot be recovered automatically)",
				entity.TabName, property.ColName, actual)
		}
		definition := `(?i)(?:^|[\s,(])` + regexp.QuoteMeta(quoteIdent(property.ColName)) + `\s+TEAQL_DECIMAL_TEXT\s+COLLATE\s+TEAQL_DECIMAL\b`
		if !regexp.MustCompile(definition).MatchString(createSQL) {
			return fmt.Errorf("SQLite Decimal column %s.%s must declare COLLATE TEAQL_DECIMAL; rebuild the table with the generated Decimal definition", entity.TabName, property.ColName)
		}
		column := regexp.QuoteMeta(quoteIdent(property.ColName))
		valueCheck := `(?i)\bCHECK\s*\(\s*` + column + `\s+IS\s+NULL\s+OR\s*\(\s*typeof\(\s*` + column + `\s*\)\s*=\s*'text'\s+AND\s+teaql_decimal_valid\(\s*` + column + `\s*\)\s*\)\s*\)`
		if !regexp.MustCompile(valueCheck).MatchString(createSQL) {
			return fmt.Errorf("SQLite Decimal column %s.%s lacks the generated exact-value check; rebuild the table with the Decimal value check", entity.TabName, property.ColName)
		}
	}
	return nil
}

type SqliteMutationExecutor struct {
	db *sql.DB
}

func NewSqliteMutationExecutor(db *sql.DB) *SqliteMutationExecutor {
	return &SqliteMutationExecutor{db: db}
}

func (e *SqliteMutationExecutor) NextId(context stdcontext.Context, entity string) (uint64, error) {
	return teaql_sql.NextOptimisticId(context, e, &SqliteDialect{}, entity)
}

func (e *SqliteMutationExecutor) EnsureIdFloor(context stdcontext.Context, entity string, floor uint64) error {
	return teaql_sql.EnsureOptimisticIdFloor(context, e, &SqliteDialect{}, entity, floor)
}

func (e *SqliteMutationExecutor) GenerateId(entity string) (uint64, error) {
	return e.NextId(stdcontext.Background(), entity)
}

func (e *SqliteMutationExecutor) FetchAllSql(context stdcontext.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
	rows, err := e.db.QueryContext(context, query.SqlWithComment(), params...)
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
		record, err := decodeSqliteRowWithDecimalAliases(cols, colTypes, vals, query.ExactDecimalResultAliases)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *SqliteMutationExecutor) StreamSql(context stdcontext.Context, query *teaql_sql.CompiledQuery, chunkSize int, yield func([]core.Record) error) (streamErr error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return err
	}
	rows, err := e.db.QueryContext(context, query.SqlWithComment(), params...)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && streamErr == nil {
			streamErr = closeErr
		}
	}()
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
		record, err := decodeSqliteRowWithDecimalAliases(cols, colTypes, vals, query.ExactDecimalResultAliases)
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

func (e *SqliteMutationExecutor) ExecuteSql(context stdcontext.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
	res, err := e.db.ExecContext(context, query.SqlWithComment(), params...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(affected), nil
}

func (e *SqliteMutationExecutor) BeginSql(context stdcontext.Context) (teaql_sql.SqlTransactionTransportTx, error) {
	tx, err := e.db.BeginTx(context, nil)
	if err != nil {
		return nil, err
	}
	return &SqliteTransactionExecutor{tx: tx}, nil
}

type SqliteTransactionExecutor struct {
	tx *sql.Tx
}

func (e *SqliteTransactionExecutor) FetchAllSql(context stdcontext.Context, query *teaql_sql.CompiledQuery) ([]core.Record, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return nil, err
	}
	rows, err := e.tx.QueryContext(context, query.SqlWithComment(), params...)
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
		record, err := decodeSqliteRowWithDecimalAliases(cols, colTypes, vals, query.ExactDecimalResultAliases)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (e *SqliteTransactionExecutor) ExecuteSql(context stdcontext.Context, query *teaql_sql.CompiledQuery) (uint64, error) {
	params, err := bindValues(query.Params)
	if err != nil {
		return 0, err
	}
	res, err := e.tx.ExecContext(context, query.SqlWithComment(), params...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return uint64(affected), nil
}

func (e *SqliteTransactionExecutor) CommitSql(context stdcontext.Context) error {
	return e.tx.Commit()
}

func (e *SqliteTransactionExecutor) RollbackSql(context stdcontext.Context) error {
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
		return exactDecimalText(v), nil
	case time.Time:
		// TeaQL time.Time values are dates (timestamps use epoch milliseconds).
		// Match SQLite DATE text storage so inclusive bounds include the day.
		return v.Format("2006-01-02"), nil
	case core.Timestamp:
		return int64(v), nil
	case string:
		return v, nil
	}
	return nil, fmt.Errorf("unsupported sqlite bind value type %T", value.V)
}

func decodeSqliteRow(cols []string, colTypes []*sql.ColumnType, vals []interface{}) (core.Record, error) {
	return decodeSqliteRowWithDecimalAliases(cols, colTypes, vals, nil)
}

func decodeSqliteRowWithDecimalAliases(cols []string, colTypes []*sql.ColumnType, vals []interface{}, exactAliases map[string]bool) (core.Record, error) {
	record := make(core.Record)
	for i, colName := range cols {
		val := vals[i]
		dbType := ""
		if i < len(colTypes) {
			dbType = strings.ToUpper(colTypes[i].DatabaseTypeName())
		}
		if val == nil {
			record[colName] = core.ValNull()
			continue
		}
		switch v := val.(type) {
		case int64:
			if dbType == "BOOLEAN" || dbType == "BOOL" {
				record[colName] = core.ValBool(v != 0)
			} else if dbType == "NUMERIC" || dbType == "DECIMAL" {
				return nil, fmt.Errorf("legacy SQLite NUMERIC Decimal column %s may have lost precision; migrate to TEAQL_DECIMAL_TEXT", colName)
			} else {
				record[colName] = core.ValI64(v)
			}
		case float64:
			if dbType == "NUMERIC" || dbType == "DECIMAL" {
				return nil, fmt.Errorf("legacy SQLite NUMERIC Decimal column %s may have lost precision; migrate to TEAQL_DECIMAL_TEXT", colName)
			}
			record[colName] = core.ValF64(v)
		case []byte:
			decoded, err := decodeSqliteText(colName, dbType, string(v), exactAliases[colName])
			if err != nil {
				return nil, err
			}
			record[colName] = decoded
		case string:
			decoded, err := decodeSqliteText(colName, dbType, v, exactAliases[colName])
			if err != nil {
				return nil, err
			}
			record[colName] = decoded
		case time.Time:
			// go-sqlite3 decodes DATE/TIMESTAMP columns as time.Time. Preserve the
			// value so generated date and timestamp accessors can convert it.
			record[colName] = core.Value{V: v}
		}
	}
	return record, nil
}

func decodeSqliteText(column, dbType, value string, exactAlias bool) (core.Value, error) {
	if dbType == "NUMERIC" || dbType == "DECIMAL" {
		return core.Value{}, fmt.Errorf("legacy SQLite NUMERIC Decimal column %s may have lost precision; migrate to TEAQL_DECIMAL_TEXT", column)
	}
	if dbType != "TEAQL_DECIMAL_TEXT" && !exactAlias {
		return core.ValText(value), nil
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return core.Value{}, fmt.Errorf("invalid SQLite Decimal in column %s: %w", column, err)
	}
	return core.ValDecimal(parsed), nil
}

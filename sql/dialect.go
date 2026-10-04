package sql

import (
	"fmt"
	"sort"
	"strings"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

var SqlKeywords = []string{
	"all", "alter", "and", "as", "asc", "between", "by", "case", "create", "delete", "desc",
	"distinct", "drop", "exists", "false", "from", "group", "having", "in", "insert", "into", "is",
	"join", "like", "limit", "not", "null", "offset", "on", "or", "order", "select", "set",
	"table", "true", "type", "union", "update", "values", "where",
}

func QuoteIdentifierIfNeeded(ident string, quote rune) string {
	if isWrappedIdentifier(ident) {
		return ident
	}
	if needsQuotedIdentifier(ident) {
		quoteStr := string(quote)
		escaped := strings.ReplaceAll(ident, quoteStr, quoteStr+quoteStr)
		return fmt.Sprintf("%c%s%c", quote, escaped, quote)
	}
	return ident
}

func isWrappedIdentifier(ident string) bool {
	return (strings.HasPrefix(ident, "\"") && strings.HasSuffix(ident, "\"")) ||
		(strings.HasPrefix(ident, "`") && strings.HasSuffix(ident, "`")) ||
		(strings.HasPrefix(ident, "[") && strings.HasSuffix(ident, "]"))
}

func isSqlKeyword(ident string) bool {
	lower := strings.ToLower(ident)
	for _, kw := range SqlKeywords {
		if lower == kw {
			return true
		}
	}
	return false
}

func needsQuotedIdentifier(ident string) bool {
	if ident == "" || isSqlKeyword(ident) {
		return true
	}

	runes := []rune(ident)
	if len(runes) > 0 {
		first := runes[0]
		if first != '_' && !(first >= 'a' && first <= 'z') && !(first >= 'A' && first <= 'Z') {
			return true
		}
	}
	for _, ch := range runes {
		if ch != '_' && !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') {
			return true
		}
	}
	return false
}

type SqlDialect interface {
	Kind() DatabaseKind
	QuoteIdent(ident string) string
	Placeholder(index int) string
	SchemaSetupSqls() []string
	SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error)
	CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error)
}

type DefaultSqlDialect struct {
	Dialect SqlDialect
	// Fresh per top-level compilation, never shared across requests.
	logPolicies  map[int]string
	generatedSQL bool
	likeOperands map[int]core.Value
}

func (d *DefaultSqlDialect) SchemaSetupSqls() []string {
	return []string{}
}

func (d *DefaultSqlDialect) SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error) {
	switch dataType {
	case core.TypeBool:
		return "BOOLEAN", nil
	case core.TypeI64, core.TypeU64:
		return "INTEGER", nil
	case core.TypeF64:
		return "REAL", nil
	case core.TypeDecimal:
		return "NUMERIC", nil
	case core.TypeText:
		return "VARCHAR(255)", nil
	case core.TypeLargeText, core.TypeJson, core.TypeDate, core.TypeTimestamp:
		return "TEXT", nil
	}
	return "", ErrUnsupportedSchemaType(dataType)
}

func (d *DefaultSqlDialect) ColumnDefinitionSql(property *core.PropertyDescriptor) (string, error) {
	quotedName := d.Dialect.QuoteIdent(property.ColName)
	schemaType, err := d.Dialect.SchemaTypeSql(property.DataType, property)
	if err != nil {
		return "", err
	}

	parts := []string{quotedName, schemaType}

	if property.IsId {
		parts = append(parts, "PRIMARY KEY")
	}
	if property.IsId || !property.Nullable {
		parts = append(parts, "NOT NULL")
	}

	return strings.Join(parts, " "), nil
}

func (d *DefaultSqlDialect) CompileCreateTable(entity *core.EntityDescriptor) (string, error) {
	var columns []string
	for _, prop := range entity.Properties {
		colDef, err := d.ColumnDefinitionSql(prop)
		if err != nil {
			return "", err
		}
		columns = append(columns, colDef)
	}

	tableName := d.Dialect.QuoteIdent(entity.TabName)
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columns, ", ")), nil
}

func (d *DefaultSqlDialect) SchemaIndexesSqls(entity *core.EntityDescriptor) ([]string, error) {
	var sqls []string
	tableNameUpper := strings.ToUpper(entity.TabName)
	quotedTable := d.Dialect.QuoteIdent(entity.TabName)

	versionCol := entity.VersionProperty()
	if versionCol != nil {
		idCol := entity.IdProperty()
		idColName := "id"
		if idCol != nil {
			idColName = idCol.ColName
		}

		idxName := fmt.Sprintf("PK_%s_ID_VERSION", tableNameUpper)
		sqls = append(sqls, fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s, %s)",
			d.Dialect.QuoteIdent(idxName),
			quotedTable,
			d.Dialect.QuoteIdent(idColName),
			d.Dialect.QuoteIdent(versionCol.ColName)))
	}

	for _, p := range entity.Properties {
		if strings.HasSuffix(p.Name, "Id") || strings.HasSuffix(p.Name, "_id") {
			idColumn := "id"
			if id := entity.IdProperty(); id != nil {
				idColumn = id.ColName
			}
			idxName := fmt.Sprintf("IDX_%s_%s_%s", tableNameUpper, strings.ToUpper(p.ColName), strings.ToUpper(idColumn))
			sqls = append(sqls, fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s, %s DESC)",
				d.Dialect.QuoteIdent(idxName), quotedTable,
				d.Dialect.QuoteIdent(p.ColName), d.Dialect.QuoteIdent(idColumn)))
			continue
		}
		if strings.HasSuffix(p.Name, "Time") || strings.HasSuffix(p.Name, "_time") || p.Name == "create_time" || p.Name == "update_time" {
			idxName := fmt.Sprintf("IDX_%s_%s", tableNameUpper, strings.ToUpper(p.ColName))
			sqls = append(sqls, fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
				d.Dialect.QuoteIdent(idxName),
				quotedTable,
				d.Dialect.QuoteIdent(p.ColName)))
		}
	}

	return sqls, nil
}

func (d *DefaultSqlDialect) FallbackDefaultValueSql(dataType core.DataType) string {
	switch dataType {
	case core.TypeBool:
		return "FALSE"
	case core.TypeI64, core.TypeU64, core.TypeF64, core.TypeDecimal:
		return "0"
	case core.TypeText, core.TypeLargeText:
		return "''"
	case core.TypeJson:
		return "'{}'"
	case core.TypeDate:
		return "'1970-01-01'"
	case core.TypeTimestamp:
		return "'1970-01-01 00:00:00Z'"
	}
	return "''"
}

func (d *DefaultSqlDialect) CompileAddColumn(entity *core.EntityDescriptor, property *core.PropertyDescriptor) (string, error) {
	def, err := d.ColumnDefinitionSql(property)
	if err != nil {
		return "", err
	}
	if !property.Nullable && !property.IsId {
		def += " DEFAULT " + d.FallbackDefaultValueSql(property.DataType)
	}
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", d.Dialect.QuoteIdent(entity.TabName), def), nil
}

func (d *DefaultSqlDialect) CompileSelect(entity *core.EntityDescriptor, query *core.SelectQuery) (*CompiledQuery, error) {
	d = d.forCompilation()
	var params []core.Value
	resultSql, err := d.compileSelectSql(entity, query, &params)
	if err != nil {
		return nil, err
	}
	return &CompiledQuery{
		Sql:                  resultSql,
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
		Comment:        query.CommentText,
		intentOperands: d.intentOperands(params),
	}, nil
}

func (d *DefaultSqlDialect) compileSelectSql(entity *core.EntityDescriptor, query *core.SelectQuery, params *[]core.Value) (string, error) {
	if query.RawSql != nil || len(query.RawSqlSearchCriteria)+len(query.RawProjections)+len(query.DynamicProperties) > 0 {
		d.generatedSQL = false
	}
	if query.RawSql != nil {
		return *query.RawSql, nil
	}

	projection, err := d.compileProjection(entity, query, params)
	if err != nil {
		return "", err
	}
	partitioned := query.PartitionBy != nil && query.Slice != nil
	if partitioned {
		partitionColumn, err := d.columnSql(entity, *query.PartitionBy)
		if err != nil {
			return "", err
		}
		windowOrder := ""
		if len(query.OrderBy) > 0 {
			var orderByParts []string
			for _, order := range query.OrderBy {
				orderSql, err := d.orderBySql(entity, order, params)
				if err != nil {
					return "", err
				}
				orderByParts = append(orderByParts, orderSql)
			}
			windowOrder = " ORDER BY " + strings.Join(orderByParts, ", ")
		}
		projection += fmt.Sprintf(", ROW_NUMBER() OVER (PARTITION BY %s%s) AS %s", partitionColumn, windowOrder, d.Dialect.QuoteIdent("__teaql_partition_rank"))
	}

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", projection, d.Dialect.QuoteIdent(entity.TabName)))

	var whereParts []string
	if query.Filter != nil {
		filterSql, err := d.compileExpr(entity, query.Filter, params)
		if err != nil {
			return "", err
		}
		whereParts = append(whereParts, filterSql)
	}

	if query.SearchWithText != nil {
		var orParts []string
		likeValue := fmt.Sprintf("%%%s%%", *query.SearchWithText)
		for _, property := range entity.Properties {
			if property.DataType == core.TypeText || property.DataType == core.TypeLargeText {
				d.bindField(params, core.ValText(likeValue), entity, property.Name)
				orParts = append(orParts, fmt.Sprintf("%s LIKE %s", d.Dialect.QuoteIdent(property.ColName), d.Dialect.Placeholder(len(*params))))
			}
		}
		if len(orParts) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("(%s)", strings.Join(orParts, " OR ")))
		}
	}

	whereParts = append(whereParts, query.RawSqlSearchCriteria...)
	if len(whereParts) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(strings.Join(whereParts, " AND "))
	}

	if partitioned {
		rank := d.Dialect.QuoteIdent("__teaql_partition_rank")
		predicates := []string{fmt.Sprintf("%s > %d", rank, query.Slice.Offset)}
		if query.Slice.Limit != nil {
			predicates = append(predicates, fmt.Sprintf("%s <= %d", rank, query.Slice.Offset+*query.Slice.Limit))
		}
		return fmt.Sprintf("SELECT * FROM (%s) AS %s WHERE %s ORDER BY %s", sqlBuilder.String(), d.Dialect.QuoteIdent("__teaql_partitioned"), strings.Join(predicates, " AND "), rank), nil
	}

	if len(query.GroupBy) > 0 {
		var groupByParts []string
		for _, field := range query.GroupBy {
			colSql, err := d.columnSql(entity, field)
			if err != nil {
				return "", err
			}
			groupByParts = append(groupByParts, colSql)
		}
		sqlBuilder.WriteString(" GROUP BY ")
		sqlBuilder.WriteString(strings.Join(groupByParts, ", "))
	}

	if query.Having != nil {
		havingSql, err := d.compileExpr(entity, query.Having, params)
		if err != nil {
			return "", err
		}
		sqlBuilder.WriteString(" HAVING ")
		sqlBuilder.WriteString(havingSql)
	}

	if len(query.OrderBy) > 0 {
		var orderByParts []string
		for _, order := range query.OrderBy {
			orderSql, err := d.orderBySql(entity, order, params)
			if err != nil {
				return "", err
			}
			orderByParts = append(orderByParts, orderSql)
		}
		sqlBuilder.WriteString(" ORDER BY ")
		sqlBuilder.WriteString(strings.Join(orderByParts, ", "))
	}

	if query.Slice != nil {
		if query.Slice.Limit != nil {
			sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", *query.Slice.Limit))
		}
		if query.Slice.Offset > 0 {
			sqlBuilder.WriteString(fmt.Sprintf(" OFFSET %d", query.Slice.Offset))
		}
	}

	return sqlBuilder.String(), nil
}

func (d *DefaultSqlDialect) CompileInsert(entity *core.EntityDescriptor, command *core.InsertCommand) (*CompiledQuery, error) {
	d = d.forCompilation()
	var columns []string
	var placeholders []string
	var params []core.Value

	for _, property := range entity.Properties {
		if value, ok := command.Values[property.Name]; ok {
			columns = append(columns, d.Dialect.QuoteIdent(property.ColName))
			if value.V == nil {
				value = core.ValTypedNull(property.DataType)
			}
			d.bindField(&params, value, entity, property.Name)
			placeholders = append(placeholders, d.Dialect.Placeholder(len(params)))
		}
	}

	if len(columns) == 0 {
		return nil, ErrEmptyMutation("insert")
	}

	return &CompiledQuery{
		Sql: fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			d.Dialect.QuoteIdent(entity.TabName),
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", ")),
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
	}, nil
}

func (d *DefaultSqlDialect) CompileBatchInsert(entity *core.EntityDescriptor, command *core.BatchInsertCommand) (*CompiledQuery, error) {
	d = d.forCompilation()
	if len(command.BatchValues) == 0 {
		return nil, ErrEmptyMutation("batch_insert")
	}

	var columns []*core.PropertyDescriptor
	firstRecord := command.BatchValues[0]

	for _, property := range entity.Properties {
		if _, ok := firstRecord[property.Name]; ok {
			columns = append(columns, property)
		}
	}

	if len(columns) == 0 {
		return nil, ErrEmptyMutation("batch_insert")
	}

	var columnNames []string
	for _, p := range columns {
		columnNames = append(columnNames, d.Dialect.QuoteIdent(p.ColName))
	}

	var params []core.Value
	var valuesClauses []string

	for _, record := range command.BatchValues {
		var rowPlaceholders []string
		for _, property := range columns {
			value, ok := record[property.Name]
			if !ok {
				value = core.ValNull()
			}
			if value.V == nil {
				value = core.ValTypedNull(property.DataType)
			}
			d.bindField(&params, value, entity, property.Name)
			rowPlaceholders = append(rowPlaceholders, d.Dialect.Placeholder(len(params)))
		}
		valuesClauses = append(valuesClauses, fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", ")))
	}

	return &CompiledQuery{
		Sql: fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			d.Dialect.QuoteIdent(entity.TabName),
			strings.Join(columnNames, ", "),
			strings.Join(valuesClauses, ", ")),
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
	}, nil
}

func (d *DefaultSqlDialect) CompileUpdate(entity *core.EntityDescriptor, command *core.UpdateCommand) (*CompiledQuery, error) {
	d = d.forCompilation()
	idProperty := entity.IdProperty()
	if idProperty == nil {
		return nil, ErrMissingIdProperty(entity.Name)
	}

	var assignments []string
	var params []core.Value

	for _, property := range entity.Properties {
		if property.IsId {
			continue
		}
		if property.IsVersion && command.ExpectedVersion != nil {
			continue
		}
		if value, ok := command.Values[property.Name]; ok {
			if value.V == nil {
				value = core.ValTypedNull(property.DataType)
			}
			d.bindField(&params, value, entity, property.Name)
			assignments = append(assignments, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(property.ColName), d.Dialect.Placeholder(len(params))))
		}
	}

	if command.ExpectedVersion != nil {
		versionProperty := entity.VersionProperty()
		if versionProperty == nil {
			return nil, ErrMissingVersionProperty(entity.Name)
		}
		d.bindField(&params, core.ValI64(*command.ExpectedVersion+1), entity, versionProperty.Name)
		assignments = append(assignments, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), d.Dialect.Placeholder(len(params))))
	}
	if len(assignments) == 0 {
		return nil, ErrEmptyMutation("update")
	}

	d.bindField(&params, command.Id, entity, idProperty.Name)
	predicates := []string{fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(idProperty.ColName), d.Dialect.Placeholder(len(params)))}

	if command.ExpectedVersion != nil {
		versionProperty := entity.VersionProperty()
		d.bindField(&params, core.ValI64(*command.ExpectedVersion), entity, versionProperty.Name)
		predicates = append(predicates, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), d.Dialect.Placeholder(len(params))))
	}
	var err error
	predicates, params, err = d.appendMutationGuards(entity, predicates, params, command.Guards)
	if err != nil {
		return nil, err
	}

	return &CompiledQuery{
		Sql: fmt.Sprintf("UPDATE %s SET %s WHERE %s",
			d.Dialect.QuoteIdent(entity.TabName),
			strings.Join(assignments, ", "),
			strings.Join(predicates, " AND ")),
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
	}, nil
}

func (d *DefaultSqlDialect) CompileBatchUpdate(entity *core.EntityDescriptor, command *core.BatchUpdateCommand) (*CompiledQuery, error) {
	d = d.forCompilation()
	if len(command.BatchValues) == 0 {
		return nil, ErrEmptyMutation("batch_update")
	}

	idProperty := entity.IdProperty()
	if idProperty == nil {
		return nil, ErrMissingIdProperty(entity.Name)
	}

	var params []core.Value
	var setClauses []string

	for _, fieldName := range command.UpdateFields {
		property := entity.PropertyByName(fieldName)
		if property == nil {
			return nil, ErrUnknownField(fieldName)
		}

		caseParts := []string{fmt.Sprintf("CASE %s", d.Dialect.QuoteIdent(idProperty.ColName))}

		for i, record := range command.BatchValues {
			id := command.BatchIds[i]
			val, ok := record[fieldName]
			if !ok {
				val = core.ValNull()
			}
			if val.V == nil {
				val = core.ValTypedNull(property.DataType)
			}
			d.bindField(&params, id, entity, idProperty.Name)
			idPh := d.Dialect.Placeholder(len(params))

			d.bindField(&params, val, entity, property.Name)
			valPh := d.Dialect.Placeholder(len(params))

			caseParts = append(caseParts, fmt.Sprintf("WHEN %s THEN %s", idPh, valPh))
		}

		caseParts = append(caseParts, fmt.Sprintf("ELSE %s END", d.Dialect.QuoteIdent(property.ColName)))
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(property.ColName), strings.Join(caseParts, " ")))
	}

	hasVersions := false
	versionProperty := entity.VersionProperty()
	if versionProperty != nil {
		caseParts := []string{fmt.Sprintf("CASE %s", d.Dialect.QuoteIdent(idProperty.ColName))}
		for i, expVerOpt := range command.BatchExpectedVersions {
			if expVerOpt != nil {
				hasVersions = true
				id := command.BatchIds[i]

				d.bindField(&params, id, entity, idProperty.Name)
				idPh := d.Dialect.Placeholder(len(params))

				d.bindField(&params, core.ValI64(*expVerOpt+1), entity, versionProperty.Name)
				valPh := d.Dialect.Placeholder(len(params))

				caseParts = append(caseParts, fmt.Sprintf("WHEN %s THEN %s", idPh, valPh))
			}
		}

		if hasVersions {
			caseParts = append(caseParts, fmt.Sprintf("ELSE %s END", d.Dialect.QuoteIdent(versionProperty.ColName)))
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), strings.Join(caseParts, " ")))
		}
	}

	if len(setClauses) == 0 {
		return nil, ErrEmptyMutation("batch_update")
	}

	var inPlaceholders []string
	for _, id := range command.BatchIds {
		d.bindField(&params, id, entity, idProperty.Name)
		inPlaceholders = append(inPlaceholders, d.Dialect.Placeholder(len(params)))
	}

	predicates := []string{fmt.Sprintf("%s IN (%s)", d.Dialect.QuoteIdent(idProperty.ColName), strings.Join(inPlaceholders, ", "))}

	if hasVersions {
		caseParts := []string{fmt.Sprintf("CASE %s", d.Dialect.QuoteIdent(idProperty.ColName))}
		for i, expVerOpt := range command.BatchExpectedVersions {
			if expVerOpt != nil {
				id := command.BatchIds[i]

				d.bindField(&params, id, entity, idProperty.Name)
				idPh := d.Dialect.Placeholder(len(params))

				d.bindField(&params, core.ValI64(*expVerOpt), entity, versionProperty.Name)
				valPh := d.Dialect.Placeholder(len(params))

				caseParts = append(caseParts, fmt.Sprintf("WHEN %s THEN %s", idPh, valPh))
			}
		}
		caseParts = append(caseParts, fmt.Sprintf("ELSE %s END", d.Dialect.QuoteIdent(versionProperty.ColName)))
		predicates = append(predicates, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), strings.Join(caseParts, " ")))
	}

	return &CompiledQuery{
		Sql: fmt.Sprintf("UPDATE %s SET %s WHERE %s",
			d.Dialect.QuoteIdent(entity.TabName),
			strings.Join(setClauses, ", "),
			strings.Join(predicates, " AND ")),
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
	}, nil
}

func (d *DefaultSqlDialect) CompileDelete(entity *core.EntityDescriptor, command *core.DeleteCommand) (*CompiledQuery, error) {
	d = d.forCompilation()
	idProperty := entity.IdProperty()
	if idProperty == nil {
		return nil, ErrMissingIdProperty(entity.Name)
	}

	var params []core.Value

	if command.SoftDelete {
		versionProperty := entity.VersionProperty()
		if versionProperty == nil {
			return nil, ErrMissingVersionProperty(entity.Name)
		}

		if command.ExpectedVersion != nil {
			d.bindField(&params, core.ValI64(-(*command.ExpectedVersion + 1)), entity, versionProperty.Name)
		} else {
			d.bindField(&params, core.ValI64(-1), entity, versionProperty.Name)
		}
		d.bindField(&params, command.Id, entity, idProperty.Name)
		predicates := []string{fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(idProperty.ColName), d.Dialect.Placeholder(len(params)))}

		if command.ExpectedVersion != nil {
			d.bindField(&params, core.ValI64(*command.ExpectedVersion), entity, versionProperty.Name)
			predicates = append(predicates, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), d.Dialect.Placeholder(len(params))))
		}
		var err error
		predicates, params, err = d.appendMutationGuards(entity, predicates, params, command.Guards)
		if err != nil {
			return nil, err
		}

		return &CompiledQuery{
			Sql: fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s",
				d.Dialect.QuoteIdent(entity.TabName),
				d.Dialect.QuoteIdent(versionProperty.ColName),
				d.Dialect.Placeholder(1),
				strings.Join(predicates, " AND ")),
			Params:               params,
			ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
		}, nil
	}

	d.bindField(&params, command.Id, entity, idProperty.Name)
	predicates := []string{fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(idProperty.ColName), d.Dialect.Placeholder(len(params)))}

	if command.ExpectedVersion != nil {
		versionProperty := entity.VersionProperty()
		if versionProperty == nil {
			return nil, ErrMissingVersionProperty(entity.Name)
		}
		d.bindField(&params, core.ValI64(*command.ExpectedVersion), entity, versionProperty.Name)
		predicates = append(predicates, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), d.Dialect.Placeholder(len(params))))
	}
	var err error
	predicates, params, err = d.appendMutationGuards(entity, predicates, params, command.Guards)
	if err != nil {
		return nil, err
	}

	return &CompiledQuery{
		Sql:                  fmt.Sprintf("DELETE FROM %s WHERE %s", d.Dialect.QuoteIdent(entity.TabName), strings.Join(predicates, " AND ")),
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
	}, nil
}

func (d *DefaultSqlDialect) CompileRecover(entity *core.EntityDescriptor, command *core.RecoverCommand) (*CompiledQuery, error) {
	d = d.forCompilation()
	if command.ExpectedVersion >= 0 {
		return nil, ErrInvalidRecoverVersion(command.ExpectedVersion)
	}

	idProperty := entity.IdProperty()
	if idProperty == nil {
		return nil, ErrMissingIdProperty(entity.Name)
	}
	versionProperty := entity.VersionProperty()
	if versionProperty == nil {
		return nil, ErrMissingVersionProperty(entity.Name)
	}

	var params []core.Value
	d.bindField(&params, core.ValI64(-command.ExpectedVersion+1), entity, versionProperty.Name)
	d.bindField(&params, command.Id, entity, idProperty.Name)
	d.bindField(&params, core.ValI64(command.ExpectedVersion), entity, versionProperty.Name)

	predicates := []string{
		fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(idProperty.ColName), d.Dialect.Placeholder(2)),
		fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(versionProperty.ColName), d.Dialect.Placeholder(3)),
	}
	var err error
	predicates, params, err = d.appendMutationGuards(entity, predicates, params, command.Guards)
	if err != nil {
		return nil, err
	}
	return &CompiledQuery{
		Sql: fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s",
			d.Dialect.QuoteIdent(entity.TabName),
			d.Dialect.QuoteIdent(versionProperty.ColName),
			d.Dialect.Placeholder(1),
			strings.Join(predicates, " AND "),
		),
		Params:               params,
		ParameterLogPolicies: d.parameterPolicies(params), GeneratedSQL: d.generatedSQL,
	}, nil
}

func (d *DefaultSqlDialect) appendMutationGuards(entity *core.EntityDescriptor, predicates []string, params []core.Value, guards core.Record) ([]string, []core.Value, error) {
	keys := make([]string, 0, len(guards))
	for field := range guards {
		keys = append(keys, field)
	}
	sort.Strings(keys)
	for _, field := range keys {
		property := entity.PropertyByName(field)
		if property == nil {
			return nil, nil, ErrUnknownField(field)
		}
		d.bindField(&params, guards[field], entity, field)
		predicates = append(predicates, fmt.Sprintf("%s = %s", d.Dialect.QuoteIdent(property.ColName), d.Dialect.Placeholder(len(params))))
	}
	return predicates, params, nil
}

func (d *DefaultSqlDialect) columnSql(entity *core.EntityDescriptor, field string) (string, error) {
	property := entity.PropertyByName(field)
	if property == nil {
		return "", ErrUnknownField(field)
	}
	return d.Dialect.QuoteIdent(property.ColName), nil
}

func (d *DefaultSqlDialect) orderBySql(entity *core.EntityDescriptor, orderBy *core.OrderBy, params *[]core.Value) (string, error) {
	var field string
	var err error
	if orderBy.Expr != nil {
		field, err = d.compileExpr(entity, orderBy.Expr, params)
	} else {
		field, err = d.columnSql(entity, orderBy.Field)
	}
	if err != nil {
		return "", err
	}
	direction := "ASC"
	if orderBy.Direction == core.SortDesc {
		direction = "DESC"
	}
	return fmt.Sprintf("%s %s", field, direction), nil
}

func (d *DefaultSqlDialect) selectProjection(entity *core.EntityDescriptor, query *core.SelectQuery, params *[]core.Value) (string, error) {
	propertyProjection := func(property *core.PropertyDescriptor) string {
		return d.columnWithAlias(property)
	}

	if len(query.Projection) == 0 && len(query.ExprProjection) == 0 && len(query.RawProjections) == 0 && len(query.DynamicProperties) == 0 {
		var parts []string
		for _, prop := range entity.Properties {
			parts = append(parts, propertyProjection(prop))
		}
		return strings.Join(parts, ", "), nil
	}

	var parts []string
	for _, field := range query.Projection {
		property := entity.PropertyByName(field)
		if property == nil {
			return "", ErrUnknownField(field)
		}
		parts = append(parts, propertyProjection(property))
	}
	for _, projection := range query.ExprProjection {
		expr, err := d.compileExpr(entity, projection.Expr, params)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s AS %s", expr, d.Dialect.QuoteIdent(projection.Alias)))
	}
	for _, projection := range query.RawProjections {
		parts = append(parts, fmt.Sprintf("%s AS %s", projection.RawSqlSegment, d.Dialect.QuoteIdent(projection.PropertyName)))
	}
	for _, projection := range query.DynamicProperties {
		parts = append(parts, fmt.Sprintf("%s AS %s", projection.RawSqlSegment, d.Dialect.QuoteIdent(projection.PropertyName)))
	}
	return strings.Join(parts, ", "), nil
}

func (d *DefaultSqlDialect) aggregateProjection(entity *core.EntityDescriptor, query *core.SelectQuery, params *[]core.Value) (string, error) {
	var parts []string

	contains := func(slice []string, val string) bool {
		for _, item := range slice {
			if item == val {
				return true
			}
		}
		return false
	}

	for _, field := range query.GroupBy {
		colSql, err := d.columnSql(entity, field)
		if err != nil {
			return "", err
		}
		if !contains(parts, colSql) {
			parts = append(parts, colSql)
		}
	}
	for _, field := range query.Projection {
		colSql, err := d.columnSql(entity, field)
		if err != nil {
			return "", err
		}
		if !contains(parts, colSql) {
			parts = append(parts, colSql)
		}
	}
	for _, projection := range query.ExprProjection {
		expr, err := d.compileExpr(entity, projection.Expr, params)
		if err != nil {
			return "", err
		}
		aliased := fmt.Sprintf("%s AS %s", expr, d.Dialect.QuoteIdent(projection.Alias))
		if !contains(parts, aliased) {
			parts = append(parts, aliased)
		}
	}
	for _, projection := range query.RawProjections {
		aliased := fmt.Sprintf("%s AS %s", projection.RawSqlSegment, d.Dialect.QuoteIdent(projection.PropertyName))
		if !contains(parts, aliased) {
			parts = append(parts, aliased)
		}
	}
	for _, projection := range query.DynamicProperties {
		aliased := fmt.Sprintf("%s AS %s", projection.RawSqlSegment, d.Dialect.QuoteIdent(projection.PropertyName))
		if !contains(parts, aliased) {
			parts = append(parts, aliased)
		}
	}
	for _, aggregate := range query.Aggregates {
		field, err := d.resolveAggregateField(entity, aggregate)
		if err != nil {
			return "", err
		}
		call := d.aggregateCallSql(aggregate.Function, field)
		parts = append(parts, fmt.Sprintf("%s AS %s", call, d.Dialect.QuoteIdent(aggregate.Alias)))
	}

	return strings.Join(parts, ", "), nil
}

func (d *DefaultSqlDialect) aggregateCallSql(function core.AggregateFunction, field string) string {
	return fmt.Sprintf("%s(%s)", d.aggregateFunctionSql(function), field)
}

func (d *DefaultSqlDialect) aggregateFunctionSql(function core.AggregateFunction) string {
	switch function {
	case core.AggCount:
		return "COUNT"
	case core.AggSum:
		return "SUM"
	case core.AggAvg:
		return "AVG"
	case core.AggMin:
		return "MIN"
	case core.AggMax:
		return "MAX"
	case core.AggStddev:
		return "STDDEV"
	case core.AggStddevPop:
		return "STDDEV_POP"
	case core.AggVarSamp:
		return "VAR_SAMP"
	case core.AggVarPop:
		return "VAR_POP"
	case core.AggBitAnd:
		return "BIT_AND"
	case core.AggBitOr:
		return "BIT_OR"
	case core.AggBitXor:
		return "BIT_XOR"
	}
	return "UNKNOWN"
}

func (d *DefaultSqlDialect) compileExpr(entity *core.EntityDescriptor, expr *core.Expr, params *[]core.Value) (string, error) {
	start := len(*params)
	result, err := d.compileExprSQL(entity, expr, params)
	if policy := d.expressionPolicy(entity, expr); d.logPolicies != nil && policy != "" {
		for i := start; i < len(*params); i++ {
			if _, exists := d.logPolicies[i]; !exists {
				d.logPolicies[i] = policy
			}
		}
	}
	return result, err
}

func (d *DefaultSqlDialect) compileExprSQL(entity *core.EntityDescriptor, expr *core.Expr, params *[]core.Value) (string, error) {
	switch expr.Type {
	case core.ExprTypeColumn:
		return d.columnSql(entity, expr.Column)
	case core.ExprTypeValue:
		*params = append(*params, expr.Value)
		return d.Dialect.Placeholder(len(*params)), nil
	case core.ExprTypeFunctionCall:
		return d.compileFunction(entity, expr.Function, expr.Args, params)
	case core.ExprTypeBinary:
		op := expr.Op
		if op == core.OpIn || op == core.OpNotIn || op == core.OpInLarge || op == core.OpNotInLarge {
			return d.compileIn(entity, expr.Left, op, expr.Right, params)
		}
		lhs, err := d.compileExpr(entity, expr.Left, params)
		if err != nil {
			return "", err
		}
		rhs, err := d.compileExpr(entity, expr.Right, params)
		if err != nil {
			return "", err
		}
		if operand, ok := logprivacy.ReadIntentSource(expr.DiagnosticLikeOperand()).(core.Value); ok && d.likeOperands != nil {
			d.likeOperands[len(*params)-1] = operand
		}
		opStr := ""
		switch op {
		case core.OpEq:
			opStr = "="
		case core.OpNe:
			opStr = "!="
		case core.OpGt:
			opStr = ">"
		case core.OpGte:
			opStr = ">="
		case core.OpLt:
			opStr = "<"
		case core.OpLte:
			opStr = "<="
		case core.OpLike:
			opStr = "LIKE"
		case core.OpNotLike:
			opStr = "NOT LIKE"
		}
		return fmt.Sprintf("(%s %s %s)", lhs, opStr, rhs), nil
	case core.ExprTypeSubQuery:
		return d.compileSubquery(entity, expr.Left, expr.Op, expr.Entity, expr.Query, params)
	case core.ExprTypeBetween:
		exprSql, err := d.compileExpr(entity, expr.Left, params)
		if err != nil {
			return "", err
		}
		lowerSql, err := d.compileExpr(entity, expr.Lower, params)
		if err != nil {
			return "", err
		}
		upperSql, err := d.compileExpr(entity, expr.Upper, params)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s BETWEEN %s AND %s)", exprSql, lowerSql, upperSql), nil
	case core.ExprTypeIsNull:
		exprSql, err := d.compileExpr(entity, expr.Left, params)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s IS NULL)", exprSql), nil
	case core.ExprTypeIsNotNull:
		exprSql, err := d.compileExpr(entity, expr.Left, params)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s IS NOT NULL)", exprSql), nil
	case core.ExprTypeAnd:
		return d.compileJoined(entity, expr.Parts, "AND", params)
	case core.ExprTypeOr:
		return d.compileJoined(entity, expr.Parts, "OR", params)
	case core.ExprTypeNot:
		exprSql, err := d.compileExpr(entity, expr.Left, params)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(NOT %s)", exprSql), nil
	}
	return "", fmt.Errorf("unsupported expr type: %d", expr.Type)
}

func (d *DefaultSqlDialect) compileFunction(entity *core.EntityDescriptor, function core.ExprFunction, args []*core.Expr, params *[]core.Value) (string, error) {
	switch function {
	case core.FuncSoundex:
		if len(args) != 1 {
			return "", ErrInvalidFunctionArguments("SOUNDEX expects exactly one argument")
		}
		arg, err := d.compileExpr(entity, args[0], params)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("SOUNDEX(%s)", arg), nil
	case core.FuncGbk:
		return d.Dialect.CompileGbkFunction(entity, args, params)
	case core.FuncCount:
		if len(args) == 0 {
			return "COUNT(*)", nil
		}
		return d.compileSingleArgFunction(entity, "COUNT", args, params)
	case core.FuncSum:
		return d.compileSingleArgFunction(entity, "SUM", args, params)
	case core.FuncAvg:
		return d.compileSingleArgFunction(entity, "AVG", args, params)
	case core.FuncMin:
		return d.compileSingleArgFunction(entity, "MIN", args, params)
	case core.FuncMax:
		return d.compileSingleArgFunction(entity, "MAX", args, params)
	case core.FuncStddev:
		return d.compileSingleArgFunction(entity, "STDDEV", args, params)
	case core.FuncStddevPop:
		return d.compileSingleArgFunction(entity, "STDDEV_POP", args, params)
	case core.FuncVarSamp:
		return d.compileSingleArgFunction(entity, "VAR_SAMP", args, params)
	case core.FuncVarPop:
		return d.compileSingleArgFunction(entity, "VAR_POP", args, params)
	case core.FuncBitAnd:
		return d.compileSingleArgFunction(entity, "BIT_AND", args, params)
	case core.FuncBitOr:
		return d.compileSingleArgFunction(entity, "BIT_OR", args, params)
	case core.FuncBitXor:
		return d.compileSingleArgFunction(entity, "BIT_XOR", args, params)
	}
	return "", ErrInvalidFunctionArguments(fmt.Sprintf("unknown function: %v", function))
}

func (d *DefaultSqlDialect) compileSingleArgFunction(entity *core.EntityDescriptor, function string, args []*core.Expr, params *[]core.Value) (string, error) {
	if len(args) != 1 {
		return "", ErrInvalidFunctionArguments(fmt.Sprintf("%s expects exactly one argument", function))
	}
	arg, err := d.compileExpr(entity, args[0], params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", function, arg), nil
}

func (d *DefaultSqlDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	if len(args) != 1 {
		return "", ErrInvalidFunctionArguments("GBK expects exactly one argument")
	}
	return d.compileExpr(entity, args[0], params)
}

func (d *DefaultSqlDialect) compileSubquery(entity *core.EntityDescriptor, left *core.Expr, op core.BinaryOp, subEntity *core.EntityDescriptor, query *core.SelectQuery, params *[]core.Value) (string, error) {
	lhs, err := d.compileExpr(entity, left, params)
	if err != nil {
		return "", err
	}
	operator := ""
	switch op {
	case core.OpIn, core.OpInLarge:
		operator = "IN"
	case core.OpNotIn, core.OpNotInLarge:
		operator = "NOT IN"
	default:
		return "", ErrInvalidSubQueryOperator(fmt.Sprintf("%v", op))
	}
	subquery, err := d.compileSelectSql(subEntity, query, params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s %s (%s))", lhs, operator, subquery), nil
}

func (d *DefaultSqlDialect) compileJoined(entity *core.EntityDescriptor, parts []*core.Expr, joiner string, params *[]core.Value) (string, error) {
	var compiled []string
	for _, part := range parts {
		expr, err := d.compileExpr(entity, part, params)
		if err != nil {
			return "", err
		}
		compiled = append(compiled, expr)
	}
	return fmt.Sprintf("(%s)", strings.Join(compiled, fmt.Sprintf(" %s ", joiner))), nil
}

func (d *DefaultSqlDialect) compileIn(entity *core.EntityDescriptor, left *core.Expr, op core.BinaryOp, right *core.Expr, params *[]core.Value) (string, error) {
	lhs, err := d.compileExpr(entity, left, params)
	if err != nil {
		return "", err
	}
	operator := ""
	switch op {
	case core.OpIn, core.OpInLarge:
		operator = "IN"
	case core.OpNotIn, core.OpNotInLarge:
		operator = "NOT IN"
	}

	if right.Type == core.ExprTypeValue {
		if list, ok := right.Value.V.([]core.Value); ok {
			if len(list) == 0 {
				return "", ErrEmptyInList()
			}
			var placeholders []string
			for _, value := range list {
				*params = append(*params, value)
				placeholders = append(placeholders, d.Dialect.Placeholder(len(*params)))
			}
			return fmt.Sprintf("(%s %s (%s))", lhs, operator, strings.Join(placeholders, ", ")), nil
		}
	}

	rhs, err := d.compileExpr(entity, right, params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s %s (%s))", lhs, operator, rhs), nil
}

func (d *DefaultSqlDialect) compileProjection(entity *core.EntityDescriptor, query *core.SelectQuery, params *[]core.Value) (string, error) {
	if len(query.Aggregates) == 0 {
		return d.selectProjection(entity, query, params)
	}
	return d.aggregateProjection(entity, query, params)
}

func (d *DefaultSqlDialect) resolveOrderField(entity *core.EntityDescriptor, orderBy *core.OrderBy, params *[]core.Value) (string, error) {
	if orderBy.Expr != nil {
		return d.compileExpr(entity, orderBy.Expr, params)
	}
	return d.columnSql(entity, orderBy.Field)
}

func (d *DefaultSqlDialect) columnWithAlias(property *core.PropertyDescriptor) string {
	column := d.Dialect.QuoteIdent(property.ColName)
	if property.ColName == property.Name {
		return column
	}
	return fmt.Sprintf("%s AS %s", column, d.Dialect.QuoteIdent(property.Name))
}

func (d *DefaultSqlDialect) resolveAggregateField(entity *core.EntityDescriptor, aggregate *core.Aggregate) (string, error) {
	if aggregate.Function == core.AggCount && aggregate.Field == "*" {
		return "*", nil
	}
	return d.columnSql(entity, aggregate.Field)
}

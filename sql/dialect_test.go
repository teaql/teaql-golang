package sql

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

type TestDialect struct{}

func (d *TestDialect) Kind() DatabaseKind {
	return DatabaseKindPostgreSQL
}

func (d *TestDialect) QuoteIdent(ident string) string {
	return fmt.Sprintf("\"%s\"", ident)
}

func (d *TestDialect) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (d *TestDialect) SchemaSetupSqls() []string {
	return []string{}
}

func (d *TestDialect) SchemaTypeSql(dataType core.DataType, property *core.PropertyDescriptor) (string, error) {
	return "TEST", nil
}

func (d *TestDialect) CompileGbkFunction(entity *core.EntityDescriptor, args []*core.Expr, params *[]core.Value) (string, error) {
	if len(args) != 1 {
		return "", ErrInvalidFunctionArguments("GBK expects exactly one argument")
	}
	defaultDialect := &DefaultSqlDialect{Dialect: d}
	arg, err := defaultDialect.compileExpr(entity, args[0], params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("convert_to(%s, 'GBK')", arg), nil
}

func entity() *core.EntityDescriptor {
	return core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).ColumnName("version").Version().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
}

func lineEntity() *core.EntityDescriptor {
	return core.NewEntityDescriptor("OrderLine").
		TableName("orderline").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("order_id", core.TypeU64).ColumnName("order_id")).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
}

const OrderDefaultProjection = "\"id\", \"version\", \"name\""

func TestQuotesIdentifiersOnlyWhenNeeded(t *testing.T) {
	assert.Equal(t, "stock_item_data", QuoteIdentifierIfNeeded("stock_item_data", '"'))
	assert.Equal(t, "\"select\"", QuoteIdentifierIfNeeded("select", '"'))
	assert.Equal(t, "`order`", QuoteIdentifierIfNeeded("order", '`'))
	assert.Equal(t, "\"has space\"", QuoteIdentifierIfNeeded("has space", '"'))
	assert.Equal(t, "\"already_wrapped\"", QuoteIdentifierIfNeeded("\"already_wrapped\"", '"'))
}

func TestCompilesSelectWithFiltersOrderAndLimit(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	query := core.NewSelectQuery("Order").
		Project("id").
		Project("name").
		WithFilter(core.ExprEq("name", core.ValText("A"))).
		OrderDesc("id").
		Limit(10).
		Offset(5)
		
	compiled, err := defaultDialect.CompileSelect(entity(), query)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT \"id\", \"name\" FROM \"orders\" WHERE (\"name\" = $1) ORDER BY \"id\" DESC LIMIT 10 OFFSET 5", compiled.Sql)
	assert.Equal(t, []core.Value{core.ValText("A")}, compiled.Params)
}

func TestCompilesAggregateProjection(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	query := core.NewSelectQuery("Order").CountField("id", "count")
	compiled, err := defaultDialect.CompileSelect(entity(), query)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT COUNT(\"id\") AS \"count\" FROM \"orders\"", compiled.Sql)
}

func TestCompilesGroupedAggregateAndExtendedPredicates(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	query := core.NewSelectQuery("Order").
		WithGroupBy("name").
		Count("total").
		Sum("version", "versionSum").
		WithFilter(
			core.ExprBetweenNode("version", core.ValI64(1), core.ValI64(9)).
				AndExpr(core.ExprNotLike("name", "tmp%")).
				AndExpr(core.ExprNotInList("name", []core.Value{core.ValText("x"), core.ValText("y")})).
				AndExpr(core.ExprIsNotNullNode("name")),
		).
		OrderAsc("name")
		
	compiled, err := defaultDialect.CompileSelect(entity(), query)
	assert.NoError(t, err)
	
	assert.Equal(t, "SELECT \"name\", COUNT(*) AS \"total\", SUM(\"version\") AS \"versionSum\" FROM \"orders\" WHERE ((\"version\" BETWEEN $1 AND $2) AND (\"name\" NOT LIKE $3) AND (\"name\" NOT IN ($4, $5)) AND (\"name\" IS NOT NULL)) GROUP BY \"name\" ORDER BY \"name\" ASC", compiled.Sql)
	assert.Equal(t, []core.Value{core.ValI64(1), core.ValI64(9), core.ValText("tmp%"), core.ValText("x"), core.ValText("y")}, compiled.Params)
}

func TestCompilesInsertUpdateDeleteAndRecover(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	insertCmd := core.NewInsertCommand("Order").Value("id", core.ValU64(1)).Value("name", core.ValText("A"))
	insert, err := defaultDialect.CompileInsert(entity(), insertCmd)
	assert.NoError(t, err)
	// Output order may differ because map iteration, but they should be there.
	// Since order isn't guaranteed by map, let's just check length and contains.
	assert.Contains(t, insert.Sql, "INSERT INTO \"orders\" (")
	assert.Equal(t, 2, len(insert.Params))
	
	updateCmd := core.NewUpdateCommand("Order", core.ValU64(1)).WithExpectedVersion(3).Value("name", core.ValText("B"))
	update, err := defaultDialect.CompileUpdate(entity(), updateCmd)
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE \"orders\" SET \"name\" = $1, \"version\" = $2 WHERE \"id\" = $3 AND \"version\" = $4", update.Sql)
	
	deleteCmd := core.NewDeleteCommand("Order", core.ValU64(1)).WithExpectedVersion(3)
	deleteSql, err := defaultDialect.CompileDelete(entity(), deleteCmd)
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE \"orders\" SET \"version\" = $1 WHERE \"id\" = $2 AND \"version\" = $3", deleteSql.Sql)
	
	recoverCmd := core.NewRecoverCommand("Order", core.ValU64(1), -4)
	recoverSql, err := defaultDialect.CompileRecover(entity(), recoverCmd)
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE \"orders\" SET \"version\" = $1 WHERE \"id\" = $2 AND \"version\" = $3", recoverSql.Sql)
}

func TestDefaultDialect_Schema(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	sqls := defaultDialect.SchemaSetupSqls()
	assert.Equal(t, 0, len(sqls))
	
	typeSql, err := defaultDialect.SchemaTypeSql(core.TypeU64, core.NewPropertyDescriptor("id", core.TypeU64))
	assert.NoError(t, err)
	assert.Equal(t, "INTEGER", typeSql)
	
	colSql, err := defaultDialect.ColumnDefinitionSql(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").NotNull())
	assert.NoError(t, err)
	assert.Equal(t, "\"id\" TEST NOT NULL", colSql)
	
	createTable, err := defaultDialect.CompileCreateTable(entity())
	assert.NoError(t, err)
	assert.Contains(t, createTable, "CREATE TABLE IF NOT EXISTS \"orders\"")
	
	idxSqls, err := defaultDialect.SchemaIndexesSqls(entity())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(idxSqls))
	
	defaultVal := defaultDialect.FallbackDefaultValueSql(core.TypeU64)
	assert.Equal(t, "0", defaultVal)
	
	addColumn, err := defaultDialect.CompileAddColumn(entity(), core.NewPropertyDescriptor("new_col", core.TypeText).ColumnName("new_col"))
	assert.NoError(t, err)
	assert.Contains(t, addColumn, "ALTER TABLE \"orders\" ADD COLUMN \"new_col\"")
}

func TestDefaultDialect_BatchInsertUpdate(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	batchInsert := core.NewBatchInsertCommand("Order")
	batchInsert.BatchValues = append(batchInsert.BatchValues, core.Record{"id": core.ValU64(1), "name": core.ValText("A")})
	batchInsert.BatchValues = append(batchInsert.BatchValues, core.Record{"id": core.ValU64(2), "name": core.ValText("B")})
	
	insertSql, err := defaultDialect.CompileBatchInsert(entity(), batchInsert)
	assert.NoError(t, err)
	assert.Contains(t, insertSql.Sql, "INSERT INTO \"orders\"")
	
	batchUpdate := core.NewBatchUpdateCommand("Order", []string{"name"})
	batchUpdate.BatchIds = append(batchUpdate.BatchIds, core.ValU64(1), core.ValU64(2))
	batchUpdate.BatchValues = append(batchUpdate.BatchValues, core.Record{"name": core.ValText("C")}, core.Record{"name": core.ValText("D")})
	v1 := int64(1)
	batchUpdate.BatchExpectedVersions = append(batchUpdate.BatchExpectedVersions, &v1, &v1)
	
	updateSql, err := defaultDialect.CompileBatchUpdate(entity(), batchUpdate)
	assert.NoError(t, err)
	assert.Contains(t, updateSql.Sql, "UPDATE \"orders\" SET")
}

func TestDefaultDialect_Expressions(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// Test function compilation (like Gbk)
	gbkQuery := core.NewSelectQuery("Order").WithFilter(core.ExprEq(core.ExprGbk(core.ExprColumnNode("name")).Column, core.ValText("a")))
	// Wait, ExprEq takes (string, Value). We need ExprEqNode(left, right)
	
	// Let's use ExprFunctionNode manually
	funcExpr := core.ExprGbk(core.ExprColumnNode("name"))
	eqExpr := core.ExprBinaryNode(funcExpr, core.OpEq, core.ExprValueNode(core.ValText("a")))
	gbkQuery = core.NewSelectQuery("Order").WithFilter(eqExpr)
	
	compiled, err := defaultDialect.CompileSelect(entity(), gbkQuery)
	assert.NoError(t, err)
	assert.Contains(t, compiled.Sql, "convert_to(\"name\", 'GBK')")
	
	// Test Subquery
	subq := core.NewSelectQuery("OrderLine").Project("order_id").WithFilter(core.ExprEq("id", core.ValU64(1)))
	subqExpr := core.ExprInSubQuery("id", lineEntity(), subq, "order_id")
	subqQuery := core.NewSelectQuery("Order").WithFilter(subqExpr)
	compiled2, err := defaultDialect.CompileSelect(entity(), subqQuery)
	assert.NoError(t, err)
	assert.Contains(t, compiled2.Sql, "\"id\" IN (SELECT \"order_id\" FROM \"orderline\" WHERE (\"id\" = $1))")
}


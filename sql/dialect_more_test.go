package sql

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
)

func TestDialectCoverage_SchemaTypes(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// Test SchemaTypeSql
	types := []core.DataType{core.TypeBool, core.TypeI64, core.TypeU64, core.TypeF64, core.TypeDecimal, core.TypeText, core.TypeLargeText, core.TypeJson, core.TypeDate, core.TypeTimestamp}
	for _, dt := range types {
		_, err := defaultDialect.SchemaTypeSql(dt, core.NewPropertyDescriptor("id", dt))
		assert.NoError(t, err)
	}
	
	// Test FallbackDefaultValueSql coverage
	assert.Equal(t, "FALSE", defaultDialect.FallbackDefaultValueSql(core.TypeBool))
	assert.Equal(t, "0", defaultDialect.FallbackDefaultValueSql(core.TypeF64))
	assert.Equal(t, "0", defaultDialect.FallbackDefaultValueSql(core.TypeDecimal))
	assert.Equal(t, "''", defaultDialect.FallbackDefaultValueSql(core.TypeText))
	assert.Equal(t, "''", defaultDialect.FallbackDefaultValueSql(core.TypeLargeText))
	assert.Equal(t, "'{}'", defaultDialect.FallbackDefaultValueSql(core.TypeJson))
	assert.Equal(t, "'1970-01-01'", defaultDialect.FallbackDefaultValueSql(core.TypeDate))
	assert.Equal(t, "'1970-01-01 00:00:00Z'", defaultDialect.FallbackDefaultValueSql(core.TypeTimestamp))
	assert.Equal(t, "''", defaultDialect.FallbackDefaultValueSql(core.DataType(99))) // unknown
}

func TestDialectCoverage_SchemaIndexes(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// Test SchemaIndexesSqls for special suffixes
	ent := core.NewEntityDescriptor("Suffix").
		TableName("suffix").
		Property(core.NewPropertyDescriptor("myId", core.TypeU64).ColumnName("my_id")).
		Property(core.NewPropertyDescriptor("myTime", core.TypeU64).ColumnName("my_time")).
		Property(core.NewPropertyDescriptor("my_time", core.TypeU64).ColumnName("my_time_2")).
		Property(core.NewPropertyDescriptor("create_time", core.TypeU64).ColumnName("create_time")).
		Property(core.NewPropertyDescriptor("update_time", core.TypeU64).ColumnName("update_time"))
		
	sqls, err := defaultDialect.SchemaIndexesSqls(ent)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(sqls))
}

func TestDialectCoverage_CompileUpdateErrors(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	noIdEnt := core.NewEntityDescriptor("NoId").
		TableName("noid").
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		
	// CompileUpdate missing id
	_, err := defaultDialect.CompileUpdate(noIdEnt, core.NewUpdateCommand("NoId", core.ValI64(1)))
	assert.Error(t, err)
	
	// CompileDelete missing id
	_, err = defaultDialect.CompileDelete(noIdEnt, core.NewDeleteCommand("NoId", core.ValI64(1)))
	assert.Error(t, err)
	
	// CompileRecover missing id
	_, err = defaultDialect.CompileRecover(noIdEnt, core.NewRecoverCommand("NoId", core.ValI64(1), -1))
	assert.Error(t, err)
	
	// CompileRecover positive expectedVersion
	_, err = defaultDialect.CompileRecover(entity(), core.NewRecoverCommand("Order", core.ValI64(1), 1))
	assert.Error(t, err)
	
	// CompileUpdate expected version missing version property
	updCmd := core.NewUpdateCommand("Order", core.ValU64(1)).Value("name", core.ValText("B")).WithExpectedVersion(1)
	entNoVer := core.NewEntityDescriptor("Order").
		TableName("orders").
		Property(core.NewPropertyDescriptor("id", core.TypeU64).ColumnName("id").Id().NotNull()).
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
	_, err = defaultDialect.CompileUpdate(entNoVer, updCmd)
	assert.Error(t, err)
	
	// CompileDelete soft delete missing version property
	delCmd := core.NewDeleteCommand("Order", core.ValI64(1))
	_, err = defaultDialect.CompileDelete(entNoVer, delCmd)
	assert.Error(t, err)
	
	// CompileRecover missing version property
	_, err = defaultDialect.CompileRecover(entNoVer, core.NewRecoverCommand("Order", core.ValI64(1), -1))
	assert.Error(t, err)
	
	// CompileDelete soft delete with expected version and valid entity
	delCmdSoft := core.NewDeleteCommand("Order", core.ValI64(1)).WithExpectedVersion(2)
	softDelSql, err := defaultDialect.CompileDelete(entity(), delCmdSoft)
	assert.NoError(t, err)
	assert.Contains(t, softDelSql.Sql, "UPDATE")
	
	// CompileDelete hard delete
	delCmdHard := core.NewDeleteCommand("Order", core.ValI64(1)).HardDelete()
	hardDelSql, err := defaultDialect.CompileDelete(entity(), delCmdHard)
	assert.NoError(t, err)
	assert.Contains(t, hardDelSql.Sql, "DELETE")
	
	// CompileDelete hard delete with expected version
	delCmdHardVer := core.NewDeleteCommand("Order", core.ValI64(1)).HardDelete().WithExpectedVersion(2)
	hardDelVerSql, err := defaultDialect.CompileDelete(entity(), delCmdHardVer)
	assert.NoError(t, err)
	assert.Contains(t, hardDelVerSql.Sql, "DELETE")
	
	// CompileDelete hard delete expected version missing version property
	_, err = defaultDialect.CompileDelete(entNoVer, delCmdHardVer)
	assert.Error(t, err)
}

func TestDialectCoverage_BatchUpdate(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	noIdEnt := core.NewEntityDescriptor("NoId").
		TableName("noid").
		Property(core.NewPropertyDescriptor("name", core.TypeText).ColumnName("name"))
		
	batchUpdate := core.NewBatchUpdateCommand("NoId", []string{"name"})
	batchUpdate.BatchValues = append(batchUpdate.BatchValues, core.Record{"name": core.ValText("C")})
	batchUpdate.BatchIds = append(batchUpdate.BatchIds, core.ValU64(1))
	
	// Missing id property
	_, err := defaultDialect.CompileBatchUpdate(noIdEnt, batchUpdate)
	assert.Error(t, err)
	
	// Empty mutation
	emptyUpdate := core.NewBatchUpdateCommand("Order", []string{"name"})
	_, err = defaultDialect.CompileBatchUpdate(entity(), emptyUpdate)
	assert.Error(t, err)
	
	// Unknown field
	unknownFieldUpdate := core.NewBatchUpdateCommand("Order", []string{"unknown"})
	unknownFieldUpdate.BatchValues = append(unknownFieldUpdate.BatchValues, core.Record{"name": core.ValText("C")})
	unknownFieldUpdate.BatchIds = append(unknownFieldUpdate.BatchIds, core.ValU64(1))
	_, err = defaultDialect.CompileBatchUpdate(entity(), unknownFieldUpdate)
	assert.Error(t, err)
	
	// Missing field value falls back to null
	batchUpdateMiss := core.NewBatchUpdateCommand("Order", []string{"name"})
	batchUpdateMiss.BatchValues = append(batchUpdateMiss.BatchValues, core.Record{})
	batchUpdateMiss.BatchIds = append(batchUpdateMiss.BatchIds, core.ValU64(1))
	updateSqlMiss, err := defaultDialect.CompileBatchUpdate(entity(), batchUpdateMiss)
	assert.NoError(t, err)
	assert.Contains(t, updateSqlMiss.Sql, "UPDATE")
	
	// Expected version without versions in batch (nil in BatchExpectedVersions)
	batchUpdateNoVer := core.NewBatchUpdateCommand("Order", []string{"name"})
	batchUpdateNoVer.BatchValues = append(batchUpdateNoVer.BatchValues, core.Record{"name": core.ValText("C")})
	batchUpdateNoVer.BatchIds = append(batchUpdateNoVer.BatchIds, core.ValU64(1))
	batchUpdateNoVer.BatchExpectedVersions = append(batchUpdateNoVer.BatchExpectedVersions, nil)
	updateSqlNoVer, err := defaultDialect.CompileBatchUpdate(entity(), batchUpdateNoVer)
	assert.NoError(t, err)
	assert.Contains(t, updateSqlNoVer.Sql, "UPDATE")
}

func TestDialectCoverage_EmptyMutations(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// Empty Insert
	insertCmd := core.NewInsertCommand("Order")
	_, err := defaultDialect.CompileInsert(entity(), insertCmd)
	assert.Error(t, err)
	
	// Empty BatchInsert
	batchInsertCmd := core.NewBatchInsertCommand("Order")
	_, err = defaultDialect.CompileBatchInsert(entity(), batchInsertCmd)
	assert.Error(t, err)
	
	// Empty BatchInsert with empty columns
	batchInsertCmdCol := core.NewBatchInsertCommand("Order")
	batchInsertCmdCol.BatchValues = append(batchInsertCmdCol.BatchValues, core.Record{"unknown": core.ValText("A")})
	_, err = defaultDialect.CompileBatchInsert(entity(), batchInsertCmdCol)
	assert.Error(t, err)
	
	// Empty Update
	updateCmd := core.NewUpdateCommand("Order", core.ValU64(1))
	_, err = defaultDialect.CompileUpdate(entity(), updateCmd)
	assert.Error(t, err)
	
	// Update with unknown field
	updateCmd2 := core.NewUpdateCommand("Order", core.ValU64(1)).Value("unknown", core.ValText("A"))
	_, err = defaultDialect.CompileUpdate(entity(), updateCmd2)
	assert.Error(t, err)
}

func TestDialectCoverage_CompileFunctions(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	funcs := []struct{
		fn core.ExprFunction
		name string
	}{
		{core.FuncCount, "COUNT"},
		{core.FuncSum, "SUM"},
		{core.FuncAvg, "AVG"},
		{core.FuncMin, "MIN"},
		{core.FuncMax, "MAX"},
		{core.FuncStddev, "STDDEV"},
		{core.FuncStddevPop, "STDDEV_POP"},
		{core.FuncVarSamp, "VAR_SAMP"},
		{core.FuncVarPop, "VAR_POP"},
		{core.FuncBitAnd, "BIT_AND"},
		{core.FuncBitOr, "BIT_OR"},
		{core.FuncBitXor, "BIT_XOR"},
	}
	
	for _, f := range funcs {
		q := core.NewSelectQuery("Order").WithFilter(
			core.ExprBinaryNode(core.ExprFunctionNode(f.fn, core.ExprColumnNode("id")), core.OpEq, core.ExprValueNode(core.ValU64(1))),
		)
		compiled, err := defaultDialect.CompileSelect(entity(), q)
		assert.NoError(t, err)
		assert.Contains(t, compiled.Sql, f.name)
	}
	
	// COUNT(*)
	qCountAll := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprCountAll(), core.OpEq, core.ExprValueNode(core.ValU64(1))),
	)
	compiledCountAll, err := defaultDialect.CompileSelect(entity(), qCountAll)
	assert.NoError(t, err)
	assert.Contains(t, compiledCountAll.Sql, "COUNT(*)")
	
	// SOUNDEX
	qSoundex := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprSoundex(core.ExprColumnNode("name")), core.OpEq, core.ExprValueNode(core.ValText("a"))),
	)
	compiledSoundex, err := defaultDialect.CompileSelect(entity(), qSoundex)
	assert.NoError(t, err)
	assert.Contains(t, compiledSoundex.Sql, "SOUNDEX(")
	
	// SOUNDEX invalid args
	qSoundexErr := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprFunctionNode(core.FuncSoundex, core.ExprColumnNode("name"), core.ExprColumnNode("id")), core.OpEq, core.ExprValueNode(core.ValText("a"))),
	)
	_, err = defaultDialect.CompileSelect(entity(), qSoundexErr)
	assert.Error(t, err)
	
	// GBK invalid args
	qGbkErr := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprFunctionNode(core.FuncGbk, core.ExprColumnNode("name"), core.ExprColumnNode("id")), core.OpEq, core.ExprValueNode(core.ValText("a"))),
	)
	_, err = defaultDialect.CompileSelect(entity(), qGbkErr)
	assert.Error(t, err)
	
	// Unknown function
	qUnknownFunc := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprFunctionNode(core.ExprFunction(999), core.ExprColumnNode("name")), core.OpEq, core.ExprValueNode(core.ValText("a"))),
	)
	_, err = defaultDialect.CompileSelect(entity(), qUnknownFunc)
	assert.Error(t, err)
}

func TestDialectCoverage_Select(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// compile joined OR
	qOr := core.NewSelectQuery("Order").WithFilter(
		core.ExprOrNode(
			core.ExprEq("name", core.ValText("A")),
			core.ExprEq("name", core.ValText("B")),
		),
	)
	compiledOr, err := defaultDialect.CompileSelect(entity(), qOr)
	assert.NoError(t, err)
	assert.Contains(t, compiledOr.Sql, " OR ")
	
	// compile joined AND (fallback)
	qAnd := core.NewSelectQuery("Order").WithFilter(
		core.ExprAndNode(
			core.ExprEq("name", core.ValText("A")),
			core.ExprEq("name", core.ValText("B")),
		),
	)
	compiledAnd, err := defaultDialect.CompileSelect(entity(), qAnd)
	assert.NoError(t, err)
	assert.Contains(t, compiledAnd.Sql, " AND ")
	
	// compile Not
	qNot := core.NewSelectQuery("Order").WithFilter(
		core.ExprNegate(core.ExprEq("name", core.ValText("A"))),
	)
	compiledNot, err := defaultDialect.CompileSelect(entity(), qNot)
	assert.NoError(t, err)
	assert.Contains(t, compiledNot.Sql, "(NOT")
	
	// select projections
	qProj := core.NewSelectQuery("Order").Project("id")
	qProj.ExprProjection = append(qProj.ExprProjection, core.NewNamedExpr("is_one", core.ExprEq("id", core.ValU64(1))))
	qProj.RawProjections = append(qProj.RawProjections, core.NewRawSqlProjection("one", "1"))
	qProj.DynamicProperties = append(qProj.DynamicProperties, core.NewRawSqlProjection("two", "2"))
	compiledProj, err := defaultDialect.CompileSelect(entity(), qProj)
	assert.NoError(t, err)
	assert.Contains(t, compiledProj.Sql, "is_one")
	assert.Contains(t, compiledProj.Sql, "one")
	assert.Contains(t, compiledProj.Sql, "two")
	
	// compile In list
	qIn := core.NewSelectQuery("Order").WithFilter(
		core.ExprInLarge("id", []core.Value{core.ValU64(1), core.ValU64(2)}),
	)
	compiledIn, err := defaultDialect.CompileSelect(entity(), qIn)
	assert.NoError(t, err)
	assert.Contains(t, compiledIn.Sql, "IN")
	
	// In Subquery
	subq := core.NewSelectQuery("OrderLine").Project("order_id").WithFilter(core.ExprEq("id", core.ValU64(1)))
	subqExpr := core.ExprNotInSubQuery("id", lineEntity(), subq, "order_id")
	subqQuery := core.NewSelectQuery("Order").WithFilter(subqExpr)
	compiledSubq, err := defaultDialect.CompileSelect(entity(), subqQuery)
	assert.NoError(t, err)
	assert.Contains(t, compiledSubq.Sql, "NOT IN")
	
	// invalid subquery op
	invalidSubq := core.ExprSubQueryNode(core.ExprColumnNode("id"), core.OpEq, lineEntity(), subq, "order_id")
	_, err = defaultDialect.CompileSelect(entity(), core.NewSelectQuery("Order").WithFilter(invalidSubq))
	assert.Error(t, err)
	
	// In with non-list right side
	qInNonList := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprColumnNode("id"), core.OpIn, core.ExprColumnNode("name")),
	)
	compiledInNonList, err := defaultDialect.CompileSelect(entity(), qInNonList)
	assert.NoError(t, err)
	assert.Contains(t, compiledInNonList.Sql, "IN")
	
	// Order by expr
	qOrderExpr := core.NewSelectQuery("Order").OrderExprAsc(core.ExprColumnNode("id"))
	compiledOrderExpr, err := defaultDialect.CompileSelect(entity(), qOrderExpr)
	assert.NoError(t, err)
	assert.Contains(t, compiledOrderExpr.Sql, "ORDER BY")
}

func TestDialectCoverage_SearchWithText(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	txt := "hello"
	q := core.NewSelectQuery("Order")
	q.SearchWithText = &txt
	q.RawSqlSearchCriteria = append(q.RawSqlSearchCriteria, "1=1")
	
	compiled, err := defaultDialect.CompileSelect(entity(), q)
	assert.NoError(t, err)
	assert.Contains(t, compiled.Sql, "LIKE")
	assert.Contains(t, compiled.Sql, "1=1")
}

func TestDialectCoverage_AggregateProjection(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	q := core.NewSelectQuery("Order").
		WithGroupBy("id").
		Project("id") // Group by and Project
	q.ExprProjection = append(q.ExprProjection, core.NewNamedExpr("is_one", core.ExprEq("id", core.ValU64(1))))
	q.RawProjections = append(q.RawProjections, core.NewRawSqlProjection("one", "1"))
	q.DynamicProperties = append(q.DynamicProperties, core.NewRawSqlProjection("two", "2"))
	q.Aggregates = append(q.Aggregates, &core.Aggregate{
		Function: core.AggSum,
		Field:    "version",
		Alias:    "sum_ver",
	})
	q.Aggregates = append(q.Aggregates, &core.Aggregate{
		Function: core.AggCount,
		Field:    "*",
		Alias:    "cnt",
	})
	
	aggFuncs := []core.AggregateFunction{core.AggAvg, core.AggMin, core.AggMax, core.AggStddev, core.AggStddevPop, core.AggVarSamp, core.AggVarPop, core.AggBitAnd, core.AggBitOr, core.AggBitXor}
	for _, af := range aggFuncs {
		q.Aggregates = append(q.Aggregates, &core.Aggregate{
			Function: af,
			Field:    "version",
			Alias:    "agg",
		})
	}
	
	compiled, err := defaultDialect.CompileSelect(entity(), q)
	assert.NoError(t, err)
	assert.Contains(t, compiled.Sql, "sum_ver")
	assert.Contains(t, compiled.Sql, "cnt")
}

func TestDialectCoverage_MissingBinaryOps(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	ops := []core.BinaryOp{core.OpNe, core.OpGt, core.OpGte, core.OpLt, core.OpLte, core.OpLike, core.OpNotLike}
	for _, op := range ops {
		q := core.NewSelectQuery("Order").WithFilter(core.ExprBinaryNode(core.ExprColumnNode("id"), op, core.ExprValueNode(core.ValU64(1))))
		_, err := defaultDialect.CompileSelect(entity(), q)
		assert.NoError(t, err)
	}
	
	// Between
	qBetween := core.NewSelectQuery("Order").WithFilter(core.ExprBetweenNode("id", core.ValU64(1), core.ValU64(2)))
	_, err := defaultDialect.CompileSelect(entity(), qBetween)
	assert.NoError(t, err)
	
	// IsNull
	qIsNull := core.NewSelectQuery("Order").WithFilter(core.ExprIsNullNode("id"))
	_, err = defaultDialect.CompileSelect(entity(), qIsNull)
	assert.NoError(t, err)
	
	// IsNotNull
	qIsNotNull := core.NewSelectQuery("Order").WithFilter(core.ExprIsNotNullNode("id"))
	_, err = defaultDialect.CompileSelect(entity(), qIsNotNull)
	assert.NoError(t, err)
}

func TestDialectCoverage_EdgeCases(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// CompileGbkFunction directly
	_, err := defaultDialect.CompileGbkFunction(entity(), []*core.Expr{core.ExprColumnNode("id")}, &[]core.Value{})
	assert.NoError(t, err)
	
	// CompileGbkFunction with multiple args error
	_, err = defaultDialect.CompileGbkFunction(entity(), []*core.Expr{core.ExprColumnNode("id"), core.ExprColumnNode("name")}, &[]core.Value{})
	assert.Error(t, err)
	
	// compileSingleArgFunction with 0 args
	qErr := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprFunctionNode(core.FuncSum), core.OpEq, core.ExprValueNode(core.ValU64(1))),
	)
	_, err = defaultDialect.CompileSelect(entity(), qErr)
	assert.Error(t, err)
	
	// compileSingleArgFunction with 2 args
	qErr2 := core.NewSelectQuery("Order").WithFilter(
		core.ExprBinaryNode(core.ExprFunctionNode(core.FuncSum, core.ExprColumnNode("id"), core.ExprColumnNode("id")), core.OpEq, core.ExprValueNode(core.ValU64(1))),
	)
	_, err = defaultDialect.CompileSelect(entity(), qErr2)
	assert.Error(t, err)
	
	// compileSubquery with invalid operator
	subq := core.NewSelectQuery("OrderLine").Project("order_id").WithFilter(core.ExprEq("id", core.ValU64(1)))
	invalidOpSubq := core.ExprSubQueryNode(core.ExprColumnNode("id"), core.OpGt, lineEntity(), subq, "order_id")
	_, err = defaultDialect.CompileSelect(entity(), core.NewSelectQuery("Order").WithFilter(invalidOpSubq))
	assert.Error(t, err)
	
	// Empty IN list
	emptyIn := core.NewSelectQuery("Order").WithFilter(core.ExprInList("id", []core.Value{}))
	_, err = defaultDialect.CompileSelect(entity(), emptyIn)
	assert.Error(t, err)
	
	// ResolveOrderField with field string instead of Expr
	qOrderField := core.NewSelectQuery("Order").OrderAsc("id")
	_, err = defaultDialect.CompileSelect(entity(), qOrderField)
	assert.NoError(t, err)
	
	// aggregateProjection with no aggregates
	qProj := core.NewSelectQuery("Order")
	qProj.Aggregates = []*core.Aggregate{}
	_, err = defaultDialect.CompileSelect(entity(), qProj)
	assert.NoError(t, err)
}

func TestDialectCoverage_UnsupportedSchemaType(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	_, err := defaultDialect.SchemaTypeSql(core.DataType(99), core.NewPropertyDescriptor("id", core.DataType(99)).ColumnName("id"))
	assert.Error(t, err)
}

func TestDialectCoverage_UnsupportedExprType(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	q := core.NewSelectQuery("Order").WithFilter(
		&core.Expr{Type: core.ExprType(99)},
	)
	_, err := defaultDialect.CompileSelect(entity(), q)
	assert.Error(t, err)
}

func TestDialectCoverage_UnknownFieldErrors(t *testing.T) {
	dialect := &TestDialect{}
	defaultDialect := &DefaultSqlDialect{Dialect: dialect}
	
	// projection unknown
	_, err := defaultDialect.CompileSelect(entity(), core.NewSelectQuery("Order").Project("unknown"))
	assert.Error(t, err)
	
	// groupby unknown
	_, err = defaultDialect.CompileSelect(entity(), core.NewSelectQuery("Order").WithGroupBy("unknown"))
	assert.Error(t, err)
	
	// orderby unknown
	_, err = defaultDialect.CompileSelect(entity(), core.NewSelectQuery("Order").OrderAsc("unknown"))
	assert.Error(t, err)
	
	// aggregate unknown field
	qAgg := core.NewSelectQuery("Order")
	qAgg.Aggregates = append(qAgg.Aggregates, &core.Aggregate{
		Function: core.AggSum,
		Field:    "unknown",
		Alias:    "u",
	})
	_, err = defaultDialect.CompileSelect(entity(), qAgg)
	assert.Error(t, err)
}

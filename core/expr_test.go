package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCompareColumnsBuildsPropertyToPropertyFilter(t *testing.T) {
	expr := ExprCompareColumns("updated_at", OpGte, "created_at")
	
	assert.Equal(t, ExprTypeBinary, expr.Type)
	assert.Equal(t, OpGte, expr.Op)
	assert.Equal(t, ExprTypeColumn, expr.Left.Type)
	assert.Equal(t, "updated_at", expr.Left.Column)
	assert.Equal(t, ExprTypeColumn, expr.Right.Type)
	assert.Equal(t, "created_at", expr.Right.Column)
}

func TestSoundLikeBuildsSoundexEquality(t *testing.T) {
	expr := ExprSoundLike("name", ValText("Robert"))
	
	assert.Equal(t, ExprTypeBinary, expr.Type)
	assert.Equal(t, OpEq, expr.Op)
	assert.Equal(t, ExprTypeFunctionCall, expr.Left.Type)
	assert.Equal(t, FuncSoundex, expr.Left.Function)
	assert.Equal(t, ExprTypeColumn, expr.Left.Args[0].Type)
	assert.Equal(t, "name", expr.Left.Args[0].Column)
}

func TestJavaStyleStringMatchBuildersExpandLikePatterns(t *testing.T) {
	contain := ExprContain("name", "tea")
	assert.Equal(t, OpLike, contain.Op)
	assert.Equal(t, "%tea%", contain.Right.Value.V)

	notContain := ExprNotContain("name", "tea")
	assert.Equal(t, OpNotLike, notContain.Op)
	assert.Equal(t, "%tea%", notContain.Right.Value.V)

	beginWith := ExprBeginWith("name", "tea")
	assert.Equal(t, OpLike, beginWith.Op)
	assert.Equal(t, "tea%", beginWith.Right.Value.V)

	notBeginWith := ExprNotBeginWith("name", "tea")
	assert.Equal(t, OpNotLike, notBeginWith.Op)
	assert.Equal(t, "tea%", notBeginWith.Right.Value.V)

	endWith := ExprEndWith("name", "tea")
	assert.Equal(t, OpLike, endWith.Op)
	assert.Equal(t, "%tea", endWith.Right.Value.V)

	notEndWith := ExprNotEndWith("name", "tea")
	assert.Equal(t, OpNotLike, notEndWith.Op)
	assert.Equal(t, "%tea", notEndWith.Right.Value.V)
}

func TestLargeInBuildersUseLargeBinaryOps(t *testing.T) {
	inLarge := ExprInLarge("id", []Value{ValU64(1)})
	assert.Equal(t, OpInLarge, inLarge.Op)
	
	notInLarge := ExprNotInLarge("id", []Value{ValU64(1)})
	assert.Equal(t, OpNotInLarge, notInLarge.Op)
}

func TestSubqueryBuilderProjectsRequestedField(t *testing.T) {
	query := &SelectQuery{}
	entity := NewEntityDescriptor("OrderLine")
	expr := ExprInSubQuery("id", entity, query, "order_id")

	assert.Equal(t, ExprTypeSubQuery, expr.Type)
	assert.Equal(t, OpIn, expr.Op)
	assert.Equal(t, "OrderLine", expr.Entity.Name)
	assert.Equal(t, []string{"order_id"}, expr.Query.Projection)
	assert.Equal(t, "id", expr.Left.Column)
}

func TestExprFunctions(t *testing.T) {
	col := ExprColumnNode("col")
	
	gbk := ExprGbk(col)
	assert.Equal(t, FuncGbk, gbk.Function)
	
	countAll := ExprCountAll()
	assert.Equal(t, FuncCount, countAll.Function)
	assert.Len(t, countAll.Args, 0)
	
	count := ExprCountExpr(col)
	assert.Equal(t, FuncCount, count.Function)
	assert.Equal(t, col, count.Args[0])
	
	sum := ExprSumExpr(col)
	assert.Equal(t, FuncSum, sum.Function)
	
	avg := ExprAvgExpr(col)
	assert.Equal(t, FuncAvg, avg.Function)
	
	min := ExprMinExpr(col)
	assert.Equal(t, FuncMin, min.Function)
	
	max := ExprMaxExpr(col)
	assert.Equal(t, FuncMax, max.Function)
	
	stddev := ExprStddevExpr(col)
	assert.Equal(t, FuncStddev, stddev.Function)
	
	stddevPop := ExprStddevPopExpr(col)
	assert.Equal(t, FuncStddevPop, stddevPop.Function)
	
	varSamp := ExprVarSampExpr(col)
	assert.Equal(t, FuncVarSamp, varSamp.Function)
	
	varPop := ExprVarPopExpr(col)
	assert.Equal(t, FuncVarPop, varPop.Function)
	
	bitAnd := ExprBitAndExpr(col)
	assert.Equal(t, FuncBitAnd, bitAnd.Function)
	
	bitOr := ExprBitOrExpr(col)
	assert.Equal(t, FuncBitOr, bitOr.Function)
	
	bitXor := ExprBitXorExpr(col)
	assert.Equal(t, FuncBitXor, bitXor.Function)
}

func TestExprComparisons(t *testing.T) {
	val := ValI64(1)
	
	ne := ExprNe("col", val)
	assert.Equal(t, OpNe, ne.Op)
	
	gte := ExprGte("col", val)
	assert.Equal(t, OpGte, gte.Op)
	
	lt := ExprLt("col", val)
	assert.Equal(t, OpLt, lt.Op)
	
	lte := ExprLte("col", val)
	assert.Equal(t, OpLte, lte.Op)
	
	inList := ExprInList("col", []Value{val})
	assert.Equal(t, OpIn, inList.Op)
	
	notInList := ExprNotInList("col", []Value{val})
	assert.Equal(t, OpNotIn, notInList.Op)
}

func TestExprSubqueries(t *testing.T) {
	query := &SelectQuery{}
	entity := NewEntityDescriptor("OrderLine")
	notInSub := ExprNotInSubQuery("id", entity, query, "order_id")
	
	assert.Equal(t, ExprTypeSubQuery, notInSub.Type)
	assert.Equal(t, OpNotIn, notInSub.Op)
}

func TestExprBetweenNullAndNot(t *testing.T) {
	between := ExprBetweenNode("col", ValI64(1), ValI64(10))
	assert.Equal(t, ExprTypeBetween, between.Type)
	
	isNull := ExprIsNullNode("col")
	assert.Equal(t, ExprTypeIsNull, isNull.Type)
	
	isNotNull := ExprIsNotNullNode("col")
	assert.Equal(t, ExprTypeIsNotNull, isNotNull.Type)
	
	negate := ExprNegate(isNull)
	assert.Equal(t, ExprTypeNot, negate.Type)
	assert.Equal(t, isNull, negate.Left)
}

func TestExprAndOr(t *testing.T) {
	e1 := ExprEq("c1", ValI64(1))
	e2 := ExprEq("c2", ValI64(2))
	e3 := ExprEq("c3", ValI64(3))
	
	andNode := ExprAndNode(e1, e2)
	assert.Equal(t, ExprTypeAnd, andNode.Type)
	
	orNode := ExprOrNode(e1, e2)
	assert.Equal(t, ExprTypeOr, orNode.Type)
	
	// Test AndExpr deduplication
	andNode2 := andNode.AndExpr(e3)
	assert.Len(t, andNode2.Parts, 3)
	
	// Duplicate And
	andNode3 := andNode2.AndExpr(e2)
	assert.Len(t, andNode3.Parts, 3)
	
	// Self And
	andSelf := e1.AndExpr(e1)
	assert.Equal(t, e1, andSelf)
	
	// Non-And node And
	andNew := e1.AndExpr(e2)
	assert.Equal(t, ExprTypeAnd, andNew.Type)
	
	// Test OrExpr deduplication
	orNode2 := orNode.OrExpr(e3)
	assert.Len(t, orNode2.Parts, 3)
	
	// Duplicate Or
	orNode3 := orNode2.OrExpr(e2)
	assert.Len(t, orNode3.Parts, 3)
	
	// Self Or
	orSelf := e1.OrExpr(e1)
	assert.Equal(t, e1, orSelf)
	
	// Non-Or node Or
	orNew := e1.OrExpr(e2)
	assert.Equal(t, ExprTypeOr, orNew.Type)
}

func TestExprsEqualDeduplication(t *testing.T) {
	e1 := ExprColumnNode("col")
	e2 := ExprColumnNode("col")
	e3 := ExprEq("col", ValI64(1))
	assert.True(t, exprsEqual(e1, e2))
	
	var nilExpr1 *Expr
	var nilExpr2 *Expr
	
	assert.True(t, exprsEqual(nilExpr1, nilExpr1))
	assert.True(t, exprsEqual(nilExpr1, nilExpr2))
	assert.False(t, exprsEqual(nilExpr1, e1))
	assert.False(t, exprsEqual(e1, nilExpr1))
	
	dedup := removeDuplicateExprs([]*Expr{e1, e1, e1, e2, e3})
	assert.Len(t, dedup, 2) // e1 and e2 are equal, e3 is different
}

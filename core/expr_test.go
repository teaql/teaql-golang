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

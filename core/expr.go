package core

import "fmt"

type BinaryOp int

const (
	OpEq BinaryOp = iota
	OpNe
	OpGt
	OpGte
	OpLt
	OpLte
	OpLike
	OpNotLike
	OpIn
	OpNotIn
	OpInLarge
	OpNotInLarge
)

type ExprFunction int

const (
	FuncSoundex ExprFunction = iota
	FuncGbk
	FuncCount
	FuncSum
	FuncAvg
	FuncMin
	FuncMax
	FuncStddev
	FuncStddevPop
	FuncVarSamp
	FuncVarPop
	FuncBitAnd
	FuncBitOr
	FuncBitXor
)

type ExprType int

const (
	ExprTypeColumn ExprType = iota
	ExprTypeValue
	ExprTypeFunctionCall
	ExprTypeBinary
	ExprTypeSubQuery
	ExprTypeBetween
	ExprTypeIsNull
	ExprTypeIsNotNull
	ExprTypeAnd
	ExprTypeOr
	ExprTypeNot
)

type Expr struct {
	Type     ExprType
	Column   string
	Value    Value
	Function ExprFunction
	Args     []*Expr
	Left     *Expr
	Op       BinaryOp
	Right    *Expr
	Entity   *EntityDescriptor
	// Using interface{} to avoid cyclic dependency before SelectQuery is fully defined,
	// or we can just define SelectQuery in query.go. Let's define a pointer to SelectQuery.
	Query *SelectQuery
	Lower *Expr
	Upper *Expr
	Parts []*Expr
}

func ExprColumnNode(name string) *Expr {
	return &Expr{Type: ExprTypeColumn, Column: name}
}

func ExprValueNode(value Value) *Expr {
	return &Expr{Type: ExprTypeValue, Value: value}
}

func ExprFunctionNode(function ExprFunction, args ...*Expr) *Expr {
	return &Expr{Type: ExprTypeFunctionCall, Function: function, Args: args}
}

func ExprSoundex(expr *Expr) *Expr {
	return ExprFunctionNode(FuncSoundex, expr)
}

func ExprGbk(expr *Expr) *Expr {
	return ExprFunctionNode(FuncGbk, expr)
}

func ExprCountAll() *Expr {
	return ExprFunctionNode(FuncCount)
}

func ExprCountExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncCount, expr)
}

func ExprSumExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncSum, expr)
}

func ExprAvgExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncAvg, expr)
}

func ExprMinExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncMin, expr)
}

func ExprMaxExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncMax, expr)
}

func ExprStddevExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncStddev, expr)
}

func ExprStddevPopExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncStddevPop, expr)
}

func ExprVarSampExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncVarSamp, expr)
}

func ExprVarPopExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncVarPop, expr)
}

func ExprBitAndExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncBitAnd, expr)
}

func ExprBitOrExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncBitOr, expr)
}

func ExprBitXorExpr(expr *Expr) *Expr {
	return ExprFunctionNode(FuncBitXor, expr)
}

func ExprBinaryNode(left *Expr, op BinaryOp, right *Expr) *Expr {
	return &Expr{Type: ExprTypeBinary, Left: left, Op: op, Right: right}
}

func ExprSoundLike(column string, value Value) *Expr {
	return ExprBinaryNode(
		ExprSoundex(ExprColumnNode(column)),
		OpEq,
		ExprSoundex(ExprValueNode(value)),
	)
}

func ExprEq(column string, value Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpEq, ExprValueNode(value))
}

func ExprNe(column string, value Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpNe, ExprValueNode(value))
}

func ExprGt(column string, value Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpGt, ExprValueNode(value))
}

func ExprGte(column string, value Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpGte, ExprValueNode(value))
}

func ExprLt(column string, value Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpLt, ExprValueNode(value))
}

func ExprLte(column string, value Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpLte, ExprValueNode(value))
}

func ExprLike(column string, pattern string) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpLike, ExprValueNode(ValText(pattern)))
}

func ExprNotLike(column string, pattern string) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpNotLike, ExprValueNode(ValText(pattern)))
}

func ExprContain(column string, value string) *Expr {
	return ExprLike(column, fmt.Sprintf("%%%s%%", value))
}

func ExprNotContain(column string, value string) *Expr {
	return ExprNotLike(column, fmt.Sprintf("%%%s%%", value))
}

func ExprBeginWith(column string, value string) *Expr {
	return ExprLike(column, fmt.Sprintf("%s%%", value))
}

func ExprNotBeginWith(column string, value string) *Expr {
	return ExprNotLike(column, fmt.Sprintf("%s%%", value))
}

func ExprEndWith(column string, value string) *Expr {
	return ExprLike(column, fmt.Sprintf("%%%s", value))
}

func ExprNotEndWith(column string, value string) *Expr {
	return ExprNotLike(column, fmt.Sprintf("%%%s", value))
}

func ExprCompareColumns(leftColumn string, op BinaryOp, rightColumn string) *Expr {
	return ExprBinaryNode(ExprColumnNode(leftColumn), op, ExprColumnNode(rightColumn))
}

func ExprInList(column string, values []Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpIn, ExprValueNode(ValList(values)))
}

func ExprNotInList(column string, values []Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpNotIn, ExprValueNode(ValList(values)))
}

func ExprInLarge(column string, values []Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpInLarge, ExprValueNode(ValList(values)))
}

func ExprNotInLarge(column string, values []Value) *Expr {
	return ExprBinaryNode(ExprColumnNode(column), OpNotInLarge, ExprValueNode(ValList(values)))
}

func ExprInSubQuery(column string, entity *EntityDescriptor, query *SelectQuery, field string) *Expr {
	return ExprSubQueryNode(ExprColumnNode(column), OpIn, entity, query, field)
}

func ExprNotInSubQuery(column string, entity *EntityDescriptor, query *SelectQuery, field string) *Expr {
	return ExprSubQueryNode(ExprColumnNode(column), OpNotIn, entity, query, field)
}

func ExprSubQueryNode(left *Expr, op BinaryOp, entity *EntityDescriptor, query *SelectQuery, field string) *Expr {
	query = query.Clone()
	if op == OpNotIn || op == OpNotInLarge {
		// NULL in a NOT IN projection makes every outer comparison UNKNOWN.
		// Ignore orphan relation keys here; they remain queryable via IsNull.
		query.AndFilter(ExprIsNotNullNode(field))
	}
	query.Projection = []string{field}
	return &Expr{
		Type:   ExprTypeSubQuery,
		Left:   left,
		Op:     op,
		Entity: entity,
		Query:  query,
	}
}

func ExprBetweenNode(column string, lower Value, upper Value) *Expr {
	return &Expr{
		Type:  ExprTypeBetween,
		Left:  ExprColumnNode(column),
		Lower: ExprValueNode(lower),
		Upper: ExprValueNode(upper),
	}
}

func ExprIsNullNode(column string) *Expr {
	return &Expr{Type: ExprTypeIsNull, Left: ExprColumnNode(column)}
}

func ExprIsNotNullNode(column string) *Expr {
	return &Expr{Type: ExprTypeIsNotNull, Left: ExprColumnNode(column)}
}

func ExprAndNode(parts ...*Expr) *Expr {
	return &Expr{Type: ExprTypeAnd, Parts: removeDuplicateExprs(parts)}
}

func ExprOrNode(parts ...*Expr) *Expr {
	return &Expr{Type: ExprTypeOr, Parts: removeDuplicateExprs(parts)}
}

func ExprNegate(expr *Expr) *Expr {
	return &Expr{Type: ExprTypeNot, Left: expr}
}

func (e *Expr) AndExpr(other *Expr) *Expr {
	if exprsEqual(e, other) {
		return e
	}
	if e.Type == ExprTypeAnd {
		for _, p := range e.Parts {
			if exprsEqual(p, other) {
				return e
			}
		}
		e.Parts = append(e.Parts, other)
		return e
	}
	return ExprAndNode(e, other)
}

func (e *Expr) OrExpr(other *Expr) *Expr {
	if exprsEqual(e, other) {
		return e
	}
	if e.Type == ExprTypeOr {
		for _, p := range e.Parts {
			if exprsEqual(p, other) {
				return e
			}
		}
		e.Parts = append(e.Parts, other)
		return e
	}
	return ExprOrNode(e, other)
}

func removeDuplicateExprs(parts []*Expr) []*Expr {
	unique := make([]*Expr, 0)
	for _, p := range parts {
		found := false
		for _, u := range unique {
			if exprsEqual(p, u) {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, p)
		}
	}
	return unique
}

func exprsEqual(a, b *Expr) bool {
	// A simple deep equality check for deduplication
	// In Go, deep equality of structs with pointers requires custom logic or reflect.DeepEqual
	// For now, we'll do pointer equality or a quick struct value compare for simplicity.
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// For thoroughness we could implement full deep equal, but for AST deduplication
	// simple pointer eq might not catch everything. We'll use fmt.Sprintf("%v") as a hacky hash for now.
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

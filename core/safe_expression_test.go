package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSafeExpressionEvalWithUsesTheSuppliedRoot(t *testing.T) {
	expr := NewSafeExpression(2, func(root int) (int, bool) {
		return root * 3, true
	})

	val, ok := expr.Eval()
	assert.True(t, ok)
	assert.Equal(t, 6, val)

	valWith, okWith := expr.EvalWith(4)
	assert.True(t, okWith)
	assert.Equal(t, 12, valWith)
}

func TestSafeExpressionApplyOptionalShortCircuitsRemainingMappers(t *testing.T) {
	optionalCalls := 0
	remainingCalls := 0

	expr := ApplySafeExpression(
		ApplyOptionalSafeExpression(ValueSafeExpression(5), func(val int) (int, bool) {
			optionalCalls++
			return 0, false
		}),
		func(val int) int {
			remainingCalls++
			return val * 2
		},
	)

	_, ok := expr.Eval()
	assert.False(t, ok)
	assert.Equal(t, 1, optionalCalls)
	assert.Equal(t, 0, remainingCalls)
}

func TestSafeExpressionLazyFallbackAndErrorOnlyRunForMissingValues(t *testing.T) {
	presentFallbackCalls := 0
	present := ValueSafeExpression(7)
	
	val := present.OrElseWith(func() int {
		presentFallbackCalls++
		return 9
	})
	assert.Equal(t, 7, val)
	assert.Equal(t, 0, presentFallbackCalls)
	
	val2, err := present.OrElseThrow(func() any {
		return "unused error"
	})
	assert.Nil(t, err) // Actually error is string here, let's just check value
	assert.Equal(t, 7, val2)

	missing := NewSafeExpression((any)(nil), func(root any) (int, bool) {
		return 0, false
	})
	missingFallbackCalls := 0
	
	mVal := missing.OrElseWith(func() int {
		missingFallbackCalls++
		return 9
	})
	assert.Equal(t, 9, mVal)
	assert.Equal(t, 1, missingFallbackCalls)

	_, mErr := missing.OrElseThrow(func() any {
		return "missing value"
	})
	assert.Equal(t, "missing value", mErr)
}

func TestSafeExpressionCallbacksOnlyRunForTheirMatchingBranch(t *testing.T) {
	present := ValueSafeExpression("teaql")
	presentNullCalls := 0
	var presentValue *string

	present.WhenIsNull(func() {
		presentNullCalls++
	})
	present.WhenIsNotNull(func(val string) {
		v := val
		presentValue = &v
	})

	assert.Equal(t, 0, presentNullCalls)
	assert.NotNil(t, presentValue)
	assert.Equal(t, "teaql", *presentValue)

	missing := NewSafeExpression((any)(nil), func(root any) (string, bool) {
		return "", false
	})
	missingNullCalls := 0
	missingValueCalls := 0

	missing.WhenIsNull(func() {
		missingNullCalls++
	})
	missing.WhenIsNotNull(func(val string) {
		missingValueCalls++
	})

	assert.Equal(t, 1, missingNullCalls)
	assert.Equal(t, 0, missingValueCalls)
}

func TestSafeExpressionOrElseAndIsNotNull(t *testing.T) {
	present := ValueSafeExpression(42)
	assert.Equal(t, 42, present.OrElse(0))
	assert.True(t, present.IsNotNull())
	assert.False(t, present.IsNull())

	missing := NewSafeExpression((any)(nil), func(root any) (int, bool) {
		return 0, false
	})
	assert.Equal(t, 0, missing.OrElse(0))
	assert.False(t, missing.IsNotNull())
	assert.True(t, missing.IsNull())
}

func TestApplySafeExpressionMissingBranch(t *testing.T) {
	missing := NewSafeExpression((any)(nil), func(root any) (int, bool) {
		return 0, false
	})

	applied := ApplySafeExpression(missing, func(val int) string {
		return "test"
	})
	val, ok := applied.Eval()
	assert.False(t, ok)
	assert.Equal(t, "", val)

	appliedOpt := ApplyOptionalSafeExpression(missing, func(val int) (string, bool) {
		return "test", true
	})
	valOpt, okOpt := appliedOpt.Eval()
	assert.False(t, okOpt)
	assert.Equal(t, "", valOpt)
}

func TestApplySafeExpressionSuccessBranch(t *testing.T) {
	present := ValueSafeExpression(42)
	applied := ApplySafeExpression(present, func(val int) string {
		return "success"
	})
	val, ok := applied.Eval()
	assert.True(t, ok)
	assert.Equal(t, "success", val)
}

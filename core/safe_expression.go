package core

type SafeExpression[R any, T any] struct {
	Root      R
	Evaluator func(R) (T, bool)
}

func NewSafeExpression[R any, T any](root R, evaluator func(R) (T, bool)) SafeExpression[R, T] {
	return SafeExpression[R, T]{
		Root:      root,
		Evaluator: evaluator,
	}
}

func ValueSafeExpression[R any](root R) SafeExpression[R, R] {
	return NewSafeExpression(root, func(r R) (R, bool) {
		return r, true
	})
}

func (s SafeExpression[R, T]) Eval() (T, bool) {
	return s.Evaluator(s.Root)
}

func (s SafeExpression[R, T]) EvalWith(root R) (T, bool) {
	return s.Evaluator(root)
}

func (s SafeExpression[R, T]) OrElse(defaultValue T) T {
	if val, ok := s.Eval(); ok {
		return val
	}
	return defaultValue
}

func (s SafeExpression[R, T]) OrElseWith(defaultFn func() T) T {
	if val, ok := s.Eval(); ok {
		return val
	}
	return defaultFn()
}

func (s SafeExpression[R, T]) OrElseThrow(errorFn func() any) (T, any) {
	if val, ok := s.Eval(); ok {
		return val, nil
	}
	var zero T
	return zero, errorFn()
}

func (s SafeExpression[R, T]) IsNull() bool {
	_, ok := s.Eval()
	return !ok
}

func (s SafeExpression[R, T]) IsNotNull() bool {
	_, ok := s.Eval()
	return ok
}

func (s SafeExpression[R, T]) WhenIsNull(fn func()) {
	if s.IsNull() {
		fn()
	}
}

func (s SafeExpression[R, T]) WhenIsNotNull(fn func(T)) {
	if val, ok := s.Eval(); ok {
		fn(val)
	}
}

// Package-level functions for operations that change type

func ApplySafeExpression[R any, T any, U any](s SafeExpression[R, T], mapper func(T) U) SafeExpression[R, U] {
	return NewSafeExpression(s.Root, func(r R) (U, bool) {
		val, ok := s.Evaluator(r)
		if !ok {
			var zero U
			return zero, false
		}
		return mapper(val), true
	})
}

func ApplyOptionalSafeExpression[R any, T any, U any](s SafeExpression[R, T], mapper func(T) (U, bool)) SafeExpression[R, U] {
	return NewSafeExpression(s.Root, func(r R) (U, bool) {
		val, ok := s.Evaluator(r)
		if !ok {
			var zero U
			return zero, false
		}
		return mapper(val)
	})
}

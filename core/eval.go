package core

import "fmt"

type LoadStateType int

const (
	LoadStateNotLoaded LoadStateType = iota
	LoadStatePartial
	LoadStateFullyLoaded
)

type LoadState struct {
	Type   LoadStateType
	Fields map[string]bool
}

func NewLoadStateNotLoaded() *LoadState {
	return &LoadState{Type: LoadStateNotLoaded}
}

func NewLoadStateFullyLoaded() *LoadState {
	return &LoadState{Type: LoadStateFullyLoaded}
}

func NewLoadStatePartial(fields []string) *LoadState {
	set := make(map[string]bool)
	for _, f := range fields {
		set[f] = true
	}
	return &LoadState{Type: LoadStatePartial, Fields: set}
}

func (l *LoadState) IsLoaded(fieldOrRelation string) bool {
	switch l.Type {
	case LoadStateNotLoaded:
		return false
	case LoadStateFullyLoaded:
		return true
	case LoadStatePartial:
		return l.Fields[fieldOrRelation]
	}
	return false
}

type EvalResultType int

const (
	EvalResultValue EvalResultType = iota
	EvalResultNull
	EvalResultNotLoaded
)

type EvalResult[T any] struct {
	Type          EvalResultType
	Value         T
	FailedNode    string
	AttemptedPath string
}

func EvalValue[T any](value T) *EvalResult[T] {
	return &EvalResult[T]{Type: EvalResultValue, Value: value}
}

func EvalNull[T any]() *EvalResult[T] {
	return &EvalResult[T]{Type: EvalResultNull}
}

func EvalNotLoaded[T any](failedNode, attemptedPath string) *EvalResult[T] {
	return &EvalResult[T]{
		Type:          EvalResultNotLoaded,
		FailedNode:    failedNode,
		AttemptedPath: attemptedPath,
	}
}

func EvalAndThen[T, U any](res *EvalResult[T], fieldName string, f func(T) *EvalResult[U]) *EvalResult[U] {
	switch res.Type {
	case EvalResultValue:
		nextRes := f(res.Value)
		if nextRes.Type == EvalResultNotLoaded {
			newPath := nextRes.AttemptedPath
			if nextRes.AttemptedPath != fieldName && nextRes.AttemptedPath != "" {
				newPath = fmt.Sprintf("%s.%s", fieldName, nextRes.AttemptedPath)
			} else if nextRes.AttemptedPath == "" {
				newPath = fieldName
			}
			return EvalNotLoaded[U](nextRes.FailedNode, newPath)
		}
		return nextRes
	case EvalResultNull:
		return EvalNull[U]()
	case EvalResultNotLoaded:
		return EvalNotLoaded[U](res.FailedNode, res.AttemptedPath)
	}
	return EvalNull[U]()
}

func EvalMap[T, U any](res *EvalResult[T], f func(T) U) *EvalResult[U] {
	switch res.Type {
	case EvalResultValue:
		return EvalValue(f(res.Value))
	case EvalResultNull:
		return EvalNull[U]()
	case EvalResultNotLoaded:
		return EvalNotLoaded[U](res.FailedNode, res.AttemptedPath)
	}
	return EvalNull[U]()
}

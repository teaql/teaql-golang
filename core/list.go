package core

type Record map[string]Value

type SmartList[T any] struct {
	Data         []T
	TotalCount   *uint64
	Aggregations Record
	Summary      Record
	Facets       map[string]*SmartList[Record]
	IsLoaded     bool
}

func EmptySmartList[T any]() *SmartList[T] {
	return &SmartList[T]{
		Data:         make([]T, 0),
		Aggregations: make(Record),
		Summary:      make(Record),
		Facets:       make(map[string]*SmartList[Record]),
		IsLoaded:     false,
	}
}

func NewSmartList[T any](data []T) *SmartList[T] {
	return &SmartList[T]{
		Data:         data,
		Aggregations: make(Record),
		Summary:      make(Record),
		Facets:       make(map[string]*SmartList[Record]),
		IsLoaded:     true,
	}
}

func (s *SmartList[T]) WithTotalCount(count uint64) *SmartList[T] {
	s.TotalCount = &count
	return s
}

func (s *SmartList[T]) WithAggregation(key string, value Value) *SmartList[T] {
	s.Aggregations[key] = value
	return s
}

func (s *SmartList[T]) WithSummary(key string, value Value) *SmartList[T] {
	s.Summary[key] = value
	return s
}

func (s *SmartList[T]) WithFacet(key string, facet *SmartList[Record]) *SmartList[T] {
	s.Facets[key] = facet
	return s
}

func (s *SmartList[T]) AddFacet(key string, facet *SmartList[Record]) {
	s.Facets[key] = facet
}

func (s *SmartList[T]) Push(item T) {
	s.Data = append(s.Data, item)
}

func (s *SmartList[T]) Extend(items []T) {
	s.Data = append(s.Data, items...)
}

func (s *SmartList[T]) Len() int {
	return len(s.Data)
}

func (s *SmartList[T]) MergeBy(incoming []T, keyFn func(T) any) {
	positions := make(map[any]int)
	for i, item := range s.Data {
		positions[keyFn(item)] = i
	}
	
	for _, item := range incoming {
		k := keyFn(item)
		if idx, ok := positions[k]; ok {
			s.Data[idx] = item
		} else {
			positions[k] = len(s.Data)
			s.Data = append(s.Data, item)
		}
	}
}

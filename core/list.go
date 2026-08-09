package core

import "fmt"

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

func (s *SmartList[T]) FacetsMap() map[string]*SmartList[Record] {
	return s.Facets
}

func (s *SmartList[T]) Facet(key string) (*SmartList[Record], bool) {
	facet, ok := s.Facets[key]
	return facet, ok
}

func (s *SmartList[T]) RemoveFacet(key string) {
	delete(s.Facets, key)
}

func (s *SmartList[T]) TakeFacets() map[string]*SmartList[Record] {
	facets := s.Facets
	s.Facets = make(map[string]*SmartList[Record])
	return facets
}

func (s *SmartList[T]) Set(index int, value T) {
	if index >= 0 && index < len(s.Data) {
		s.Data[index] = value
	}
}

func (s *SmartList[T]) Get(index int) (T, bool) {
	if index >= 0 && index < len(s.Data) {
		return s.Data[index], true
	}
	var zero T
	return zero, false
}

func (s *SmartList[T]) Last() (T, bool) {
	if len(s.Data) > 0 {
		return s.Data[len(s.Data)-1], true
	}
	var zero T
	return zero, false
}

func (s *SmartList[T]) IsEmpty() bool {
	return len(s.Data) == 0
}

func (s *SmartList[T]) First() (T, bool) {
	if len(s.Data) > 0 {
		return s.Data[0], true
	}
	var zero T
	return zero, false
}

func (s *SmartList[T]) Retain(filter func(T) bool) {
	var newData []T
	for _, item := range s.Data {
		if filter(item) {
			newData = append(newData, item)
		}
	}
	s.Data = newData
}

func (s *SmartList[T]) TotalCountOrLen() uint64 {
	if s.TotalCount != nil {
		return *s.TotalCount
	}
	return uint64(len(s.Data))
}

func (s *SmartList[T]) Aggregation(key string) (Value, bool) {
	val, ok := s.Aggregations[key]
	return val, ok
}

func (s *SmartList[T]) SummaryValue(key string) (Value, bool) {
	val, ok := s.Summary[key]
	return val, ok
}

func (s *SmartList[T]) IntoVec() []T {
	return s.Data
}

func MapSmartList[T, U any](s *SmartList[T], mapper func(T) U) *SmartList[U] {
	data := make([]U, len(s.Data))
	for i, v := range s.Data {
		data[i] = mapper(v)
	}
	return &SmartList[U]{
		Data:         data,
		TotalCount:   s.TotalCount,
		Aggregations: s.Aggregations,
		Summary:      s.Summary,
		Facets:       s.Facets,
		IsLoaded:     s.IsLoaded,
	}
}

func ToList[T, U any](s *SmartList[T], mapper func(T) U) []U {
	data := make([]U, len(s.Data))
	for i, v := range s.Data {
		data[i] = mapper(v)
	}
	return data
}

func ToSet[T any, U comparable](s *SmartList[T], mapper func(T) U) map[U]struct{} {
	set := make(map[U]struct{})
	for _, v := range s.Data {
		set[mapper(v)] = struct{}{}
	}
	return set
}

func IdentityMap[T any, K comparable](s *SmartList[T], keyFn func(T) K) map[K]T {
	result := make(map[K]T)
	for _, item := range s.Data {
		result[keyFn(item)] = item
	}
	return result
}

func GroupBy[T any, K comparable](s *SmartList[T], keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, item := range s.Data {
		k := keyFn(item)
		groups[k] = append(groups[k], item)
	}
	return groups
}

func IntoRecords[T Entity](s *SmartList[T]) *SmartList[Record] {
	data := make([]Record, len(s.Data))
	for i, v := range s.Data {
		data[i] = v.IntoRecord()
	}
	return &SmartList[Record]{
		Data:         data,
		TotalCount:   s.TotalCount,
		Aggregations: s.Aggregations,
		Summary:      s.Summary,
		Facets:       s.Facets,
		IsLoaded:     s.IsLoaded,
	}
}

func Ids[T IdentifiableEntity](s *SmartList[T]) []Value {
	ids := make([]Value, len(s.Data))
	for i, v := range s.Data {
		ids[i] = v.IdValue()
	}
	return ids
}

func IdKey(value Value) string {
	if value.V == nil {
		return "null"
	}
	switch v := value.V.(type) {
	case bool:
		return fmt.Sprintf("b:%v", v)
	case int64:
		return fmt.Sprintf("i:%v", v)
	case uint64:
		return fmt.Sprintf("u:%v", v)
	case float64:
		return fmt.Sprintf("f:%v", v)
	case string:
		return fmt.Sprintf("t:%v", v)
	case DataType:
		return "null"
	default:
		return fmt.Sprintf("object:%v", v)
	}
}

func MapById[T IdentifiableEntity](s *SmartList[T]) map[string]T {
	result := make(map[string]T)
	for _, v := range s.Data {
		result[IdKey(v.IdValue())] = v
	}
	return result
}

func Versions[T VersionedEntity](s *SmartList[T]) []int64 {
	versions := make([]int64, len(s.Data))
	for i, v := range s.Data {
		versions[i] = v.Version()
	}
	return versions
}


package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestItem struct {
	Id    uint64
	Value string
}

func TestSmartListMergeByReplacesInPlaceAndAppendsNewKeys(t *testing.T) {
	items := NewSmartList([]TestItem{
		{Id: 1, Value: "one"},
		{Id: 2, Value: "old two"},
		{Id: 4, Value: "four"},
	})

	items.MergeBy([]TestItem{
		{Id: 2, Value: "new two"},
		{Id: 3, Value: "three"},
	}, func(item TestItem) any {
		return item.Id
	})

	assert.Equal(t, []TestItem{
		{Id: 1, Value: "one"},
		{Id: 2, Value: "new two"},
		{Id: 4, Value: "four"},
		{Id: 3, Value: "three"},
	}, items.Data)
}

func TestSmartListMergeByKeepsOnePositionForRepeatedIncomingKeys(t *testing.T) {
	items := NewSmartList([]TestItem{
		{Id: 1, Value: "one"},
	})

	items.MergeBy([]TestItem{
		{Id: 2, Value: "first two"},
		{Id: 2, Value: "final two"},
	}, func(item TestItem) any {
		return item.Id
	})

	assert.Equal(t, []TestItem{
		{Id: 1, Value: "one"},
		{Id: 2, Value: "final two"},
	}, items.Data)
}

func TestSmartListAdditionalMethods(t *testing.T) {
	items := NewSmartList([]TestItem{
		{Id: 1, Value: "one"},
	})
	
	items.WithTotalCount(100)
	assert.Equal(t, uint64(100), *items.TotalCount)
	
	items.WithAggregation("sum", ValI64(50))
	assert.Equal(t, ValI64(50), items.Aggregations["sum"])
	
	items.WithSummary("avg", ValI64(5))
	assert.Equal(t, ValI64(5), items.Summary["avg"])
	
	facetList := NewSmartList([]Record{})
	items.WithFacet("tags", facetList)
	assert.Equal(t, facetList, items.Facets["tags"])
	
	items.Push(TestItem{Id: 2, Value: "two"})
	assert.Equal(t, 2, items.Len())
	
	items.Extend([]TestItem{
		{Id: 3, Value: "three"},
		{Id: 4, Value: "four"},
	})
	assert.Equal(t, 4, items.Len())
	assert.Equal(t, "four", items.Data[3].Value)
}

type DummyListEntity struct {
	id      Value
	version int64
	record  Record
}


func (d *DummyListEntity) IntoRecord() Record {
	return d.record
}

func (d *DummyListEntity) IdValue() Value {
	return d.id
}

func (d *DummyListEntity) Version() int64 {
	return d.version
}

func (d *DummyListEntity) DirtyFields() []string {
	return nil
}

func (d *DummyListEntity) EntityDescriptor() *EntityDescriptor {
	return nil
}

func (d *DummyListEntity) EntityName() string {
	return "DummyListEntity"
}

func (d *DummyListEntity) FromRecord(r Record) error {
	return nil
}

func (d *DummyListEntity) IsMarkedAsDelete() bool {
	return false
}

func (d *DummyListEntity) IsNew() bool {
	return false
}

func (d *DummyListEntity) MarkAsNew() {
}

func (d *DummyListEntity) GetComment() *string {
	return nil
}

func (d *DummyListEntity) SetComment(comment string) {
}

func (d *DummyListEntity) OriginalValues() Record {
	return d.record
}

func (d *DummyListEntity) OnLoaded(context any) {
}

func (d *DummyListEntity) IntoJson() any {
	return nil
}

func TestSmartList_Facets(t *testing.T) {
	list := EmptySmartList[int]()
	
	facetList := EmptySmartList[Record]()
	list.AddFacet("f1", facetList)

	if len(list.FacetsMap()) != 1 {
		t.Errorf("Expected 1 facet")
	}

	f, ok := list.Facet("f1")
	if !ok || f != facetList {
		t.Errorf("Facet not found or mismatch")
	}

	_, ok = list.Facet("f2")
	if ok {
		t.Errorf("Expected false for missing facet")
	}

	list.RemoveFacet("f1")
	if len(list.FacetsMap()) != 0 {
		t.Errorf("Expected 0 facets")
	}

	list.WithFacet("f2", facetList)
	facets := list.TakeFacets()
	if len(facets) != 1 {
		t.Errorf("Expected 1 taken facet")
	}
	if len(list.FacetsMap()) != 0 {
		t.Errorf("Expected list facets to be cleared")
	}
}

func TestSmartList_BasicOps(t *testing.T) {
	list := NewSmartList([]int{1, 2, 3})
	
	if list.IsEmpty() {
		t.Errorf("Expected not empty")
	}

	list.Set(1, 10)
	val, ok := list.Get(1)
	if !ok || val != 10 {
		t.Errorf("Get failed")
	}

	_, ok = list.Get(100)
	if ok {
		t.Errorf("Expected false")
	}

	last, ok := list.Last()
	if !ok || last != 3 {
		t.Errorf("Last failed")
	}

	first, ok := list.First()
	if !ok || first != 1 {
		t.Errorf("First failed")
	}

	emptyList := EmptySmartList[int]()
	if !emptyList.IsEmpty() {
		t.Errorf("Expected empty")
	}
	_, ok = emptyList.First()
	if ok {
		t.Errorf("Expected false")
	}
	_, ok = emptyList.Last()
	if ok {
		t.Errorf("Expected false")
	}

	list.Retain(func(v int) bool { return v > 2 })
	if list.Len() != 2 || list.Data[0] != 10 || list.Data[1] != 3 {
		t.Errorf("Retain failed: expected [10, 3], got %v", list.Data)
	}
}

func TestSmartList_TotalCount(t *testing.T) {
	list := NewSmartList([]int{1, 2, 3})
	if list.TotalCountOrLen() != 3 {
		t.Errorf("TotalCountOrLen expected 3")
	}

	list.WithTotalCount(100)
	if list.TotalCountOrLen() != 100 {
		t.Errorf("TotalCountOrLen expected 100")
	}
}

func TestSmartList_SummaryAggregation(t *testing.T) {
	list := EmptySmartList[int]()
	list.WithAggregation("agg1", ValI64(10))
	list.WithSummary("sum1", ValI64(20))

	val, ok := list.Aggregation("agg1")
	if !ok || val.V != int64(10) {
		t.Errorf("Aggregation failed")
	}
	_, ok = list.Aggregation("agg2")
	if ok {
		t.Errorf("Expected false")
	}

	val, ok = list.SummaryValue("sum1")
	if !ok || val.V != int64(20) {
		t.Errorf("SummaryValue failed")
	}
	_, ok = list.SummaryValue("sum2")
	if ok {
		t.Errorf("Expected false")
	}
}

func TestSmartList_IntoVec(t *testing.T) {
	list := NewSmartList([]int{1, 2})
	vec := list.IntoVec()
	if len(vec) != 2 {
		t.Errorf("IntoVec failed")
	}
}

func TestSmartList_MapAndTransform(t *testing.T) {
	list := NewSmartList([]int{1, 2, 3})
	
	mapped := MapSmartList(list, func(v int) int64 { return int64(v) * 2 })
	if len(mapped.Data) != 3 || mapped.Data[0] != 2 {
		t.Errorf("MapSmartList failed")
	}

	slice := ToList(list, func(v int) int64 { return int64(v) * 2 })
	if len(slice) != 3 || slice[0] != 2 {
		t.Errorf("ToList failed")
	}

	set := ToSet(list, func(v int) int { return v * 2 })
	if _, ok := set[2]; !ok {
		t.Errorf("ToSet failed")
	}

	idMap := IdentityMap(list, func(v int) int { return v })
	if val, ok := idMap[2]; !ok || val != 2 {
		t.Errorf("IdentityMap failed")
	}

	groups := GroupBy(list, func(v int) int { return v % 2 })
	if len(groups[1]) != 2 {
		t.Errorf("GroupBy failed")
	}
}

func TestSmartList_Entities(t *testing.T) {
	e1 := &DummyListEntity{id: ValI64(1), version: 10, record: Record{"a": ValI64(1)}}
	e2 := &DummyListEntity{id: ValI64(2), version: 20, record: Record{"a": ValI64(2)}}

	list := NewSmartList([]*DummyListEntity{e1, e2})
	
	records := IntoRecords(list)
	if len(records.Data) != 2 {
		t.Errorf("IntoRecords failed")
	}

	ids := Ids(list)
	if len(ids) != 2 || ids[0].V != int64(1) {
		t.Errorf("Ids failed")
	}

	mapped := MapById(list)
	if _, ok := mapped["i:1"]; !ok {
		t.Errorf("MapById failed")
	}

	versions := Versions(list)
	if len(versions) != 2 || versions[0] != int64(10) {
		t.Errorf("Versions failed")
	}
}

func TestIdKey(t *testing.T) {
	if IdKey(Value{V: nil}) != "null" {
		t.Errorf("IdKey nil failed")
	}
	if IdKey(Value{V: true}) != "b:true" {
		t.Errorf("IdKey bool failed")
	}
	if IdKey(Value{V: int64(1)}) != "i:1" {
		t.Errorf("IdKey i64 failed")
	}
	if IdKey(Value{V: uint64(1)}) != "u:1" {
		t.Errorf("IdKey u64 failed")
	}
	if IdKey(Value{V: float64(1.5)}) != "f:1.5" {
		t.Errorf("IdKey float failed")
	}
	if IdKey(Value{V: "test"}) != "t:test" {
		t.Errorf("IdKey string failed")
	}
	if IdKey(Value{V: TypeText}) != "null" {
		t.Errorf("IdKey DataType failed")
	}
	if IdKey(Value{V: []int{}}) != "object:[]" {
		t.Errorf("IdKey object failed")
	}
}

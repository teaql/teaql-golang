package core

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

type keyedListRow struct {
	Key   string
	Value int
}
type listFixture struct {
	States []struct {
		Name   string
		Loaded bool
		Rows   []int
		Total  *uint64
	}
	Transform struct {
		Rows, MapTimesTen, FilterOdd, FlatMapTwice []int
		Total                                      uint64
	}
	Merge    struct{ Target, Incoming, Expected []keyedListRow }
	Identity struct {
		Rows []struct {
			ID      any
			Version int64
		}
		Versions         []int64
		IndexedPositions []int
	}
	Metadata struct{ Decimal struct{ Value string } }
}

func listVectors(t *testing.T) listFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/smart-list-semantics.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture listFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestSmartListSharedStatesAndDerivedStatistics(t *testing.T) {
	fixture := listVectors(t)
	for _, state := range fixture.States {
		t.Run(state.Name, func(t *testing.T) {
			list := NewSmartList(state.Rows)
			if !state.Loaded {
				list = EmptySmartList[int]()
			}
			list.TotalCount = state.Total
			if list.IsLoaded != state.Loaded || !reflect.DeepEqual(list.Data, state.Rows) {
				t.Fatal("state mismatch", list)
			}
			expected := uint64(list.Len())
			if state.Total != nil {
				expected = *state.Total
			}
			if list.TotalCountOrLen() != expected || !reflect.DeepEqual(list.TotalCount, state.Total) {
				t.Fatal("total mismatch")
			}
		})
	}
	source := NewSmartList(fixture.Transform.Rows).WithTotalCount(fixture.Transform.Total).
		WithAggregation("source", ValI64(5)).WithSummary("source", ValBool(true)).WithFacet("empty", NewSmartList([]Record{}))
	mapped := MapSmartList(source, func(v int) int { return v * 10 })
	filtered := FilterSmartList(source, func(v int) bool { return v%2 == 1 })
	expanded := FlatMapSmartList(source, func(v int) []int { return []int{v, v} })
	if !reflect.DeepEqual(mapped.Data, fixture.Transform.MapTimesTen) || !reflect.DeepEqual(filtered.Data, fixture.Transform.FilterOdd) || !reflect.DeepEqual(expanded.Data, fixture.Transform.FlatMapTwice) {
		t.Fatal("derived rows differ")
	}
	for _, result := range []*SmartList[int]{filtered, expanded, FilterSmartList(source, func(v int) bool { return true })} {
		if !result.IsLoaded || result.TotalCount != nil || len(result.Aggregations)+len(result.Summary)+len(result.Facets) != 0 {
			t.Fatal("derived statistics retained")
		}
	}
	if MapSmartList(EmptySmartList[int](), func(v int) int { return v }).IsLoaded || FilterSmartList(EmptySmartList[int](), func(v int) bool { return true }).IsLoaded || FlatMapSmartList(EmptySmartList[int](), func(v int) []int { return []int{v} }).IsLoaded {
		t.Fatal("unloaded state was manufactured")
	}
}

func TestSmartListMapOwnsNestedMetadataFacetRecordsAndCountPointer(t *testing.T) {
	value := decimal.RequireFromString(listVectors(t).Metadata.Decimal.Value)
	source := NewSmartList([]int{1, 2}).WithTotalCount(9).
		WithAggregation("nested", ValJson(map[string]any{"values": []any{1, 2}})).WithAggregation("decimal", ValDecimal(value)).WithAggregation("nullable", ValNull()).
		WithSummary("nested", ValJson(map[string]any{"values": []any{3}})).
		WithFacet("status", NewSmartList([]Record{{"count": ValI64(3)}})).WithFacet("empty", NewSmartList([]Record{}))
	mapped := MapSmartList(source, func(v int) int { return v * 10 })
	if !reflect.DeepEqual(source.Aggregations, mapped.Aggregations) || !reflect.DeepEqual(source.Summary, mapped.Summary) || !reflect.DeepEqual(source.Facets, mapped.Facets) {
		t.Fatal("metadata differs")
	}
	*mapped.TotalCount = 99
	mapped.Aggregations["nested"].V.(map[string]any)["values"].([]any)[0] = 99
	mapped.Summary["nested"].V.(map[string]any)["values"].([]any)[0] = 99
	mapped.Facets["status"].Data[0]["count"] = ValI64(99)
	mapped.RemoveFacet("empty")
	if *source.TotalCount != 9 || source.Aggregations["nested"].V.(map[string]any)["values"].([]any)[0] != 1 || source.Summary["nested"].V.(map[string]any)["values"].([]any)[0] != 3 || source.Facets["status"].Data[0]["count"] != ValI64(3) {
		t.Fatal("mapped mutation changed source")
	}
	if _, found := source.Facet("empty"); !found {
		t.Fatal("empty facet removed from source")
	}
	if _, found := source.Aggregation("missing"); found {
		t.Fatal("missing alias invented")
	}
	if mapped.Aggregations["decimal"].V.(decimal.Decimal).String() != value.String() {
		t.Fatal("decimal changed")
	}
	native := source.IntoVec()
	native[0] = 99
	if source.Data[0] != 1 {
		t.Fatal("native copy aliases source positions")
	}
}

func TestSmartListDuplicateMergeOrderAndSafeAccess(t *testing.T) {
	fixture := listVectors(t)
	source := NewSmartList(fixture.Merge.Target).WithTotalCount(9)
	source.MergeBy(fixture.Merge.Incoming, func(v keyedListRow) any { return v.Key })
	if !reflect.DeepEqual(source.Data, fixture.Merge.Expected) {
		t.Fatal(source.Data)
	}
	if IdentityMap(source, func(v keyedListRow) string { return v.Key })["a"].Value != 4 {
		t.Fatal("last duplicate must win")
	}
	if !reflect.DeepEqual(GroupBy(source, func(v keyedListRow) string { return v.Key })["a"], fixture.Merge.Expected[:2]) {
		t.Fatal("group order")
	}
	for _, index := range []int{-1, source.Len(), 100} {
		if _, found := source.Get(index); found {
			t.Fatal("invalid safe index")
		}
		source.Set(index, keyedListRow{})
	}
	if !reflect.DeepEqual(source.Data, fixture.Merge.Expected) || *source.TotalCount != 9 {
		t.Fatal("invalid set changed source")
	}
	if _, found := EmptySmartList[int]().First(); found {
		t.Fatal("empty first")
	}
	if _, found := EmptySmartList[int]().Last(); found {
		t.Fatal("empty last")
	}
	facets := NewSmartList([]int{}).WithFacet("empty", NewSmartList([]Record{}))
	facets.RemoveFacet("absent")
	if len(facets.TakeFacets()) != 1 || len(facets.Facets) != 0 {
		t.Fatal("facet transfer")
	}
}

func TestSmartListMissingIDsAndVersionsFromSharedVectors(t *testing.T) {
	fixture := listVectors(t)
	data := make([]*DummyListEntity, len(fixture.Identity.Rows))
	for index, row := range fixture.Identity.Rows {
		id := Value{V: row.ID}
		if number, ok := row.ID.(float64); ok {
			id = ValI64(int64(number))
		}
		data[index] = &DummyListEntity{id: id, version: row.Version}
	}
	source := NewSmartList(data)
	indexed := MapById(source)
	if len(indexed) != 4 {
		t.Fatal(indexed)
	}
	if _, found := indexed["null"]; found {
		t.Fatal("missing ID indexed")
	}
	for index, key := range []string{"i:0", "i:1", "t:1", "t:"} {
		if indexed[key] != data[fixture.Identity.IndexedPositions[index]] {
			t.Fatal("ID association")
		}
	}
	if !reflect.DeepEqual(Versions(source), fixture.Identity.Versions) || Ids(source)[1].V != nil {
		t.Fatal("identity extraction")
	}
}

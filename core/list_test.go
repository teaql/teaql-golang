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

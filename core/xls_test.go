package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestXlsBlockContextMatchesJavaNavigationModel(t *testing.T) {
	context := NewXlsBlockBuildContext("orders", 2, 3)
	
	header := context.ToBlock("Order No").
		AddProperty("bold", true).
		Span(2, 1)

	next := context.Next().ToBlock("Amount")
	nextLine := context.NextLine().ToBlock("SO-1")
	newLine := context.NewLine().ToBlock("reset-left")

	assert.Equal(t, "orders", header.Page)
	assert.Equal(t, int32(2), header.Left)
	assert.Equal(t, int32(3), header.Top)
	assert.Equal(t, int32(3), header.Right)
	assert.Equal(t, int32(3), header.Bottom)
	assert.Equal(t, int32(2), header.Width())
	assert.Equal(t, int32(1), header.Height())
	assert.Equal(t, "Order No", header.Value)
	assert.Equal(t, true, header.Properties["bold"])
	assert.True(t, header.Contains(2, 3))
	assert.True(t, header.Contains(3, 3))
	assert.False(t, header.Contains(4, 3))

	assert.Equal(t, int32(3), next.Left)
	assert.Equal(t, int32(3), next.Top)
	assert.Equal(t, "Amount", next.Value)

	assert.Equal(t, int32(2), nextLine.Left)
	assert.Equal(t, int32(4), nextLine.Top)

	assert.Equal(t, int32(0), newLine.Left)
	assert.Equal(t, int32(4), newLine.Top)

	page := NewXlsPage("orders").
		AddBlock(header).
		AddBlock(next).
		AddBlock(nextLine)

	assert.Equal(t, header, page.BlockAt(3, 3))
	assert.Equal(t, nextLine, page.BlockAt(2, 4))
	assert.Nil(t, page.BlockAt(0, 0))

	workbook := NewXlsWorkbook().AddPage(page)
	assert.NotNil(t, workbook.Page("orders"))
	assert.Nil(t, workbook.Page("missing"))

	jsonObj := workbook.ToJsonValue()
	pages := jsonObj["pages"].([]any)
	assert.Equal(t, 1, len(pages))
	pageJson := pages[0].(map[string]any)
	assert.Equal(t, "orders", pageJson["name"])
	blocks := pageJson["blocks"].([]any)
	assert.Equal(t, 3, len(blocks))
	
	b0 := blocks[0].(map[string]any)
	assert.Equal(t, "Order No", b0["value"])
	props := b0["properties"].(map[string]any)
	assert.Equal(t, true, props["bold"])
}

package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

func TestXlsAdditional(t *testing.T) {
	b := NewXlsBlock("page1", 10, 20, "val")
	b.Region(1, 2, 3, 4)
	assert.Equal(t, int32(1), b.Left)
	assert.Equal(t, int32(2), b.Top)
	assert.Equal(t, int32(3), b.Right)
	assert.Equal(t, int32(4), b.Bottom)

	// test span with negative w, h
	b.Span(0, 0)
	assert.Equal(t, int32(1), b.Right)
	assert.Equal(t, int32(2), b.Bottom)

	b.WithValue("newVal")
	assert.Equal(t, "newVal", b.Value)

	b.SetProperty("prop1", "pval")
	assert.Equal(t, "pval", b.Properties["prop1"])

	styleBlock := NewXlsBlock("page1", 0, 0, nil)
	b.Style(styleBlock)
	assert.Equal(t, styleBlock, b.StyleReferBlock)

	bJson := b.ToJsonValue()
	assert.Equal(t, "newVal", bJson["value"])

	context := NewXlsBlockBuildContext("page2", -1, -2)
	assert.Equal(t, int32(0), context.X)
	assert.Equal(t, int32(0), context.Y)

	ctxPage := XlsBlockBuildContextPage("page3")
	assert.Equal(t, "page3", ctxPage.Page)
	assert.Equal(t, int32(0), ctxPage.X)

	page := NewXlsPage("p1")
	page.PushBlock(b)
	assert.Equal(t, 1, len(page.Blocks))

	wb := NewXlsWorkbook()
	wb.PushPage(page)
	assert.Equal(t, 1, len(wb.Pages))
}

package core

import "encoding/json"

type XlsBlock struct {
	Page            string         `json:"page"`
	Top             int32          `json:"top"`
	Bottom          int32          `json:"bottom"`
	Left            int32          `json:"left"`
	Right           int32          `json:"right"`
	StyleReferBlock *XlsBlock      `json:"styleReferBlock,omitempty"`
	Value           any            `json:"value,omitempty"`
	Properties      map[string]any `json:"properties,omitempty"`
}

func NewXlsBlock(page string, x, y int32, value any) *XlsBlock {
	return &XlsBlock{
		Page:       page,
		Top:        y,
		Bottom:     y,
		Left:       x,
		Right:      x,
		Value:      value,
		Properties: make(map[string]any),
	}
}

func XlsBlockFromContext(context *XlsBlockBuildContext, value any) *XlsBlock {
	return NewXlsBlock(context.Page, context.X, context.Y, value)
}

func (x *XlsBlock) Region(left, top, right, bottom int32) *XlsBlock {
	x.Left = left
	x.Top = top
	x.Right = right
	x.Bottom = bottom
	return x
}

func (x *XlsBlock) Span(width, height int32) *XlsBlock {
	w := width - 1
	if w < 0 {
		w = 0
	}
	h := height - 1
	if h < 0 {
		h = 0
	}
	x.Right = x.Left + w
	x.Bottom = x.Top + h
	return x
}

func (x *XlsBlock) WithValue(value any) *XlsBlock {
	x.Value = value
	return x
}

func (x *XlsBlock) AddProperty(name string, value any) *XlsBlock {
	x.Properties[name] = value
	return x
}

func (x *XlsBlock) SetProperty(name string, value any) {
	x.Properties[name] = value
}

func (x *XlsBlock) Style(style *XlsBlock) *XlsBlock {
	x.StyleReferBlock = style
	return x
}

func (x *XlsBlock) Width() int32 {
	return x.Right - x.Left + 1
}

func (x *XlsBlock) Height() int32 {
	return x.Bottom - x.Top + 1
}

func (x *XlsBlock) Contains(posX, posY int32) bool {
	return posX >= x.Left && posX <= x.Right && posY >= x.Top && posY <= x.Bottom
}

func (x *XlsBlock) ToJsonValue() map[string]any {
	b, _ := json.Marshal(x)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

type XlsBlockBuildContext struct {
	Page   string `json:"page"`
	StartX int32  `json:"startX"`
	X      int32  `json:"x"`
	Y      int32  `json:"y"`
}

func NewXlsBlockBuildContext(page string, x, y int32) *XlsBlockBuildContext {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return &XlsBlockBuildContext{
		Page:   page,
		StartX: x,
		X:      x,
		Y:      y,
	}
}

func XlsBlockBuildContextPage(page string) *XlsBlockBuildContext {
	return NewXlsBlockBuildContext(page, 0, 0)
}

func (x *XlsBlockBuildContext) Next() *XlsBlockBuildContext {
	return &XlsBlockBuildContext{
		Page:   x.Page,
		StartX: x.StartX,
		X:      x.X + 1,
		Y:      x.Y,
	}
}

func (x *XlsBlockBuildContext) NewLine() *XlsBlockBuildContext {
	return &XlsBlockBuildContext{
		Page:   x.Page,
		StartX: x.StartX,
		X:      0,
		Y:      x.Y + 1,
	}
}

func (x *XlsBlockBuildContext) NextLine() *XlsBlockBuildContext {
	return &XlsBlockBuildContext{
		Page:   x.Page,
		StartX: x.StartX,
		X:      x.StartX,
		Y:      x.Y + 1,
	}
}

func (x *XlsBlockBuildContext) ToBlock(value any) *XlsBlock {
	return XlsBlockFromContext(x, value)
}

type XlsPage struct {
	Name   string      `json:"name"`
	Blocks []*XlsBlock `json:"blocks"`
}

func NewXlsPage(name string) *XlsPage {
	return &XlsPage{
		Name:   name,
		Blocks: make([]*XlsBlock, 0),
	}
}

func (p *XlsPage) AddBlock(block *XlsBlock) *XlsPage {
	p.Blocks = append(p.Blocks, block)
	return p
}

func (p *XlsPage) PushBlock(block *XlsBlock) {
	p.Blocks = append(p.Blocks, block)
}

func (p *XlsPage) BlockAt(x, y int32) *XlsBlock {
	for _, block := range p.Blocks {
		if block.Contains(x, y) {
			return block
		}
	}
	return nil
}

type XlsWorkbook struct {
	Pages []*XlsPage `json:"pages"`
}

func NewXlsWorkbook() *XlsWorkbook {
	return &XlsWorkbook{
		Pages: make([]*XlsPage, 0),
	}
}

func (w *XlsWorkbook) AddPage(page *XlsPage) *XlsWorkbook {
	w.Pages = append(w.Pages, page)
	return w
}

func (w *XlsWorkbook) PushPage(page *XlsPage) {
	w.Pages = append(w.Pages, page)
}

func (w *XlsWorkbook) Page(name string) *XlsPage {
	for _, page := range w.Pages {
		if page.Name == name {
			return page
		}
	}
	return nil
}

func (w *XlsWorkbook) ToJsonValue() map[string]any {
	b, _ := json.Marshal(w)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

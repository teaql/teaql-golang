package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/shopspring/decimal"
)

func TestWebResponseConstructorsAndJsonShape(t *testing.T) {
	successResp := SuccessWebResponse()
	assert.Equal(t, int32(0), successResp.ResultCode)
	assert.Equal(t, "YES", *successResp.Status)
	assert.Nil(t, successResp.Message)
	assert.Equal(t, WebResponseVersion, successResp.Version)
	assert.Equal(t, 0, len(successResp.Data))
	assert.Equal(t, uint64(0), successResp.RecordCount)

	failResp := FailWebResponse("Internal Error")
	assert.Equal(t, int32(1), failResp.ResultCode)
	assert.Equal(t, "NO", *failResp.Status)
	assert.Equal(t, "Internal Error", *failResp.Message)
	assert.Equal(t, 0, len(failResp.Data))
	assert.Equal(t, uint64(0), failResp.RecordCount)

	emptyListResp := EmptyListWebResponse("No items found")
	assert.Equal(t, int32(0), emptyListResp.ResultCode)
	assert.Nil(t, emptyListResp.Status)
	assert.Equal(t, "No items found", *emptyListResp.Message)
	assert.Equal(t, 0, len(emptyListResp.Data))
	assert.Equal(t, uint64(0), emptyListResp.RecordCount)

	jsonObj := successResp.ToJsonValue()
	assert.Equal(t, float64(0), jsonObj["resultCode"])
	assert.Equal(t, "YES", jsonObj["status"])
	_, ok := jsonObj["message"]
	assert.False(t, ok) // skipped because omitempty
	assert.Equal(t, WebResponseVersion, jsonObj["version"])
	assert.Equal(t, float64(0), jsonObj["recordCount"])
	assert.Empty(t, jsonObj["data"])
	assert.Empty(t, jsonObj["facets"])
}

func TestWebStyleAndActionBindFrontendMetadata(t *testing.T) {
	base := NewBaseEntityData()
	WebStyleWithBackgroundColor("#ffeecc").
		WithFontColor("#111111").
		BindBase(base)

	ViewWebAction().BindBase(base)
	ModifyWebAction("EDIT", "/orders/1/edit").BindBase(base)

	styleVal, ok := base.GetDynamic(StyleKey)
	assert.True(t, ok)
	styleJson := ValueToJson(styleVal).(map[string]any)
	assert.Equal(t, "#ffeecc", styleJson["backgroundColor"])
	assert.Equal(t, "#111111", styleJson["color"])

	actionVal, ok := base.GetDynamic(ActionListKey)
	assert.True(t, ok)
	actionList := ValueToJson(actionVal).([]any)
	assert.Equal(t, 2, len(actionList))
	
	action0 := actionList[0].(map[string]any)
	assert.Equal(t, "switchview", action0["execute"])
	assert.Equal(t, "detail", action0["target"])

	action1 := actionList[1].(map[string]any)
	assert.Equal(t, "EDIT", action1["name"])
	assert.Equal(t, "/orders/1/edit", action1["requestURL"])
}

func TestWebStyleAdditional(t *testing.T) {
	ws := WebStyleWithFontColor("red").WithClassNames("class1")
	assert.Equal(t, "red", *ws.Color)
	assert.Equal(t, "class1", *ws.ClassNames)

	ws2 := WebStyleWithClassNames("class2")
	assert.Equal(t, "class2", *ws2.ClassNames)

	r := make(Record)
	ws.BindRecord(r)
	assert.NotNil(t, r[StyleKey])
}

func TestWebActionAdditional(t *testing.T) {
	wa := DefaultModifyWebAction().
		WithComponent("comp").
		WithWarningMessage("warn").
		WithRoleForList("role")
	
	assert.Equal(t, "comp", *wa.Component)
	assert.Equal(t, "warn", *wa.WarningMessage)
	assert.Equal(t, "role", *wa.RoleForList)

	r := make(Record)
	wa.BindRecord(r)
	assert.NotNil(t, r[ActionListKey])
	
	// Test AppendActionToDynamic with existing non-list json value
	r2 := make(Record)
	r2[ActionListKey] = ValJson([]any{map[string]any{"a": 1}})
	wa.BindRecord(r2)
	assert.NotNil(t, r2[ActionListKey])
	
	// Test AppendActionToDynamic with existing string json value
	r3 := make(Record)
	r3[ActionListKey] = ValJson("some string")
	wa.BindRecord(r3)
	assert.NotNil(t, r3[ActionListKey])

	// Test AppendActionToDynamic with existing Value list
	r4 := make(Record)
	r4[ActionListKey] = ValList([]Value{ValJson(map[string]any{"a": 1})})
	wa.BindRecord(r4)
	assert.NotNil(t, r4[ActionListKey])

	// Test AppendActionToDynamic with existing non-json value
	r5 := make(Record)
	r5[ActionListKey] = ValI64(123)
	wa.BindRecord(r5)
	assert.NotNil(t, r5[ActionListKey])

	// Test AppendActionToDynamic with existing null value
	r6 := make(Record)
	r6[ActionListKey] = ValNull()
	wa.BindRecord(r6)
	assert.NotNil(t, r6[ActionListKey])
}

func TestWebResponseAdditional(t *testing.T) {
	wr := SuccessWebResponse().WithData([]any{"test"})
	assert.Equal(t, 1, len(wr.Data))
	assert.Equal(t, uint64(1), wr.RecordCount)

	entity := &webMockEntity{BaseEntityData: NewBaseEntityData()}
	resp := WebResponseFromEntity(entity)
	assert.Equal(t, 1, len(resp.Data))
}

type webMockEntity struct {
	*BaseEntityData
}

func (m *webMockEntity) EntityName() string { return "Mock" }
func (m *webMockEntity) EntityDescriptor() *EntityDescriptor { return nil }
func (m *webMockEntity) FromRecord(record Record) error { return nil }
func (m *webMockEntity) IntoRecord() Record {
	r := make(Record)
	r["id"] = ValU64(m.Id)
	return r
}
func (m *webMockEntity) DirtyFields() []string { return nil }
func (m *webMockEntity) IsMarkedAsDelete() bool { return false }
func (m *webMockEntity) IsNew() bool { return false }
func (m *webMockEntity) MarkAsNew() {}
func (m *webMockEntity) GetComment() *string { return nil }
func (m *webMockEntity) SetComment(comment string) {}
func (m *webMockEntity) OriginalValues() Record { return nil }
func (m *webMockEntity) OnLoaded(context any) {}
func (m *webMockEntity) IntoJson() any { return nil }

func TestValueToJson(t *testing.T) {
	v1 := ValueToJson(ValI64(1))
	assert.Equal(t, int64(1), v1)

	v2 := ValueToJson(ValText("test"))
	assert.Equal(t, "test", v2)

	v3 := ValueToJson(ValDecimal(decimal.NewFromInt(10)))
	assert.Equal(t, "10", v3)

	v4 := ValueToJson(ValList([]Value{ValI64(1), ValI64(2)}))
	assert.Equal(t, 2, len(v4.([]any)))

	r := make(Record)
	r["key"] = ValI64(3)
	v5 := ValueToJson(ValObject(r))
	m := v5.(map[string]any)
	assert.Equal(t, int64(3), m["key"])
    
    // test nil value
    assert.Nil(t, ValueToJson(Value{}))
}


package core

import (
	"testing"
	"github.com/stretchr/testify/assert"
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

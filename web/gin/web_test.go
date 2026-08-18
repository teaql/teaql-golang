package gin

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebResponse(t *testing.T) {
	succ := Success()
	assert.Equal(t, 0, succ.ResultCode)
	assert.Equal(t, "YES", succ.Status)

	fail := Fail("error message")
	assert.Equal(t, 1, fail.ResultCode)
	assert.Equal(t, "NO", fail.Status)
	assert.Equal(t, "error message", *fail.Message)

	list := OfList([]interface{}{"a", "b"})
	assert.Equal(t, 2, list.RecordCount)
	assert.Equal(t, []interface{}{"a", "b"}, list.Data)

	single := OfSingle("a")
	assert.Equal(t, 1, single.RecordCount)
	assert.Equal(t, []interface{}{"a"}, single.Data)
}

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorHandler(c, errors.New("some error"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mod := runtime.NewRuntimeModule()
	middleware := GinMiddleware(mod)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// mock request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-Id", "user123")
	req.Header.Set("User-Agent", "TestAgent")
	c.Request = req

	middleware(c)

	context, ok := GetTeaContext(c)
	assert.True(t, ok)
	assert.NotNil(t, context)

	reqInfo := context.GetResource("WebRequestInfo").(*WebRequestInfo)
	assert.NotNil(t, reqInfo)
	assert.Equal(t, "TestAgent", *reqInfo.UserAgent)
	assert.Equal(t, "user123", context.GetResource("UserId"))

	// Test GetTeaContext failure
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx2, ok2 := GetTeaContext(c2)
	assert.False(t, ok2)
	assert.Nil(t, ctx2)
}

func TestRecordToJsonMap(t *testing.T) {
	r := core.Record{
		"a": core.ValI64(1),
		"b": core.ValText("text"),
	}
	m := RecordToJsonMap(r)
	assert.Equal(t, int64(1), m["a"])
	assert.Equal(t, "text", m["b"])
}

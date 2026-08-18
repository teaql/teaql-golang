package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

// WebResponse a unified response format
type WebResponse struct {
	Data        []interface{}          `json:"data"`
	ResultCode  int                    `json:"resultCode"`
	Status      string                 `json:"status"`
	Message     *string                `json:"message,omitempty"`
	RecordCount int                    `json:"recordCount"`
	Facets      map[string]interface{} `json:"facets,omitempty"`
	Version     string                 `json:"version"`
	TraceId     *string                `json:"traceId,omitempty"`
}

func Success() *WebResponse {
	return &WebResponse{
		Data:        []interface{}{},
		ResultCode:  0,
		Status:      "YES",
		RecordCount: 0,
		Version:     "1.001",
	}
}

func Fail(message string) *WebResponse {
	return &WebResponse{
		Data:        []interface{}{},
		ResultCode:  1,
		Status:      "NO",
		Message:     &message,
		RecordCount: 0,
		Version:     "1.001",
	}
}

func OfList(list []interface{}) *WebResponse {
	res := Success()
	res.Data = list
	res.RecordCount = len(list)
	return res
}

func OfSingle(entity interface{}) *WebResponse {
	res := Success()
	res.Data = []interface{}{entity}
	res.RecordCount = 1
	return res
}

func ErrorHandler(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Fail(err.Error()))
}

type WebRequestInfo struct {
	ClientIp   *string
	UserAgent  *string
	RequestUri string
	Method     string
}

func GinMiddleware(module *runtime.RuntimeModule) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := module.IntoContext()

		clientIp := c.ClientIP()
		userAgent := c.Request.UserAgent()

		info := &WebRequestInfo{
			ClientIp:   &clientIp,
			UserAgent:  &userAgent,
			RequestUri: c.Request.RequestURI,
			Method:     c.Request.Method,
		}

		context.InsertResource("WebRequestInfo", info)

		userId := c.GetHeader("X-User-Id")
		if userId != "" {
			// context.SetUserIdentifier(userId)
			context.InsertResource("UserId", userId)
		}

		c.Set("TeaContext", context)

		c.Next()
	}
}

func GetTeaContext(c *gin.Context) (*runtime.UserContext, bool) {
	val, ok := c.Get("TeaContext")
	if !ok {
		return nil, false
	}
	context, ok := val.(*runtime.UserContext)
	return context, ok
}

// Convert core.Record to standard map
func RecordToJsonMap(r core.Record) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range r {
		m[k] = v.V
	}
	return m
}

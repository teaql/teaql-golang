package core

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

const (
	StyleKey           = "style"
	ActionListKey      = "actionList"
	WebResponseVersion = "1.001"
)

type WebStyle struct {
	BackgroundColor *string `json:"backgroundColor,omitempty"`
	Color           *string `json:"color,omitempty"`
	ClassNames      *string `json:"classNames,omitempty"`
}

func NewWebStyle() *WebStyle {
	return &WebStyle{}
}

func WebStyleWithBackgroundColor(color string) *WebStyle {
	return NewWebStyle().WithBackgroundColor(color)
}

func WebStyleWithFontColor(color string) *WebStyle {
	return NewWebStyle().WithFontColor(color)
}

func WebStyleWithClassNames(classNames string) *WebStyle {
	return NewWebStyle().WithClassNames(classNames)
}

func (w *WebStyle) WithBackgroundColor(color string) *WebStyle {
	w.BackgroundColor = &color
	return w
}

func (w *WebStyle) WithFontColor(color string) *WebStyle {
	w.Color = &color
	return w
}

func (w *WebStyle) WithClassNames(classNames string) *WebStyle {
	w.ClassNames = &classNames
	return w
}

func (w *WebStyle) ToJsonValue() map[string]any {
	b, _ := json.Marshal(w)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

func (w *WebStyle) BindBase(entity *BaseEntityData) {
	entity.PutDynamic(StyleKey, ValJson(w.ToJsonValue()))
}

func (w *WebStyle) BindRecord(record Record) {
	record[StyleKey] = ValJson(w.ToJsonValue())
}

type WebAction struct {
	Key            *string `json:"key,omitempty"`
	Name           *string `json:"name,omitempty"`
	Level          *string `json:"level,omitempty"`
	Execute        *string `json:"execute,omitempty"`
	Target         *string `json:"target,omitempty"`
	Component      *string `json:"component,omitempty"`
	WarningMessage *string `json:"warningMessage,omitempty"`
	RoleForList    *string `json:"roleForList,omitempty"`
	RequestURL     *string `json:"requestURL,omitempty"`
}

func NewWebAction() *WebAction {
	return &WebAction{}
}

func strPtr(s string) *string {
	return &s
}

func ViewWebAction() *WebAction {
	return NewWebAction().
		WithName("VIEW DETAIL").
		WithLevel("view").
		WithExecute("switchview").
		WithTarget("detail")
}

func ModifyWebAction(name string, url string) *WebAction {
	return NewWebAction().
		WithName(name).
		WithKey(name).
		WithLevel("modify").
		WithExecute("switchview").
		WithTarget("modify").
		WithRequestURL(url)
}

func DefaultModifyWebAction() *WebAction {
	return NewWebAction().
		WithName("UPDATE").
		WithLevel("modify").
		WithExecute("switchview").
		WithTarget("modify")
}

func (w *WebAction) WithKey(key string) *WebAction {
	w.Key = &key
	return w
}

func (w *WebAction) WithName(name string) *WebAction {
	w.Name = &name
	return w
}

func (w *WebAction) WithLevel(level string) *WebAction {
	w.Level = &level
	return w
}

func (w *WebAction) WithExecute(execute string) *WebAction {
	w.Execute = &execute
	return w
}

func (w *WebAction) WithTarget(target string) *WebAction {
	w.Target = &target
	return w
}

func (w *WebAction) WithComponent(component string) *WebAction {
	w.Component = &component
	return w
}

func (w *WebAction) WithWarningMessage(warningMessage string) *WebAction {
	w.WarningMessage = &warningMessage
	return w
}

func (w *WebAction) WithRoleForList(roleForList string) *WebAction {
	w.RoleForList = &roleForList
	return w
}

func (w *WebAction) WithRequestURL(requestURL string) *WebAction {
	w.RequestURL = &requestURL
	return w
}

func (w *WebAction) ToJsonValue() map[string]any {
	b, _ := json.Marshal(w)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

func (w *WebAction) BindBase(entity *BaseEntityData) {
	AppendActionToDynamic(&entity.Dynamic, w.ToJsonValue())
}

func (w *WebAction) BindRecord(record Record) {
	AppendActionToDynamic(&record, w.ToJsonValue())
}

func AppendActionToDynamic(dynamic *Record, action map[string]any) {
	if val, ok := (*dynamic)[ActionListKey]; ok {
		if list, isList := val.TryList(); isList {
			list = append(list, ValJson(action))
			(*dynamic)[ActionListKey] = ValList(list)
			return
		}
		
		var list []Value
		if jVal, isJson := val.TryJson(); isJson {
			if arr, ok := jVal.([]any); ok {
				for _, item := range arr {
					list = append(list, ValJson(item))
				}
			} else {
				list = append(list, val)
			}
		} else {
			list = append(list, val)
		}
		list = append(list, ValJson(action))
		(*dynamic)[ActionListKey] = ValList(list)
	} else {
		(*dynamic)[ActionListKey] = ValList([]Value{ValJson(action)})
	}
}

type WebResponse struct {
	Data        []any                     `json:"data"`
	ResultCode  int32                     `json:"resultCode"`
	Status      *string                   `json:"status,omitempty"`
	Message     *string                   `json:"message,omitempty"`
	RecordCount uint64                    `json:"recordCount"`
	Version     string                    `json:"version"`
	Facets      map[string]any            `json:"facets"`
}

func SuccessWebResponse() *WebResponse {
	return &WebResponse{
		Data:        make([]any, 0),
		ResultCode:  0,
		Status:      strPtr("YES"),
		Message:     nil,
		RecordCount: 0,
		Version:     WebResponseVersion,
		Facets:      make(map[string]any),
	}
}

func FailWebResponse(message string) *WebResponse {
	return &WebResponse{
		Data:        make([]any, 0),
		ResultCode:  1,
		Status:      strPtr("NO"),
		Message:     &message,
		RecordCount: 0,
		Version:     WebResponseVersion,
		Facets:      make(map[string]any),
	}
}

func EmptyListWebResponse(message string) *WebResponse {
	return &WebResponse{
		Data:        make([]any, 0),
		ResultCode:  0,
		Status:      nil,
		Message:     &message,
		RecordCount: 0,
		Version:     WebResponseVersion,
		Facets:      make(map[string]any),
	}
}

func (w *WebResponse) ToJsonValue() map[string]any {
	b, _ := json.Marshal(w)
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

func (w *WebResponse) WithData(data []any) *WebResponse {
	w.Data = data
	w.RecordCount = uint64(len(data))
	return w
}

// Convert record to JSON recursively handles Value -> any mapping
func RecordToJson(record Record) map[string]any {
	m := make(map[string]any)
	for k, v := range record {
		m[k] = ValueToJson(v)
	}
	return m
}

func ValueToJson(v Value) any {
	if v.V == nil {
		return nil
	}
	switch val := v.V.(type) {
	case bool, int64, uint64, float64, string:
		return val
	case decimal.Decimal:
		return val.String()
	case []Value:
		var arr []any
		for _, item := range val {
			arr = append(arr, ValueToJson(item))
		}
		return arr
	case Record:
		return RecordToJson(val)
	default:
		return val
	}
}

func WebResponseFromEntity(entity Entity) *WebResponse {
	record := entity.IntoRecord()
	return SuccessWebResponse().WithData([]any{RecordToJson(record)})
}

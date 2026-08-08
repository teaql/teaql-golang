package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TrimmedFields struct {
	Required TrimmedString  `json:"required"`
	Optional *TrimmedString `json:"optional"`
}

func NewTrimmedString(s string) *TrimmedString {
	ts := TrimmedString(s)
	return &ts
}

func TestTrimmedStringHelpersTrimDuringSerialization(t *testing.T) {
	fields := TrimmedFields{
		Required: "  required value\n",
		Optional: NewTrimmedString("\toptional value  "),
	}

	b, err := json.Marshal(fields)
	assert.NoError(t, err)

	var m map[string]any
	json.Unmarshal(b, &m)

	assert.Equal(t, "required value", m["required"])
	assert.Equal(t, "optional value", m["optional"])
}

func TestTrimmedOptionalStringPreservesNoneDuringSerialization(t *testing.T) {
	fields := TrimmedFields{
		Required: " value ",
		Optional: nil,
	}

	b, err := json.Marshal(fields)
	assert.NoError(t, err)

	var m map[string]any
	json.Unmarshal(b, &m)

	assert.Equal(t, "value", m["required"])
	assert.Nil(t, m["optional"])
}

func TestTrimmedStringHelpersTrimDuringDeserialization(t *testing.T) {
	js := `{"required": "  required value\n", "optional": "\toptional value  "}`

	var fields TrimmedFields
	err := json.Unmarshal([]byte(js), &fields)
	assert.NoError(t, err)

	assert.Equal(t, TrimmedString("required value"), fields.Required)
	assert.Equal(t, NewTrimmedString("optional value"), fields.Optional)
}

func TestTrimmedOptionalStringKeepsWhitespaceOnlyInputAsSomeEmpty(t *testing.T) {
	js := `{"required": " value ", "optional": " \t\n "}`

	var fields TrimmedFields
	err := json.Unmarshal([]byte(js), &fields)
	assert.NoError(t, err)

	assert.Equal(t, TrimmedString("value"), fields.Required)
	assert.Equal(t, NewTrimmedString(""), fields.Optional)
}

func TestTrimmedStringUnmarshalError(t *testing.T) {
	var ts TrimmedString
	err := ts.UnmarshalJSON([]byte(`{"invalid":`)) // Invalid JSON string
	assert.Error(t, err)
}

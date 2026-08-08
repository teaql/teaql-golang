package core

import (
	"encoding/json"
	"strings"
)

// TrimmedString is a string type that automatically trims whitespace
// during JSON serialization and deserialization.
type TrimmedString string

// MarshalJSON implements json.Marshaler
func (t TrimmedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.TrimSpace(string(t)))
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TrimmedString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = TrimmedString(strings.TrimSpace(s))
	return nil
}

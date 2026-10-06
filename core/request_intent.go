package core

import (
	"fmt"
	"strings"
	"unicode"
)

// RequestIntentError names the request boundary without echoing caller text or
// payload. Logging controls never affect this validation.
type RequestIntentError struct {
	Code        string
	Field       string
	RequestKind string
}

func (e *RequestIntentError) Error() string {
	return fmt.Sprintf("%s: %s request requires a non-blank %s; supply it on this request", e.Code, e.RequestKind, e.Field)
}

func requiredIntentText(text *string, kind, field string) (string, error) {
	if text == nil || strings.TrimFunc(*text, unicode.IsSpace) == "" {
		code := "REQUEST_COMMENT_REQUIRED"
		if field == "purpose" {
			code = "QUERY_PURPOSE_REQUIRED"
		}
		return "", &RequestIntentError{Code: code, Field: field, RequestKind: kind}
	}
	return *text, nil
}

// QueryIntent is an immutable captured value. Its zero value is invalid.
type QueryIntent struct{ comment, purpose string }

func NewQueryIntent(comment, purpose *string) (QueryIntent, error) {
	what, err := requiredIntentText(comment, "query", "comment")
	if err != nil {
		return QueryIntent{}, err
	}
	why, err := requiredIntentText(purpose, "query", "purpose")
	if err != nil {
		return QueryIntent{}, err
	}
	return QueryIntent{comment: what, purpose: why}, nil
}

func (i QueryIntent) Comment() string  { return i.comment }
func (i QueryIntent) Purpose() string  { return i.purpose }
func (i QueryIntent) Validate() error  { _, err := NewQueryIntent(&i.comment, &i.purpose); return err }
func (i QueryIntent) String() string   { return "QueryIntent(<redacted>)" }
func (i QueryIntent) GoString() string { return i.String() }

// MutationIntent.comment is also the root audit reason; callers do not supply
// a duplicate auditReason or a second purpose.
type MutationIntent struct{ comment string }

func NewMutationIntent(comment *string) (MutationIntent, error) {
	what, err := requiredIntentText(comment, "mutation", "comment")
	if err != nil {
		return MutationIntent{}, err
	}
	return MutationIntent{comment: what}, nil
}

func (i MutationIntent) Comment() string     { return i.comment }
func (i MutationIntent) AuditReason() string { return i.comment }
func (i MutationIntent) Validate() error     { _, err := NewMutationIntent(&i.comment); return err }
func (i MutationIntent) String() string      { return "MutationIntent(<redacted>)" }
func (i MutationIntent) GoString() string    { return i.String() }

func (i MutationIntent) ReadbackIntent() (QueryIntent, error) {
	purpose := "verify the persisted mutation result"
	return NewQueryIntent(&i.comment, &purpose)
}

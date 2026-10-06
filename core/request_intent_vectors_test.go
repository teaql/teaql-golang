package core

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
)

//go:embed testdata/request-intent-v1.json
var requestIntentVectors []byte

func TestSharedRequestIntentVectors(t *testing.T) {
	digest := sha256.Sum256(requestIntentVectors)
	if hex.EncodeToString(digest[:]) != "3b911b0edb1b6634204a41c199f67f709d6408405b87d4f2f72a02110cb0b38b" {
		t.Fatal("shared request construction fixture drifted")
	}
	var fixture struct {
		Cases []struct {
			ID, Kind string
			Input    map[string]any
			Error    struct{ Code, Field string }
			Expected struct{ Comment, Purpose string }
		}
	}
	if err := json.Unmarshal(requestIntentVectors, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 20 {
		t.Fatal("incorrect construction vector inventory")
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.ID, func(t *testing.T) {
			text := func(key string) *string {
				value, ok := vector.Input[key].(string)
				if !ok {
					return nil
				}
				return &value
			}
			var err error
			var comment, purpose string
			if vector.Kind == "query" {
				var intent QueryIntent
				intent, err = NewQueryIntent(text("comment"), text("purpose"))
				comment, purpose = intent.Comment(), intent.Purpose()
			} else {
				var intent MutationIntent
				intent, err = NewMutationIntent(text("comment"))
				comment = intent.Comment()
			}
			if vector.Error.Code != "" {
				var required *RequestIntentError
				if !errors.As(err, &required) || required.Code != vector.Error.Code || required.Field != vector.Error.Field || required.RequestKind != vector.Kind {
					t.Fatalf("wrong construction error: %v", err)
				}
			} else if err != nil || comment != vector.Expected.Comment || purpose != vector.Expected.Purpose {
				t.Fatalf("valid request intent changed: %v", err)
			}
		})
	}
}

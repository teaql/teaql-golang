package core

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRequestIntentRequiredText(t *testing.T) {
	for _, blank := range []string{"", " \t\r\n", "\u0085", "\u00a0", "\u2003", "\u202f", "\u3000"} {
		t.Run(fmt.Sprintf("%q", blank), func(t *testing.T) {
			for _, kind := range []string{"query", "mutation"} {
				var err error
				if kind == "query" {
					_, err = NewQueryIntent(&blank, intentTestText("render orders"))
				} else {
					_, err = NewMutationIntent(&blank)
				}
				var required *RequestIntentError
				if !errors.As(err, &required) || required.Code != "REQUEST_COMMENT_REQUIRED" || required.Field != "comment" || required.RequestKind != kind {
					t.Fatalf("incorrect required comment error: %v", err)
				}
			}
			_, err := NewQueryIntent(intentTestText("load orders"), &blank)
			var required *RequestIntentError
			if !errors.As(err, &required) || required.Code != "QUERY_PURPOSE_REQUIRED" || required.Field != "purpose" {
				t.Fatalf("incorrect required purpose error: %v", err)
			}
		})
	}
}

func TestRequestIntentCapturesWithoutTrimmingOrFormattingSecrets(t *testing.T) {
	comment, purpose := " load SECRET-CANARY ", "render SECRET-CANARY"
	query, err := NewQueryIntent(&comment, &purpose)
	if err != nil {
		t.Fatal(err)
	}
	mutation, err := NewMutationIntent(&comment)
	if err != nil {
		t.Fatal(err)
	}
	comment, purpose = "changed", "changed"
	if query.Comment() != " load SECRET-CANARY " || query.Purpose() != "render SECRET-CANARY" || mutation.Comment() != query.Comment() {
		t.Fatal("request intent did not capture original text")
	}
	for _, value := range []any{query, mutation} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			if strings.Contains(fmt.Sprintf(format, value), "SECRET-CANARY") {
				t.Fatal("intent formatter exposed text")
			}
		}
	}
	for _, text := range []string{"\ufeff", "\u200b"} {
		if _, err := NewMutationIntent(&text); err != nil {
			t.Fatal("non White_Space text rejected", err)
		}
	}
	var queryZero QueryIntent
	var mutationZero MutationIntent
	if queryZero.Validate() == nil || mutationZero.Validate() == nil {
		t.Fatal("zero-value intent was valid")
	}
}

func intentTestText(text string) *string { return &text }

package runtime

import (
	"errors"
	"strings"
	"testing"
)

func TestI18nFifteenByFive(t *testing.T) {
	rules := []CheckResult{{RuleID: "required", Location: "name"}, {RuleID: "min", Location: "age", InputValue: 1, SystemValue: 2}, {RuleID: "max", Location: "age", InputValue: 3, SystemValue: 2}, {RuleID: "min_length", Location: "name", InputValue: "a", SystemValue: 2}, {RuleID: "max_length", Location: "name", InputValue: "abc", SystemValue: 2}}
	cells := 0
	for _, locale := range Locales {
		for _, source := range rules {
			result := source
			BuiltinI18nCatalog().Translate(&result, locale)
			if result.Message == "" || strings.HasPrefix(result.Message, "checker.") {
				t.Fatalf("missing %s %s", locale, result.RuleID)
			}
			cells++
		}
	}
	if cells != 75 {
		t.Fatalf("cells=%d", cells)
	}
}
func TestLocaleAliasesAndPreservation(t *testing.T) {
	c := NewUserContext()
	if err := c.SetLocaleCode("ZH_hans"); err != nil {
		t.Fatal(err)
	}
	err := c.SetLocaleCode("xx")
	var unsupported *UnsupportedLocaleError
	if !errors.As(err, &unsupported) {
		t.Fatalf("expected typed error: %v", err)
	}
	if c.Language() != LocaleChineseSimplified {
		t.Fatal("locale changed after error")
	}
}

package runtime

import (
	"errors"
	"reflect"
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

func TestRelocatedCheckerMessagesPreserveLocaleAndMetadata(t *testing.T) {
	rules := []CheckResult{
		{RuleID: "required", Location: "product_name"},
		{RuleID: "min", Location: "unit_price", InputValue: 0, SystemValue: 1},
		{RuleID: "max", Location: "unit_price", InputValue: 3, SystemValue: 2},
		{RuleID: "min_length", Location: "product_name", InputValue: "a", SystemValue: 2},
		{RuleID: "max_length", Location: "product_name", InputValue: "abc", SystemValue: 2},
	}
	prefix := Location().Property("order_item_list").At(2)
	for _, locale := range Locales {
		for _, rule := range rules {
			ctx := NewUserContext()
			if err := ctx.SetLocaleCode(string(locale)); err != nil {
				t.Fatal(err)
			}
			rule.EntityType = "ReviewedItem"
			rule.SourceInstancePath = "/old~1items/2/product~0name"
			original := []CheckResult{rule}
			ctx.TranslateCheckResults(original)
			retained := original[0]
			relocated := []CheckResult{original[0].PrefixedBy(prefix)}
			ctx.TranslateCheckResults(relocated)
			expected := rule.PrefixedBy(prefix)
			BuiltinI18nCatalog().Translate(&expected, locale)
			if relocated[0].Message != expected.Message || !strings.Contains(relocated[0].Message, relocated[0].NativePath()) {
				t.Fatalf("%s %s: stale message %q", locale, rule.RuleID, relocated[0].Message)
			}
			if relocated[0].EntityType != rule.EntityType || relocated[0].SourceInstancePath != rule.SourceInstancePath ||
				!reflect.DeepEqual(relocated[0].InputValue, rule.InputValue) || !reflect.DeepEqual(relocated[0].SystemValue, rule.SystemValue) ||
				!reflect.DeepEqual(original[0], retained) {
				t.Fatal("relocation changed the Checker metadata or original result")
			}
		}
	}
}

func TestRelocatedCheckerUsesCustomerCatalogAndPreservesUnknownMessage(t *testing.T) {
	catalog, err := ParseI18nCatalog([]byte(`{"schema":"teaql.i18n/v1","defaultLocale":"en","locales":{"en":{"messages":{"checker.domain_rule":"Policy at {location}: {system}"}},"zh-CN":{"messages":{"checker.domain_rule":"业务规则 {location}: {system}"}}}}`), BuiltinI18nCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewUserContext()
	if err := ctx.SetLocaleCode("zh-CN"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.InstallI18nCatalog(catalog); err != nil {
		t.Fatal(err)
	}
	prefix := Location().Property("order_item_list").At(2)
	child := []CheckResult{{RuleID: "domain_rule", Location: "unit_price", EntityType: "ReviewedItem", SystemValue: "approved ceiling"}}
	ctx.TranslateCheckResults(child)
	relocated := []CheckResult{child[0].PrefixedBy(prefix)}
	ctx.TranslateCheckResults(relocated)
	if relocated[0].Message != "业务规则 OrderItemList[2].UnitPrice: approved ceiling" {
		t.Fatal(relocated[0].Message)
	}
	opaque := CheckResult{RuleID: "opaque_policy", Location: "product_name", EntityType: "ReviewedItem", Message: "Customer policy rejected this value"}
	results := []CheckResult{opaque.PrefixedBy(prefix)}
	ctx.TranslateCheckResults(results)
	if results[0].Message != opaque.Message || results[0].ModelPath() != "order_item_list[2].product_name" {
		t.Fatal("unknown custom diagnostics were overwritten", results)
	}
}

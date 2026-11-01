package runtime

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func catalogRootResult(rule string) CheckResult {
	return CheckResult{RuleID: rule, EntityType: "OrderItem", SourceInstancePath: "/external~1root",
		CanonicalLocation: Location().Property("purchase_order_list").At(0).Property("order_item_list").At(0).Property("tenant")}
}

func TestContextRootCatalogBuiltinCoverage(t *testing.T) {
	for _, locale := range Locales {
		for _, rule := range []string{"required", "min", "max", "min_length", "max_length", "context_root_missing", "context_root_mismatch"} {
			result := catalogRootResult(rule)
			BuiltinI18nCatalog().Translate(&result, locale)
			if strings.TrimSpace(result.Message) == "" || strings.HasPrefix(result.Message, "checker.") || strings.ContainsAny(result.Message, "{}") {
				t.Errorf("untranslated %s %s: %q", locale, rule, result.Message)
			}
		}
	}
}

func TestContextRootCatalogMessagePrivacy(t *testing.T) {
	for _, rule := range []string{"context_root_missing", "context_root_mismatch"} {
		result := catalogRootResult(rule)
		result.InputValue, result.SystemValue = "PRIVATE-INPUT-ROOT", "PRIVATE-SYSTEM-ROOT"
		BuiltinI18nCatalog().Translate(&result, LocaleEnglish)
		if strings.HasPrefix(result.Message, "checker.") || !strings.Contains(result.Message, result.NativePath()) ||
			strings.Contains(result.Message, "PRIVATE-INPUT-ROOT") || strings.Contains(result.Message, "PRIVATE-SYSTEM-ROOT") {
			t.Errorf("unsafe or unreadable %s message: %q", rule, result.Message)
		}
	}
}

func TestContextRootCatalogFallback(t *testing.T) {
	catalog, err := ParseI18nCatalog([]byte(`{"schema":"teaql.i18n/v1","defaultLocale":"en","locales":{"en":{"messages":{},"vocabulary":{}}}}`), BuiltinI18nCatalog())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range []string{"context_root_missing", "context_root_mismatch"} {
		english, french, chinese := catalogRootResult(rule), catalogRootResult(rule), catalogRootResult(rule)
		BuiltinI18nCatalog().Translate(&english, LocaleEnglish)
		catalog.Translate(&french, LocaleFrench)
		catalog.Translate(&chinese, LocaleChineseSimplified)
		if french.Message != english.Message || strings.HasPrefix(french.Message, "checker.") || chinese.Message == english.Message {
			t.Errorf("missing localized/fallback message for %s: en=%q fr=%q zh=%q", rule, english.Message, french.Message, chinese.Message)
		}
	}
}

func TestContextRootCatalogProjectOverride(t *testing.T) {
	catalog, err := ParseI18nCatalog([]byte(`{"schema":"teaql.i18n/v1","defaultLocale":"en","locales":{"en":{"messages":{"checker.context_root_missing":"Project root required at {location}","checker.context_root_mismatch":"Project root mismatch at {location}"},"vocabulary":{}}}}`), BuiltinI18nCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewUserContext()
	if err := ctx.InstallI18nCatalog(catalog); err != nil {
		t.Fatal(err)
	}
	for _, rule := range []string{"context_root_missing", "context_root_mismatch"} {
		results := []CheckResult{catalogRootResult(rule)}
		ctx.TranslateCheckResults(results)
		if !strings.HasPrefix(results[0].Message, "Project root ") || !strings.Contains(results[0].Message, results[0].NativePath()) {
			t.Fatal("customer catalog did not take precedence", results[0].Message)
		}
	}
}

func TestContextRootCatalogExplicitMessage(t *testing.T) {
	for _, rule := range []string{"context_root_missing", "context_root_mismatch"} {
		ctx := NewUserContext()
		results := []CheckResult{catalogRootResult(rule)}
		results[0].Message = "Customer-owned contextual refusal"
		ctx.TranslateCheckResults(results)
		if results[0].Message != "Customer-owned contextual refusal" {
			t.Fatal("default translation erased customer explanation", results[0].Message)
		}
	}
}

func TestContextRootCatalogAutomaticRerender(t *testing.T) {
	ctx := NewUserContext()
	results := []CheckResult{catalogRootResult("context_root_mismatch")}
	ctx.TranslateCheckResults(results)
	if strings.HasPrefix(results[0].Message, "checker.") {
		t.Fatal("automatic message was never rendered", results[0].Message)
	}
	original := results[0]
	relocated := []CheckResult{results[0].PrefixedBy(Location().Property("orders").At(2))}
	if err := ctx.SetLocaleCode("zh-CN"); err != nil {
		t.Fatal(err)
	}
	ctx.TranslateCheckResults(relocated)
	expected := catalogRootResult("context_root_mismatch").PrefixedBy(Location().Property("orders").At(2))
	BuiltinI18nCatalog().Translate(&expected, LocaleChineseSimplified)
	if relocated[0].Message != expected.Message || relocated[0].Message == original.Message || !strings.Contains(relocated[0].Message, relocated[0].NativePath()) || !reflect.DeepEqual(results[0], original) {
		t.Fatal("automatic message froze a locale/path or changed its source", relocated[0].Message)
	}
	relocated[0].Message = "Replacement business explanation"
	ctx.TranslateCheckResults(relocated)
	if relocated[0].Message != "Replacement business explanation" {
		t.Fatal("automatic ownership survived an explicit customer replacement")
	}
}

func TestContextRootCatalogLegacyKey(t *testing.T) {
	result := catalogRootResult("context_root_mismatch")
	result.Message = "checker.context_root_mismatch"
	BuiltinI18nCatalog().Translate(&result, LocaleEnglish)
	if strings.HasPrefix(result.Message, "checker.") || !strings.Contains(result.Message, result.NativePath()) {
		t.Fatal("an old untranslated key was treated as customer-owned text", result.Message)
	}
}

func TestContextRootCatalogWireProfiles(t *testing.T) {
	profiles := []struct {
		name    string
		profile JsonFieldNamingProfile
		pointer string
	}{
		{"camelCase", JsonFieldCamelCase, "/purchaseOrderList/0/orderItemList/0/tenant"},
		{"snake_case", JsonFieldSnakeCase, "/purchase_order_list/0/order_item_list/0/tenant"},
		{"PascalCase", JsonFieldPascalCase, "/PurchaseOrderList/0/OrderItemList/0/Tenant"},
	}
	for _, locale := range Locales {
		ctx := NewUserContext()
		if err := ctx.SetLocaleCode(string(locale)); err != nil {
			t.Fatal(err)
		}
		for _, rule := range []string{"context_root_missing", "context_root_mismatch"} {
			results := []CheckResult{catalogRootResult(rule)}
			ctx.TranslateCheckResults(results)
			for _, profile := range profiles {
				wire := results[0].ToWire(profile.profile)
				if wire.RuleID != rule || wire.EntityType != "OrderItem" || wire.InstancePath != profile.pointer || wire.SourceInstancePath != "/external~1root" ||
					strings.HasPrefix(wire.Message, "checker.") || wire.InputValue != nil || wire.SystemValue != nil {
					t.Errorf("incomplete root diagnostic %s/%s/%s: %+v", locale, rule, profile.name, wire)
				}
				frame, err := json.Marshal(map[string]any{"language": locale, "profile": profile.name, "wire": wire})
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println("CHECKER_CONTEXT_CATALOG_CASE=" + string(frame))
			}
		}
	}
}

package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

const checkerPrivacySecret = "private-contact-314159"

type privacyChecker func(*UserContext, *CheckAndFixInput) []CheckResult

func (f privacyChecker) CheckAndFix(c *UserContext, in *CheckAndFixInput) []CheckResult {
	return f(c, in)
}

func privacyContext(results []CheckResult, masks []string) *UserContext {
	metadata := NewInMemoryMetadataStore()
	metadata.Register(core.NewEntityDescriptor("OrderItem").
		Property(core.NewPropertyDescriptor("contact_phone", core.TypeText)).AuditMaskFields(masks))
	ctx := NewUserContext().WithCheckerRegistry(privacyChecker(func(*UserContext, *CheckAndFixInput) []CheckResult { return results }))
	ctx.Metadata = metadata
	return ctx
}

func privacyResult(field string, input any) CheckResult {
	return CheckResult{RuleID: "opaque_policy", EntityType: "OrderItem", CanonicalLocation: Location().Property(field), InputValue: input}
}

func privacyRejected(t *testing.T, ctx *UserContext) *RuntimeError {
	t.Helper()
	values := core.Record{"contact_phone": core.ValText(checkerPrivacySecret)}
	err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "OrderItem", Operation: core.MutationInsert, Values: values})
	var checked *RuntimeError
	if !errors.As(err, &checked) || checked.Type != "Check" || len(checked.CheckResults) == 0 {
		t.Fatalf("expected structured Checker rejection: %v", err)
	}
	if values["contact_phone"] != core.ValText(checkerPrivacySecret) {
		t.Fatal("presentation rewrote application input")
	}
	return checked
}

func TestCheckerPrivacyModelMask(t *testing.T) {
	source := privacyResult("contact_phone", checkerPrivacySecret)
	err := privacyRejected(t, privacyContext([]CheckResult{source}, []string{"contact_phone"}))
	result := err.CheckResults[0]
	if result.InputValue != MaskAuditValue(checkerPrivacySecret) {
		t.Fatal("model mask missing")
	}
	for _, value := range []any{err, result, result.ToWire(JsonFieldCamelCase)} {
		encoded, e := json.Marshal(value)
		if e != nil {
			t.Fatal(e)
		}
		for _, text := range []string{string(encoded), fmt.Sprintf("%v", value), fmt.Sprintf("%+v", value), fmt.Sprintf("%#v", value)} {
			if strings.Contains(text, checkerPrivacySecret) {
				t.Fatal("sensitive value escaped presentation")
			}
		}
	}
	if source.InputValue != checkerPrivacySecret {
		t.Fatal("customer-owned result changed")
	}
}

func TestCheckerPrivacyCredentialsWithoutMetadata(t *testing.T) {
	values := []CheckResult{privacyResult("password", checkerPrivacySecret)}
	values[0].SystemValue = "expected-private-credential"
	NewUserContext().TranslateCheckResults(values)
	if values[0].InputValue != "[REDACTED]" || values[0].SystemValue != "[REDACTED]" {
		t.Fatal("credential or expected constraint leaked")
	}
}

func TestCheckerPrivacyOriginalUnicodeLength(t *testing.T) {
	for _, locale := range []string{"en", "zh-CN"} {
		ctx := NewUserContext()
		if err := ctx.SetLocaleCode(locale); err != nil {
			t.Fatal(err)
		}
		result := privacyResult("password", "密钥🔑abc")
		result.RuleID = "min_length"
		result.SystemValue = 12
		results := []CheckResult{result}
		ctx.TranslateCheckResults(results)
		phrase := "is 6"
		if locale == "zh-CN" {
			phrase = "实际长度为 6"
		}
		if results[0].InputValue != "[REDACTED]" || results[0].SystemValue != 12 || !strings.Contains(results[0].Message, phrase) {
			t.Fatal("original length or numeric bound lost")
		}
	}
}

func TestCheckerPrivacyCanonicalPaths(t *testing.T) {
	r := privacyResult("contact_phone", checkerPrivacySecret)
	r.CanonicalLocation = Location().Property("order_item_list").At(2).Property("contact_phone")
	r.SourceInstancePath = "/old~1items/2/phone~0"
	got := privacyRejected(t, privacyContext([]CheckResult{r}, []string{"contact_phone"})).CheckResults[0]
	if got.InputValue != MaskAuditValue(checkerPrivacySecret) || !reflect.DeepEqual(got.ObjectLocation(), r.ObjectLocation()) || got.SourceInstancePath != r.SourceInstancePath {
		t.Fatal("mask or canonical provenance changed")
	}
	for profile, path := range map[JsonFieldNamingProfile]string{JsonFieldCamelCase: "/orderItemList/2/contactPhone", JsonFieldSnakeCase: "/order_item_list/2/contact_phone", JsonFieldPascalCase: "/OrderItemList/2/ContactPhone"} {
		if got.ToWire(profile).InstancePath != path {
			t.Fatal("wire profile changed")
		}
	}
}

func TestCheckerPrivacyOrdinaryValuesAndExplanation(t *testing.T) {
	results := []CheckResult{{RuleID: "min", Location: "quantity", InputValue: 1, SystemValue: 2}, {RuleID: "opaque_policy", Location: "name", InputValue: "ordinary", Message: "Customer-owned explanation"}}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != 1 || results[0].SystemValue != 2 || results[1].Message != "Customer-owned explanation" || results[1].InputValue != "ordinary" {
		t.Fatal("ordinary diagnostic changed")
	}
}

func TestCheckerPrivacySiblingMessages(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		results := []CheckResult{privacyResult("contact_phone", checkerPrivacySecret), {RuleID: "opaque_policy", Location: "name", Message: "Fix " + checkerPrivacySecret + " first"}}
		if reverse {
			results[0], results[1] = results[1], results[0]
		}
		err := privacyRejected(t, privacyContext(results, []string{"contact_phone"}))
		if strings.Contains(err.Error(), checkerPrivacySecret) {
			t.Fatal("sibling message leaked")
		}
	}
}

func TestCheckerPrivacyCachedRegistryOwnership(t *testing.T) {
	shared := []CheckResult{privacyResult("contact_phone", checkerPrivacySecret)}
	shared[0].RuleID = "min_length"
	shared[0].SystemValue = 25
	original := shared[0]
	ctx := privacyContext(shared, []string{"contact_phone"})
	first := privacyRejected(t, ctx).CheckResults[0]
	if err := ctx.SetLocaleCode("zh-CN"); err != nil {
		t.Fatal(err)
	}
	second := privacyRejected(t, ctx).CheckResults[0]
	if !reflect.DeepEqual(shared[0], original) || first.InputValue != MaskAuditValue(checkerPrivacySecret) || second.InputValue != first.InputValue || first.Message == second.Message {
		t.Fatal("cached registry or locale snapshot changed")
	}
}

func TestCheckerPrivacyDynamicField(t *testing.T) {
	results := []CheckResult{privacyResult("#external_reference", checkerPrivacySecret)}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != MaskAuditValue(checkerPrivacySecret) || results[0].ModelPath() != "#external_reference" {
		t.Fatal("dynamic field was not safely projected")
	}
}

func TestCheckerPrivacyNestedCredentials(t *testing.T) {
	input := map[string]any{"nested": []any{map[string]any{"access_token": checkerPrivacySecret}}}
	results := []CheckResult{privacyResult("profile", input)}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != "[REDACTED]" || input["nested"].([]any)[0].(map[string]any)["access_token"] != checkerPrivacySecret {
		t.Fatal("structured credential escaped or source changed")
	}
}

func TestCheckerPrivacyRetranslatedLength(t *testing.T) {
	ctx := NewUserContext()
	r := privacyResult("password", "密钥🔑abc")
	r.RuleID = "min_length"
	r.SystemValue = 12
	first := []CheckResult{r}
	ctx.TranslateCheckResults(first)
	if err := ctx.SetLocaleCode("zh-CN"); err != nil {
		t.Fatal(err)
	}
	relocated := []CheckResult{first[0].PrefixedBy(Location().Property("order_item_list").At(2))}
	ctx.TranslateCheckResults(relocated)
	if relocated[0].InputValue != "[REDACTED]" || !strings.Contains(relocated[0].Message, "实际长度为 6") || !strings.Contains(relocated[0].Message, "OrderItemList[2].Password") {
		t.Fatal("cached length or relocated path is wrong")
	}
}

func TestCheckerPrivacyShortCredentialBounds(t *testing.T) {
	r := privacyResult("password", "1")
	r.RuleID = "min_length"
	r.SystemValue = 12
	results := []CheckResult{r}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != "[REDACTED]" || !strings.Contains(results[0].Message, "12") || !strings.Contains(results[0].Message, "is 1") {
		t.Fatal("short credential corrupted a bound or validation length")
	}
}

func TestCheckerPrivacyRootMessageOnly(t *testing.T) {
	r := CheckResult{RuleID: "opaque_policy", CanonicalLocation: Location(), Message: "Cannot accept " + checkerPrivacySecret}
	got := privacyRejected(t, privacyContext([]CheckResult{r}, []string{"contact_phone"})).CheckResults[0]
	if got.Message != "Cannot accept [REDACTED]" || got.InputValue != nil || len(got.ObjectLocation().Segments) != 0 {
		t.Fatal("root error leaked a checked value or invented a path")
	}
}

func TestCheckerPrivacyMaskAlgorithmAndTypedValues(t *testing.T) {
	for _, pair := range [][2]any{{"12345", "*****"}, {"密钥🔑abc", "******"}, {"short", "*****"}, {nil, nil}, {12345, "*****"}, {core.ValText(checkerPrivacySecret), core.ValText(MaskAuditValue(checkerPrivacySecret))}} {
		got := privacyRejected(t, privacyContext([]CheckResult{privacyResult("contact_phone", pair[0])}, []string{"contact_phone"})).CheckResults[0]
		if !reflect.DeepEqual(got.InputValue, pair[1]) {
			t.Fatal("existing audit algorithm or typed carrier not preserved")
		}
	}
}

func TestCheckerPrivacyCustomCatalogLiteral(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"schema": "teaql.i18n/v1", "defaultLocale": "en", "locales": map[string]any{"en": map[string]any{"messages": map[string]string{"checker.minLength": "Private " + checkerPrivacySecret + ": {location} expects {system}; got {input_len}"}}}})
	catalog, err := ParseI18nCatalog(data, BuiltinI18nCatalog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewUserContext()
	if err = ctx.InstallI18nCatalog(catalog); err != nil {
		t.Fatal(err)
	}
	r := privacyResult("password", checkerPrivacySecret)
	r.RuleID = "min_length"
	r.SystemValue = 12
	results := []CheckResult{r}
	ctx.TranslateCheckResults(results)
	if results[0].Message != fmt.Sprintf("Private [REDACTED]: Password expects 12; got %d", len(checkerPrivacySecret)) {
		t.Fatal("custom literal leaked or typed substitutions changed")
	}
}

func TestCheckerPrivacyPublicRuleVocabulary(t *testing.T) {
	r := privacyResult("password", "required")
	r.RuleID = "required"
	r.SystemValue = "required"
	results := []CheckResult{r}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != "[REDACTED]" || results[0].SystemValue != "[REDACTED]" || results[0].Message != "The Password is required" {
		t.Fatal("credential escaped or public rule text was corrupted")
	}
}

func TestCheckerPrivacyCachedPublicBatch(t *testing.T) {
	ctx := NewUserContext()
	first := []CheckResult{{RuleID: "opaque_policy", Location: "name", Message: "Fix " + checkerPrivacySecret}}
	ctx.TranslateCheckResults(first)
	next := []CheckResult{first[0], privacyResult("password", checkerPrivacySecret)}
	ctx.TranslateCheckResults(next)
	if next[0].Message != "Fix [REDACTED]" || first[0].Message != "Fix "+checkerPrivacySecret {
		t.Fatal("new batch borrowed a stale safe-message decision")
	}
}

func TestCheckerPrivacyDebugOptIn(t *testing.T) {
	for _, value := range []string{"", "true", plaintextLogAck} {
		t.Setenv(plaintextLogEnv, value)
		results := []CheckResult{privacyResult("contact_phone", checkerPrivacySecret), privacyResult("password", checkerPrivacySecret)}
		got := privacyRejected(t, privacyContext(results, []string{"contact_phone"})).CheckResults
		want := MaskAuditValue(checkerPrivacySecret)
		if value == plaintextLogAck {
			want = checkerPrivacySecret
		}
		if got[0].InputValue != want || got[1].InputValue != "[REDACTED]" {
			t.Fatal("debug switch is not exact or exposed a credential")
		}
	}
}

func TestCheckerPrivacyIndependentContexts(t *testing.T) {
	r := []CheckResult{privacyResult("contact_phone", checkerPrivacySecret)}
	first := privacyRejected(t, privacyContext(r, []string{"contact_phone"})).CheckResults[0]
	second := privacyRejected(t, privacyContext(r, []string{})).CheckResults[0]
	if first.InputValue != MaskAuditValue(checkerPrivacySecret) || second.InputValue != checkerPrivacySecret || r[0].InputValue != checkerPrivacySecret {
		t.Fatal("independent metadata or caller state leaked")
	}
}

func TestCheckerPrivacyCredentialAncestor(t *testing.T) {
	r := privacyResult("label", checkerPrivacySecret)
	r.CanonicalLocation = Location().Property("api_key").Property("label")
	results := []CheckResult{r}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != "[REDACTED]" || !reflect.DeepEqual(results[0].ObjectLocation(), r.ObjectLocation()) {
		t.Fatal("credential ancestor not protected")
	}
}

func TestCheckerPrivacyEscapedMessage(t *testing.T) {
	input := "line\n\"private-contact\"🔑"
	r := privacyResult("password", input)
	r.Message = fmt.Sprintf("Invalid %q", input)
	results := []CheckResult{r}
	NewUserContext().TranslateCheckResults(results)
	if results[0].Message != "Invalid \"[REDACTED]\"" || results[0].InputValue != "[REDACTED]" {
		t.Fatal("quoted secret escaped a custom message")
	}
}

func TestCheckerPrivacyCanonicalRelationMask(t *testing.T) {
	descriptor := core.NewEntityDescriptor("OrderItem").Property(core.NewPropertyDescriptor("school_type_id", core.TypeU64)).
		Relation(core.NewRelationDescriptor("schoolTypeEntity", "SchoolType").CanonicalKey("school_type").LocalKey("school_type_id").ForeignKey("id")).AuditMaskFields([]string{"school_type_id"})
	ctx := NewUserContext()
	metadata := NewInMemoryMetadataStore()
	metadata.Register(descriptor)
	ctx.Metadata = metadata
	r := privacyResult("school_type", uint64(1001))
	results := []CheckResult{r}
	ctx.TranslateCheckResults(results)
	if results[0].InputValue != "****" || results[0].ModelPath() != "school_type" {
		t.Fatal("canonical relation did not use the trusted storage-field mask")
	}
}

func TestCheckerPrivacyConcurrentTranslation(t *testing.T) {
	ctx := privacyContext(nil, []string{"contact_phone"})
	source := privacyResult("contact_phone", checkerPrivacySecret)
	var workers sync.WaitGroup
	failures := make(chan bool, 16)
	for i := 0; i < 16; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			own := []CheckResult{source}
			ctx.TranslateCheckResults(own)
			failures <- own[0].InputValue == MaskAuditValue(checkerPrivacySecret)
		}()
	}
	workers.Wait()
	close(failures)
	for ok := range failures {
		if !ok {
			t.Fatal("concurrent projection did not protect its owned result")
		}
	}
	if source.InputValue != checkerPrivacySecret {
		t.Fatal("source mutated")
	}
}

func TestCheckerPrivacyRootBeforeAndAfterFix(t *testing.T) {
	const after = "private-after-fix-271828"
	ctx := privacyContext(nil, []string{"contact_phone"})
	ctx.WithCheckerRegistry(privacyChecker(func(_ *UserContext, input *CheckAndFixInput) []CheckResult {
		input.Values["contact_phone"] = core.ValText(after)
		return []CheckResult{{RuleID: "opaque_policy", Message: "Changed " + checkerPrivacySecret + " to " + after}}
	}))
	values := core.Record{"contact_phone": core.ValText(checkerPrivacySecret)}
	var checked *RuntimeError
	err := ctx.CheckAndFix(&CheckAndFixInput{Entity: "OrderItem", Operation: core.MutationInsert, Values: values})
	if !errors.As(err, &checked) || len(checked.CheckResults) != 1 || checked.CheckResults[0].Message != "Changed [REDACTED] to [REDACTED]" {
		t.Fatal("root message did not protect both sides of Fix")
	}
	if values["contact_phone"] != core.ValText(after) {
		t.Fatal("presentation undid the real Fix")
	}
	public := []CheckResult{{RuleID: "opaque_policy", Message: "Public " + checkerPrivacySecret}}
	ctx.TranslateCheckResults(public)
	if public[0].Message != "Public "+checkerPrivacySecret {
		t.Fatal("completed invocation retained raw redaction state")
	}
}

func TestCheckerPrivacyCyclicInput(t *testing.T) {
	input := map[string]any{"access_token": checkerPrivacySecret}
	input["self"] = input
	results := []CheckResult{privacyResult("profile", input)}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != "[REDACTED]" || input["access_token"] != checkerPrivacySecret {
		t.Fatal("cyclic credential did not fail closed")
	}
	if _, err := json.Marshal(results[0].ToWire(JsonFieldCamelCase)); err != nil {
		t.Fatal("cyclic source survived projection")
	}
}

func TestCheckerPrivacyCollectionsOwnTheirCachedSnapshot(t *testing.T) {
	input := []core.Value{core.ValText(checkerPrivacySecret), core.ValText("short"), core.ValNull()}
	ctx := privacyContext(nil, []string{"contact_phone"})
	results := []CheckResult{privacyResult("contact_phone", core.Value{V: input})}
	ctx.TranslateCheckResults(results)
	want := core.Value{V: []core.Value{core.ValText(MaskAuditValue(checkerPrivacySecret)), core.ValText("*****"), core.ValNull()}}
	if !reflect.DeepEqual(results[0].InputValue, want) || input[0] != core.ValText(checkerPrivacySecret) {
		t.Fatal("collection mask changed the input or did not use audit rules")
	}
	ctx.TranslateCheckResults(results)
	if !reflect.DeepEqual(results[0].InputValue, want) {
		t.Fatal("collection projection was not idempotent")
	}
	results[0].InputValue.(core.Value).V.([]core.Value)[0] = core.ValText("private-reinjected-123456")
	ctx.TranslateCheckResults(results)
	if results[0].InputValue.(core.Value).V.([]core.Value)[0] != core.ValText(MaskAuditValue("private-reinjected-123456")) {
		t.Fatal("mutable output corrupted the private safe snapshot")
	}
}

func TestCheckerPrivacyDebugSnapshotReturnsToSafeMode(t *testing.T) {
	t.Setenv(plaintextLogEnv, plaintextLogAck)
	ctx := privacyContext(nil, []string{"contact_phone"})
	first := []CheckResult{privacyResult("contact_phone", checkerPrivacySecret)}
	ctx.TranslateCheckResults(first)
	if first[0].InputValue != checkerPrivacySecret {
		t.Fatal("explicit debug opt-in not honored")
	}
	t.Setenv(plaintextLogEnv, "")
	next := []CheckResult{first[0]}
	ctx.TranslateCheckResults(next)
	if next[0].InputValue != MaskAuditValue(checkerPrivacySecret) || first[0].InputValue != checkerPrivacySecret {
		t.Fatal("debug snapshot was not reprojected independently")
	}
}

func TestCheckerPrivacyTypedJSONCredentialsAndNamedBounds(t *testing.T) {
	type profile struct {
		Token string `json:"access_token"`
		Note  string
	}
	type bound int64
	input := profile{Token: checkerPrivacySecret, Note: "public note"}
	r := privacyResult("profile", input)
	r.SystemValue = bound(12)
	r.Message = "Invalid " + checkerPrivacySecret
	results := []CheckResult{r}
	NewUserContext().TranslateCheckResults(results)
	if results[0].InputValue != "[REDACTED]" || results[0].SystemValue != bound(12) || results[0].Message != "Invalid [REDACTED]" || input.Token != checkerPrivacySecret {
		t.Fatal("typed JSON credentials escaped, named bound changed, or input was rewritten")
	}
}

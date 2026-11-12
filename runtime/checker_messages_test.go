package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func naturalFixture(t *testing.T) map[string]any {
	t.Helper()
	bytes, err := os.ReadFile("testdata/natural-messages/cases-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]any
	if json.Unmarshal(bytes, &fixture) != nil {
		t.Fatal("invalid fixture")
	}
	return fixture
}
func naturalModel(t *testing.T) *CheckerPresentationModel {
	t.Helper()
	fixture := naturalFixture(t)
	data, _ := json.Marshal(map[string]any{"schema": "teaql.checker-model/v1", "revision": fixture["revision"], "entities": fixture["entities"], "currentObject": fixture["currentObject"]})
	model, err := ParseCheckerPresentationModel(data)
	if err != nil {
		t.Fatal(err)
	}
	return model
}
func naturalContext(t *testing.T) *UserContext {
	t.Helper()
	c := NewUserContext()
	if err := c.InstallCheckerPresentationModel(naturalModel(t)); err != nil {
		t.Fatal(err)
	}
	_ = c.InstallCheckerMessageWarning(func(CheckerMessageWarning) {})
	return c
}
func naturalLocation(path []any) ObjectLocation {
	result := Location()
	for _, segment := range path {
		if name, ok := segment.(string); ok {
			result = result.Property(name)
		} else {
			result = result.At(int(segment.(float64)))
		}
	}
	return result
}
func naturalCatalog(data string) []byte {
	return []byte(`{"schema":"teaql.checker-messages/v1","defaultLocale":"en","locales":` + data + `}`)
}

func TestNaturalCheckerSharedTwentyFourCases(t *testing.T) {
	fixture := naturalFixture(t)
	source, err := os.ReadFile("testdata/natural-messages/root-cases-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var roots map[string]any
	_ = json.Unmarshal(source, &roots)
	cases := append(fixture["cases"].([]any), roots["cases"].([]any)...)
	if len(cases) != 24 {
		t.Fatal(len(cases))
	}
	for _, value := range cases {
		item := value.(map[string]any)
		t.Run(item["name"].(string), func(t *testing.T) {
			origin := "Catalog"
			if value, present := item["origin"]; present {
				origin, _ = value.(string)
			}
			location := naturalLocation(item["path"].([]any))
			snapshot := CaptureCheckerPresentation(naturalModel(t), origin, location)
			if strings.ToUpper(string(snapshot.Shape())) != item["shape"] {
				t.Fatal(snapshot.Shape())
			}
			locale, err := ParseLocale(item["locale"].(string))
			if err != nil {
				t.Fatal(err)
			}
			args := map[string]string{}
			if values, ok := item["arguments"].(map[string]any); ok {
				for key, value := range values {
					args[key] = fmt.Sprint(value)
				}
			}
			output := snapshot.Render(builtinCheckerMessages, builtinCheckerMessages.Select(locale, item["rule"].(string), snapshot.Shape()), args)
			if output.Message != item["expected"] {
				t.Fatalf("got %q want %q", output.Message, item["expected"])
			}
			if pointer, ok := item["instancePath"].(string); ok && location.InstancePath() != pointer {
				t.Fatal(location.InstancePath())
			}
			fmt.Println("GO_NATURAL_GOLDEN=" + item["name"].(string))
		})
	}
}
func TestNaturalCheckerAllFifteenLocalesAnd270Templates(t *testing.T) {
	count := 0
	for _, locale := range Locales {
		for _, rule := range []string{"checker.required", "checker.min", "checker.max", "checker.minLength", "checker.maxLength", "checker.invalid"} {
			for _, shape := range []CheckerMessageShape{CheckerMessageOne, CheckerMessageTwo, CheckerMessageDeep} {
				selected, ok := builtinCheckerMessages.Lookup(locale, rule, shape)
				if !ok {
					t.Fatalf("missing %s %s %s", locale, rule, shape)
				}
				text := selected.Render(map[string]string{"field": "F", "object": "O", "objectPath": "P", "system": "8", "input_len": "9", "unit": "U"})
				if !strings.Contains(text, "F") || strings.Contains(text, "{") {
					t.Fatal(text)
				}
				if shape == CheckerMessageTwo && !strings.Contains(text, "O") {
					t.Fatal(text)
				}
				if shape == CheckerMessageDeep && !strings.Contains(text, "P") {
					t.Fatal(text)
				}
				count++
			}
		}
	}
	if count != 270 {
		t.Fatal(count)
	}
}
func TestNaturalCheckerCatalogRejectionIsAtomic(t *testing.T) {
	c := naturalContext(t)
	good := naturalCatalog(`{"en":{"rules":{"checker.required":{"one":"Need {field}."}}}}`)
	if err := c.InstallCheckerMessageCatalog(good); err != nil {
		t.Fatal(err)
	}
	bad := []string{
		`{"en":{"rules":{"checker.required":{"one":"{field} {input}"}}}}`,
		`{"en":{"rules":{"checker.min":{"one":"{field}"}}}}`,
		`{"en":{"rules":{"checker.maxLength":{"one":"{field} {system} {input_len} {unit} {input}"}}}}`,
		`{"en":{"rules":{"checker.required":{"one":"{unknown}"}}}}`,
		`{"en":{"rules":{"checker.required":{"two":"{field}"}}}}`,
		`{"en":{"rules":{"checker.required":{"deep":"{field} {object}"}}}}`,
		`{"en":{"rules":{"checker.required":{"four":"{field}","one":""}}}}`,
		`{"en":{},"en-US":{}}`,
		`{"en":{"rules":{"checker.required":{"one":"{field}","one":"{field}"}}}}`,
		`{"en":{"locationPatterns":{"reference":"{relation}","item":"{relation}","separator":" → "}}}`,
		`{"en":{"locationPatterns":{"reference":"{relation}","item":"{ordinal} {relation}","separator":"{input}"}}}`,
	}
	for _, value := range bad {
		if err := c.InstallCheckerMessageCatalog(naturalCatalog(value)); err == nil {
			t.Fatal("invalid catalog admitted", value)
		}
	}
	if err := c.InstallCheckerMessageCatalog(append(append([]byte{}, good...), []byte(` {}`)...)); err == nil {
		t.Fatal("trailing JSON admitted")
	}
	if err := c.InstallCheckerMessageCatalog([]byte(strings.Replace(string(good), "teaql.checker-messages/v1", "teaql.i18n/v1", 1))); err == nil {
		t.Fatal("unsupported schema")
	}
	values := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("name")}}
	c.TranslateCheckResultsFrom("Catalog", values)
	if values[0].Message != "Need name." {
		t.Fatal(values[0].Message)
	}
}
func TestNaturalCheckerMetadataCompositionAndAliases(t *testing.T) {
	model := naturalModel(t)
	same, err := model.Merge(model)
	if err != nil || same != model {
		t.Fatal("idempotent merge failed")
	}
	fixture := naturalFixture(t)
	entities := fixture["entities"].(map[string]any)
	catalog := entities["Catalog"].(map[string]any)
	catalog["labels"].(map[string]any)["en"] = "different"
	source, _ := json.Marshal(map[string]any{"schema": "teaql.checker-model/v1", "revision": "conflict", "entities": entities})
	other, err := ParseCheckerPresentationModel(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = model.Merge(other); err == nil {
		t.Fatal("conflicting identity admitted")
	}
	fields := catalog["fields"].([]any)
	fields[0].(map[string]any)["aliases"] = []any{"entryList"}
	source, _ = json.Marshal(map[string]any{"schema": "teaql.checker-model/v1", "revision": "duplicate", "entities": entities})
	if _, err = ParseCheckerPresentationModel(source); err == nil {
		t.Fatal("duplicate alias admitted")
	}
	c := NewRuntimeModule().CheckerPresentation(model).IntoContext()
	_ = c.InstallCheckerMessageWarning(func(CheckerMessageWarning) {})
	values := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("entryList").At(1).Property("name")}}
	c.TranslateCheckResultsFrom("Catalog", values)
	if values[0].Message != "The name of the 2nd entry is required." {
		t.Fatal(values[0].Message)
	}
}
func TestNaturalCheckerOrdinalsAndLiteralBraces(t *testing.T) {
	for value, want := range map[string]string{"1": "1st", "2": "2nd", "3": "3rd", "11": "11th", "12": "12th", "13": "13th", "21": "21st"} {
		if checkerOrdinal(value) != want {
			t.Fatal(value)
		}
	}
	c := naturalContext(t)
	if err := c.InstallCheckerMessageCatalog(naturalCatalog(`{"en":{"rules":{"checker.required":{"one":"{{literal}} {field}"}}}}`)); err != nil {
		t.Fatal(err)
	}
	result := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("name")}}
	c.TranslateCheckResultsFrom("Catalog", result)
	if result[0].Message != "{literal} name" {
		t.Fatal(result[0].Message)
	}
}

type naturalRules func(*UserContext, *CheckAndFixInput) []CheckResult

func (f naturalRules) CheckAndFix(c *UserContext, input *CheckAndFixInput) []CheckResult {
	return f(c, input)
}
func TestNaturalCheckerActualEntryAllLocalesPrivacyAndWire(t *testing.T) {
	c := naturalContext(t).WithCheckerRegistry(naturalRules(func(*UserContext, *CheckAndFixInput) []CheckResult {
		owner := Location().Property("entry_list").At(1)
		return []CheckResult{
			{RuleID: "required", EntityType: "Entry", CanonicalLocation: owner.Property("name")},
			{RuleID: "min", EntityType: "Entry", CanonicalLocation: owner.Property("amount"), InputValue: -1, SystemValue: 0},
			{RuleID: "max", EntityType: "Entry", CanonicalLocation: owner.Property("amount"), InputValue: 10, SystemValue: 9},
			{RuleID: "min_length", EntityType: "Entry", CanonicalLocation: owner.Property("password"), InputValue: "密钥🔑abc", SystemValue: 8},
			{RuleID: "max_length", EntityType: "Entry", CanonicalLocation: owner.Property("password"), InputValue: "密钥🔑abc", SystemValue: 4},
		}
	}))
	if c.Language() != LocaleEnglish {
		t.Fatal("default is not English")
	}
	for _, locale := range Locales {
		_ = c.SetLocaleCode(string(locale))
		err := c.CheckAndFix(&CheckAndFixInput{Entity: "Catalog", Operation: core.MutationInsert, Values: core.Record{}})
		var failure *RuntimeError
		if !errors.As(err, &failure) || len(failure.CheckResults) != 5 {
			t.Fatal(err)
		}
		for _, result := range failure.CheckResults {
			wire := result.ToWire(JsonFieldCamelCase)
			if result.EntityType != "Entry" || !strings.HasPrefix(wire.InstancePath, "/entryList/1/") || wire.Message == "" {
				t.Fatal(wire)
			}
			bytes, _ := json.Marshal(wire)
			if strings.Contains(string(bytes), "密钥🔑abc") {
				t.Fatal("secret escaped")
			}
		}
		if locale == LocaleEnglish {
			if failure.CheckResults[0].Message != "The name of the 2nd entry is required." || !strings.Contains(failure.CheckResults[4].Message, "current length is 6.") {
				t.Fatal(failure.CheckResults)
			}
		}
		if locale == LocaleChineseSimplified && failure.CheckResults[0].Message != "请填写第2条明细的名称。" {
			t.Fatal(failure.CheckResults[0].Message)
		}
	}
}
func TestNaturalCheckerLegacyAndLiteralPriority(t *testing.T) {
	c := naturalContext(t)
	legacy, err := ParseI18nCatalog([]byte(`{"schema":"teaql.i18n/v1","defaultLocale":"en","locales":{"en":{"messages":{"checker.required":"Legacy {location}: {input}"}}}}`), builtinCatalog)
	if err != nil {
		t.Fatal(err)
	}
	_ = c.InstallI18nCatalog(legacy)
	makeResult := func() []CheckResult {
		return []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("name")}}
	}
	values := makeResult()
	c.TranslateCheckResultsFrom("Catalog", values)
	if values[0].Message != "Legacy Name: <nil>" {
		t.Fatal(values[0].Message)
	}
	_ = c.InstallCheckerMessageCatalog(naturalCatalog(`{"en":{"rules":{"checker.required":{"one":"New {field}"}}}}`))
	values = makeResult()
	c.TranslateCheckResultsFrom("Catalog", values)
	if values[0].Message != "New name" {
		t.Fatal(values[0].Message)
	}
	_ = c.SetLocaleCode("zh-CN")
	values = makeResult()
	c.TranslateCheckResultsFrom("Catalog", values)
	if values[0].Message != "请填写名称。" {
		t.Fatal(values[0].Message)
	}
	values[0].Message = "Customer literal {field}"
	c.TranslateCheckResults(values)
	if values[0].Message != "Customer literal {field}" {
		t.Fatal(values[0].Message)
	}
}
func TestNaturalCheckerRendererFailureAndPanicCannotReplaceValidation(t *testing.T) {
	for _, panicMode := range []bool{false, true} {
		c := naturalContext(t)
		_ = c.InstallCheckerMessageRenderer(CheckerMessageRendererFunc(func(CheckerMessageRequest) (CheckerMessageOutput, error) {
			if panicMode {
				panic("PRIVATE_PROVIDER_CAUSE")
			}
			return CheckerMessageOutput{}, errors.New("PRIVATE_PROVIDER_CAUSE")
		}))
		values := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("entry_list").At(1).Property("name")}}
		c.TranslateCheckResultsFrom("Catalog", values)
		if values[0].Message != "The name of the 2nd entry is invalid." {
			t.Fatal(values[0].Message)
		}
	}
}
func TestNaturalCheckerGraphReusesSafePlansWithoutCallingProviderAgain(t *testing.T) {
	secret := "graph-provider-private"
	calls := 0
	c := naturalContext(t)
	_ = c.InstallCheckerMessageRenderer(CheckerMessageRendererFunc(func(CheckerMessageRequest) (CheckerMessageOutput, error) {
		calls++
		return CheckerMessageOutput{Message: "Provider echoed " + secret}, nil
	}))
	values := []CheckResult{{RuleID: "min_length", CanonicalLocation: Location().Property("password"), InputValue: secret, SystemValue: 50}}
	c.TranslateCheckResultsFrom("Entry", values)
	if strings.Contains(values[0].Message, secret) {
		t.Fatal("initial secret escaped")
	}
	failure := &RuntimeError{Type: "Check", CheckResults: []CheckResult{values[0].PrefixedBy(Location().Property("entry_list").At(1))}}
	state := &preparedGraphSnapshot{checkerMessages: c.captureCheckerMessages(), checkerOrigin: "Catalog"}
	rerenderGraphCheckerError(failure, state)
	if calls != 1 || strings.Contains(failure.CheckResults[0].Message, secret) {
		t.Fatal("provider was reinvoked without original redactions", calls)
	}
}
func TestNaturalCheckerGraphPreprojectsLaterDepthTemplates(t *testing.T) {
	secret := "graph-template-private"
	c := naturalContext(t)
	if err := c.InstallCheckerMessageCatalog(naturalCatalog(`{"en":{"rules":{"checker.minLength":{"two":"{field} of {object} needs {system} {unit}; got {input_len}. Private ` + secret + `"}}}}`)); err != nil {
		t.Fatal(err)
	}
	values := []CheckResult{{RuleID: "min_length", CanonicalLocation: Location().Property("password"), InputValue: secret, SystemValue: 50}}
	c.TranslateCheckResultsFrom("Entry", values)
	failure := &RuntimeError{Type: "Check", CheckResults: []CheckResult{values[0].PrefixedBy(Location().Property("entry_list").At(1))}}
	rerenderGraphCheckerError(failure, &preparedGraphSnapshot{checkerMessages: c.captureCheckerMessages(), checkerOrigin: "Catalog"})
	text := failure.CheckResults[0].Message
	if strings.Contains(text, secret) || !strings.Contains(text, "Private [REDACTED]") || !strings.Contains(text, "2nd entry") {
		t.Fatal(text)
	}
}

func TestNaturalCheckerRetainedLocaleAndOwnerPatternsAreProjected(t *testing.T) {
	secret := "later-locale-private"
	c := naturalContext(t)
	if err := c.InstallCheckerMessageCatalog(naturalCatalog(`{"zh-CN":{"rules":{"checker.minLength":{"two":"{object}的{field}至少{system}{unit}，当前{input_len}；` + secret + `"}}}}`)); err != nil {
		t.Fatal(err)
	}
	values := []CheckResult{{RuleID: "min_length", CanonicalLocation: Location().Property("password"), InputValue: secret, SystemValue: 50}}
	c.TranslateCheckResultsFrom("Entry", values)
	_ = c.SetLocaleCode("zh-CN")
	failure := &RuntimeError{Type: "Check", CheckResults: []CheckResult{values[0].PrefixedBy(Location().Property("entry_list").At(1))}}
	rerenderGraphCheckerError(failure, &preparedGraphSnapshot{checkerMessages: c.captureCheckerMessages(), checkerOrigin: "Catalog"})
	text := failure.CheckResults[0].Message
	if strings.Contains(text, secret) || !strings.Contains(text, "第2条明细") || !strings.Contains(text, "[REDACTED]") || !strings.Contains(text, fmt.Sprint(len(secret))) {
		t.Fatal(text)
	}

	c = naturalContext(t)
	if err := c.InstallCheckerMessageCatalog(naturalCatalog(`{"en":{"locationPatterns":{"reference":"Private 1 {relation}","item":"Private 1 the {ordinal} {relation}","separator":" → "}}}`)); err != nil {
		t.Fatal(err)
	}
	values = []CheckResult{{RuleID: "min_length", CanonicalLocation: Location().Property("password"), InputValue: "1", SystemValue: 12}}
	c.TranslateCheckResultsFrom("Entry", values)
	failure = &RuntimeError{Type: "Check", CheckResults: []CheckResult{values[0].PrefixedBy(Location().Property("entry_list").At(1))}}
	rerenderGraphCheckerError(failure, &preparedGraphSnapshot{checkerMessages: c.captureCheckerMessages(), checkerOrigin: "Catalog"})
	text = failure.CheckResults[0].Message
	if !strings.Contains(text, "Private [REDACTED] the 2nd entry") || !strings.Contains(text, "12 Unicode scalar values") || !strings.Contains(text, "current length is 1.") {
		t.Fatal(text)
	}
}

func TestNaturalCheckerLegacyPlanAndCustomerReplacement(t *testing.T) {
	secret := "legacy-plan-private"
	c := naturalContext(t)
	legacy, err := ParseI18nCatalog([]byte(`{"schema":"teaql.i18n/v1","defaultLocale":"en","locales":{"en":{"messages":{"checker.minLength":"Private `+secret+`: {location} expects {system}; got {input_len}"}}}}`), builtinCatalog)
	if err != nil {
		t.Fatal(err)
	}
	_ = c.InstallI18nCatalog(legacy)
	values := []CheckResult{{RuleID: "min_length", CanonicalLocation: Location().Property("password"), InputValue: secret, SystemValue: 50}}
	c.TranslateCheckResultsFrom("Entry", values)
	failure := &RuntimeError{Type: "Check", CheckResults: []CheckResult{values[0].PrefixedBy(Location().Property("entry_list").At(1))}}
	rerenderGraphCheckerError(failure, &preparedGraphSnapshot{checkerMessages: c.captureCheckerMessages(), checkerOrigin: "Catalog"})
	want := fmt.Sprintf("Private [REDACTED]: EntryList[1].Password expects 50; got %d", len(secret))
	if failure.CheckResults[0].Message != want {
		t.Fatalf("got %q want %q", failure.CheckResults[0].Message, want)
	}
	failure.CheckResults[0].Message = "Customer replacement"
	c.TranslateCheckResults(failure.CheckResults)
	if failure.CheckResults[0].Message != "Customer replacement" {
		t.Fatal("customer text was overwritten")
	}
}

func TestNaturalCheckerRetainedSPIContainsOnlyProjectedLiteralNodes(t *testing.T) {
	secret := "retained-selection-private"
	c := naturalContext(t)
	if err := c.InstallCheckerMessageCatalog(naturalCatalog(`{"en":{"rules":{"checker.required":{"one":"Private ` + secret + ` {field}"}}}}`)); err != nil {
		t.Fatal(err)
	}
	var retained CheckerMessageRequest
	_ = c.InstallCheckerMessageRenderer(CheckerMessageRendererFunc(func(request CheckerMessageRequest) (CheckerMessageOutput, error) {
		retained = request
		return request.Presentation.Render(request.Catalog, request.Selection, request.Arguments), nil
	}))
	values := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("password"), InputValue: secret}}
	c.TranslateCheckResultsFrom("Entry", values)
	if values[0].Message != "Private [REDACTED] password" || retained.Selection.Render(map[string]string{"field": "password"}) != "Private [REDACTED] password" {
		t.Fatal("raw application template survived projection")
	}
}

func TestNaturalCheckerOriginUsesExplicitAncestryNotFirstWritableChild(t *testing.T) {
	c := naturalContext(t)
	c.startGraphPreparation()
	defer c.finishGraphPreparation()
	child := core.EntityKey{Entity: "Entry", ID: core.ValU64(1)}
	view := &preparedViewProbe{key: child, values: core.Record{}}
	if err := c.AdmitPreparedGraphEntityView(child, view); err != nil {
		t.Fatal(err)
	}
	if c.preparedGraphSnapshot.checkerOrigin != "" {
		t.Fatal("writable child was guessed to be root")
	}
	root := core.EntityKey{Entity: "Catalog", ID: core.ValU64(1)}
	if err := c.PrepareGraphMutationLineage(core.NewEntityRoot(), root, nil, graphTestIntent(), nil); err != nil {
		t.Fatal(err)
	}
	values := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("name")}}
	c.TranslateCheckResultsFrom("Entry", values)
	failure := &RuntimeError{Type: "Check", CheckResults: []CheckResult{values[0].PrefixedBy(Location().Property("entry_list").At(1))}}
	rerenderGraphCheckerError(failure, c.preparedGraphSnapshot)
	if failure.CheckResults[0].Message != "The name of the 2nd entry is required." {
		t.Fatal(failure.CheckResults[0].Message)
	}
}

func TestNaturalCheckerRetainedPlanReprojectsNewSensitiveValues(t *testing.T) {
	secret := "newly-sensitive-literal"
	c := naturalContext(t)
	if err := c.InstallCheckerMessageCatalog(naturalCatalog(`{"en":{"rules":{"checker.minLength":{"one":"{field} needs {system} {unit}; got {input_len}. Private ` + secret + `"}}}}`)); err != nil {
		t.Fatal(err)
	}
	values := []CheckResult{{RuleID: "min_length", CanonicalLocation: Location().Property("password"), InputValue: "old-private", SystemValue: 50}}
	c.TranslateCheckResultsFrom("Entry", values)
	if !strings.Contains(values[0].Message, secret) {
		t.Fatal("expected ordinary application text in initial invocation")
	}
	values[0].InputValue = secret
	c.TranslateCheckResults(values)
	if strings.Contains(values[0].Message, secret) || !strings.Contains(values[0].Message, "Private [REDACTED]") || !strings.Contains(values[0].Message, fmt.Sprint(len(secret))) || values[0].InputValue != "[REDACTED]" {
		t.Fatal("retained custom plan restored a newly sensitive value")
	}
}
func TestNaturalCheckerConcurrentOriginsAndRetainedMetadata(t *testing.T) {
	c := naturalContext(t)
	var wait sync.WaitGroup
	for _, origin := range []string{"Catalog", "Person"} {
		origin := origin
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := 0; i < 40; i++ {
				values := []CheckResult{{RuleID: "required", CanonicalLocation: Location()}}
				c.TranslateCheckResultsFrom(origin, values)
				want := "catalog"
				if origin == "Person" {
					want = "person"
				}
				if !strings.Contains(values[0].Message, want) {
					t.Error(values[0].Message)
				}
			}
		}()
	}
	wait.Wait()
	values := []CheckResult{{RuleID: "required", CanonicalLocation: Location().Property("name")}}
	c.TranslateCheckResultsFrom("Catalog", values)
	_ = c.InstallCheckerPresentationModel(EmptyCheckerPresentationModel())
	moved := []CheckResult{values[0].PrefixedBy(Location().Property("payer"))}
	c.TranslateCheckResults(moved)
	if moved[0].Message != "The name of the payer is required." {
		t.Fatal(moved[0].Message)
	}
}

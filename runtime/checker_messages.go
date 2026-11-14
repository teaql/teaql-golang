package runtime

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

type CheckerMessageOutput struct {
	Message  string
	Warnings []CheckerMessageWarning
}

// Renderer inputs contain projected scalars and immutable model data only.
// There is no callback retaining a Context or validation invocation.
type CheckerMessageRequest struct {
	Presentation *CheckerPresentation
	Catalog      *CheckerMessageCatalog
	Selection    CheckerMessageSelection
	Arguments    map[string]string
}
type CheckerMessageRenderer interface {
	RenderCheckerMessage(CheckerMessageRequest) (CheckerMessageOutput, error)
}
type CheckerMessageRendererFunc func(CheckerMessageRequest) (CheckerMessageOutput, error)

func (f CheckerMessageRendererFunc) RenderCheckerMessage(r CheckerMessageRequest) (CheckerMessageOutput, error) {
	return f(r)
}

type standardCheckerMessageRenderer struct{}

func (standardCheckerMessageRenderer) RenderCheckerMessage(r CheckerMessageRequest) (CheckerMessageOutput, error) {
	return r.Presentation.Render(r.Catalog, r.Selection, r.Arguments), nil
}

type checkerMessageConfig struct {
	model    *CheckerPresentationModel
	catalog  *CheckerMessageCatalog
	legacy   *I18nCatalog
	locale   Locale
	renderer CheckerMessageRenderer
	warning  func(CheckerMessageWarning)
}
type naturalCheckerPresentation struct {
	snapshot  *CheckerPresentation
	automatic [32]byte
	plan      *checkerMessagePlan
}

func naturalCheckerState(result CheckResult) (naturalCheckerPresentation, bool) {
	value, ok := logprivacy.ReadProjectionState(result.naturalPresentation).(naturalCheckerPresentation)
	return value, ok
}
func naturalCheckerAutomatic(result CheckResult) bool {
	if state, ok := naturalCheckerState(result); ok && state.automatic == sha256.Sum256([]byte(result.Message)) {
		return true
	}
	previous, ok := logprivacy.ReadProjectionState(result.messagePresentation).([32]byte)
	return ok && previous == sha256.Sum256([]byte(result.Message))
}
func bindNaturalCheckerOrigin(results []CheckResult, model *CheckerPresentationModel, origin string) {
	for i := range results {
		r := &results[i]
		state := naturalCheckerPresentation{snapshot: CaptureCheckerPresentation(model, origin, r.ObjectLocation())}
		if naturalCheckerAutomatic(*r) {
			state.automatic = sha256.Sum256([]byte(r.Message))
			if previous, ok := naturalCheckerState(*r); ok {
				state.plan = previous.plan
			}
		}
		r.naturalPresentation = logprivacy.NewProjectionState(state)
	}
}

func (c *UserContext) InstallCheckerMessageCatalog(source []byte) error {
	catalog, err := ParseCheckerMessageCatalog(source)
	if err != nil {
		return err
	}
	c.checkerMessageMu.Lock()
	c.checkerMessageCatalog = catalog
	c.checkerMessageMu.Unlock()
	return nil
}
func (c *UserContext) InstallCheckerPresentationModel(model *CheckerPresentationModel) error {
	if model == nil {
		return errors.New("Checker presentation model is required")
	}
	c.checkerMessageMu.Lock()
	c.checkerPresentationModel = model
	c.checkerMessageMu.Unlock()
	return nil
}
func (c *UserContext) InstallCheckerMessageRenderer(renderer CheckerMessageRenderer) error {
	if renderer == nil {
		return errors.New("Checker message renderer is required")
	}
	c.checkerMessageMu.Lock()
	c.checkerMessageRenderer = renderer
	c.checkerMessageMu.Unlock()
	return nil
}
func (c *UserContext) InstallCheckerMessageWarning(sink func(CheckerMessageWarning)) error {
	if sink == nil {
		return errors.New("Checker warning sink is required")
	}
	c.checkerMessageMu.Lock()
	c.checkerMessageWarning = sink
	c.checkerMessageMu.Unlock()
	return nil
}
func (c *UserContext) captureCheckerMessages() checkerMessageConfig {
	c.checkerMessageMu.RLock()
	defer c.checkerMessageMu.RUnlock()
	model, catalog, renderer, warning := c.checkerPresentationModel, c.checkerMessageCatalog, c.checkerMessageRenderer, c.checkerMessageWarning
	if model == nil {
		model = emptyCheckerPresentationModel
	}
	if catalog == nil {
		catalog = builtinCheckerMessages
	}
	if renderer == nil {
		renderer = standardCheckerMessageRenderer{}
	}
	if warning == nil {
		warning = func(code CheckerMessageWarning) { fmt.Fprintln(os.Stderr, string(code)) }
	}
	return checkerMessageConfig{model, catalog, c.I18nCatalog(), c.Language().(Locale), renderer, warning}
}

// TranslateCheckResultsFrom is for callers who own a known validation origin.
// It must not receive a declaring type guessed from an individual violation.
func (c *UserContext) TranslateCheckResultsFrom(origin string, results []CheckResult) {
	config := c.captureCheckerMessages()
	bindNaturalCheckerOrigin(results, config.model, origin)
	c.translateCheckerResults(results, "", nil, plaintextLogsEnabled(), config)
}
func checkerMessageRule(rule string) string {
	switch strings.ToUpper(rule) {
	case "REQUIRED":
		return "checker.required"
	case "MIN":
		return "checker.min"
	case "MAX":
		return "checker.max"
	case "MIN_STR_LEN", "MIN_LENGTH":
		return "checker.minLength"
	case "MAX_STR_LEN", "MAX_LENGTH":
		return "checker.maxLength"
	}
	return "checker." + strings.ToLower(rule)
}
func checkerMessageScalar(value any) string {
	if typed, ok := value.(core.Value); ok {
		return checkerMessageScalar(typed.V)
	}
	switch value := value.(type) {
	case nil:
		return "<nil>"
	case string:
		return sanitizeCheckerLabel(value)
	case bool:
		return strconv.FormatBool(value)
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64)
	case decimal.Decimal:
		return value.String()
	case time.Time:
		return value.Format(time.RFC3339Nano)
	case core.DataType:
		return "null"
	default:
		return "[value omitted]"
	}
}
func safeCheckerMessageRenderer(renderer CheckerMessageRenderer, request CheckerMessageRequest) (out CheckerMessageOutput, err error) {
	defer func() {
		if recover() != nil {
			out = CheckerMessageOutput{}
			err = errors.New("Checker renderer failed")
		}
	}()
	return renderer.RenderCheckerMessage(request)
}
func safeCheckerMessageWarning(sink func(CheckerMessageWarning), code CheckerMessageWarning) {
	defer func() { _ = recover() }()
	sink(code)
}
func renderNaturalChecker(config checkerMessageConfig, result *CheckResult, secrets []string, automatic bool, seen map[CheckerMessageWarning]bool, forced *CheckerPresentation) {
	if strings.TrimSpace(result.Message) != "" && !automatic {
		return
	}
	warn := func(code CheckerMessageWarning) {
		switch code {
		case CheckerMessageLabelFallback, CheckerMessageGenericRule, CheckerMessageRenderFallback:
		default:
			code = CheckerMessageRenderFallback
		}
		if !seen[code] {
			seen[code] = true
			safeCheckerMessageWarning(config.warning, code)
		}
	}
	presentation := forced
	if presentation == nil {
		if state, ok := naturalCheckerState(*result); ok {
			presentation = state.snapshot.Relocate(result.ObjectLocation())
		} else {
			presentation = CaptureCheckerPresentation(config.model, "", result.ObjectLocation())
		}
	}
	if ready, output, plan := renderPreparedCheckerPlan(config, result, presentation, automatic, secrets); ready {
		for _, code := range output.Warnings {
			warn(code)
		}
		state, _ := naturalCheckerState(*result)
		state.snapshot = presentation
		state.plan = plan
		state.automatic = sha256.Sum256([]byte(output.Message))
		result.Message = output.Message
		result.naturalPresentation = logprivacy.NewProjectionState(state)
		return
	}
	rule := checkerMessageRule(result.RuleID)
	var selection CheckerMessageSelection
	found := false
	for _, locale := range []Locale{config.locale, LocaleEnglish} {
		if selection, found = config.catalog.override(locale, rule, presentation.shape); found {
			break
		}
		if config.legacy != nil && config.legacy != builtinCatalog {
			if template := config.legacy.find(string(locale), rule); template != "" {
				args := map[string]string{"{location}": result.NativePath(), "{system}": checkerMessageScalar(result.SystemValue), "{input}": checkerMessageScalar(result.InputValue), "{input_len}": strconv.Itoa(checkerInputLength(*result))}
				var out strings.Builder
				last := 0
				for _, token := range checkerTemplateToken.FindAllStringIndex(template, -1) {
					out.WriteString(scrubCheckerText(template[last:token[0]], secrets))
					out.WriteString(args[template[token[0]:token[1]]])
					last = token[1]
				}
				out.WriteString(scrubCheckerText(template[last:], secrets))
				result.Message = out.String()
				plan := prepareCheckerMessagePlan(config, result, secrets)
				result.naturalPresentation = logprivacy.NewProjectionState(naturalCheckerPresentation{snapshot: presentation, automatic: sha256.Sum256([]byte(result.Message)), plan: plan})
				return
			}
		}
		if selection, found = builtinCheckerMessages.Lookup(locale, rule, presentation.shape); found {
			break
		}
	}
	if !found {
		warn(CheckerMessageGenericRule)
		selection = builtinCheckerMessages.Select(config.locale, "checker.invalid", presentation.shape)
	}
	args := map[string]string{"system": checkerMessageScalar(result.SystemValue), "input": checkerMessageScalar(result.InputValue), "input_len": strconv.Itoa(checkerInputLength(*result)), "unit": checkerPresentationWords[selection.locale]["scalar"]}
	var output CheckerMessageOutput
	var err error
	var plan *checkerMessagePlan
	if _, standard := config.renderer.(standardCheckerMessageRenderer); standard {
		plan = prepareCheckerMessagePlan(config, result, secrets)
		output = plan.render(presentation, result, config.locale)
	} else {
		if selection.application {
			selection.template = projectCheckerNodes(selection.template, secrets)
		}
		output, err = safeCheckerMessageRenderer(config.renderer, CheckerMessageRequest{presentation, safeCheckerPatterns(config.catalog, selection.locale, secrets), selection, args})
		output.Message = scrubCheckerText(output.Message, secrets)
	}
	if err != nil || strings.TrimSpace(output.Message) == "" {
		warn(CheckerMessageRenderFallback)
		output = presentation.Render(builtinCheckerMessages, builtinCheckerMessages.Select(config.locale, "checker.invalid", presentation.shape), nil)
	}
	for _, code := range output.Warnings {
		warn(code)
	}
	result.Message = output.Message
	if plan == nil {
		plan = prepareCheckerMessagePlan(config, result, secrets)
	}
	result.naturalPresentation = logprivacy.NewProjectionState(naturalCheckerPresentation{snapshot: presentation, automatic: sha256.Sum256([]byte(result.Message)), plan: plan})
}

func rerenderGraphCheckerError(err error, state *preparedGraphSnapshot) {
	if state == nil {
		return
	}
	var failure *RuntimeError
	if !errors.As(err, &failure) || failure.Type != "Check" {
		return
	}
	results := append([]CheckResult(nil), failure.CheckResults...)
	seen := map[CheckerMessageWarning]bool{}
	for i := range results {
		result := &results[i]
		if !naturalCheckerAutomatic(*result) {
			continue
		}
		snapshot := CaptureCheckerPresentation(state.checkerMessages.model, state.checkerOrigin, result.ObjectLocation())
		renderNaturalChecker(state.checkerMessages, result, nil, true, seen, snapshot)
	}
	failure.CheckResults = results
}

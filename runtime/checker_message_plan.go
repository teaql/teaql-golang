package runtime

import "strings"

// A constant-size, already-projected render plan. Graph relocation must not call
// a customer provider again after its transient raw redactions have been erased.
// These nodes contain safe literals only, never a closure over raw secrets.
type checkerMessagePlan struct {
	shapes  map[Locale]map[CheckerMessageShape]checkerPreparedMessage
	literal string
	custom  bool
}
type checkerPreparedMessage struct {
	selection CheckerMessageSelection
	catalog   *CheckerMessageCatalog
	legacy    bool
}

func projectCheckerNodes(nodes checkerMessageTemplate, secrets []string) checkerMessageTemplate {
	result := make(checkerMessageTemplate, len(nodes))
	for i, node := range nodes {
		result[i] = node
		result[i].literal = scrubCheckerText(node.literal, secrets)
	}
	return result
}
func safeCheckerPatterns(catalog *CheckerMessageCatalog, locale Locale, secrets []string) *CheckerMessageCatalog {
	if catalog == nil || !catalog.application || catalog.locales[locale].patterns == nil {
		return builtinCheckerMessages
	}
	original := catalog.locales[locale].patterns
	patterns := &checkerMessagePatterns{projectCheckerNodes(original.reference, secrets), projectCheckerNodes(original.item, secrets), scrubCheckerText(original.separator, secrets)}
	return &CheckerMessageCatalog{locales: map[Locale]checkerMessageLocale{locale: {patterns: patterns}}}
}
func prepareCheckerMessagePlan(config checkerMessageConfig, result *CheckResult, secrets []string) *checkerMessagePlan {
	plan := &checkerMessagePlan{shapes: map[Locale]map[CheckerMessageShape]checkerPreparedMessage{}}
	if _, standard := config.renderer.(standardCheckerMessageRenderer); !standard {
		plan.custom = true
		plan.literal = result.Message
		return plan
	}
	rule := checkerMessageRule(result.RuleID)
	for _, requested := range Locales {
		plan.shapes[requested] = map[CheckerMessageShape]checkerPreparedMessage{}
		for _, shape := range []CheckerMessageShape{CheckerMessageOne, CheckerMessageTwo, CheckerMessageDeep} {
			var selected CheckerMessageSelection
			found := false
			legacy := false
			for _, locale := range []Locale{requested, LocaleEnglish} {
				if selected, found = config.catalog.override(locale, rule, shape); found {
					break
				}
				if config.legacy != nil && config.legacy != builtinCatalog {
					if template := config.legacy.find(string(locale), rule); template != "" {
						nodes := checkerMessageTemplate{}
						last := 0
						for _, token := range checkerTemplateToken.FindAllStringIndex(template, -1) {
							nodes = append(nodes, checkerMessageNode{literal: scrubCheckerText(template[last:token[0]], secrets)}, checkerMessageNode{argument: template[token[0]+1 : token[1]-1]})
							last = token[1]
						}
						nodes = append(nodes, checkerMessageNode{literal: scrubCheckerText(template[last:], secrets)})
						selected = CheckerMessageSelection{locale: locale, template: nodes, application: true}
						legacy = true
						found = true
						break
					}
				}
				if selected, found = builtinCheckerMessages.Lookup(locale, rule, shape); found {
					break
				}
			}
			if !found {
				selected = builtinCheckerMessages.Select(requested, "checker.invalid", shape)
			}
			if selected.application {
				selected.template = projectCheckerNodes(selected.template, secrets)
			}
			plan.shapes[requested][shape] = checkerPreparedMessage{selected, safeCheckerPatterns(config.catalog, selected.locale, secrets), legacy}
		}
	}
	return plan
}
func (plan *checkerMessagePlan) render(snapshot *CheckerPresentation, result *CheckResult, locale Locale) CheckerMessageOutput {
	if plan.custom {
		return CheckerMessageOutput{Message: plan.literal}
	}
	item := plan.shapes[locale][snapshot.shape]
	args := map[string]string{"system": checkerMessageScalar(result.SystemValue), "input": checkerMessageScalar(result.InputValue), "input_len": checkerMessageScalar(checkerInputLength(*result)), "unit": checkerPresentationWords[item.selection.locale]["scalar"]}
	if item.legacy {
		args["location"] = result.NativePath()
		return CheckerMessageOutput{Message: item.selection.Render(args)}
	}
	return snapshot.Render(item.catalog, item.selection, args)
}

func renderPreparedCheckerPlan(config checkerMessageConfig, result *CheckResult, presentation *CheckerPresentation, automatic bool) (bool, CheckerMessageOutput) {
	if !automatic {
		return false, CheckerMessageOutput{}
	}
	state, ok := naturalCheckerState(*result)
	if !ok || state.plan == nil {
		return false, CheckerMessageOutput{}
	}
	// New validations get fresh plans. An already-projected customer message
	// lacks original redactions; preserve its safe plan rather than recovering
	// a raw template/provider output from a later mutable Context.
	output := state.plan.render(presentation, result, config.locale)
	if strings.TrimSpace(output.Message) == "" {
		return false, CheckerMessageOutput{}
	}
	return true, output
}

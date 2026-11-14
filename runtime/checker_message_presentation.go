package runtime

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
)

type CheckerMessageWarning string

const (
	CheckerMessageLabelFallback  CheckerMessageWarning = "CHECKER_MESSAGE_LABEL_FALLBACK"
	CheckerMessageGenericRule    CheckerMessageWarning = "CHECKER_MESSAGE_GENERIC_RULE"
	CheckerMessageRenderFallback CheckerMessageWarning = "CHECKER_MESSAGE_RENDER_FALLBACK"
)

type checkerPresentationField struct {
	name, target        string
	labels, classifiers map[Locale]string
	aliases             []string
}
type checkerPresentationEntity struct {
	labels map[Locale]string
	fields map[string]*checkerPresentationField
}

// CheckerPresentationModel contains only immutable descriptor/vocabulary data.
// Keep one shared instance per model revision, never one copy per entity row.
type CheckerPresentationModel struct {
	revision string
	entities map[string]*checkerPresentationEntity
	current  map[Locale]string
}

//go:embed natural-messages/presentation-defaults-v1.json
var checkerPresentationDefaultsJSON []byte
var checkerPresentationWords = loadCheckerPresentationWords()
var emptyCheckerPresentationModel = makeEmptyCheckerPresentationModel()

func loadCheckerPresentationWords() map[Locale]map[string]string {
	var data struct {
		Schema  string                       `json:"schema"`
		Locales map[Locale]map[string]string `json:"locales"`
	}
	if json.Unmarshal(checkerPresentationDefaultsJSON, &data) != nil || data.Schema != "teaql.checker-presentation/v1" {
		panic("invalid embedded Checker vocabulary")
	}
	for _, locale := range Locales {
		for _, key := range []string{"currentObject", "utf16", "scalar", "byte"} {
			value := data.Locales[locale][key]
			if strings.TrimSpace(value) == "" || sanitizeCheckerLabel(value) != value {
				panic("invalid embedded Checker vocabulary")
			}
		}
	}
	return data.Locales
}
func makeEmptyCheckerPresentationModel() *CheckerPresentationModel {
	current := map[Locale]string{}
	for _, locale := range Locales {
		current[locale] = checkerPresentationWords[locale]["currentObject"]
	}
	return &CheckerPresentationModel{revision: "unconfigured", entities: map[string]*checkerPresentationEntity{}, current: current}
}
func EmptyCheckerPresentationModel() *CheckerPresentationModel { return emptyCheckerPresentationModel }
func checkerLabels(value any) (map[Locale]string, error) {
	result := map[Locale]string{}
	if value == nil {
		return result, nil
	}
	data, err := checkerObject(value)
	if err != nil {
		return nil, err
	}
	for code, value := range data {
		locale, err := ParseLocale(code)
		if err != nil {
			return nil, err
		}
		if _, duplicate := result[locale]; duplicate {
			return nil, checkerMessageError()
		}
		label, err := checkerText(value)
		if err != nil {
			return nil, err
		}
		result[locale] = label
	}
	return result, nil
}
func ParseCheckerPresentationModel(source []byte) (*CheckerPresentationModel, error) {
	value, err := strictCheckerJSON(source)
	if err != nil {
		return nil, err
	}
	root, err := checkerObject(value)
	if err != nil {
		return nil, err
	}
	if !checkerKeys(root, "schema", "revision", "entities", "currentObject") || root["schema"] != "teaql.checker-model/v1" {
		return nil, checkerMessageError()
	}
	revision, err := checkerText(root["revision"])
	if err != nil {
		return nil, err
	}
	entries, err := checkerObject(root["entities"])
	if err != nil {
		return nil, err
	}
	current := map[Locale]string{}
	for locale, value := range emptyCheckerPresentationModel.current {
		current[locale] = value
	}
	custom, err := checkerLabels(root["currentObject"])
	if err != nil {
		return nil, err
	}
	for locale, value := range custom {
		current[locale] = value
	}
	model := &CheckerPresentationModel{revision: revision, entities: map[string]*checkerPresentationEntity{}, current: current}
	for identity, value := range entries {
		if _, err := checkerText(identity); err != nil {
			return nil, err
		}
		entry, err := checkerObject(value)
		if err != nil || !checkerKeys(entry, "labels", "fields") {
			return nil, checkerMessageError()
		}
		labels, err := checkerLabels(entry["labels"])
		if err != nil {
			return nil, err
		}
		fields, ok := entry["fields"].([]any)
		if !ok {
			return nil, checkerMessageError()
		}
		entity := &checkerPresentationEntity{labels: labels, fields: map[string]*checkerPresentationField{}}
		for _, value := range fields {
			field, err := checkerObject(value)
			if err != nil || !checkerKeys(field, "name", "targetType", "labels", "classifiers", "aliases") {
				return nil, checkerMessageError()
			}
			name, err := checkerText(field["name"])
			if err != nil {
				return nil, err
			}
			target := ""
			if value, present := field["targetType"]; present {
				target, err = checkerText(value)
				if err != nil {
					return nil, err
				}
			}
			labels, err := checkerLabels(field["labels"])
			if err != nil {
				return nil, err
			}
			classifiers, err := checkerLabels(field["classifiers"])
			if err != nil {
				return nil, err
			}
			aliases := []string{}
			if value, present := field["aliases"]; present {
				values, ok := value.([]any)
				if !ok {
					return nil, checkerMessageError()
				}
				for _, value := range values {
					alias, err := checkerText(value)
					if err != nil {
						return nil, err
					}
					aliases = append(aliases, alias)
				}
			}
			descriptor := &checkerPresentationField{name, target, labels, classifiers, aliases}
			for _, key := range append([]string{name}, aliases...) {
				if _, duplicate := entity.fields[key]; duplicate {
					return nil, checkerMessageError()
				}
				entity.fields[key] = descriptor
			}
		}
		model.entities[identity] = entity
	}
	return model, nil
}
func (m *CheckerPresentationModel) Revision() string {
	if m == nil {
		return "unconfigured"
	}
	return m.revision
}
func (m *CheckerPresentationModel) Merge(other *CheckerPresentationModel) (*CheckerPresentationModel, error) {
	if m == nil || m == emptyCheckerPresentationModel {
		if other == nil {
			return emptyCheckerPresentationModel, nil
		}
		return other, nil
	}
	if other == nil || other == emptyCheckerPresentationModel || m == other {
		return m, nil
	}
	entities := map[string]*checkerPresentationEntity{}
	for key, value := range m.entities {
		entities[key] = value
	}
	changed := false
	for key, value := range other.entities {
		if previous, exists := entities[key]; exists {
			if !reflect.DeepEqual(previous, value) {
				return nil, checkerMessageError()
			}
		} else {
			entities[key] = value
			changed = true
		}
	}
	if !reflect.DeepEqual(m.current, other.current) {
		return nil, checkerMessageError()
	}
	if !changed {
		return m, nil
	}
	return &CheckerPresentationModel{m.revision + "+" + other.revision, entities, m.current}, nil
}

type checkerPresentationStep struct {
	identity string
	field    *checkerPresentationField
	indices  []int
}

// CheckerPresentation retains a finite path and model, never a Context or graph.
type CheckerPresentation struct {
	model  *CheckerPresentationModel
	origin string
	steps  []checkerPresentationStep
	shape  CheckerMessageShape
}

func CaptureCheckerPresentation(model *CheckerPresentationModel, origin string, location ObjectLocation) *CheckerPresentation {
	if model == nil {
		model = emptyCheckerPresentationModel
	}
	entity := model.entities[origin]
	unusual := entity == nil
	steps := []checkerPresentationStep{}
	for _, segment := range location.Segments {
		if segment.Property != nil {
			var descriptor *checkerPresentationField
			if entity != nil {
				descriptor = entity.fields[*segment.Property]
			}
			entity = nil
			if descriptor != nil {
				entity = model.entities[descriptor.target]
			}
			steps = append(steps, checkerPresentationStep{identity: *segment.Property, field: descriptor})
		} else if segment.Index != nil && *segment.Index >= 0 {
			if len(steps) == 0 {
				steps = append(steps, checkerPresentationStep{})
				unusual = true
			}
			last := &steps[len(steps)-1]
			last.indices = append(last.indices, *segment.Index)
			if len(last.indices) > 1 {
				unusual = true
			}
		} else {
			unusual = true
		}
	}
	if len(steps) == 0 || len(steps[len(steps)-1].indices) > 0 {
		unusual = true
	}
	shape := CheckerMessageTwo
	if unusual || len(steps) > 2 {
		shape = CheckerMessageDeep
	} else if len(steps) == 1 {
		shape = CheckerMessageOne
	}
	return &CheckerPresentation{model, origin, steps, shape}
}
func (p *CheckerPresentation) Shape() CheckerMessageShape { return p.shape }
func (p *CheckerPresentation) ModelRevision() string      { return p.model.revision }
func (p *CheckerPresentation) Relocate(location ObjectLocation) *CheckerPresentation {
	return CaptureCheckerPresentation(p.model, p.origin, location)
}
func sanitizeCheckerLabel(label string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029' {
			return ' '
		}
		return r
	}, label))
}
func (p *CheckerPresentation) Render(catalog *CheckerMessageCatalog, selection CheckerMessageSelection, safe map[string]string) CheckerMessageOutput {
	return p.render(catalog, selection, safe, nil)
}
func (p *CheckerPresentation) render(catalog *CheckerMessageCatalog, selection CheckerMessageSelection, safe map[string]string, project func(string) string) CheckerMessageOutput {
	locale := selection.locale
	warnings := map[CheckerMessageWarning]bool{}
	vocabulary := func(labels map[Locale]string, identity string) string {
		label := labels[locale]
		if strings.TrimSpace(label) == "" {
			warnings[CheckerMessageLabelFallback] = true
			label = labels[LocaleEnglish]
		}
		if strings.TrimSpace(label) == "" {
			label = strings.ReplaceAll(identity, "_", " ")
		}
		result := sanitizeCheckerLabel(label)
		if result == "" {
			warnings[CheckerMessageLabelFallback] = true
			return checkerPresentationWords[locale]["currentObject"]
		}
		return result
	}
	current := func() string { return vocabulary(p.model.current, "current object") }
	root := current()
	if entity := p.model.entities[p.origin]; entity != nil {
		root = vocabulary(entity.labels, p.origin)
	}
	label := func(step checkerPresentationStep) string {
		if step.field != nil {
			return vocabulary(step.field.labels, step.field.name)
		}
		return vocabulary(nil, step.identity)
	}
	args := map[string]string{}
	for _, key := range []string{"system", "input", "input_len", "unit"} {
		if value, ok := safe[key]; ok {
			args[key] = value
		}
	}
	leaf := checkerPresentationStep{}
	if len(p.steps) > 0 {
		leaf = p.steps[len(p.steps)-1]
	}
	args["field"] = root
	if leaf.identity != "" {
		args["field"] = label(leaf)
	}
	owners := []string{}
	count := len(p.steps)
	if leaf.identity != "" && len(leaf.indices) == 0 {
		count--
	}
	for _, step := range p.steps[:count] {
		relation := root
		if step.identity != "" {
			relation = label(step)
		}
		if len(step.indices) == 0 {
			owners = append(owners, catalog.reference(locale, relation))
		} else {
			for offset, index := range step.indices {
				name := relation
				if offset > 0 {
					name = current()
				}
				classifier := ""
				if locale == LocaleChineseSimplified || locale == LocaleChineseTraditional {
					classifier = "项"
				}
				if step.field != nil {
					if value, ok := step.field.classifiers[locale]; ok {
						classifier = value
					}
				}
				owners = append(owners, catalog.item(locale, name, index, sanitizeCheckerLabel(classifier)))
			}
		}
	}
	args["object"], args["objectPath"] = root, root
	if len(owners) > 0 {
		args["object"] = owners[0]
		args["objectPath"] = strings.Join(owners, catalog.patterns(locale).separator)
	}
	message := renderCheckerTemplate(selection.template, args, project)
	codes := []CheckerMessageWarning{}
	for _, code := range []CheckerMessageWarning{CheckerMessageLabelFallback, CheckerMessageGenericRule, CheckerMessageRenderFallback} {
		if warnings[code] {
			codes = append(codes, code)
		}
	}
	return CheckerMessageOutput{Message: message, Warnings: codes}
}

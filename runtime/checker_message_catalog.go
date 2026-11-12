package runtime

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

type CheckerMessageShape string

const (
	CheckerMessageOne  CheckerMessageShape = "one"
	CheckerMessageTwo  CheckerMessageShape = "two"
	CheckerMessageDeep CheckerMessageShape = "deep"
)

type checkerMessageNode struct{ literal, argument string }
type checkerMessageTemplate []checkerMessageNode
type checkerMessagePatterns struct {
	reference, item checkerMessageTemplate
	separator       string
}
type checkerMessageLocale struct {
	rules    map[string]map[CheckerMessageShape]checkerMessageTemplate
	patterns *checkerMessagePatterns
}

// CheckerMessageCatalog is immutable after parsing. Its maps never escape.
type CheckerMessageCatalog struct {
	locales     map[Locale]checkerMessageLocale
	application bool
}
type CheckerMessageSelection struct {
	locale      Locale
	template    checkerMessageTemplate
	application bool
}

func (s CheckerMessageSelection) Locale() Locale { return s.locale }
func (s CheckerMessageSelection) Render(args map[string]string) string {
	return renderCheckerTemplate(s.template, args, nil)
}

//go:embed natural-messages/builtin-v1.json
var naturalMessagesJSON []byte
var builtinCheckerMessages = mustCheckerMessageCatalog(naturalMessagesJSON)

func BuiltinCheckerMessageCatalog() *CheckerMessageCatalog { return builtinCheckerMessages }
func checkerMessageError() error                           { return errors.New("invalid Checker message catalog") }

// Decode tokens before constructing a map, rejecting duplicates at every depth.
func strictCheckerJSON(source []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	var read func() (any, error)
	read = func() (any, error) {
		token, err := decoder.Token()
		if err != nil {
			return nil, checkerMessageError()
		}
		if delim, ok := token.(json.Delim); ok {
			switch delim {
			case '{':
				values := map[string]any{}
				for decoder.More() {
					key, err := decoder.Token()
					if err != nil {
						return nil, checkerMessageError()
					}
					name, ok := key.(string)
					if !ok {
						return nil, checkerMessageError()
					}
					if _, duplicate := values[name]; duplicate {
						return nil, checkerMessageError()
					}
					value, err := read()
					if err != nil {
						return nil, err
					}
					values[name] = value
				}
				close, err := decoder.Token()
				if err != nil || close != json.Delim('}') {
					return nil, checkerMessageError()
				}
				return values, nil
			case '[':
				values := []any{}
				for decoder.More() {
					value, err := read()
					if err != nil {
						return nil, err
					}
					values = append(values, value)
				}
				close, err := decoder.Token()
				if err != nil || close != json.Delim(']') {
					return nil, checkerMessageError()
				}
				return values, nil
			default:
				return nil, checkerMessageError()
			}
		}
		return token, nil
	}
	result, err := read()
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, checkerMessageError()
	}
	return result, nil
}
func checkerObject(value any) (map[string]any, error) {
	result, ok := value.(map[string]any)
	if !ok {
		return nil, checkerMessageError()
	}
	return result, nil
}
func checkerText(value any) (string, error) {
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", checkerMessageError()
	}
	return text, nil
}
func checkerKeys(value map[string]any, allowed ...string) bool {
	for key := range value {
		found := false
		for _, name := range allowed {
			if key == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func compileCheckerTemplate(value any, allowed, required []string) (checkerMessageTemplate, error) {
	source, err := checkerText(value)
	if err != nil {
		return nil, err
	}
	nodes := checkerMessageTemplate{}
	var literal strings.Builder
	seen := map[string]bool{}
	for offset := 0; offset < len(source); {
		char := source[offset]
		if (char == '{' || char == '}') && offset+1 < len(source) && source[offset+1] == char {
			literal.WriteByte(char)
			offset += 2
			continue
		}
		if char == '}' {
			return nil, checkerMessageError()
		}
		if char != '{' {
			literal.WriteByte(char)
			offset++
			continue
		}
		end := strings.IndexByte(source[offset+1:], '}')
		if end < 0 {
			return nil, checkerMessageError()
		}
		end += offset + 1
		name := source[offset+1 : end]
		valid := false
		for _, arg := range allowed {
			if arg == name {
				valid = true
				break
			}
		}
		if !valid {
			return nil, checkerMessageError()
		}
		if literal.Len() > 0 {
			nodes = append(nodes, checkerMessageNode{literal: literal.String()})
			literal.Reset()
		}
		nodes = append(nodes, checkerMessageNode{argument: name})
		seen[name] = true
		offset = end + 1
	}
	if literal.Len() > 0 {
		nodes = append(nodes, checkerMessageNode{literal: literal.String()})
	}
	for _, name := range required {
		if !seen[name] {
			return nil, checkerMessageError()
		}
	}
	return nodes, nil
}
func renderCheckerTemplate(template checkerMessageTemplate, args map[string]string, project func(string) string) string {
	var out strings.Builder
	for _, node := range template {
		if node.argument != "" {
			value, ok := args[node.argument]
			if !ok {
				value = "-"
			}
			out.WriteString(value)
		} else {
			value := node.literal
			if project != nil {
				value = project(value)
			}
			out.WriteString(value)
		}
	}
	return out.String()
}
func ParseCheckerMessageCatalog(source []byte) (*CheckerMessageCatalog, error) {
	return parseCheckerMessageCatalog(source, false)
}
func mustCheckerMessageCatalog(source []byte) *CheckerMessageCatalog {
	catalog, err := parseCheckerMessageCatalog(source, true)
	if err != nil {
		panic("invalid embedded natural Checker catalog")
	}
	return catalog
}
func parseCheckerMessageCatalog(source []byte, complete bool) (*CheckerMessageCatalog, error) {
	value, err := strictCheckerJSON(source)
	if err != nil {
		return nil, err
	}
	root, err := checkerObject(value)
	if err != nil {
		return nil, err
	}
	if !checkerKeys(root, "schema", "defaultLocale", "locales") || root["schema"] != "teaql.checker-messages/v1" || root["defaultLocale"] != "en" {
		return nil, checkerMessageError()
	}
	entries, err := checkerObject(root["locales"])
	if err != nil {
		return nil, err
	}
	catalog := &CheckerMessageCatalog{locales: map[Locale]checkerMessageLocale{}, application: !complete}
	for code, value := range entries {
		locale, err := ParseLocale(code)
		if err != nil {
			return nil, err
		}
		if _, duplicate := catalog.locales[locale]; duplicate {
			return nil, checkerMessageError()
		}
		entry, err := checkerObject(value)
		if err != nil || !checkerKeys(entry, "rules", "locationPatterns") {
			return nil, checkerMessageError()
		}
		data := checkerMessageLocale{rules: map[string]map[CheckerMessageShape]checkerMessageTemplate{}}
		rules := map[string]any{}
		if value, present := entry["rules"]; present {
			rules, err = checkerObject(value)
			if err != nil {
				return nil, err
			}
		}
		for rule, value := range rules {
			if !strings.HasPrefix(rule, "checker.") || len(rule) == 8 {
				return nil, checkerMessageError()
			}
			shapes, err := checkerObject(value)
			if err != nil {
				return nil, err
			}
			templates := map[CheckerMessageShape]checkerMessageTemplate{}
			for name, text := range shapes {
				shape := CheckerMessageShape(name)
				if shape != CheckerMessageOne && shape != CheckerMessageTwo && shape != CheckerMessageDeep {
					return nil, checkerMessageError()
				}
				allowed, required := []string{"field", "system", "input", "input_len", "unit"}, []string{"field"}
				switch rule {
				case "checker.required", "checker.invalid":
					allowed = []string{"field"}
				case "checker.min", "checker.max":
					allowed = []string{"field", "system", "input"}
					required = append(required, "system")
				case "checker.minLength", "checker.maxLength":
					allowed = []string{"field", "system", "input_len", "unit"}
					required = append(required, "system", "input_len", "unit")
				}
				if shape != CheckerMessageOne {
					owner := "object"
					if shape == CheckerMessageDeep {
						owner = "objectPath"
					}
					allowed = append(allowed, owner)
					required = append(required, owner)
				}
				template, err := compileCheckerTemplate(text, allowed, required)
				if err != nil {
					return nil, err
				}
				templates[shape] = template
			}
			data.rules[rule] = templates
		}
		if value, present := entry["locationPatterns"]; present {
			patterns, err := checkerObject(value)
			if err != nil || !checkerKeys(patterns, "reference", "item", "separator") {
				return nil, checkerMessageError()
			}
			reference, err := compileCheckerTemplate(patterns["reference"], []string{"relation"}, []string{"relation"})
			if err != nil {
				return nil, err
			}
			item, err := compileCheckerTemplate(patterns["item"], []string{"relation", "position", "ordinal", "classifier"}, []string{"relation"})
			if err != nil {
				return nil, err
			}
			positioned := false
			for _, node := range item {
				if node.argument == "ordinal" || node.argument == "position" {
					positioned = true
				}
			}
			if !positioned {
				return nil, checkerMessageError()
			}
			separator, err := checkerText(patterns["separator"])
			if err != nil || strings.ContainsAny(separator, "{}") || strings.IndexFunc(separator, unicode.IsControl) >= 0 {
				return nil, checkerMessageError()
			}
			data.patterns = &checkerMessagePatterns{reference, item, separator}
		}
		catalog.locales[locale] = data
	}
	if complete {
		for _, locale := range Locales {
			entry, ok := catalog.locales[locale]
			if !ok || entry.patterns == nil {
				return nil, checkerMessageError()
			}
			for _, rule := range []string{"checker.required", "checker.min", "checker.max", "checker.minLength", "checker.maxLength", "checker.invalid"} {
				for _, shape := range []CheckerMessageShape{CheckerMessageOne, CheckerMessageTwo, CheckerMessageDeep} {
					if len(entry.rules[rule][shape]) == 0 {
						return nil, checkerMessageError()
					}
				}
			}
		}
	}
	return catalog, nil
}
func (c *CheckerMessageCatalog) Lookup(locale Locale, rule string, shape CheckerMessageShape) (CheckerMessageSelection, bool) {
	if c == nil {
		return CheckerMessageSelection{}, false
	}
	template, ok := c.locales[locale].rules[rule][shape]
	return CheckerMessageSelection{locale, template, c.application}, ok
}
func (c *CheckerMessageCatalog) override(locale Locale, rule string, shape CheckerMessageShape) (CheckerMessageSelection, bool) {
	if c == nil || !c.application {
		return CheckerMessageSelection{}, false
	}
	return c.Lookup(locale, rule, shape)
}
func (c *CheckerMessageCatalog) Select(locale Locale, rule string, shape CheckerMessageShape) CheckerMessageSelection {
	for _, candidate := range []Locale{locale, LocaleEnglish} {
		if selected, ok := c.Lookup(candidate, rule, shape); ok {
			return selected
		}
		if selected, ok := builtinCheckerMessages.Lookup(candidate, rule, shape); ok {
			return selected
		}
	}
	selected, _ := builtinCheckerMessages.Lookup(locale, "checker.invalid", shape)
	return selected
}
func (c *CheckerMessageCatalog) patterns(locale Locale) *checkerMessagePatterns {
	if c != nil {
		if p := c.locales[locale].patterns; p != nil {
			return p
		}
	}
	return builtinCheckerMessages.locales[locale].patterns
}
func (c *CheckerMessageCatalog) reference(locale Locale, relation string) string {
	return renderCheckerTemplate(c.patterns(locale).reference, map[string]string{"relation": relation}, nil)
}
func (c *CheckerMessageCatalog) item(locale Locale, relation string, index int, classifier string) string {
	// A big integer avoids overflow at the native index ceiling.
	position := new(big.Int).SetInt64(int64(index))
	position.Add(position, big.NewInt(1))
	value := position.String()
	return renderCheckerTemplate(c.patterns(locale).item, map[string]string{"relation": relation, "position": value, "ordinal": checkerOrdinal(value), "classifier": classifier}, nil)
}
func checkerOrdinal(position string) string {
	suffix := "th"
	last := position
	if len(last) > 2 {
		last = last[len(last)-2:]
	}
	tail, _ := strconv.Atoi(last)
	if tail < 11 || tail > 13 {
		switch tail % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return position + suffix
}

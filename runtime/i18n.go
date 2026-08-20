package runtime

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

type Locale string

const (
	LocaleEnglish            Locale = "en"
	LocaleChineseSimplified  Locale = "zh-CN"
	LocaleChineseTraditional Locale = "zh-TW"
	LocaleJapanese           Locale = "ja"
	LocaleKorean             Locale = "ko"
	LocaleGerman             Locale = "de"
	LocaleFrench             Locale = "fr"
	LocaleSpanish            Locale = "es"
	LocalePortuguese         Locale = "pt"
	LocaleArabic             Locale = "ar"
	LocaleThai               Locale = "th"
	LocaleIndonesian         Locale = "id"
	LocaleFilipino           Locale = "fil"
	LocaleUkrainian          Locale = "uk"
	LocaleVietnamese         Locale = "vi"
)

var Locales = []Locale{LocaleEnglish, LocaleChineseSimplified, LocaleChineseTraditional, LocaleJapanese, LocaleKorean, LocaleGerman, LocaleFrench, LocaleSpanish, LocalePortuguese, LocaleArabic, LocaleThai, LocaleIndonesian, LocaleFilipino, LocaleUkrainian, LocaleVietnamese}

type UnsupportedLocaleError struct{ Code string }

func (e *UnsupportedLocaleError) Error() string { return fmt.Sprintf("unsupported locale: %s", e.Code) }

var localeAliases = map[string]Locale{"en-us": LocaleEnglish, "en-gb": LocaleEnglish, "zh": LocaleChineseSimplified, "zh-hans": LocaleChineseSimplified, "zh-sg": LocaleChineseSimplified, "cn": LocaleChineseSimplified, "zh-hant": LocaleChineseTraditional, "zh-hk": LocaleChineseTraditional, "zh-mo": LocaleChineseTraditional, "tw": LocaleChineseTraditional, "ja-jp": LocaleJapanese, "ko-kr": LocaleKorean, "de-de": LocaleGerman, "fr-fr": LocaleFrench, "es-mx": LocaleSpanish, "pt-br": LocalePortuguese, "pt-pt": LocalePortuguese, "ar-sa": LocaleArabic, "th-th": LocaleThai, "id-id": LocaleIndonesian, "tl": LocaleFilipino, "fil-ph": LocaleFilipino, "uk-ua": LocaleUkrainian, "vi-vn": LocaleVietnamese}

func ParseLocale(code string) (Locale, error) {
	n := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(code), "_", "-"))
	for _, l := range Locales {
		if strings.ToLower(string(l)) == n {
			return l, nil
		}
	}
	if l, ok := localeAliases[n]; ok {
		return l, nil
	}
	return "", &UnsupportedLocaleError{code}
}

type catalogLocale struct {
	Messages   map[string]string `json:"messages"`
	Vocabulary map[string]string `json:"vocabulary"`
}
type I18nCatalog struct {
	Schema        string                   `json:"schema"`
	DefaultLocale string                   `json:"defaultLocale"`
	Locales       map[string]catalogLocale `json:"locales"`
	fallback      *I18nCatalog
}

//go:embed builtin-messages-v1.json
var builtinCatalogJSON []byte
var builtinCatalog = mustCatalog(builtinCatalogJSON, nil)

func mustCatalog(data []byte, fallback *I18nCatalog) *I18nCatalog {
	var c I18nCatalog
	if err := json.Unmarshal(data, &c); err != nil || c.Schema != "teaql.i18n/v1" {
		panic("invalid built-in i18n catalog")
	}
	c.fallback = fallback
	return &c
}
func ParseI18nCatalog(data []byte, fallback *I18nCatalog) (*I18nCatalog, error) {
	var c I18nCatalog
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Schema != "teaql.i18n/v1" {
		return nil, fmt.Errorf("unsupported i18n schema")
	}
	for code := range c.Locales {
		if _, err := ParseLocale(code); err != nil {
			return nil, err
		}
	}
	c.fallback = fallback
	return &c, nil
}
func BuiltinI18nCatalog() *I18nCatalog { return builtinCatalog }
func (c *I18nCatalog) find(code, key string) string {
	if l, ok := c.Locales[code]; ok {
		return l.Messages[key]
	}
	return ""
}
func (c *I18nCatalog) Message(locale Locale, key string) string {
	if v := c.find(string(locale), key); v != "" {
		return v
	}
	if c.fallback != nil {
		if v := c.fallback.find(string(locale), key); v != "" {
			return v
		}
	}
	if v := c.find("en", key); v != "" {
		return v
	}
	if c.fallback != nil {
		if v := c.fallback.find("en", key); v != "" {
			return v
		}
	}
	return key
}
func (c *I18nCatalog) Translate(result *CheckResult, locale Locale) {
	keys := map[string]string{"REQUIRED": "checker.required", "MIN": "checker.min", "MAX": "checker.max", "MIN_STR_LEN": "checker.minLength", "MIN_LENGTH": "checker.minLength", "MAX_STR_LEN": "checker.maxLength", "MAX_LENGTH": "checker.maxLength"}
	key := keys[strings.ToUpper(result.RuleID)]
	if key == "" {
		key = "checker." + strings.ToLower(result.RuleID)
	}
	msg := c.Message(locale, key)
	values := map[string]string{"location": result.Location, "system": fmt.Sprint(result.SystemValue), "input": fmt.Sprint(result.InputValue), "input_len": "0"}
	if s, ok := result.InputValue.(string); ok {
		values["input_len"] = fmt.Sprint(len([]rune(s)))
	}
	for k, v := range values {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	result.Message = msg
}

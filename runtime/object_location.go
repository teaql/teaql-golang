package runtime

import (
	"strconv"
	"strings"
	"unicode"
)

type JsonFieldNamingProfile string

const (
	JsonFieldCamelCase  JsonFieldNamingProfile = "camelCase"
	JsonFieldSnakeCase  JsonFieldNamingProfile = "snake_case"
	JsonFieldPascalCase JsonFieldNamingProfile = "PascalCase"
)

func (p JsonFieldNamingProfile) Render(name string) string {
	if p == JsonFieldSnakeCase {
		return name
	}
	if p == JsonFieldPascalCase {
		return upperCamel(name)
	}
	return lowerCamel(name)
}

type ObjectLocationSegment struct {
	Property *string
	Index    *int
}

// ObjectLocation keeps canonical KSML names and renders them for each boundary.
type ObjectLocation struct{ Segments []ObjectLocationSegment }

func Location() ObjectLocation { return ObjectLocation{} }
func LocationFromModelPath(path string) ObjectLocation {
	location := Location()
	for cursor := 0; cursor < len(path); {
		if path[cursor] == '.' {
			cursor++
			continue
		}
		if path[cursor] == '[' {
			end := strings.IndexByte(path[cursor:], ']')
			if end < 0 {
				location = location.Property(path[cursor:])
				break
			}
			end += cursor
			if index, err := strconv.Atoi(path[cursor+1 : end]); err == nil {
				location = location.At(index)
			}
			cursor = end + 1
			continue
		}
		end := cursor
		for end < len(path) && path[end] != '.' && path[end] != '[' {
			end++
		}
		if end > cursor {
			location = location.Property(path[cursor:end])
		}
		cursor = end
	}
	return location
}
func (l ObjectLocation) Property(name string) ObjectLocation {
	segments := append([]ObjectLocationSegment{}, l.Segments...)
	segments = append(segments, ObjectLocationSegment{Property: &name})
	return ObjectLocation{Segments: segments}
}
func (l ObjectLocation) At(index int) ObjectLocation {
	segments := append([]ObjectLocationSegment{}, l.Segments...)
	segments = append(segments, ObjectLocationSegment{Index: &index})
	return ObjectLocation{Segments: segments}
}
func (l ObjectLocation) PrefixedBy(prefix ObjectLocation) ObjectLocation {
	segments := append([]ObjectLocationSegment{}, prefix.Segments...)
	segments = append(segments, l.Segments...)
	return ObjectLocation{Segments: segments}
}
func (l ObjectLocation) ModelPath() string  { return l.render(func(s string) string { return s }) }
func (l ObjectLocation) NativePath() string { return l.render(upperCamel) }
func (l ObjectLocation) InstancePath() string {
	return l.InstancePathWith(JsonFieldCamelCase)
}
func (l ObjectLocation) InstancePathWith(profile JsonFieldNamingProfile) string {
	parts := make([]string, 0, len(l.Segments))
	for _, segment := range l.Segments {
		value := ""
		if segment.Property != nil {
			value = profile.Render(*segment.Property)
		} else {
			value = strconv.Itoa(*segment.Index)
		}
		value = strings.ReplaceAll(strings.ReplaceAll(value, "~", "~0"), "/", "~1")
		parts = append(parts, value)
	}
	if len(parts) == 0 {
		return ""
	}
	return "/" + strings.Join(parts, "/")
}
func (l ObjectLocation) String() string { return l.NativePath() }
func (l ObjectLocation) render(transform func(string) string) string {
	var result strings.Builder
	for _, segment := range l.Segments {
		if segment.Property != nil {
			if result.Len() > 0 {
				result.WriteByte('.')
			}
			result.WriteString(transform(*segment.Property))
		} else {
			result.WriteString("[")
			result.WriteString(strconv.Itoa(*segment.Index))
			result.WriteString("]")
		}
	}
	return result.String()
}
func lowerCamel(value string) string {
	result := upperCamel(value)
	if result == "" {
		return result
	}
	chars := []rune(result)
	chars[0] = unicode.ToLower(chars[0])
	return string(chars)
}
func upperCamel(value string) string {
	parts := strings.Split(value, "_")
	for i, p := range parts {
		if p != "" {
			chars := []rune(p)
			chars[0] = unicode.ToUpper(chars[0])
			parts[i] = string(chars)
		}
	}
	return strings.Join(parts, "")
}

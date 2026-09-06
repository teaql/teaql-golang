package runtime

import (
	"fmt"
	"strings"
)

type WireFieldMetadata struct {
	CanonicalName string
	WireName      string
	Aliases       []string
}
type WireEntityMetadata struct {
	EntityType string
	Profile    JsonFieldNamingProfile
	Fields     map[string]WireFieldMetadata
}
type NormalizedWireInput struct {
	Values              map[string]any
	SourceInstancePaths map[string]string
}
type WireInputError struct{ Code, InstancePath, Message string }

func (e *WireInputError) Error() string { return e.Message }
func CreateWireEntityMetadata(entityType string, canonicalFields []string, profile JsonFieldNamingProfile, aliases map[string][]string) (WireEntityMetadata, error) {
	fields := map[string]WireFieldMetadata{}
	spellings := map[string]string{}
	for _, canonical := range canonicalFields {
		field := WireFieldMetadata{canonical, profile.Render(canonical), aliases[canonical]}
		for _, spelling := range append([]string{field.WireName}, field.Aliases...) {
			if previous, ok := spellings[spelling]; ok && previous != canonical {
				return WireEntityMetadata{}, fmt.Errorf("wire field spelling %q maps to both %q and %q", spelling, previous, canonical)
			}
			spellings[spelling] = canonical
		}
		fields[canonical] = field
	}
	return WireEntityMetadata{entityType, profile, fields}, nil
}
func MustCreateWireEntityMetadata(entityType string, canonicalFields []string, profile JsonFieldNamingProfile, aliases map[string][]string) WireEntityMetadata {
	metadata, err := CreateWireEntityMetadata(entityType, canonicalFields, profile, aliases)
	if err != nil {
		panic(err)
	}
	return metadata
}
func MustCreateWireEntityMetadataWithCanonicalAliases(entityType string, canonicalFields []string, profile JsonFieldNamingProfile) WireEntityMetadata {
	aliases := map[string][]string{}
	for _, field := range canonicalFields {
		aliases[field] = []string{field}
	}
	return MustCreateWireEntityMetadata(entityType, canonicalFields, profile, aliases)
}
func NormalizeWireInput(input map[string]any, metadata WireEntityMetadata, parentPointer string) (NormalizedWireInput, error) {
	lookup := map[string]WireFieldMetadata{}
	for _, field := range metadata.Fields {
		lookup[field.WireName] = field
		for _, alias := range field.Aliases {
			lookup[alias] = field
		}
	}
	values := map[string]any{}
	paths := map[string]string{}
	submitted := map[string]string{}
	for name, value := range input {
		pointer := parentPointer + "/" + strings.ReplaceAll(strings.ReplaceAll(name, "~", "~0"), "/", "~1")
		field, ok := lookup[name]
		if !ok {
			return NormalizedWireInput{}, &WireInputError{"WIRE_UNKNOWN_FIELD", pointer, fmt.Sprintf("Unknown %s field '%s'", metadata.EntityType, name)}
		}
		if previous, ok := submitted[field.CanonicalName]; ok {
			return NormalizedWireInput{}, &WireInputError{"WIRE_FIELD_COLLISION", pointer, fmt.Sprintf("Fields '%s' and '%s' both map to canonical field '%s'", previous, name, field.CanonicalName)}
		}
		submitted[field.CanonicalName] = name
		values[field.CanonicalName] = value
		if name != field.WireName {
			paths[field.CanonicalName] = pointer
		}
	}
	return NormalizedWireInput{values, paths}, nil
}
func RetainSubmittedPaths(results []CheckResult, normalized NormalizedWireInput) []CheckResult {
	output := append([]CheckResult(nil), results...)
	for i := range output {
		for _, segment := range output[i].ObjectLocation().Segments {
			if segment.Property != nil {
				if path, ok := normalized.SourceInstancePaths[*segment.Property]; ok {
					output[i].SourceInstancePath = path
				}
				break
			}
		}
	}
	return output
}

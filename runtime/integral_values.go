package runtime

import "github.com/teaql/teaql-golang/core"

// Model-owned storage types are checked after custom Fix, before the generated
// graph can allocate IDs or enter a provider. This does not enforce business
// ranges, positive-identity policies or a narrower modeled Integer width.
func (c *UserContext) checkIntegralValues(input *CheckAndFixInput, references *deferredReferenceState) []CheckResult {
	entity := c.Entity(input.Entity)
	if entity == nil {
		return nil
	}
	var results []CheckResult
	for _, property := range entity.Properties {
		if property.DataType != core.TypeI64 && property.DataType != core.TypeU64 {
			continue
		}
		value, exists := input.Values[property.Name]
		if !exists || value.IsNull() {
			continue
		}
		var valid bool
		message := "Expected an exact signed 64-bit integer"
		if property.DataType == core.TypeU64 {
			_, valid = value.TryU64()
			if !valid {
				valid = references.allows(property.Name, value)
			}
			message = "Expected an exact unsigned 64-bit integer"
		} else {
			_, valid = value.TryI64()
		}
		if !valid {
			results = append(results, CheckResult{RuleID: "invalid_type", EntityType: input.Entity,
				CanonicalLocation: Location().Property(property.Name), Message: message})
		}
	}
	return results
}

package runtime

import "github.com/teaql/teaql-golang/core"

// A customer Fix may repair a value, but cannot turn a modeled Boolean into
// integer/string/JSON truthiness. Requiredness stays with the model checker.
func (c *UserContext) checkBooleanValues(input *CheckAndFixInput, existing []CheckResult) []CheckResult {
	entity := c.Entity(input.Entity)
	if entity == nil {
		return nil
	}
	var results []CheckResult
	for _, property := range entity.Properties {
		if property.DataType != core.TypeBool {
			continue
		}
		value, present := input.Values[property.Name]
		if !present || value.IsNull() {
			continue
		}
		if _, valid := value.TryBool(); valid {
			continue
		}
		duplicate := false
		for _, result := range existing {
			if result.RuleID == "invalid_type" && result.ModelPath() == property.Name {
				duplicate = true
				break
			}
		}
		if !duplicate {
			results = append(results, CheckResult{RuleID: "invalid_type", EntityType: input.Entity,
				CanonicalLocation: Location().Property(property.Name), Message: "Expected a Boolean value"})
		}
	}
	return results
}

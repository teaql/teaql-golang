package runtime

import (
	"github.com/teaql/teaql-golang/core"
	"math"
)

// The trusted model, not user input or alias spelling, owns scalar types.
// This final boundary complements custom Checker/Fix; it does not replace it.
func (c *UserContext) checkFiniteFloatingValues(input *CheckAndFixInput) []CheckResult {
	entity := c.Entity(input.Entity)
	if entity == nil {
		return nil
	}
	var results []CheckResult
	for _, property := range entity.Properties {
		if property.DataType != core.TypeF64 {
			continue
		}
		value, exists := input.Values[property.Name]
		if !exists || value.IsNull() {
			continue
		}
		number, valid := value.TryF64()
		if valid && !math.IsNaN(number) && !math.IsInf(number, 0) {
			continue
		}
		// Do not put NaN/Infinity in an error payload: JSON cannot encode them.
		results = append(results, CheckResult{RuleID: "invalid_type", EntityType: input.Entity,
			CanonicalLocation: Location().Property(property.Name), Message: "Expected a finite floating-point value"})
	}
	return results
}

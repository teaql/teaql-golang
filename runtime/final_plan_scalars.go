package runtime

import "github.com/teaql/teaql-golang/core"

// A final plan is already past customer Fix. Reuse intrinsic representation
// checks here without calling Fix twice or changing the caller's plan. This
// must stay separate from validateMutationPlanFields: that namespace check
// also runs before Fix, where valid customer repairs are still permitted.
func (c *UserContext) validateFinalMutationPlanScalars(plan *MutationPlan) error {
	var failures []CheckResult
	for _, operation := range plan.Operations {
		input := &CheckAndFixInput{Entity: operation.Entity, Operation: operation.Kind,
			Values: core.CloneRecord(operation.ChangedValues)}
		var references *deferredReferenceState
		if plan.ledger != nil {
			references = c.captureDeferredReferences(input, plan.ledger, core.NewEntityKey(operation.Entity, operation.ID))
			references.revalidate()
		}
		results := c.checkFiniteFloatingValues(input)
		results = append(results, c.checkIntegralValues(input, references)...)
		results = append(results, c.checkBooleanValues(input, results)...)
		if len(results) != 0 {
			c.translateCheckerResults(results, operation.Entity, nil, plaintextLogsEnabled())
			failures = append(failures, results...)
		}
	}
	if len(failures) != 0 {
		return &RuntimeError{Type: "Check", CheckResults: failures}
	}
	return nil
}

package runtime

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/teaql/teaql-golang/core"
)

// This state exists only inside one serialized save invocation. It is neither
// a Context-owned mutation ledger nor a cache shared by independent saves.
type preparedGraphSnapshot struct {
	sealed      bool
	plan        *MutationPlan
	checks      []preparedGraphCheck
	assignments map[string]core.EntityKey
}

func preparedEntityKey(key core.EntityKey) string {
	return fmt.Sprintf("%s\x00%T:%v", key.Entity, key.ID.V, key.ID.V)
}

type preparedGraphCheck struct {
	ledger *core.EntityRoot
	owner  core.EntityKey
	input  CheckAndFixInput // canonical, detached, post-Fix values
}

func cloneCheckerInput(input *CheckAndFixInput) CheckAndFixInput {
	result := *input
	result.Values = core.CloneRecord(input.Values)
	result.OldValues = core.CloneRecord(input.OldValues)
	result.changedFields = nil
	return result
}

func preparedMutationChanged(entity, field string) error {
	location := Location()
	if field != "" {
		location = location.Property(field)
	}
	return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
		RuleID: "mutation_changed_after_preparation", EntityType: entity,
		CanonicalLocation: location, Message: "mutation differs from its prepared and reviewed state; prepare and review it again",
	}}}
}

func changedPreparedField(expected, actual core.Record) string {
	fields := make(map[string]bool, len(expected)+len(actual))
	for field := range expected {
		fields[field] = true
	}
	for field := range actual {
		fields[field] = true
	}
	ordered := make([]string, 0, len(fields))
	for field := range fields {
		ordered = append(ordered, field)
	}
	sort.Strings(ordered)
	for _, field := range ordered {
		before, beforePresent := expected[field]
		after, afterPresent := actual[field]
		if beforePresent != afterPresent || !reflect.DeepEqual(before, after) {
			return field
		}
	}
	return ""
}

// Revalidate after all policy/approval/warning callbacks. Reviewing a detached
// plan must not let a callback silently replace its still-live source graph.
func validateReviewedLedger(plan *MutationPlan) error {
	if plan == nil || plan.ledger == nil {
		return nil
	}
	actual := MutationPlanFromEntityRoot(plan.ledger, plan.RootEntity, plan.AuditReason)
	if len(actual.Operations) != len(plan.Operations) {
		return preparedMutationChanged(plan.RootEntity, "")
	}
	for index, expected := range plan.Operations {
		got := actual.Operations[index]
		if got.Entity != expected.Entity || got.Kind != expected.Kind || !reflect.DeepEqual(got.ID, expected.ID) {
			return preparedMutationChanged(expected.Entity, "")
		}
		if !reflect.DeepEqual(got.OriginalVersion, expected.OriginalVersion) {
			return preparedMutationChanged(expected.Entity, "version")
		}
		if field := changedPreparedField(expected.ChangedValues, got.ChangedValues); field != "" {
			return preparedMutationChanged(expected.Entity, field)
		}
	}
	return nil
}

func (c *UserContext) replayPreparedCheck(input *CheckAndFixInput, ledger *core.EntityRoot, owner core.EntityKey) (bool, error) {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	state := c.preparedGraphSnapshot
	if state == nil || ledger == nil {
		return false, nil
	}
	for _, check := range state.checks {
		key := check.owner
		if assigned, ok := state.assignments[preparedEntityKey(key)]; ok {
			key = assigned
		}
		if check.ledger != ledger || !reflect.DeepEqual(key, owner) {
			continue
		}
		if check.input.Entity != input.Entity || check.input.Operation != input.Operation {
			return true, preparedMutationChanged(input.Entity, "")
		}
		expected := cloneCheckerInput(&check.input)
		c.resolvePreparedReferences(&expected, state)
		if field := changedPreparedField(expected.Values, input.Values); field != "" {
			return true, preparedMutationChanged(input.Entity, field)
		}
		if field := changedPreparedField(expected.OldValues, input.OldValues); field != "" {
			return true, preparedMutationChanged(input.Entity, field)
		}
		input.Values, input.OldValues, input.Now = expected.Values, expected.OldValues, expected.Now
		return true, nil
	}
	if state.sealed {
		return true, preparedMutationChanged(input.Entity, "")
	}
	return false, nil
}

func (c *UserContext) rememberPreparedCheck(input *CheckAndFixInput, ledger *core.EntityRoot, owner core.EntityKey) {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	if state := c.preparedGraphSnapshot; state != nil && ledger != nil && !state.sealed {
		state.checks = append(state.checks, preparedGraphCheck{ledger: ledger, owner: owner, input: cloneCheckerInput(input)})
	}
}

func (c *UserContext) resolvePreparedReferences(input *CheckAndFixInput, state *preparedGraphSnapshot) {
	entity := c.Entity(input.Entity)
	if entity == nil {
		return
	}
	for _, relation := range entity.Relations {
		if relation.IsMany || relation.ForKey != "id" || relation.LocKey == "id" {
			continue
		}
		field, err := core.CanonicalLedgerField(entity, relation.LocKey)
		if err != nil {
			continue
		}
		if value, present := input.Values[field]; present {
			if assigned, ok := state.assignments[preparedEntityKey(core.NewEntityKey(relation.TargetEntity, value))]; ok {
				input.Values[field] = assigned.ID
			}
		}
	}
}

// RekeyPreparedGraphEntity is generated save infrastructure. It records the
// exact typed temporary-to-assigned identity, so a child may resolve only its
// own already-reviewed parent reference, not an arbitrary newly inserted row.
func (c *UserContext) RekeyPreparedGraphEntity(ledger *core.EntityRoot, oldKey, newKey core.EntityKey) error {
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	if state := c.preparedGraphSnapshot; state != nil && state.sealed {
		id, assigned := newKey.ID.V.(uint64)
		if ledger == nil || state.plan == nil || ledger != state.plan.ledger || oldKey.Entity != newKey.Entity || !assigned || id == 0 {
			return preparedMutationChanged(oldKey.Entity, "id")
		}
		found := false
		for _, operation := range state.plan.Operations {
			if operation.Entity == oldKey.Entity && reflect.DeepEqual(operation.ID, oldKey.ID) && operation.Kind == core.MutationInsert {
				found = true
			}
			if operation.Entity == newKey.Entity && reflect.DeepEqual(operation.ID, newKey.ID) {
				return preparedMutationChanged(oldKey.Entity, "id")
			}
		}
		if _, duplicate := state.assignments[preparedEntityKey(oldKey)]; duplicate {
			found = false
		}
		for _, target := range state.assignments {
			if reflect.DeepEqual(target, newKey) {
				found = false
			}
		}
		_, loadedTarget := ledger.OriginalVersion(newKey)
		if !found || !ledger.IsNew(oldKey) || ledger.IsNew(newKey) || loadedTarget || len(ledger.Change(newKey)) != 0 {
			return preparedMutationChanged(oldKey.Entity, "id")
		}
		state.assignments[preparedEntityKey(oldKey)] = newKey
	}
	if ledger == nil {
		return preparedMutationChanged(oldKey.Entity, "id")
	}
	ledger.Rekey(oldKey, newKey)
	return nil
}

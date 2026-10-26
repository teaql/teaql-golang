package runtime

import (
	"fmt"

	"github.com/teaql/teaql-golang/core"
)

// CheckAndFixGraphEntity is generated save infrastructure, not an application
// mutation entry point. The graph remains caller-owned; neither UserContext nor
// the customer Checker receives its ledger or its temporary-reference proof.
func CheckAndFixGraphEntity(context *UserContext, input *CheckAndFixInput, ledger *core.EntityRoot, owner core.EntityKey) error {
	if context == nil {
		return fmt.Errorf("UserContext is required")
	}
	if input == nil {
		return nil
	}
	return context.checkAndFixStorage(input, ledger, owner)
}

type deferredReference struct {
	field string
	value int64
}

type deferredReferenceState struct {
	ledger        *core.EntityRoot
	owner         core.EntityKey
	fields        []deferredReference
	targets       []core.EntityKey
	before, after []bool
}

func (c *UserContext) captureDeferredReferences(input *CheckAndFixInput, ledger *core.EntityRoot, owner core.EntityKey) *deferredReferenceState {
	if ledger == nil || owner.Entity != input.Entity {
		return nil
	}
	entity := c.Entity(input.Entity)
	if entity == nil {
		return nil
	}
	var state *deferredReferenceState
	for _, relation := range entity.Relations {
		if relation.IsMany || relation.ForKey != "id" || relation.LocKey == "id" {
			continue
		}
		propertyIsReference := false
		for _, property := range entity.Properties {
			if property.Name == relation.LocKey && property.DataType == core.TypeU64 && !property.IsId {
				propertyIsReference = true
				break
			}
		}
		if !propertyIsReference {
			continue
		}
		field, err := core.CanonicalLedgerField(entity, relation.LocKey)
		if err != nil {
			continue // The mapping boundary rejects ambiguous relation metadata.
		}
		value, exists := input.Values[field]
		temporary, typed := value.V.(int64)
		if !exists || !typed || temporary >= 0 {
			continue
		}
		if state == nil {
			state = &deferredReferenceState{ledger: ledger, owner: owner}
		}
		state.fields = append(state.fields, deferredReference{field: field, value: temporary})
		state.targets = append(state.targets, core.NewEntityKey(relation.TargetEntity, value))
	}
	if state == nil {
		return nil
	}
	state.before = ledger.NewReferenceSnapshot(owner, state.targets)
	return state
}

func (s *deferredReferenceState) revalidate() {
	if s != nil {
		s.after = s.ledger.NewReferenceSnapshot(s.owner, s.targets)
	}
}

func (s *deferredReferenceState) allows(field string, value core.Value) bool {
	if s == nil {
		return false
	}
	temporary, typed := value.V.(int64)
	if !typed {
		return false
	}
	for index, reference := range s.fields {
		if reference.field == field && reference.value == temporary && index < len(s.before) && index < len(s.after) && s.before[index] && s.after[index] {
			return true
		}
	}
	return false
}

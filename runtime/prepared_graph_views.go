package runtime

import (
	"reflect"

	"github.com/teaql/teaql-golang/core"
)

// Object identity is scoped to one prepared save, not to Context's lifetime.
// It does not change the source-precedence rules of explicit ledger composition.
type preparedGraphView struct {
	owner  core.EntityKey
	entity interface {
		core.Entity
		EntityKey() core.EntityKey
	}
	values core.Record
}

// AdmitPreparedGraphEntityView is generated graph infrastructure. The complete
// normalized graph must admit its writable views before any Checker/Fix runs.
// A repeated pointer is one view. Two distinct writable views of the same typed
// identity are ambiguous even if their pending values happen to be equal.
func (c *UserContext) AdmitPreparedGraphEntityView(owner core.EntityKey, entity interface {
	core.Entity
	EntityKey() core.EntityKey
}) error {
	value := reflect.ValueOf(entity)
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() || entity.EntityName() != owner.Entity || !reflect.DeepEqual(entity.EntityKey(), owner) {
		return preparedMutationChanged(owner.Entity, "")
	}
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	state := c.preparedGraphSnapshot
	if state == nil || state.sealed || state.viewsSealed {
		return preparedMutationChanged(owner.Entity, "")
	}
	if state.views == nil {
		state.views = make(map[string]*preparedGraphView)
	}
	identity := preparedEntityKey(owner)
	if prior, exists := state.views[identity]; exists {
		if prior.entity != entity {
			return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{
				RuleID: "ambiguous_mutation_view", EntityType: owner.Entity,
				CanonicalLocation: Location(),
				Message:           "multiple writable object instances represent one entity; select one write view or explicitly compose their ledgers into one view",
			}}}
		}
		return nil
	}
	entry := &preparedGraphView{owner: owner, entity: entity}
	state.views[identity] = entry
	state.viewOrder = append(state.viewOrder, entry)
	return nil
}

// RememberPreparedGraphRelation snapshots pointer membership, including order
// and repeated references, before policy/approval/warning callbacks can run.
// The getter is generated, in-memory only; no database access is implicit.
func RememberPreparedGraphRelation[T interface {
	core.Entity
	comparable
}](c *UserContext, owner core.EntityKey, field string, get func() []T) error {
	expected := append([]T(nil), get()...)
	validate := func() error {
		actual := get()
		if len(expected) != len(actual) {
			return preparedMutationChanged(owner.Entity, field)
		}
		for index := range expected {
			if expected[index] != actual[index] {
				return preparedMutationChanged(owner.Entity, field)
			}
		}
		return nil
	}
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	state := c.preparedGraphSnapshot
	if state == nil || state.sealed || state.viewsSealed {
		return preparedMutationChanged(owner.Entity, field)
	}
	state.relations = append(state.relations, validate)
	return nil
}

// Do not call entity methods or generated getters while holding graphSaveMu:
// their implementation may inspect Context. They never enter Checker's input.
func (c *UserContext) capturePreparedGraphViews() error {
	c.graphSaveMu.Lock()
	state := c.preparedGraphSnapshot
	if state != nil {
		state.viewsSealed = true
	}
	c.graphSaveMu.Unlock()
	if state == nil {
		return nil
	}
	for _, entry := range state.viewOrder {
		if !reflect.DeepEqual(entry.entity.EntityKey(), entry.owner) {
			return preparedMutationChanged(entry.owner.Entity, "id")
		}
		entry.values = core.CloneRecord(entry.entity.IntoRecord())
	}
	return nil
}

func (c *UserContext) validatePreparedGraphViews() error {
	c.graphSaveMu.Lock()
	state := c.preparedGraphSnapshot
	c.graphSaveMu.Unlock()
	if state == nil {
		return nil
	}
	for _, entry := range state.viewOrder {
		if !reflect.DeepEqual(entry.entity.EntityKey(), entry.owner) {
			return preparedMutationChanged(entry.owner.Entity, "id")
		}
		if field := changedPreparedField(entry.values, entry.entity.IntoRecord()); field != "" {
			canonical, err := c.CanonicalLedgerField(entry.owner.Entity, field)
			if err != nil {
				return err
			}
			return preparedMutationChanged(entry.owner.Entity, canonical)
		}
	}
	for _, validate := range state.relations {
		if err := validate(); err != nil {
			return err
		}
	}
	return nil
}

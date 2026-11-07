package core

// EntityCreationState is process-local provenance, not a modeled or wire field.
// Its zero value is unmanaged. Reserving an ID and hydrating a read view are
// different operations; public new-entity intent cannot convert one to the other.
type EntityCreationState struct{ kind uint8 }

type EntityCreationError struct{ Entity string }

func (e *EntityCreationError) Error() string {
	return "ENTITY_CREATION_CONTEXT_REQUIRED: " + e.Entity + ": create through a Context-bound factory; hydrated values are not new entities"
}

func (s *EntityCreationState) Reserve(entity string, id uint64) error {
	if id == 0 || s.kind == 2 {
		return &EntityCreationError{Entity: entity}
	}
	s.kind = 1
	return nil
}

func (s *EntityCreationState) Hydrated() { s.kind = 2 }

func (s EntityCreationState) RequireNew(entity string, id uint64) error {
	if s.kind != 1 || id == 0 {
		return &EntityCreationError{Entity: entity}
	}
	return nil
}

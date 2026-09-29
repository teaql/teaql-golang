package logprivacy

// ProjectionState is private runtime bookkeeping, excluded from wire metadata.
// Its payload contains an already-safe alternative, never raw intent provenance.
type ProjectionState struct{ value any }

func NewProjectionState(value any) ProjectionState  { return ProjectionState{value: value} }
func ReadProjectionState(state ProjectionState) any { return state.value }
func (ProjectionState) String() string              { return "<internal safe projection>" }
func (ProjectionState) GoString() string            { return "<internal safe projection>" }

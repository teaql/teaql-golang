package logprivacy

// IntentSource carries invocation-local provenance between runtime packages.
// Its payload is deliberately unavailable through formatting or serialization.
// Safe projections must clear it before handing metadata to a sink or buffer.
type IntentSource struct{ value any }

func NewIntentSource(value any) IntentSource   { return IntentSource{value: value} }
func ReadIntentSource(source IntentSource) any { return source.value }
func (IntentSource) String() string            { return "<internal intent provenance>" }
func (IntentSource) GoString() string          { return "<internal intent provenance>" }

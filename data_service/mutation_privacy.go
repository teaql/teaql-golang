package data_service

import "github.com/teaql/teaql-golang/core"

// MutationPrivacyEntry is generated scalar provenance, never a write payload.
type MutationPrivacyEntry struct {
	entity string
	values core.Record
}

func NewMutationPrivacyEntry(entity string, values core.Record) MutationPrivacyEntry {
	return MutationPrivacyEntry{entity: entity, values: core.CloneRecord(values)}
}
func (e MutationPrivacyEntry) Entity() string      { return e.entity }
func (e MutationPrivacyEntry) Values() core.Record { return core.CloneRecord(e.values) }
func (MutationPrivacyEntry) String() string        { return "<mutation scalar entry>" }
func (MutationPrivacyEntry) GoString() string      { return "<mutation scalar entry>" }

// MutationPrivacy is an immutable invocation-local source. Its fields are not
// exported, so neither JSON nor formatting can expose or supply the records.
type MutationPrivacy struct{ entries []MutationPrivacyEntry }

func NewMutationPrivacy(entries ...MutationPrivacyEntry) *MutationPrivacy {
	result := &MutationPrivacy{}
	for _, entry := range entries {
		result.entries = append(result.entries, NewMutationPrivacyEntry(entry.entity, entry.values))
	}
	return result
}
func (*MutationPrivacy) String() string   { return "<mutation scalar provenance>" }
func (*MutationPrivacy) GoString() string { return "<mutation scalar provenance>" }

// WithMutationPrivacy captures a request without changing caller-owned state.
func WithMutationPrivacy(request MutationRequest, source *MutationPrivacy) (MutationRequest, error) {
	captured, err := CaptureMutationRequest(request)
	if err != nil {
		return nil, err
	}
	switch value := captured.(type) {
	case *InsertMutation:
		value.privacy = source
	case *UpdateMutation:
		value.privacy = source
	case *DeleteMutation:
		value.privacy = source
	case *RecoverMutation:
		value.privacy = source
	case *BatchMutation:
		value.privacy = source
	}
	return captured, nil
}

// MutationPrivacyEntries exposes copies to provider adapters, not shared maps.
func MutationPrivacyEntries(request MutationRequest) []MutationPrivacyEntry {
	var source *MutationPrivacy
	switch value := request.(type) {
	case *InsertMutation:
		source = value.privacy
	case *UpdateMutation:
		source = value.privacy
	case *DeleteMutation:
		source = value.privacy
	case *RecoverMutation:
		source = value.privacy
	case *BatchMutation:
		source = value.privacy
	}
	if source == nil {
		return nil
	}
	return append([]MutationPrivacyEntry(nil), source.entries...)
}

package runtime

import (
	"reflect"
	"strings"

	"github.com/teaql/teaql-golang/core"
)

type preparedGraphLineage struct {
	ledger            *core.EntityRoot
	owner             core.EntityKey
	parent            *core.EntityKey
	local, rootReason string
	explicit          []*core.TraceNode
}

// PrepareGraphMutationLineage captures generated graph ancestry before policy.
// It is framework infrastructure, not an application audit-writing API.
func (c *UserContext) PrepareGraphMutationLineage(ledger *core.EntityRoot, owner core.EntityKey, parent *core.EntityKey, intent core.MutationIntent, local *string) error {
	if err := intent.Validate(); err != nil {
		return err
	}
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	state := c.preparedGraphSnapshot
	if state == nil || state.sealed || ledger == nil {
		return preparedMutationChanged(owner.Entity, "traceChain")
	}
	entry := preparedGraphLineage{ledger: ledger, owner: owner, rootReason: intent.Comment(), explicit: ledger.TraceChain(owner)}
	if parent != nil {
		owned := *parent
		entry.parent = &owned
	}
	if local != nil {
		entry.local = strings.TrimSpace(*local)
	}
	identity := preparedEntityKey(owner)
	if prior, exists := state.lineage[identity]; exists {
		if !reflect.DeepEqual(prior, entry) {
			return preparedMutationChanged(owner.Entity, "traceChain")
		}
		return nil
	}
	state.lineage[identity] = entry
	return nil
}

func (s *preparedGraphSnapshot) traceFor(owner core.EntityKey, useExplicit bool, visiting map[string]bool) ([]*core.TraceNode, error) {
	identity := preparedEntityKey(owner)
	entry, exists := s.lineage[identity]
	if !exists || visiting[identity] {
		return nil, preparedMutationChanged(owner.Entity, "traceChain")
	}
	visiting[identity] = true
	defer delete(visiting, identity)
	if useExplicit && len(entry.explicit) != 0 {
		return core.CloneTraceNodes(entry.explicit), nil
	}
	var trace []*core.TraceNode
	reason := entry.rootReason
	if entry.parent != nil {
		var err error
		trace, err = s.traceFor(*entry.parent, false, visiting)
		if err != nil {
			return nil, err
		}
		if entry.local == "" {
			return trace, nil
		}
		reason = entry.local
	}
	key := owner
	if assigned, ok := s.assignments[identity]; ok {
		key = assigned
	}
	node := core.NewTypedTraceNode("auditReason", key.Entity, reason)
	if id, ok := key.ID.TryU64(); ok {
		node.EntityId = &id
	}
	return append(trace, node), nil
}

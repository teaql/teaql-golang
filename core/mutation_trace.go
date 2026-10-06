package core

// MutationTraceScope owns one immutable parent link. No trace stack lives on
// UserContext, and sibling branches share ancestors without sharing mutation.
type MutationTraceScope struct {
	parent *MutationTraceScope
	node   TraceNode
}

// MutationScopeForEntity is generated infrastructure. The root reason comes
// only from validated request intent. A blank descendant reason inherits.
func MutationScopeForEntity(parent *MutationTraceScope, key EntityKey, intent MutationIntent, local *string) (*MutationTraceScope, error) {
	if err := intent.Validate(); err != nil {
		return nil, err
	}
	reason := intent.Comment()
	if parent != nil {
		if text, err := requiredIntentText(local, "mutation", "comment"); err == nil {
			reason = text
		} else {
			return parent, nil
		}
	}
	node := TraceNode{Kind: "auditReason", Name: key.Entity, EntityType: key.Entity, Comment: reason}
	if id, ok := key.ID.TryU64(); ok {
		node.EntityId = &id
	}
	return &MutationTraceScope{parent: parent, node: node}, nil
}

func (s *MutationTraceScope) Recover() []*TraceNode {
	var nodes []*TraceNode
	for current := s; current != nil; current = current.parent {
		node := current.node
		nodes = append(nodes, &node)
	}
	for left, right := 0, len(nodes)-1; left < right; left, right = left+1, right-1 {
		nodes[left], nodes[right] = nodes[right], nodes[left]
	}
	return CloneTraceNodes(nodes)
}

// MutationTraceForEntity replaces fallback with a complete ledger-specific
// chain; a specific chain is never appended to an inherited ancestor twice.
func MutationTraceForEntity(root *EntityRoot, key EntityKey, scope *MutationTraceScope) []*TraceNode {
	if root != nil {
		if specific := root.TraceChain(key); len(specific) != 0 {
			return specific
		}
	}
	return scope.Recover()
}

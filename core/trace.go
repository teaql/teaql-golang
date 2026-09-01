package core

import "fmt"

type TraceNode struct {
	Kind       string
	Name       string
	EntityType string
	EntityId   *uint64
	Comment    string
}

func NewTraceNode(entityType string, entityId *uint64, comment string) *TraceNode {
	return &TraceNode{
		Kind:       "entity",
		Name:       entityType,
		EntityType: entityType,
		EntityId:   entityId,
		Comment:    comment,
	}
}

func NewTypedTraceNode(kind, name, comment string) *TraceNode {
	return &TraceNode{Kind: kind, Name: name, EntityType: name, Comment: comment}
}

func (n *TraceNode) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s:%s=%s", n.Kind, n.Name, n.Comment)
}

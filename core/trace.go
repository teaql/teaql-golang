package core

type TraceNode struct {
	EntityType string
	EntityId   *uint64
	Comment    string
}

func NewTraceNode(entityType string, entityId *uint64, comment string) *TraceNode {
	return &TraceNode{
		EntityType: entityType,
		EntityId:   entityId,
		Comment:    comment,
	}
}

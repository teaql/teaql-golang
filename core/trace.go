package core

import (
	"fmt"
	"strings"
)

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
	if n.EntityId != nil {
		return fmt.Sprintf("%s:%s#%d=%s", n.Kind, n.Name, *n.EntityId, n.Comment)
	}
	return fmt.Sprintf("%s:%s=%s", n.Kind, n.Name, n.Comment)
}

// CloneTraceNodes owns both the frames and their optional ID values. A copied
// slice alone would still let a builder or sink mutate another request's trace.
func CloneTraceNodes(source []*TraceNode) []*TraceNode {
	result := make([]*TraceNode, len(source))
	for i, node := range source {
		if node == nil {
			continue
		}
		value := *node
		if node.EntityId != nil {
			id := *node.EntityId
			value.EntityId = &id
		}
		result[i] = &value
	}
	return result
}

func traceKind(node *TraceNode) string {
	if node == nil {
		return ""
	}
	return strings.ReplaceAll(strings.ToLower(node.Kind), "_", "")
}

type SQLTracePath struct {
	TraceChain                    []*TraceNode
	Comment, Purpose, AuditReason *string
}

// CanonicalSQLTracePath implements the frozen Rust path algorithm. Request
// intent is extracted separately; this helper never authorizes a request.
func CanonicalSQLTracePath(source []*TraceNode, backend, operation string) SQLTracePath {
	var result SQLTracePath
	var hasOperation, hasProvider, hasSQL bool
	root, entity := "", ""
	for _, node := range source {
		if node == nil {
			continue
		}
		kind := traceKind(node)
		switch kind {
		case "operation":
			hasOperation = true
		case "provider":
			hasProvider = true
		case "sql":
			hasSQL = true
		case "comment":
			value := node.Comment
			result.Comment = &value
		case "purpose":
			value := node.Comment
			result.Purpose = &value
		case "auditreason":
			value := node.Comment
			result.AuditReason = &value
		}
		if strings.TrimSpace(node.Name) != "" {
			if root == "" {
				root = node.Name
			}
			if kind == "entity" {
				entity = node.Name
			}
		}
	}
	if hasOperation && hasProvider && hasSQL {
		for _, node := range CloneTraceNodes(source) {
			kind := traceKind(node)
			if node != nil && kind != "comment" && kind != "purpose" && kind != "auditreason" {
				result.TraceChain = append(result.TraceChain, node)
			}
		}
		return result
	}
	if root == "" {
		root = "unknown"
	}
	if entity == "" {
		entity = root
	}
	if strings.TrimSpace(backend) == "" {
		backend = "unknown"
	}
	mode, secondKind, statement := "mutation", "entity", entity
	if operation == "select" {
		mode, secondKind, statement = "query", "request", root
	}
	result.TraceChain = []*TraceNode{NewTypedTraceNode("operation", root, mode), NewTypedTraceNode(secondKind, statement, "")}
	for _, node := range CloneTraceNodes(source) {
		if traceKind(node) == "relation" {
			result.TraceChain = append(result.TraceChain, node)
		}
	}
	result.TraceChain = append(result.TraceChain, NewTypedTraceNode("provider", backend, ""), NewTypedTraceNode("sql", operation, ""))
	return result
}

// QueryTraceSource preserves an inherited root while using request-owned intent.
func QueryTraceSource(entity string, frames []*TraceNode, comment, purpose string) []*TraceNode {
	root := entity
	for _, node := range frames {
		kind := traceKind(node)
		if (kind == "comment" || kind == "purpose" || kind == "operation" || kind == "request") && strings.TrimSpace(node.Name) != "" {
			root = node.Name
			break
		}
	}
	source := []*TraceNode{NewTypedTraceNode("comment", root, comment), NewTypedTraceNode("purpose", root, purpose)}
	for _, node := range CloneTraceNodes(frames) {
		kind := traceKind(node)
		if node != nil && kind != "comment" && kind != "purpose" {
			source = append(source, node)
		}
	}
	return source
}

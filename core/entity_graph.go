package core

type EntityGraphOperation int

const (
	EntityGraphOpSave EntityGraphOperation = iota
	EntityGraphOpDelete
)

type RelationChild struct {
	Relation string
	Node     *EntityGraphNode
}

type EntityGraphNode struct {
	EntityType string
	Record     Record
	Comment    *string
	Operation  EntityGraphOperation
	Children   []RelationChild
}

type EntityGraphBuilder struct {
	node *EntityGraphNode
}

func (b *EntityGraphBuilder) Comment(comment string) *EntityGraphBuilder {
	b.node.Comment = &comment
	return b
}

func (b *EntityGraphBuilder) Delete() *EntityGraphBuilder {
	b.node.Operation = EntityGraphOpDelete
	return b
}

func (b *EntityGraphBuilder) Child(relation string, child *EntityGraphBuilder) *EntityGraphBuilder {
	b.node.Children = append(b.node.Children, RelationChild{Relation: relation, Node: child.node})
	return b
}

func (b *EntityGraphBuilder) Build() *EntityGraph {
	return &EntityGraph{Root: b.node}
}

type EntityGraph struct {
	Root *EntityGraphNode
}

func NewEntityGraph(entity Entity) *EntityGraphBuilder {
	return &EntityGraphBuilder{
		node: &EntityGraphNode{
			EntityType: entity.EntityName(),
			Record:     entity.IntoRecord(),
			Operation:  EntityGraphOpSave,
			Children:   make([]RelationChild, 0),
		},
	}
}

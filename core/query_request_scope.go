package core

import (
	"fmt"
	"strconv"
	"strings"
)

// QueryOperation is generator input, not an execution scope or query option.
// The runtime captures it by value, including the optional alias.
type QueryOperation struct {
	Kind, Method, Relation string
	Alias                  *string
	// RuntimeComment is used only when no generated native descriptor exists.
	// It must describe real framework work rather than impersonate a Q method.
	RuntimeComment string
}

func (o QueryOperation) captured() (QueryOperation, error) {
	if strings.TrimSpace(o.Kind) == "" || (strings.TrimSpace(o.Method) == "" && strings.TrimSpace(o.RuntimeComment) == "") {
		return QueryOperation{}, fmt.Errorf("query operation requires kind and native method")
	}
	if o.Alias != nil {
		alias := *o.Alias
		o.Alias = &alias
	}
	return o, nil
}

// Comment uses Go's literal spelling; the generator supplies the real method.
func (o QueryOperation) Comment() string {
	if o.Method == "" {
		return o.RuntimeComment
	}
	args := ""
	if o.Alias != nil {
		args = strconv.Quote(*o.Alias)
	}
	if o.Kind == "relation" {
		if args != "" {
			args += ", "
		}
		args += "..."
	}
	return o.Method + "(" + args + ")"
}

func (o *QueryOperation) Clone() *QueryOperation {
	if o == nil {
		return nil
	}
	copyOperation := *o
	if o.Alias != nil {
		alias := *o.Alias
		copyOperation.Alias = &alias
	}
	return &copyOperation
}

// QueryRequestScope shares only immutable small headers. It never owns the AST,
// Context, rows, mutable builder, or mutation ledger. Its zero value is invalid.
type QueryRequestScope struct {
	parent                *QueryRequestScope
	root                  *QueryRequestScope
	intent                QueryIntent
	entity, comment, kind string
	operation             *QueryOperation
}

func NewQueryRequestScope(entity string, intent QueryIntent) (*QueryRequestScope, error) {
	if err := intent.Validate(); err != nil {
		return nil, err
	}
	root := &QueryRequestScope{entity: entity, comment: intent.Comment(), intent: intent, kind: "query"}
	root.root = root
	return root, nil
}

func (s *QueryRequestScope) Validate() error {
	if s == nil || s.root == nil {
		return fmt.Errorf("query request requires an owned scope")
	}
	if err := s.root.intent.Validate(); err != nil {
		return err
	}
	_, err := requiredIntentText(&s.comment, "query", "comment")
	return err
}

func (s *QueryRequestScope) Parent() *QueryRequestScope { return s.parent }
func (s *QueryRequestScope) Root() *QueryRequestScope   { return s.root }
func (s *QueryRequestScope) Comment() string            { return s.comment }
func (s *QueryRequestScope) Purpose() string            { return s.root.intent.Purpose() }
func (s *QueryRequestScope) Relation() string {
	if s == nil || s.operation == nil {
		return ""
	}
	return s.operation.Relation
}
func (s *QueryRequestScope) IsFacetOperation() bool {
	return s != nil && s.operation != nil && s.operation.Kind == "facet"
}
func (s *QueryRequestScope) String() string   { return "QueryRequestScope(<redacted>)" }
func (s *QueryRequestScope) GoString() string { return s.String() }

func (s *QueryRequestScope) Operation(operation QueryOperation) (*QueryRequestScope, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	o, err := operation.captured()
	if err != nil {
		return nil, err
	}
	return &QueryRequestScope{parent: s, root: s.root, entity: s.entity, kind: "operation", comment: o.Comment(), operation: &o}, nil
}

func (s *QueryRequestScope) Child(entity string, comment *string) (*QueryRequestScope, error) {
	if comment == nil {
		local := "load " + entity
		comment = &local
	}
	return s.local(entity, "query", *comment)
}

func (s *QueryRequestScope) Helper(entity, comment string) (*QueryRequestScope, error) {
	return s.local(entity, "helper", comment)
}

func (s *QueryRequestScope) local(entity, kind, comment string) (*QueryRequestScope, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if _, err := requiredIntentText(&comment, "query", "comment"); err != nil {
		return nil, err
	}
	return &QueryRequestScope{parent: s, root: s.root, entity: entity, kind: kind, comment: comment}, nil
}

// Recover returns fresh diagnostic frames, never mutable aliases into a scope.
func (s *QueryRequestScope) Recover() []*TraceNode {
	if s.Validate() != nil {
		return nil
	}
	var reverse []*QueryRequestScope
	for current := s; current != nil; current = current.parent {
		reverse = append(reverse, current)
	}
	result := []*TraceNode{NewTypedTraceNode("purpose", s.root.entity, s.Purpose())}
	for i := len(reverse) - 1; i >= 0; i-- {
		node := reverse[i]
		name := node.entity
		if node.operation != nil {
			name = node.operation.Method
			if name == "" {
				name = node.operation.Relation
			}
		}
		result = append(result, NewTypedTraceNode("comment", name, node.comment))
	}
	return result
}

func (q *SelectQuery) QueryRequestScope() *QueryRequestScope { return q.requestScope }

// WithQueryRequestScope binds a fresh body snapshot; reusing a child template
// must not retain the first caller's ancestry or mutate its local comment.
func (q *SelectQuery) WithQueryRequestScope(scope *QueryRequestScope) (*SelectQuery, error) {
	if q == nil {
		return nil, fmt.Errorf("query request requires a query")
	}
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	copyQuery := q.Clone()
	comment, purpose := scope.Comment(), scope.Purpose()
	copyQuery.requestScope = scope
	copyQuery.CommentText, copyQuery.PurposeText = &comment, &purpose
	return copyQuery, nil
}

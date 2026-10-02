package sql

import (
	"fmt"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/logprivacy"
)

type compiledMutationLeaf struct {
	entity     *core.EntityDescriptor
	entityName string
	query      *CompiledQuery
	operation  ds.DataServiceOperation
}

// One invocation owns both the captured commands and their compiled bindings.
// This is sequential native batch execution, not a prepared-batch transport.
// No provenance is installed on a Context or reusable executor.
type sqlMutationPlan struct {
	leaves    map[ds.MutationRequest]compiledMutationLeaf
	inherited logprivacy.IntentSource
}

func (e *SqlDataServiceExecutor) prepareMutation(request ds.MutationRequest) (*sqlMutationPlan, error) {
	plan := &sqlMutationPlan{leaves: make(map[ds.MutationRequest]compiledMutationLeaf)}
	var sources []ds.ExecutionMetadata
	var compile func(ds.MutationRequest) error
	compile = func(request ds.MutationRequest) error {
		if batch, ok := request.(*ds.BatchMutation); ok {
			for _, child := range batch.Mutations {
				if err := compile(child); err != nil {
					return err
				}
			}
			return nil
		}
		var name string
		var oldValues core.Record
		switch leaf := request.(type) {
		case *ds.InsertMutation:
			name = leaf.Cmd.Entity
		case *ds.UpdateMutation:
			name, oldValues = leaf.Cmd.Entity, leaf.Cmd.OldValues
		case *ds.DeleteMutation:
			name = leaf.Cmd.Entity
		case *ds.RecoverMutation:
			name = leaf.Cmd.Entity
		default:
			return fmt.Errorf("unsupported mutation request type")
		}
		entity := e.SchemaProvider.GetEntity(name)
		if entity == nil {
			return fmt.Errorf("unknown entity %s", name)
		}
		dialect := &DefaultSqlDialect{Dialect: e.Dialect}
		leaf := compiledMutationLeaf{entity: entity, entityName: name}
		var err error
		switch command := request.(type) {
		case *ds.InsertMutation:
			leaf.query, err = dialect.CompileInsert(entity, command.Cmd)
			leaf.operation = ds.OpInsert
		case *ds.UpdateMutation:
			leaf.query, err = dialect.CompileUpdate(entity, command.Cmd)
			leaf.operation = ds.OpUpdate
		case *ds.DeleteMutation:
			leaf.query, err = dialect.CompileDelete(entity, command.Cmd)
			leaf.operation = ds.OpDelete
		case *ds.RecoverMutation:
			leaf.query, err = dialect.CompileRecover(entity, command.Cmd)
			leaf.operation = ds.OpRecover
		}
		if err != nil {
			return err
		}
		plan.leaves[request] = leaf
		source := ds.ExecutionMetadata{
			GeneratedSQL: leaf.query.GeneratedSQL, ParameterizedSQL: leaf.query.Sql,
			ParameterLogPolicies: append([]string(nil), leaf.query.ParameterLogPolicies...),
		}
		for _, value := range leaf.query.Params {
			source.Parameters = append(source.Parameters, core.CloneValue(value))
		}
		sources = append(sources, source)
		// An old private value may appear in the batch's root reason despite not
		// being a binding in its UPDATE. Classify it through model metadata too.
		if len(oldValues) > 0 {
			previous := ds.ExecutionMetadata{GeneratedSQL: true}
			for field, value := range oldValues {
				previous.Parameters = append(previous.Parameters, core.CloneValue(value))
				previous.ParameterLogPolicies = append(previous.ParameterLogPolicies, fieldLogPolicy(entity, field))
			}
			sources = append(sources, previous)
		}
		return nil
	}
	if err := compile(request); err != nil {
		return nil, &SqlExecutorError{CompileError: err}
	}
	plan.inherited = logprivacy.NewIntentSource(sources)
	return plan, nil
}

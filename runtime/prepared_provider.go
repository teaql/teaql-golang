package runtime

import (
	"context"
	"reflect"
	"sync"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

// A guard owns one transaction's detached plan. It is not reusable between
// saves, and never supplies a mutation ledger to Context or a provider.
type preparedProviderGuard struct {
	mu                  sync.Mutex
	context             *UserContext
	state               *preparedGraphSnapshot
	claimed             []bool
	resultVersions      map[int]int64
	fault               error
	inflight            int
	closed, allowFinish bool
}

func clonePreparedMetadata(c *UserContext) map[string]*core.EntityDescriptor {
	result := make(map[string]*core.EntityDescriptor)
	if c.Metadata == nil {
		return result
	}
	for _, source := range c.Metadata.AllEntities() {
		value := *source
		if source.Properties != nil {
			value.Properties = make([]*core.PropertyDescriptor, len(source.Properties))
		}
		for index, property := range source.Properties {
			owned := *property
			value.Properties[index] = &owned
		}
		if source.Relations != nil {
			value.Relations = make([]*core.RelationDescriptor, len(source.Relations))
		}
		for index, relation := range source.Relations {
			owned := *relation
			value.Relations[index] = &owned
		}
		if source.AuditMaskFlds != nil {
			value.AuditMaskFlds = append([]string{}, source.AuditMaskFlds...)
		}
		if source.DataSvc != nil {
			owned := *source.DataSvc
			value.DataSvc = &owned
		}
		if source.AuditValueMaxL != nil {
			owned := *source.AuditValueMaxL
			value.AuditValueMaxL = &owned
		}
		result[value.Name] = &value
	}
	return result
}

func preparedIncomplete(entity string) error {
	return &RuntimeError{Type: "Check", CheckResults: []CheckResult{{RuleID: "prepared_mutation_incomplete", EntityType: entity,
		CanonicalLocation: Location(), Message: "the complete reviewed mutation plan was not applied exactly once"}}}
}

func (c *UserContext) sealPreparedPlan(plan *MutationPlan) error {
	if err := c.capturePreparedGraphViews(); err != nil {
		return err
	}
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	state := c.preparedGraphSnapshot
	state.sealed, state.plan = true, plan
	if plan == nil {
		return nil
	}
	if !reflect.DeepEqual(state.metadata, clonePreparedMetadata(c)) {
		return preparedMutationChanged(plan.RootEntity, "metadata")
	}
	seen := make(map[string]bool)
	for index := range plan.Operations {
		operation := &plan.Operations[index]
		key := core.NewEntityKey(operation.Entity, operation.ID)
		identity := preparedEntityKey(key)
		if seen[identity] {
			return preparedMutationChanged(operation.Entity, "id")
		}
		seen[identity] = true
		if plan.ledger != nil {
			entry := preparedLedgerEntry{owner: key, values: core.CloneRecord(plan.ledger.Change(key)), new: plan.ledger.IsNew(key), deleted: plan.ledger.IsDeleted(key), trace: plan.ledger.TraceChain(key)}
			if version, known := plan.ledger.OriginalVersion(key); known {
				owned := version
				entry.version = &owned
			}
			state.source = append(state.source, entry)
		}
		if _, ok := state.lineage[identity]; ok {
			trace, err := state.traceFor(key, true, make(map[string]bool))
			if err != nil {
				return err
			}
			operation.TraceChain = trace
		}
	}
	return nil
}

func newPreparedProviderGuard(c *UserContext, state *preparedGraphSnapshot) *preparedProviderGuard {
	if state == nil || state.plan == nil {
		return nil
	}
	return &preparedProviderGuard{context: c, state: state, claimed: make([]bool, len(state.plan.Operations)), resultVersions: make(map[int]int64)}
}

func (g *preparedProviderGuard) poison(err error) error {
	if g.fault == nil {
		g.fault = err
	}
	return g.fault
}

func (g *preparedProviderGuard) canonical(entity string, values core.Record) (core.Record, error) {
	if descriptor := g.state.metadata[entity]; descriptor != nil {
		return core.CanonicalLedgerRecord(descriptor, values)
	}
	return core.CloneRecord(values), nil
}

func (g *preparedProviderGuard) metadataUnchanged() bool {
	if g.context.Metadata == nil {
		return len(g.state.metadata) == 0
	}
	current := g.context.Metadata.AllEntities()
	if len(current) != len(g.state.metadata) {
		return false
	}
	for _, entity := range current {
		if !reflect.DeepEqual(g.state.metadata[entity.Name], entity) {
			return false
		}
	}
	return true
}

func (g *preparedProviderGuard) resolvedKey(key core.EntityKey) core.EntityKey {
	if target, ok := g.state.assignments[preparedEntityKey(key)]; ok {
		return target
	}
	return key
}

func (g *preparedProviderGuard) expectedValues(operation MutationOperation) core.Record {
	values := core.CloneRecord(operation.ChangedValues)
	if operation.Kind == core.MutationInsert {
		for _, check := range g.state.checks {
			if check.owner.Entity == operation.Entity && reflect.DeepEqual(check.owner.ID, operation.ID) {
				values = core.CloneRecord(check.input.Values)
				break
			}
		}
		if values == nil {
			values = make(core.Record)
		}
		values["id"] = g.resolvedKey(core.NewEntityKey(operation.Entity, operation.ID)).ID
		if version, exists := values["version"]; exists {
			if number, ok := version.TryI64(); ok && number == 0 {
				values["version"] = core.ValI64(1)
			}
		}
	}
	entity := g.state.metadata[operation.Entity]
	if entity != nil {
		for _, relation := range entity.Relations {
			if relation.IsMany || relation.LocKey == "id" || relation.ForKey != "id" {
				continue
			}
			field, err := core.CanonicalLedgerField(entity, relation.LocKey)
			if err != nil {
				continue
			}
			if value, exists := values[field]; exists {
				values[field] = g.resolvedKey(core.NewEntityKey(relation.TargetEntity, value)).ID
			}
		}
	}
	return values
}

type preparedRequest struct {
	kind                core.MutationKind
	entity              string
	id                  core.Value
	version             *int64
	values, old, guards core.Record
	hardDelete          bool
}

func describePreparedRequest(request data_service.MutationRequest) preparedRequest {
	switch item := request.(type) {
	case *data_service.InsertMutation:
		return preparedRequest{kind: core.MutationInsert, entity: item.Cmd.Entity, id: item.Cmd.Values["id"], values: item.Cmd.Values}
	case *data_service.UpdateMutation:
		return preparedRequest{kind: core.MutationUpdate, entity: item.Cmd.Entity, id: item.Cmd.Id, version: item.Cmd.ExpectedVersion, values: item.Cmd.Values, old: item.Cmd.OldValues, guards: item.Cmd.Guards}
	case *data_service.DeleteMutation:
		return preparedRequest{kind: core.MutationDelete, entity: item.Cmd.Entity, id: item.Cmd.Id, version: item.Cmd.ExpectedVersion, hardDelete: !item.Cmd.SoftDelete, guards: item.Cmd.Guards}
	case *data_service.RecoverMutation:
		return preparedRequest{kind: core.MutationRecover, entity: item.Cmd.Entity, id: item.Cmd.Id, version: &item.Cmd.ExpectedVersion, guards: item.Cmd.Guards}
	}
	panic("request capture admitted an unsupported mutation")
}

func (g *preparedProviderGuard) authorize(request data_service.MutationRequest, batch bool, indexes *[]int) error {
	if reason := request.Comment(); reason == nil || *reason != g.state.plan.AuditReason {
		return preparedMutationChanged(g.state.plan.RootEntity, "comment")
	}
	if group, ok := request.(*data_service.BatchMutation); ok {
		if len(group.Mutations) == 0 {
			return preparedIncomplete(g.state.plan.RootEntity)
		}
		for _, child := range group.Mutations {
			if err := g.authorize(child, true, indexes); err != nil {
				return err
			}
		}
		return nil
	}
	actual := describePreparedRequest(request)
	for index, operation := range g.state.plan.Operations {
		key := g.resolvedKey(core.NewEntityKey(operation.Entity, operation.ID))
		if key.Entity != actual.entity || !reflect.DeepEqual(key.ID, actual.id) {
			continue
		}
		if g.claimed[index] {
			return preparedMutationChanged(actual.entity, "id")
		}
		for _, pending := range *indexes {
			if pending == index {
				return preparedMutationChanged(actual.entity, "id")
			}
		}
		if operation.Kind != actual.kind || operation.HardDelete != actual.hardDelete {
			return preparedMutationChanged(actual.entity, "")
		}
		if operation.Kind != core.MutationInsert && !reflect.DeepEqual(operation.OriginalVersion, actual.version) {
			return preparedMutationChanged(actual.entity, "version")
		}
		expectedWrite := g.expectedValues(operation)
		if operation.Kind == core.MutationDelete || operation.Kind == core.MutationRecover {
			expectedWrite = nil
		}
		for _, pair := range []struct{ expected, actual core.Record }{{expectedWrite, actual.values}, {operation.OldValues, actual.old}, {operation.Guards, actual.guards}} {
			canonical, err := g.canonical(actual.entity, pair.actual)
			if err != nil {
				return checkerMappingError(actual.entity, err)
			}
			if field := changedPreparedField(pair.expected, canonical); field != "" {
				return preparedMutationChanged(actual.entity, field)
			}
		}
		expectedTrace := core.CloneTraceNodes(operation.TraceChain)
		if _, registered := g.state.lineage[preparedEntityKey(core.NewEntityKey(operation.Entity, operation.ID))]; registered {
			var err error
			expectedTrace, err = g.state.traceFor(core.NewEntityKey(operation.Entity, operation.ID), true, make(map[string]bool))
			if err != nil {
				return err
			}
		}
		if batch {
			command := core.NewUpdateCommand(operation.Entity, key.ID)
			command.TraceChain = expectedTrace
			wrapped, err := data_service.NewMutationRequest(&data_service.BatchMutation{Mutations: []data_service.MutationRequest{&data_service.UpdateMutation{Cmd: command}}}, g.state.plan.AuditReason)
			if err != nil {
				return err
			}
			expectedTrace = wrapped.(*data_service.BatchMutation).Mutations[0].TraceChain()
		}
		if len(expectedTrace) != len(request.TraceChain()) || len(expectedTrace) != 0 && !reflect.DeepEqual(expectedTrace, request.TraceChain()) {
			return preparedMutationChanged(actual.entity, "trace_chain")
		}
		*indexes = append(*indexes, index)
		return nil
	}
	return preparedMutationChanged(actual.entity, "id")
}

func (g *preparedProviderGuard) mutate(ctx context.Context, request data_service.MutationRequest, next data_service.MutationExecutor) (result *data_service.MutationResult, err error) {
	owned, err := data_service.CaptureMutationRequest(request)
	g.mu.Lock()
	if g.closed && err == nil {
		err = preparedIncomplete(g.state.plan.RootEntity)
	}
	if err == nil && g.fault != nil {
		err = g.fault
	}
	if err == nil && !g.metadataUnchanged() {
		err = preparedMutationChanged(g.state.plan.RootEntity, "metadata")
	}
	var indexes []int
	if err == nil {
		g.context.graphSaveMu.Lock()
		err = g.authorize(owned, false, &indexes)
		g.context.graphSaveMu.Unlock()
	}
	if err != nil {
		err = g.poison(err)
		g.mu.Unlock()
		return nil, err
	}
	for _, index := range indexes {
		g.claimed[index] = true
	}
	g.inflight++
	g.mu.Unlock()
	finished := false
	defer func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.inflight--
		if !finished {
			g.poison(preparedIncomplete(g.state.plan.RootEntity))
		} else if err != nil {
			g.poison(err)
		} else if len(indexes) == 1 && result != nil {
			if version, present := result.PersistedRecord["version"]; present {
				if number, ok := version.TryI64(); ok {
					g.resultVersions[indexes[0]] = number
				}
			}
		}
	}()
	result, err = next.Mutate(ctx, owned)
	if err == nil && (result == nil || result.AffectedRows != uint64(len(indexes))) {
		err = preparedIncomplete(g.state.plan.RootEntity)
	}
	finished = true
	return result, err
}

func (g *preparedProviderGuard) finishWork(ctx context.Context, err error) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.closed, g.allowFinish = true, true
	if err != nil {
		return err
	}
	if g.fault != nil {
		return g.fault
	}
	if ctx.Err() != nil {
		return g.poison(ctx.Err())
	}
	if g.inflight != 0 {
		return g.poison(preparedIncomplete(g.state.plan.RootEntity))
	}
	for _, claimed := range g.claimed {
		if !claimed {
			return g.poison(preparedIncomplete(g.state.plan.RootEntity))
		}
	}
	if !g.metadataUnchanged() {
		return g.poison(preparedMutationChanged(g.state.plan.RootEntity, "metadata"))
	}
	g.context.graphSaveMu.Lock()
	err = g.validateLiveSource()
	g.context.graphSaveMu.Unlock()
	if err != nil {
		return g.poison(err)
	}
	return nil
}

func (g *preparedProviderGuard) allowClose(commit bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.allowFinish {
		return g.poison(preparedIncomplete(g.state.plan.RootEntity))
	}
	if commit && g.fault != nil {
		return g.fault
	}
	return nil
}

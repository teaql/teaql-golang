package runtime

import (
	"errors"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func graphTraceAuditRequest(t *testing.T) data_service.MutationRequest {
	t.Helper()
	request, err := data_service.NewMutationRequest(&data_service.InsertMutation{
		Cmd: core.NewInsertCommand("CustomerOrder").Value("id", core.ValU64(100)),
	}, "submit order")
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func graphTestIntent() core.MutationIntent {
	comment := "submit order"
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		panic(err)
	}
	return intent
}

func TestGraphMutationAuditWaitsForCommit(t *testing.T) {
	probe := &graphTransactionProbe{}
	sink := &capturingAppAuditSink{}
	ctx := NewUserContext().WithAppAuditEventSink(sink)
	ctx.InsertResource("dataService", probe)
	err := ctx.ExecuteGraphSave(graphTestIntent(), func() error {
		if err := ctx.EmitMutationAudit(graphTraceAuditRequest(t), &data_service.MutationResult{AffectedRows: 1}); err != nil {
			return err
		}
		if len(sink.events) != 0 {
			return errors.New("committed audit escaped before transaction commit")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if probe.commits != 1 || len(sink.events) != 1 {
		t.Fatalf("commit=%d audit=%d", probe.commits, len(sink.events))
	}
}

func TestGraphMutationAuditIsDiscardedOnRollback(t *testing.T) {
	probe := &graphTransactionProbe{}
	sink := &capturingAppAuditSink{}
	ctx := NewUserContext().WithAppAuditEventSink(sink)
	ctx.InsertResource("dataService", probe)
	err := ctx.ExecuteGraphSave(graphTestIntent(), func() error {
		if err := ctx.EmitMutationAudit(graphTraceAuditRequest(t), &data_service.MutationResult{AffectedRows: 1}); err != nil {
			return err
		}
		return errors.New("later child failed")
	})
	if err == nil {
		t.Fatal("transaction must fail")
	}
	if len(sink.events) != 0 {
		t.Fatal("rolled-back transaction emitted a committed audit")
	}
}

type failingGraphAuditSink struct{}

func (failingGraphAuditSink) OnSafeEvent(*UserContext, *SafeAuditEvent) error {
	return errors.New("audit consumer unavailable")
}

func TestGraphCommittedAuditFailureDoesNotRollbackOrSkipCleanup(t *testing.T) {
	probe := &graphTransactionProbe{}
	ctx := NewUserContext().WithAppAuditEventSink(failingGraphAuditSink{})
	ctx.InsertResource("dataService", probe)
	cleaned := false
	err := ctx.ExecuteGraphSave(graphTestIntent(), func() error {
		ctx.AfterGraphCommit(func() { cleaned = true })
		return ctx.EmitMutationAudit(graphTraceAuditRequest(t), &data_service.MutationResult{AffectedRows: 1})
	})
	var committed *GraphCommittedError
	if !errors.As(err, &committed) || probe.commits != 1 || probe.active != 0 || !cleaned {
		t.Fatalf("error=%v commits=%d active=%d cleaned=%t", err, probe.commits, probe.active, cleaned)
	}
	if ctx.GetResource("dataService") != probe || ctx.graphSaveActive {
		t.Fatal("after-commit failure leaked transaction context")
	}
	ctx.SetAppAuditEventSink(nil)
	if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestGraphSaveRejectsMissingRootIntentBeforeCallbacksAndProvider(t *testing.T) {
	probe := &graphTransactionProbe{}
	ctx := NewUserContext()
	ctx.InsertResource("dataService", probe)
	called := false
	err := ctx.ExecutePreparedGraphSave(core.MutationIntent{}, func() (*MutationPlan, error) {
		called = true
		return nil, nil
	}, func() error { called = true; return nil })
	if err == nil || called || probe.begins != 0 {
		t.Fatalf("error=%v callbacks=%t begins=%d", err, called, probe.begins)
	}
	err = ctx.ExecuteGraphSave(core.MutationIntent{}, func() error { called = true; return nil })
	if err == nil || called || probe.begins != 0 {
		t.Fatal("direct graph callback bypassed request intent")
	}
}

type rawGraphSnapshotSink struct{ event *RawAuditEvent }

func (s *rawGraphSnapshotSink) OnEvent(_ *UserContext, event *RawAuditEvent) error {
	s.event = event
	return nil
}

func TestDeferredGraphAuditOwnsUpdateValues(t *testing.T) {
	probe := &graphTransactionProbe{}
	sink := &rawGraphSnapshotSink{}
	ctx := NewUserContext()
	ctx.setStandardAuditEventSink(sink)
	ctx.InsertResource("dataService", probe)
	command := core.NewUpdateCommand("CustomerOrder", core.ValU64(100)).Value("amount", core.ValI64(20))
	command.OldValues = core.Record{"amount": core.ValI64(10)}
	request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: command}, "submit order")
	if err != nil {
		t.Fatal(err)
	}
	err = ctx.ExecuteGraphSave(graphTestIntent(), func() error {
		if err := ctx.EmitMutationAudit(request, &data_service.MutationResult{AffectedRows: 1}); err != nil {
			return err
		}
		command.Values["amount"] = core.ValI64(999)
		command.OldValues["amount"] = core.ValI64(888)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if sink.event == nil || sink.event.Values["amount"].V != int64(20) || (*sink.event.OldValues)["amount"].V != int64(10) {
		t.Fatal("deferred event read caller-mutated command maps")
	}
}

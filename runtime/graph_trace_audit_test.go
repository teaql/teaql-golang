package runtime

import (
	"errors"
	"fmt"
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
	for _, logging := range []bool{false, true} {
		t.Run(fmt.Sprintf("logging=%t", logging), func(t *testing.T) {
			probe := &graphTransactionProbe{}
			sink := &capturingAppAuditSink{}
			registryCalls := 0
			ctx := NewUserContext().WithAppAuditEventSink(sink).WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
				registryCalls++
				return nil
			}))
			if logging {
				ctx.EnableAllSqlLog()
			} else {
				ctx.DisableSqlLog()
			}
			ctx.InsertResource("dataService", probe)
			called := false
			err := ctx.ExecutePreparedGraphSave(core.MutationIntent{}, func() (*MutationPlan, error) {
				called = true
				return nil, nil
			}, func() error { called = true; return nil })
			assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED", "comment", "mutation")
			err = ctx.ExecuteGraphSave(core.MutationIntent{}, func() error { called = true; return nil })
			assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED", "comment", "mutation")
			for _, blank := range []string{"", " \t\r\n", "\u0085", "\u00a0", "\u2003"} {
				err = ctx.ExecutePlannedGraphSave(&MutationPlan{RequestKey: "CustomerOrder.saveGraph", RootEntity: "CustomerOrder", AuditReason: blank}, func() error { called = true; return nil })
				assertIntentGate(t, err, "REQUEST_COMMENT_REQUIRED", "comment", "mutation")
			}
			if called || probe.begins != 0 || registryCalls != 0 || len(sink.events) != 0 {
				t.Fatalf("rejected graph reached downstream work: callback=%t begin=%d policy=%d audit=%d", called, probe.begins, registryCalls, len(sink.events))
			}
			// Rejection must not retain the graph gate or poison the next request.
			if err := ctx.ExecuteGraphSave(graphTestIntent(), func() error { called = true; return nil }); err != nil {
				t.Fatal(err)
			}
			if !called || probe.begins != 1 || probe.commits != 1 || probe.active != 0 {
				t.Fatal("valid graph did not recover after rejection")
			}
		})
	}
}

func TestGraphWorkErrorCannotClaimAnotherTransactionCommitted(t *testing.T) {
	probe := &graphTransactionProbe{}
	ctx := NewUserContext()
	ctx.InsertResource("dataService", probe)
	rolledBack, cleanedCommit := false, false
	err := ctx.ExecuteGraphSave(graphTestIntent(), func() error {
		ctx.AfterGraphRollback(func() { rolledBack = true })
		ctx.AfterGraphCommit(func() { cleanedCommit = true })
		return &data_service.MutationCommittedError{Cause: errors.New("another unit's consumer failed")}
	})
	if err == nil || !rolledBack || cleanedCommit || probe.commits != 0 || probe.active != 0 {
		t.Fatalf("work error claimed this uncommitted graph: err=%v rollback=%v commitCleanup=%v active=%d commits=%d", err, rolledBack, cleanedCommit, probe.active, probe.commits)
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

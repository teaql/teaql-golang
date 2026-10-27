package runtime

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type preparedProviderProbe struct {
	graphTransactionProbe
	requests      []data_service.MutationRequest
	beforeCapture func(data_service.MutationRequest)
	failure       error
	panicValue    any
	affected      *uint64
}

type preparedClockFunc func() time.Time

func (clock preparedClockFunc) Now() time.Time { return clock() }

func (probe *preparedProviderProbe) Begin(ctx context.Context) (data_service.Transaction, error) {
	tx, err := probe.graphTransactionProbe.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &preparedProviderTx{Transaction: tx, probe: probe}, nil
}

type preparedProviderTx struct {
	data_service.Transaction
	probe *preparedProviderProbe
}

func (tx *preparedProviderTx) Mutate(_ context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	if tx.probe.beforeCapture != nil {
		tx.probe.beforeCapture(request)
	}
	owned, err := data_service.CaptureMutationRequest(request)
	if err != nil {
		return nil, err
	}
	tx.probe.requests = append(tx.probe.requests, owned)
	if tx.probe.panicValue != nil {
		panic(tx.probe.panicValue)
	}
	if tx.probe.failure != nil {
		return nil, tx.probe.failure
	}
	affected := preparedLeafCount(owned)
	if tx.probe.affected != nil {
		affected = *tx.probe.affected
	}
	return &data_service.MutationResult{AffectedRows: affected}, nil
}

func preparedLeafCount(request data_service.MutationRequest) uint64 {
	if batch, ok := request.(*data_service.BatchMutation); ok {
		var total uint64
		for _, child := range batch.Mutations {
			total += preparedLeafCount(child)
		}
		return total
	}
	return 1
}

func preparedProviderContext() (*UserContext, *core.EntityRoot, core.EntityKey, *preparedProviderProbe) {
	ctx, root, probe := restorationContext(), core.NewEntityRoot(), &preparedProviderProbe{}
	ctx.InsertResource("dataService", probe)
	key := core.NewEntityKey("PurchaseOrder", core.ValU64(7))
	root.SetOriginalVersion(key, 3)
	root.Set(key, "phone", core.ValText("reviewed"))
	return ctx, root, key, probe
}

func preparedProviderPlan(ctx *UserContext, root *core.EntityRoot, key core.EntityKey) (*MutationPlan, error) {
	input := &CheckAndFixInput{Entity: key.Entity, Operation: core.MutationUpdate, Values: root.Change(key)}
	if err := CheckAndFixGraphEntity(ctx, input, root, key); err != nil {
		return nil, err
	}
	return MutationPlanFromEntityRoot(root, key.Entity, "submit order"), nil
}

func preparedProviderRequest(t *testing.T, key core.EntityKey) data_service.MutationRequest {
	t.Helper()
	request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: core.NewUpdateCommand(key.Entity, key.ID).WithExpectedVersion(3).Value("phone", core.ValText("reviewed"))}, "submit order")
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestGraphCompositionProviderRejectsRequestsChangedAfterBridgeReplay(t *testing.T) {
	for _, changed := range []string{"value", "target", "type", "version", "missingVersion", "reason", "trace", "guards", "old", "metadata"} {
		t.Run(changed, func(t *testing.T) {
			ctx, root, key, probe := preparedProviderContext()
			err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error {
				input := &CheckAndFixInput{Entity: key.Entity, Operation: core.MutationUpdate, Values: root.Change(key)}
				if err := CheckAndFixGraphEntity(ctx, input, root, key); err != nil {
					return err
				}
				request := preparedProviderRequest(t, key)
				update := request.(*data_service.UpdateMutation)
				switch changed {
				case "value":
					update.Cmd.Values["phone"] = core.ValText("not reviewed")
				case "target":
					update.Cmd.Id = core.ValU64(8)
				case "type":
					update.Cmd.Entity = "OrderItem"
				case "version":
					update.Cmd.WithExpectedVersion(4)
				case "missingVersion":
					update.Cmd.ExpectedVersion = nil
				case "reason":
					request, _ = data_service.NewMutationRequest(update, "not reviewed")
				case "trace":
					update.Cmd.TraceChain = []*core.TraceNode{core.NewTypedTraceNode("auditReason", "PurchaseOrder", "not reviewed")}
				case "guards":
					update.Cmd.Guards["phone"] = core.ValText("not reviewed")
				case "old":
					update.Cmd.OldValues = core.Record{"phone": core.ValText("not reviewed")}
				case "metadata":
					ctx.Entity("PurchaseOrder").Properties[0].ColName = "not_reviewed"
				}
				_, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
				return err
			})
			if err == nil || len(probe.requests) != 0 || probe.commits != 0 || probe.active != 0 {
				t.Fatalf("changed %s reached provider/commit: err=%v requests=%d commits=%d active=%d", changed, err, len(probe.requests), probe.commits, probe.active)
			}
		})
	}
}

func TestGraphCompositionProviderRequiresTheCompleteReviewedPlan(t *testing.T) {
	ctx, root, key, probe := preparedProviderContext()
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error { return nil })
	if err == nil || probe.commits != 0 || len(probe.requests) != 0 {
		t.Fatal("incomplete reviewed graph committed", err, probe.commits)
	}
}

func TestGraphCompositionProviderDuplicateAttemptPoisonsEvenIfSwallowed(t *testing.T) {
	ctx, root, key, probe := preparedProviderContext()
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error {
		provider := ctx.GetResource("dataService").(data_service.MutationExecutor)
		if _, err := provider.Mutate(ctx, preparedProviderRequest(t, key)); err != nil {
			return err
		}
		_, _ = provider.Mutate(ctx, preparedProviderRequest(t, key))
		return nil
	})
	if err == nil || probe.commits != 0 || len(probe.requests) != 1 || probe.active != 0 {
		t.Fatalf("duplicate error was swallowed: err=%v requests=%d commits=%d", err, len(probe.requests), probe.commits)
	}
}

func TestGraphCompositionProviderAllKindsAndNestedBatch(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(fmt.Sprint(batch), func(t *testing.T) {
			ctx, probe := NewUserContext(), &preparedProviderProbe{}
			ctx.InsertResource("dataService", probe)
			v3, deleted := int64(3), int64(-4)
			plan := &MutationPlan{RequestKey: "Document.saveGraph", RootEntity: "Document", AuditReason: "submit order", Operations: []MutationOperation{
				{Kind: core.MutationInsert, Entity: "Document", ID: core.ValU64(1), ChangedValues: core.Record{"name": core.ValText("created")}},
				{Kind: core.MutationUpdate, Entity: "Document", ID: core.ValU64(2), OriginalVersion: &v3, ChangedValues: core.Record{"name": core.ValText("changed")}},
				{Kind: core.MutationDelete, Entity: "Document", ID: core.ValU64(3), OriginalVersion: &v3},
				{Kind: core.MutationDelete, Entity: "Document", ID: core.ValU64(4), OriginalVersion: &v3, HardDelete: true},
				{Kind: core.MutationRecover, Entity: "Document", ID: core.ValU64(5), OriginalVersion: &deleted},
			}}
			commands := []data_service.MutationRequest{
				&data_service.InsertMutation{Cmd: core.NewInsertCommand("Document").Value("id", core.ValU64(1)).Value("name", core.ValText("created"))},
				&data_service.UpdateMutation{Cmd: core.NewUpdateCommand("Document", core.ValU64(2)).WithExpectedVersion(3).Value("name", core.ValText("changed"))},
				&data_service.DeleteMutation{Cmd: core.NewDeleteCommand("Document", core.ValU64(3)).WithExpectedVersion(3)},
				&data_service.DeleteMutation{Cmd: core.NewDeleteCommand("Document", core.ValU64(4)).WithExpectedVersion(3).HardDelete()},
				&data_service.RecoverMutation{Cmd: core.NewRecoverCommand("Document", core.ValU64(5), -4)},
			}
			if batch {
				commands = []data_service.MutationRequest{&data_service.BatchMutation{Mutations: []data_service.MutationRequest{commands[0], &data_service.BatchMutation{Mutations: commands[1:]}}}}
			}
			err := ctx.ExecutePlannedGraphSave(plan, func() error {
				for _, command := range commands {
					request, err := data_service.NewMutationRequest(command, "submit order")
					if err != nil {
						return err
					}
					if _, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request); err != nil {
						return err
					}
				}
				return nil
			})
			want := 5
			if batch {
				want = 1
			}
			if err != nil || probe.commits != 1 || probe.active != 0 || len(probe.requests) != want {
				t.Fatal("complete valid plan did not commit", err, probe.commits, len(probe.requests))
			}
		})
	}
}

func TestGraphCompositionProviderRejectsWholeInvalidBatchBeforeAnyChild(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		t.Run(fmt.Sprint(duplicate), func(t *testing.T) {
			ctx, root, key, probe := preparedProviderContext()
			err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error {
				first, second := preparedProviderRequest(t, key), preparedProviderRequest(t, key)
				if !duplicate {
					second.(*data_service.UpdateMutation).Cmd.Values["phone"] = core.ValText("not reviewed")
				}
				request, err := data_service.NewMutationRequest(&data_service.BatchMutation{Mutations: []data_service.MutationRequest{first, second}}, "submit order")
				if err != nil {
					return err
				}
				_, _ = ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
				return nil // A swallowed batch error still poisons the transaction.
			})
			if err == nil || len(probe.requests) != 0 || probe.commits != 0 || probe.active != 0 {
				t.Fatal("invalid batch entered provider", err, len(probe.requests), probe.commits)
			}
		})
	}
}

func TestGraphCompositionProviderRejectsIgnoredFailurePanicAndZeroRows(t *testing.T) {
	for _, failure := range []string{"error", "panic", "zero", "cancel", "earlyCommit", "earlyRollback"} {
		t.Run(failure, func(t *testing.T) {
			ctx, root, key, probe := preparedProviderContext()
			parent, cancel := context.WithCancel(ctx.Context)
			defer cancel()
			ctx.Context = parent
			switch failure {
			case "error":
				probe.failure = errors.New("synthetic provider failure")
			case "panic":
				probe.panicValue = errors.New("synthetic provider panic")
			case "zero":
				zero := uint64(0)
				probe.affected = &zero
			}
			err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error {
				provider := ctx.GetResource("dataService").(data_service.Transaction)
				if failure == "earlyCommit" {
					_ = provider.Commit(ctx)
				}
				if failure == "earlyRollback" {
					_ = provider.Rollback(ctx)
				}
				func() { defer func() { _ = recover() }(); _, _ = provider.Mutate(ctx, preparedProviderRequest(t, key)) }()
				if failure == "cancel" {
					cancel()
				}
				return nil
			})
			if err == nil || probe.commits != 0 || probe.active != 0 || ctx.preparedGraphSnapshot != nil {
				t.Fatal("failed transaction committed or leaked state", err, probe.commits, probe.active)
			}
		})
	}
}

func TestGraphCompositionProviderOwnsRequestAfterHandoff(t *testing.T) {
	ctx, root, key, probe := preparedProviderContext()
	request := preparedProviderRequest(t, key)
	entered, release, changed := make(chan struct{}), make(chan struct{}), make(chan struct{})
	probe.beforeCapture = func(data_service.MutationRequest) { close(entered); <-release }
	go func() {
		defer close(changed)
		<-entered
		request.(*data_service.UpdateMutation).Cmd.Values["phone"] = core.ValText("not reviewed")
		close(release)
	}()
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error {
		_, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
		return err
	})
	<-changed
	if err != nil || len(probe.requests) != 1 || probe.requests[0].(*data_service.UpdateMutation).Cmd.Values["phone"] != core.ValText("reviewed") {
		t.Fatal("caller mutation reached provider after handoff", err)
	}
}

func TestGraphCompositionProviderCannotCommitAndClearLaterLiveIntent(t *testing.T) {
	ctx, root, key, probe := preparedProviderContext()
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error {
		ctx.AfterGraphCommit(func() { root.ClearEntity(key) })
		if _, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, preparedProviderRequest(t, key)); err != nil {
			return err
		}
		root.Set(key, "phone", core.ValText("later live intent"))
		return nil
	})
	if err == nil || probe.commits != 0 || len(probe.requests) != 1 || root.Change(key)["phone"] != core.ValText("later live intent") {
		t.Fatalf("commit consumed unreviewed intent: err=%v commits=%d pending=%v", err, probe.commits, root.Change(key))
	}
}

func TestGraphCompositionPreparationFailureRetainsFixEvidenceAndZeroClock(t *testing.T) {
	ctx, _, _, probe := preparedProviderContext()
	reads := 0
	ctx.SetBusinessClock(preparedClockFunc(func() time.Time { reads++; return time.Time{} }))
	injected := errors.New("synthetic preparation failure")
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) {
		if !ctx.FixTime().IsZero() || !ctx.FixTime().IsZero() {
			t.Fatal("zero Business Clock changed")
		}
		if err := ctx.RecordFixEvidence(FixEvidence{EntityType: "PurchaseOrder", ModelPath: "create_time", Source: FixEvidenceClock, SourceLabel: "business-clock"}); err != nil {
			return nil, err
		}
		return nil, injected
	}, func() error { t.Fatal("work after failed preparation"); return nil })
	if err != injected || reads != 1 || len(ctx.LastFixEvidence()) != 1 || probe.begins != 0 {
		t.Fatal("failed preparation lost evidence or reread its clock", err, reads, ctx.LastFixEvidence())
	}
}

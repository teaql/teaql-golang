package runtime

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

func TestGraphCompositionProviderPreservesSignedZeroAndLargeJSON(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprint(changed), func(t *testing.T) {
			ctx, probe := NewUserContext(), &preparedProviderProbe{}
			ctx.InsertResource("dataService", probe)
			items := make([]any, 5000)
			for index := range items {
				items[index] = float64(index)
			}
			values := core.Record{"payload": core.ValJson(map[string]any{"items": items, "zero": math.Copysign(0, -1)})}
			version := int64(3)
			plan := &MutationPlan{RequestKey: "Document.saveGraph", RootEntity: "Document", AuditReason: "submit order", Operations: []MutationOperation{{Kind: core.MutationUpdate, Entity: "Document", ID: core.ValU64(7), OriginalVersion: &version, ChangedValues: values}}}
			err := ctx.ExecutePlannedGraphSave(plan, func() error {
				command := core.NewUpdateCommand("Document", core.ValU64(7)).WithExpectedVersion(3)
				command.Values = core.CloneRecord(values)
				if changed {
					command.Values["payload"].V.(map[string]any)["zero"] = float64(0)
				}
				request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: command}, "submit order")
				if err != nil {
					return err
				}
				_, err = ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
				return err
			})
			if changed {
				if err == nil || probe.commits != 0 || len(probe.requests) != 0 {
					t.Fatal("signed zero changed after review", err)
				}
			} else if err != nil || probe.commits != 1 {
				t.Fatal("valid large JSON rejected", err)
			}
		})
	}
}

func TestGraphCompositionProviderMetadataChangesDuringReviewStopBeforeBegin(t *testing.T) {
	ctx, root, key, probe := preparedProviderContext()
	ctx.WithMutationPolicyRegistry(MutationPolicyRegistryFunc(func(string) MutationPolicy {
		return &testMutationPolicy{identity: NewMutationPolicyIdentity("metadata-review", "1", "sha256:metadata"), review: func(*MutationPlan) MutationDecision {
			ctx.Entity(key.Entity).TabName = "unreviewed_table"
			return AllowMutation()
		}}
	}))
	err := ctx.ExecutePreparedGraphSave(graphTestIntent(), func() (*MutationPlan, error) { return preparedProviderPlan(ctx, root, key) }, func() error { t.Fatal("work after metadata replacement"); return nil })
	if err == nil || probe.begins != 0 || len(probe.requests) != 0 || len(root.Change(key)) != 1 {
		t.Fatal("metadata replacement reached transaction", err, probe.begins)
	}
}

func TestGraphCompositionProviderSoftDeleteCannotBecomeHardDelete(t *testing.T) {
	ctx, probe, version := NewUserContext(), &preparedProviderProbe{}, int64(3)
	ctx.InsertResource("dataService", probe)
	plan := &MutationPlan{RequestKey: "Document.saveGraph", RootEntity: "Document", AuditReason: "submit order", Operations: []MutationOperation{{Kind: core.MutationDelete, Entity: "Document", ID: core.ValU64(7), OriginalVersion: &version}}}
	err := ctx.ExecutePlannedGraphSave(plan, func() error {
		request, err := data_service.NewMutationRequest(&data_service.DeleteMutation{Cmd: core.NewDeleteCommand("Document", core.ValU64(7)).WithExpectedVersion(3).HardDelete()}, "submit order")
		if err != nil {
			return err
		}
		_, err = ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request)
		return err
	})
	if err == nil || probe.commits != 0 || len(probe.requests) != 0 {
		t.Fatal("soft-delete review authorized physical removal", err)
	}
}

func TestGraphCompositionProviderTypedTargetsWithTheSameIDRemainDistinct(t *testing.T) {
	ctx, probe, version := NewUserContext(), &preparedProviderProbe{}, int64(3)
	ctx.InsertResource("dataService", probe)
	plan := &MutationPlan{RequestKey: "Document.saveGraph", RootEntity: "Document", AuditReason: "submit order"}
	for _, entity := range []string{"Document", "Payment"} {
		plan.Operations = append(plan.Operations, MutationOperation{Kind: core.MutationUpdate, Entity: entity, ID: core.ValU64(7), OriginalVersion: &version, ChangedValues: core.Record{"name": core.ValText(entity)}})
	}
	err := ctx.ExecutePlannedGraphSave(plan, func() error {
		for _, entity := range []string{"Document", "Payment"} {
			request, err := data_service.NewMutationRequest(&data_service.UpdateMutation{Cmd: core.NewUpdateCommand(entity, core.ValU64(7)).WithExpectedVersion(3).Value("name", core.ValText(entity))}, "submit order")
			if err != nil {
				return err
			}
			if _, err := ctx.GetResource("dataService").(data_service.MutationExecutor).Mutate(ctx, request); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || probe.commits != 1 || len(probe.requests) != 2 {
		t.Fatal("typed targets collapsed", err)
	}
}

type preparedCapabilities struct {
	data_service.Transaction
	ids, floors, streams int
}

func (c *preparedCapabilities) GenerateId(string) (uint64, error) { c.ids++; return 71, nil }
func (c *preparedCapabilities) EnsureIdFloor(context.Context, string, uint64) error {
	c.floors++
	return nil
}
func (c *preparedCapabilities) QueryStream(context.Context, *data_service.QueryRequest, int, func(*data_service.StreamChunk) error) error {
	c.streams++
	return nil
}

func TestGraphCompositionProviderPreservesEveryOptionalCapabilityCombination(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		t.Run(fmt.Sprint(mask), func(t *testing.T) {
			probe := &preparedProviderProbe{}
			base := &preparedProviderTx{Transaction: &graphTransactionProbeTx{probe: &probe.graphTransactionProbe}, probe: probe}
			capabilities := &preparedCapabilities{Transaction: base}
			streamed := &graphStream{Transaction: base, streamer: capabilities}
			var inner data_service.Transaction
			switch mask {
			case 0:
				inner = base
			case 1:
				inner = &graphIDs{Transaction: base, generator: capabilities}
			case 2:
				inner = &graphFloor{Transaction: base, floor: capabilities}
			case 3:
				inner = &graphIDsFloor{graphIDs: &graphIDs{Transaction: base, generator: capabilities}, floor: capabilities}
			case 4:
				inner = streamed
			case 5:
				inner = &graphStreamIDs{graphStream: streamed, generator: capabilities}
			case 6:
				inner = &graphStreamFloor{graphStream: streamed, floor: capabilities}
			case 7:
				inner = &graphStreamIDsFloor{graphStreamIDs: &graphStreamIDs{graphStream: streamed, generator: capabilities}, floor: capabilities}
			}
			guard := &preparedProviderGuard{}
			wrapped := bindGraphTelemetryTransaction(inner, context.Background(), guard)
			ids, hasIDs := wrapped.(graphIDGenerator)
			floor, hasFloor := wrapped.(graphIDFloor)
			stream, hasStream := wrapped.(data_service.StreamQueryExecutor)
			if hasIDs != (mask&1 != 0) || hasFloor != (mask&2 != 0) || hasStream != (mask&4 != 0) {
				t.Fatal("optional method set changed")
			}
			if hasIDs {
				if id, err := ids.GenerateId("Document"); err != nil || id != 71 {
					t.Fatal("ID allocator not forwarded", err)
				}
			}
			if hasFloor {
				if err := floor.EnsureIdFloor(context.Background(), "Document", 70); err != nil {
					t.Fatal(err)
				}
			}
			if hasStream {
				if err := stream.QueryStream(context.Background(), nil, 1, func(*data_service.StreamChunk) error { return nil }); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

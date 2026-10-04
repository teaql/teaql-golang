package data_service

import (
	"reflect"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestBatchResponsibilityCaptureIsIdempotentAndKeepsKnownIdentity(t *testing.T) {
	for _, repeated := range []bool{false, true} {
		for _, kind := range []string{"insert", "update", "delete", "recover"} {
			t.Run(kind+map[bool]string{false: "/local", true: "/same-root"}[repeated], func(t *testing.T) {
				reason, local := "batch root", "local item"
				if repeated {
					local = reason
				}
				intent, _ := core.NewMutationIntent(&local)
				key := core.NewEntityKey("Customer", core.ValU64(17))
				scope, _ := core.MutationScopeForEntity(nil, key, intent, nil)
				trace := core.MutationTraceForEntity(core.NewEntityRoot(), key, scope)
				original := core.CloneTraceNodes(trace)
				var leaf MutationRequest
				switch kind {
				case "insert":
					command := core.NewInsertCommand("Customer").Value("id", key.ID)
					command.TraceChain = trace
					leaf = &InsertMutation{Cmd: command}
				case "update":
					command := core.NewUpdateCommand("Customer", key.ID)
					command.TraceChain = trace
					leaf = &UpdateMutation{Cmd: command}
				case "delete":
					command := core.NewDeleteCommand("Customer", key.ID)
					command.TraceChain = trace
					leaf = &DeleteMutation{Cmd: command}
				case "recover":
					command := core.NewRecoverCommand("Customer", key.ID, -1)
					command.TraceChain = trace
					leaf = &RecoverMutation{Cmd: command}
				}
				request, err := NewMutationRequest(&BatchMutation{Mutations: []MutationRequest{
					&BatchMutation{Mutations: []MutationRequest{leaf}},
				}}, reason)
				if err != nil {
					t.Fatal(err)
				}
				for attempt := 0; attempt < 3; attempt++ {
					request, err = CaptureMutationRequest(request)
					if err != nil {
						t.Fatal(err)
					}
				}
				owned := request.(*BatchMutation).Mutations[0].(*BatchMutation).Mutations[0]
				chain := owned.TraceChain()
				wantLength := 2
				if repeated {
					wantLength = 1
				}
				if len(chain) != wantLength || chain[0].Comment != reason || chain[0].EntityType != "Customer" ||
					chain[len(chain)-1].EntityId == nil || *chain[len(chain)-1].EntityId != 17 || chain[len(chain)-1].Comment != local {
					t.Fatalf("repeated capture lost item identity or duplicated root: %+v", chain)
				}
				if !repeated && chain[0].EntityId != nil {
					t.Fatal("batch invented a container identity")
				}
				if !reflect.DeepEqual(leaf.TraceChain(), original) {
					t.Fatal("capture modified caller-owned lineage")
				}
				*chain[len(chain)-1].EntityId = 99
				if *leaf.TraceChain()[0].EntityId != 17 {
					t.Fatal("captured lineage shares mutable identity with caller")
				}
			})
		}
	}
}

package tracechain_test

import (
	"fmt"
	"testing"

	"github.com/teaql/teaql-golang/data_service"
)

type graphIdentity struct {
	Entity string `json:"entity"`
	ID     uint64 `json:"id"`
}

func exactGraphIdentities(expected, actual []graphIdentity, boundary string) error {
	if len(expected) != 6 || len(actual) != len(expected) {
		return fmt.Errorf("%s identity count mismatch", boundary)
	}
	want := make(map[graphIdentity]bool)
	for _, key := range expected {
		want[key] = true
	}
	if len(want) != len(expected) {
		return fmt.Errorf("invalid expected graph")
	}
	seen := make(map[graphIdentity]bool)
	for _, key := range actual {
		if seen[key] {
			return fmt.Errorf("%s duplicate identity", boundary)
		}
		if key.ID == 0 || !want[key] {
			return fmt.Errorf("%s missing or unknown typed identity", boundary)
		}
		seen[key] = true
	}
	return nil
}

func assertExactGraphIdentities(t *testing.T, expected, actual []graphIdentity, boundary string) {
	t.Helper()
	if err := exactGraphIdentities(expected, actual, boundary); err != nil {
		t.Fatal(err)
	}
}

func observedMutationIdentity(request data_service.MutationRequest) (graphIdentity, error) {
	var key graphIdentity
	var present bool
	switch request := request.(type) {
	case *data_service.InsertMutation:
		key.Entity = request.Cmd.Entity
		key.ID, present = request.Cmd.Values["id"].TryU64()
	case *data_service.UpdateMutation:
		key.Entity = request.Cmd.Entity
		key.ID, present = request.Cmd.Id.TryU64()
	case *data_service.DeleteMutation:
		key.Entity = request.Cmd.Entity
		key.ID, present = request.Cmd.Id.TryU64()
	case *data_service.RecoverMutation:
		key.Entity = request.Cmd.Entity
		key.ID, present = request.Cmd.Id.TryU64()
	default:
		return key, fmt.Errorf("unsupported observed mutation %T", request)
	}
	if !present || key.ID == 0 {
		return key, fmt.Errorf("observed mutation omitted assigned target ID")
	}
	return key, nil
}

func graphIdentityControls(t *testing.T) {
	t.Helper()
	want := []graphIdentity{{"Customer Order", 100}, {"Order Item", 201}, {"Order Item", 202},
		{"Payment", 100}, {"Payment Attempt", 401}, {"Shipment", 501}}
	assertExactGraphIdentities(t, want, want, "positive control")
	for _, corruption := range []string{"duplicate", "missing", "type-collapse"} {
		t.Run(corruption, func(t *testing.T) {
			actual := append([]graphIdentity(nil), want...)
			switch corruption {
			case "duplicate":
				actual[2] = actual[1]
			case "missing":
				actual[5].ID = 999
			case "type-collapse":
				actual[3].Entity = "Customer Order"
			}
			if exactGraphIdentities(want, actual, corruption) == nil {
				t.Fatal("corrupted six-object identity set passed")
			}
		})
	}
	t.Log("PASS Go graph identity controls: duplicate, missing and equal-ID type collapse rejected")
}

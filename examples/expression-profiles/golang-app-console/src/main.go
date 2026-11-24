package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	dispatch "dispatch-document-core-workspace/lib"
	"github.com/teaql/teaql-golang/runtime"
)

const (
	reference = "DOCK-042"
	note      = "Lift-gate pickup after 14:00"
	dock      = "North Receiving Terminal"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 || (os.Args[2] != "first" && os.Args[2] != "restart") {
		return errors.New("supply PROFILE {first|restart}")
	}
	profile, phase := os.Args[1], os.Args[2]
	caller, err := dispatch.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	if err := dispatch.EnsureSchema(caller); err != nil {
		return err
	}
	if phase == "first" {
		order, err := dispatch.Q.DispatchOrders().Comment("Prepare the receiving order").
			Purpose("Set its loading instruction").NewEntity(caller)
		if err != nil {
			return err
		}
		order.UpdateDistributionCenterId(1).UpdateOrderReference(reference).UpdateHandlingNote(note)
		if _, err = order.AuditAs("Register the dispatch at the receiving terminal").Save(caller); err != nil {
			return err
		}
	}
	child := dispatch.Q.DispatchOrders().WithOrderReferenceIs(reference).Limit(1).Comment("Select the dispatch order")
	root, err := dispatch.Q.DistributionCenters().WithIdIs(1).SelectDispatchOrderListWith(child).Limit(1).
		Comment("Read the dispatch desk and its order").Purpose("Check the complete loading instruction").ExecuteForOne(caller)
	if err != nil || root == nil {
		return fmt.Errorf("missing dispatch desk: %v", err)
	}
	expression := dispatch.E.DistributionCenter(root)
	dockName, dockLoaded := expression.DockName().Eval()
	referenceValue, referenceLoaded := expression.DispatchOrderList().First().OrderReference().Eval()
	noteValue, noteLoaded := expression.DispatchOrderList().First().HandlingNote().Eval()
	if !dockLoaded || !referenceLoaded || !noteLoaded || dockName != dock || referenceValue != reference || noteValue != note {
		return errors.New("loaded navigation returned a different dispatch order")
	}

	matches, err := dispatch.Q.DispatchOrders().WithOrderReferenceIs(reference).Limit(2).
		Comment("Read back the dispatch record").Purpose("Check the typed JSON round trip").ExecuteForList(caller)
	if err != nil {
		return err
	}
	if len(matches.Data) != 1 || matches.Data[0].Id() != 1 || matches.Data[0].Version() != 1 {
		return errors.New("wrong retained dispatch record")
	}
	payload, err := matches.Data[0].ToJSON()
	if err != nil {
		return err
	}
	decoded, err := dispatch.DispatchOrderFromJSON(payload)
	if err != nil {
		return err
	}
	decodedReference, present := dispatch.E.DispatchOrder(decoded).OrderReference().Eval()
	decodedNote, hasNote := dispatch.E.DispatchOrder(decoded).HandlingNote().Eval()
	if !present || !hasNote || decodedReference != reference || decodedNote != note {
		return errors.New("typed JSON changed the dispatch payload")
	}

	sparse, err := dispatch.Q.DistributionCentersMinimal().WithIdIs(1).SelectDockName().Limit(1).
		Comment("Read the desk without its dispatch orders").Purpose("Verify the modeled preload instruction").ExecuteForOne(caller)
	if err != nil || sparse == nil {
		return fmt.Errorf("missing sparse dispatch desk: %v", err)
	}
	if name, loaded := dispatch.E.DistributionCenter(sparse).DockName().Eval(); !loaded || name != dock {
		return errors.New("the sparse projection lost its dock name")
	}
	_, _, err = dispatch.E.DistributionCenter(sparse).DispatchOrderList().First().OrderReference().TryEval()
	var missing *runtime.NotLoadedError
	if !errors.As(err, &missing) {
		return fmt.Errorf("expected the modeled preload diagnostic, got %v", err)
	}
	fmt.Print("DISPATCH_DOCUMENT_PROFILE=")
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"profile": profile, "phase": phase, "reference": referenceValue, "handlingNote": noteValue,
		"id": matches.Data[0].Id(), "version": matches.Data[0].Version(), "dock": dockName,
		"document": json.RawMessage(payload), "diagnostic": missing.ToWire(), "human": missing.Error(),
	})
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	orders "first-order-service-core-workspace/lib"
	"first-order-service-core-workspace/lib/purchase_order"
	"github.com/teaql/teaql-golang/runtime"
)

func main() {
	if err := placeFirstOrder(); err != nil {
		log.Fatal(err)
	}
}

func placeFirstOrder() error {
	label := os.Getenv("TEAQL_SHOWCASE_ROUND")
	if label == "" || os.Getenv("FIRST_ORDER_SERVICE_CORE_DATABASE_URL") == "" {
		return fmt.Errorf("supply an example round label and SQLite path")
	}
	context, err := orders.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	if err := orders.EnsureSchema(context); err != nil {
		return err
	}

	description := "Tea delivery " + label
	first, err := prepareOrder(context, description)
	if err != nil {
		return err
	}
	saved, err := first.AuditAs("Record the tea order requested by the delivery desk").Save(context)
	if err != nil {
		return err
	}
	found, err := orderByDescription(context, description)
	if err != nil {
		return err
	}
	if found == nil || found.Id() != saved.Id() || found.Description() != description {
		return fmt.Errorf("the delivery desk could not read its saved order")
	}

	badDescription := strings.Repeat("x", 101)
	rejected, err := prepareOrder(context, badDescription)
	if err != nil {
		return err
	}
	_, err = rejected.AuditAs("Check the purchase order description limit").Save(context)
	var validation *runtime.RuntimeError
	if !errors.As(err, &validation) || validation.Type != "Check" {
		return fmt.Errorf("expected description validation failure, got %v", err)
	}
	rejectedField := ""
	for _, check := range validation.CheckResults {
		if check.EntityType == "PurchaseOrder" && check.InstancePath() == "/description" {
			rejectedField = check.InstancePath()
		}
	}
	if rejectedField == "" {
		return fmt.Errorf("validation did not identify the order description")
	}
	absent, err := orderByDescription(context, badDescription)
	if err != nil {
		return err
	}
	if absent != nil {
		return fmt.Errorf("the rejected purchase order persisted")
	}

	fmt.Print("FIRST_ORDER=")
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"round": label, "id": found.Id(), "description": found.Description(), "rejectedField": rejectedField,
	})
}

func prepareOrder(context *runtime.UserContext, description string) (*purchase_order.PurchaseOrder, error) {
	order, err := orders.Q.PurchaseOrders().Comment("Initialize a purchase order").
		Purpose("Reserve identity for a new order from the delivery desk").NewEntity(context)
	if err != nil {
		return nil, err
	}
	return order.UpdatePlatformId(1).UpdateDescription(description), nil
}

func orderByDescription(context *runtime.UserContext, description string) (*purchase_order.PurchaseOrder, error) {
	return orders.Q.PurchaseOrders().WithDescriptionIs(description).Limit(2).
		Comment("Read the purchase order by description").
		Purpose("Check the delivery desk's persisted order").ExecuteForOne(context)
}

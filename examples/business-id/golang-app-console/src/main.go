package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	ordering "purchase-order-numbering-core-workspace/lib"
	"purchase-order-numbering-core-workspace/lib/purchase_order"
)

const rejectedSupplier = "Unaudited Supplier"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	database := os.Getenv("PURCHASE_ORDER_NUMBERING_CORE_DATABASE_URL")
	round := os.Getenv("TEAQL_PURCHASE_NUMBERING_ROUND")
	if database == "" || round == "" {
		return fmt.Errorf("supply an example database and a round label")
	}

	context, err := ordering.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	context.WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)))
	if err := ordering.EnsureSchema(context); err != nil {
		return err
	}

	first, err := createOrder(context, "Northwind-"+round)
	if err != nil {
		return err
	}
	second, err := createOrder(context, "Contoso-"+round)
	if err != nil {
		return err
	}
	firstID, secondID := first.Id(), second.Id()
	firstNumber, secondNumber := first.OrderNumber(), second.OrderNumber()
	format := regexp.MustCompile(`^ORD-20261001-[0-9A-Z]{6}$`)
	if firstID == secondID || firstNumber == secondNumber ||
		!format.MatchString(firstNumber.String()) || !format.MatchString(secondNumber.String()) {
		return fmt.Errorf("expected distinct purchase orders numbered on the Context business day")
	}

	retry, err := first.AuditAs("Retry recording Northwind's purchase order").Save(context)
	if err != nil {
		return err
	}
	if retry.Id() != firstID || retry.OrderNumber() != firstNumber {
		return fmt.Errorf("retry changed the purchase order identity")
	}

	for _, order := range []*purchase_order.PurchaseOrder{first, second} {
		persisted, err := ordering.Q.PurchaseOrderByOrderNumber(order.OrderNumber()).
			Comment("Read a purchase order by its assigned number").
			Purpose("Verify the procurement record and allocation").ExecuteForOne(context)
		if err != nil {
			return err
		}
		if persisted == nil || persisted.Id() != order.Id() || persisted.OrderNumber() != order.OrderNumber() ||
			persisted.Supplier() != order.Supplier() {
			return fmt.Errorf("typed order-number query returned a different purchase order")
		}
	}

	rejected, err := prepareOrder(context, rejectedSupplier)
	if err != nil {
		return err
	}
	failure := rejectBlankAudit(rejected)
	intent, ok := failure.(*core.RequestIntentError)
	if !ok || intent.Code != "REQUEST_COMMENT_REQUIRED" || intent.RequestKind != "mutation" {
		return fmt.Errorf("expected missing mutation comment rejection, got %v", failure)
	}
	found, err := ordering.Q.PurchaseOrders().WithSupplierIs(rejectedSupplier).Limit(2).
		Comment("Look for the unaudited purchase order").
		Purpose("Prove the rejected operation stored no order").ExecuteForOne(context)
	if err != nil {
		return err
	}
	if found != nil {
		return fmt.Errorf("unaudited purchase order persisted")
	}

	fmt.Print("PURCHASE_ORDER_NUMBERING=")
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"round": round, "firstId": firstID, "firstNumber": firstNumber.String(),
		"secondId": secondID, "secondNumber": secondNumber.String(),
		"retryNumber": retry.OrderNumber().String(), "rejection": intent.Code, "typedReadback": true,
	})
}

// The fluent builder rejects a blank comment before a save can be made.
func rejectBlankAudit(order *purchase_order.PurchaseOrder) (failure any) {
	defer func() { failure = recover() }()
	order.AuditAs(" ")
	return nil
}

func prepareOrder(context *runtime.UserContext, supplier string) (*purchase_order.PurchaseOrder, error) {
	order, err := ordering.Q.PurchaseOrders().
		Comment("Initialize the purchase order").
		Purpose("Record a procurement commitment").NewEntity(context)
	if err != nil {
		return nil, err
	}
	return order.UpdatePlatformId(1).UpdateSupplier(supplier), nil
}

func createOrder(context *runtime.UserContext, supplier string) (*purchase_order.PurchaseOrder, error) {
	order, err := prepareOrder(context, supplier)
	if err != nil {
		return nil, err
	}
	return order.AuditAs("Record the approved purchase order").Save(context)
}

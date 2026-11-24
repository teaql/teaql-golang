package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	freight "freight-quote-settlement-core-workspace/lib"
	"freight-quote-settlement-core-workspace/lib/freight_quote"
	"github.com/shopspring/decimal"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

const firstShipment = 9_007_199_254_740_993

var preciseAmount = decimal.RequireFromString("123456789012.3456789")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	phase := os.Getenv("TEAQL_FREIGHT_QUOTE_PHASE")
	if phase != "first" && phase != "restart" {
		return fmt.Errorf("set TEAQL_FREIGHT_QUOTE_PHASE to first or restart")
	}
	context, err := freight.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	if err := freight.EnsureSchema(context); err != nil {
		return err
	}
	date := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	pickup, err := core.NewTimeOfDay(9, 30, 15, 123456000)
	if err != nil {
		return err
	}
	quotedAt := time.Date(2026, 10, 31, 1, 45, 0, 123000000, time.FixedZone("Dispatch", 8*60*60))
	if phase == "first" {
		if err := registerQuotes(context, date, pickup, quotedAt); err != nil {
			return err
		}
	}

	rows, err := freight.Q.FreightQuotes().OrderByAmountAsc().Limit(4).
		Comment("Compare outstanding freight bids").Purpose("Rank freight bids numerically").ExecuteForList(context)
	if err != nil {
		return err
	}
	if len(rows.Data) != 3 {
		return fmt.Errorf("expected three persisted freight quotes, got %d", len(rows.Data))
	}
	amounts := []string{"0.10", "0.20", "123456789012.3456789"}
	for i, quote := range rows.Data {
		expression := freight.E.FreightQuote(quote)
		amount, present, err := expression.Amount().TryEval()
		if err != nil || !present || amount.StringFixed(-amount.Exponent()) != amounts[i] ||
			quote.Reference() != fmt.Sprintf("SEA-%03d", i+1) || quote.ShipmentNumber() != firstShipment+int64(i) {
			return fmt.Errorf("quote %d changed its rank, rate or shipment number: %v", i+1, err)
		}
	}
	empty, inspection := rows.Data[1].Instructions(), rows.Data[2].Instructions()
	if rows.Data[0].Instructions() != nil || empty == nil || *empty != "" ||
		inspection == nil || *inspection != "Hold for inspection" {
		return fmt.Errorf("absent and empty delivery instructions were conflated")
	}
	high := rows.Data[2]
	match, err := freight.Q.FreightQuotes().WithAmountIs(preciseAmount).Limit(1).
		Comment("Find the revised offshore freight rate").Purpose("Match an exact Decimal amount").ExecuteForOne(context)
	if err != nil || match == nil || match.Id() != high.Id() || high.Version() != 2 {
		return fmt.Errorf("exact amount selected a different freight quote: %v", err)
	}
	summary, err := freight.Q.FreightQuotes().SumAmountAs("total").
		Comment("Sum the outstanding freight bids").Purpose("Keep exact base-ten precision").ExecuteForRows(context)
	if err != nil {
		return err
	}
	if len(summary.Data) != 1 {
		return fmt.Errorf("expected one aggregate row, got %d", len(summary.Data))
	}
	total, ok := summary.Data[0]["total"].TryDecimal()
	if !ok || total.String() != "123456789012.6456789" {
		return fmt.Errorf("unexpected freight total %s", total)
	}
	if err := verifyWire(context, high, date, pickup, quotedAt); err != nil {
		return err
	}

	high.UpdatePackageCount(int64(math.MaxInt32) + 1)
	_, err = high.AuditAs("Reject a quote with a shipment size outside its modeled Integer domain").Save(context)
	var checked *runtime.RuntimeError
	if !errors.As(err, &checked) || checked.Type != "Check" {
		return fmt.Errorf("expected a modeled package count rejection, got %v", err)
	}
	widthRejected := false
	for _, result := range checked.CheckResults {
		widthRejected = widthRejected || result.ModelPath() == "package_count" && result.RuleID == "invalid_type"
	}
	retained, err := readQuote(context, "SEA-003")
	if err != nil {
		return err
	}
	if !widthRejected || retained.Version() != 2 || !retained.Amount().Equal(preciseAmount) ||
		retained.PackageCount() != math.MaxInt32 {
		return fmt.Errorf("the rejected package count changed the retained quote or failed at a different field")
	}

	fmt.Print("FREIGHT_QUOTE=")
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"phase": phase, "quotes": len(rows.Data), "total": total.String(), "precise": retained.Amount().String(),
		"version": retained.Version(), "packageCountOverflowRejected": true, "sparseInstructionsNotLoaded": true,
	})
}

func registerQuotes(context *runtime.UserContext, date time.Time, pickup core.TimeOfDay, quotedAt time.Time) error {
	empty, inspection := "", "Hold for inspection"
	seeds := []struct {
		reference, amount string
		packages          int64
		instructions      *string
	}{
		{"SEA-001", "0.10", 2, nil},
		{"SEA-002", "0.20", 3, &empty},
		{"SEA-003", "3.50", math.MaxInt32, &inspection},
	}
	for i, seed := range seeds {
		quote, err := freight.Q.FreightQuotes().Comment("Prepare a freight quote").
			Purpose("Price the shipment for its delivery window").NewEntity(context)
		if err != nil {
			return err
		}
		quote.UpdatePlatformId(1).UpdateReference(seed.reference).UpdateAmount(decimal.RequireFromString(seed.amount)).
			UpdateShipmentNumber(firstShipment + int64(i)).UpdatePackageCount(seed.packages).UpdateDeliveryDate(date).
			UpdatePickupTime(pickup).UpdateQuotedAt(quotedAt).UpdateInstructions(seed.instructions)
		if _, err := quote.AuditAs("Register a freight quote").Save(context); err != nil {
			return err
		}
	}
	revised, err := readQuote(context, "SEA-003")
	if err != nil {
		return err
	}
	if amount := revised.Amount(); amount.Exponent() != -2 || amount.StringFixed(2) != "3.50" {
		return fmt.Errorf("the initial offshore quote lost its Decimal scale")
	}
	_, err = revised.UpdateAmount(preciseAmount).AuditAs("Accept the revised offshore freight rate").Save(context)
	return err
}

func readQuote(context *runtime.UserContext, reference string) (*freight_quote.FreightQuote, error) {
	quote, err := freight.Q.FreightQuotes().WithReferenceIs(reference).Limit(1).
		Comment("Read the selected freight quote").Purpose("Verify the retained delivery commitment").ExecuteForOne(context)
	if err != nil {
		return nil, err
	}
	if quote == nil {
		return nil, fmt.Errorf("missing freight quote %s", reference)
	}
	return quote, nil
}

func verifyWire(context *runtime.UserContext, high *freight_quote.FreightQuote, date time.Time, pickup core.TimeOfDay, quotedAt time.Time) error {
	payload, err := high.ToJSON()
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		return err
	}
	if string(fields["amount"]) != `"123456789012.3456789"` ||
		string(fields["shipmentNumber"]) != `"9007199254740995"` ||
		string(fields["packageCount"]) != "2147483647" {
		return fmt.Errorf("JSON changed Decimal, Long or Integer width")
	}
	restored, err := freight.FreightQuoteFromJSON(payload)
	if err != nil {
		return err
	}
	expression := freight.E.FreightQuote(restored)
	amount, amountPresent := expression.Amount().Eval()
	shipment, shipmentPresent := expression.ShipmentNumber().Eval()
	delivery, deliveryPresent := expression.DeliveryDate().Eval()
	clock, clockPresent := expression.PickupTime().Eval()
	quoted, quotedPresent := expression.QuotedAt().Eval()
	if !amountPresent || !shipmentPresent || !deliveryPresent || !clockPresent || !quotedPresent ||
		!amount.Equal(preciseAmount) || shipment != firstShipment+2 ||
		!delivery.Equal(date) || clock != pickup || !quoted.Equal(quotedAt) {
		return fmt.Errorf("generated E changed the decoded freight commitment")
	}

	sparse, err := freight.Q.FreightQuotesMinimal().WithIdIs(high.Id()).SelectReference().Limit(1).
		Comment("Read only the quote reference").Purpose("Keep unselected delivery instructions distinct").ExecuteForOne(context)
	if err != nil || sparse == nil {
		return fmt.Errorf("missing sparse quote: %v", err)
	}
	wire, err := sparse.ToJSON()
	if err != nil {
		return err
	}
	var selected map[string]json.RawMessage
	if err := json.Unmarshal(wire, &selected); err != nil {
		return err
	}
	decoded, err := freight.FreightQuoteFromJSON(wire)
	if err != nil {
		return err
	}
	_, _, err = freight.E.FreightQuote(decoded).Instructions().TryEval()
	var missing *runtime.NotLoadedError
	if _, included := selected["instructions"]; included || !errors.As(err, &missing) {
		return fmt.Errorf("unselected instructions became a value: %v", err)
	}
	return nil
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/teaql/teaql-golang/runtime"
	domain "merchant-order-privacy-service-core-workspace/lib"
	order "merchant-order-privacy-service-core-workspace/lib/customer_order"
)

func placeOrder(context *runtime.UserContext, merchant uint64, reference, email string) (*order.CustomerOrder, error) {
	profile, err := domain.Q.CustomerProfiles().Comment("Prepare a delivery profile").
		Purpose("Associate a customer with one merchant").NewEntity(context)
	if err != nil {
		return nil, err
	}
	profile.UpdateMerchantId(merchant)
	profile.UpdateEmail(email)
	profile, err = profile.AuditAs("Record the customer's delivery profile").Save(context)
	if err != nil {
		return nil, err
	}
	shipment, err := domain.Q.CustomerOrders().Comment("Prepare an order for dispatch").
		Purpose("Record the merchant's order and delivery profile").NewEntity(context)
	if err != nil {
		return nil, err
	}
	shipment.UpdateMerchantId(merchant)
	shipment.UpdateCustomerProfileId(profile.Id())
	shipment.UpdateOrderNumber(reference)
	shipment.UpdateFulfillmentStatus("ready")
	return shipment.AuditAs("Queue a merchant order for dispatch").Save(context)
}

func lookup(context *runtime.UserContext, merchant, id uint64, reference, email string) (*order.CustomerOrder, error) {
	return domain.Q.CustomerOrders().WithIdIs(id).WithOrderNumberIs(reference).
		WithMerchantMatching(domain.Q.MerchantsMinimal().WithIdIs(merchant)).
		WithCustomerProfileMatching(domain.Q.CustomerProfilesMinimal().WithEmailIs(email)).
		SelectCustomerProfile().Limit(1).Comment("Load the scoped order and delivery profile").
		Purpose("Review a dispatch order for the authenticated merchant").ExecuteForOne(context)
}

func run() error {
	if len(os.Args) != 5 || (os.Args[1] != "1" && os.Args[1] != "2") {
		return errors.New("usage: order-privacy ROUND DATABASE ORDINARY_SQL DEDICATED_SQL")
	}
	round := os.Args[1]
	if err := os.Setenv("MERCHANT_ORDER_PRIVACY_SERVICE_CORE_DATABASE_URL", os.Args[2]); err != nil {
		return err
	}
	ordinary, err := os.OpenFile(os.Args[3], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer ordinary.Close()
	dedicated, err := os.OpenFile(os.Args[4], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer dedicated.Close()
	context, err := domain.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	defer context.GetResource("db").(*sql.DB).Close()
	// This key and both merchant accounts belong to a disposable example.
	codec, err := runtime.NewAEADEntityReferenceCodec(1, map[uint32][]byte{1: bytes.Repeat([]byte{0x27}, 32)})
	if err != nil {
		return err
	}
	context.WithEntityReferenceCodec(codec).
		WithDiagnosticSQLLogSink(runtime.NewTextDiagnosticSQLLogSink(ordinary)).
		WithSensitiveDiagnosticSQLLogSink(runtime.NewSensitiveDiagnosticSQLLogSink(dedicated))
	context.EnableAllSqlLog()
	if err := domain.EnsureSchema(context); err != nil {
		return err
	}
	primary, err := domain.Q.MerchantsMinimal().WithIdIs(1).Limit(1).
		Comment("Read the merchant provisioned by the model").
		Purpose("Set the trusted dispatch scope").ExecuteForOne(context)
	if err != nil {
		return err
	}
	if primary == nil {
		return errors.New("the primary merchant was not provisioned")
	}
	secondary, err := domain.Q.Merchants().Comment("Prepare an independent merchant").
		Purpose("Provide a second merchant with a dispatch order").NewEntity(context)
	if err != nil {
		return err
	}
	secondary.UpdateName("Independent retailer " + round)
	secondary, err = secondary.AuditAs("Register a second merchant").Save(context)
	if err != nil {
		return err
	}

	reference, email := "SHIP-PRIVATE-"+round, "sealed-"+round+"@example.test"
	foreignRef, foreignEmail := "SHIP-OTHER-"+round, "foreign-"+round+"@example.test"
	own, err := placeOrder(context, primary.Id(), reference, email)
	if err != nil {
		return err
	}
	foreign, err := placeOrder(context, secondary.Id(), foreignRef, foreignEmail)
	if err != nil {
		return err
	}
	purpose := fmt.Sprintf("view-dispatch:%d", primary.Id())
	token, err := context.EncodeEntityReference("CustomerOrder", own.Id(), own.Version(), purpose, 5*time.Minute)
	if err != nil {
		return err
	}
	for _, deniedPurpose := range []string{"edit-dispatch", fmt.Sprintf("view-dispatch:%d", secondary.Id())} {
		var tokenError *runtime.EntityReferenceTokenError
		if _, err := context.DecodeEntityReference(token, "CustomerOrder", deniedPurpose); !errors.As(err, &tokenError) {
			return errors.New("an unauthorized dispatch reference was not rejected")
		}
	}
	claims, err := context.DecodeEntityReference(token, "CustomerOrder", purpose)
	if err != nil {
		return err
	}
	loaded, err := lookup(context, primary.Id(), claims.ID, reference, email)
	if err != nil {
		return err
	}
	if loaded == nil || loaded.Id() != own.Id() {
		return errors.New("the authorized dispatch order was not returned")
	}
	expression := order.NewCustomerOrderExpression(loaded)
	profile, present := expression.CustomerProfile().Eval()
	if !present {
		return errors.New("the selected customer profile was not loaded")
	}
	actualEmail, emailPresent := profile["email"].TryText()
	version, versionPresent := expression.Version().Eval()
	if !emailPresent || actualEmail != email || !versionPresent || version != claims.Version {
		return errors.New("the scoped dispatch data did not round-trip")
	}
	denied, err := lookup(context, primary.Id(), foreign.Id(), foreignRef, foreignEmail)
	if err != nil {
		return err
	}
	if denied != nil {
		return errors.New("the primary merchant read a second merchant's order")
	}
	result, err := json.Marshal(map[string]any{
		"round": round, "authorizedOrder": loaded.Id(), "foreignOrder": foreign.Id(),
		"merchant": primary.Id(), "foreignMerchant": secondary.Id(), "version": version,
		"purposeDenied": true, "foreignDenied": true,
	})
	if err == nil {
		fmt.Printf("ORDER_PRIVACY=%s\n", result)
	}
	return err
}

func main() {
	if err := run(); err != nil {
		// Provider errors can include SQL or application values.
		fmt.Fprintf(os.Stderr, "order privacy example failed (%T)\n", err)
		os.Exit(1)
	}
}

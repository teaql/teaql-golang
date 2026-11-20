package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	domain "customer-onboarding-service-core-workspace/lib"
	"customer-onboarding-service-core-workspace/lib/customer"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
)

const (
	blockedEmail = "blocked@example.com"
	policyKey    = "customer.saveGraph"
	denialCode   = "CUSTOMER_EMAIL_BLOCKED"
)

type approvalPolicy struct {
	identity runtime.MutationPolicyIdentity
	reviewed []int
}

func (p *approvalPolicy) Identity() runtime.MutationPolicyIdentity { return p.identity }

func (p *approvalPolicy) Review(_ *runtime.UserContext, plan *runtime.MutationPlan) runtime.MutationDecision {
	p.reviewed = append(p.reviewed, len(plan.Operations))
	if plan.RequestKey != policyKey || plan.RootEntity != "customer" || len(plan.Operations) != 2 {
		return runtime.DenyMutation("CUSTOMER_GRAPH_INCOMPLETE", "the full onboarding graph is required")
	}
	var profile, contact, blocked bool
	for _, operation := range plan.Operations {
		if operation.Kind != core.MutationInsert {
			return runtime.DenyMutation("CUSTOMER_GRAPH_INCOMPLETE", "expected new onboarding records")
		}
		switch operation.Entity {
		case "customer":
			profile = true
			blocked = operation.ChangedValues["email"] == core.ValText(blockedEmail)
		case "customer_contact":
			contact = true
		}
	}
	if !profile || !contact {
		return runtime.DenyMutation("CUSTOMER_GRAPH_INCOMPLETE", "profile and contact are both required")
	}
	if blocked {
		return runtime.DenyMutation(denialCode, "reserved customer email", "customer.email")
	}
	return runtime.AllowMutation()
}

type approvalAudit struct {
	events []*runtime.SafeAuditEvent
}

func (a *approvalAudit) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	a.events = append(a.events, event)
	return nil
}

func main() {
	if err := verifyApproval(); err != nil {
		log.Fatal(err)
	}
}

func verifyApproval() error {
	label := os.Getenv("TEAQL_SHOWCASE_ROUND")
	if label == "" || os.Getenv("CUSTOMER_ONBOARDING_SERVICE_CORE_DATABASE_URL") == "" {
		return fmt.Errorf("supply an example round label and SQLite path")
	}
	context, err := domain.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	if err := domain.EnsureSchema(context); err != nil {
		return err
	}
	identity := runtime.NewMutationPolicyIdentity("example.customer-onboarding", "1", "sha256:customer-onboarding-v1")
	policy := &approvalPolicy{identity: identity}
	audit := &approvalAudit{}
	context.WithMutationPolicyRegistry(runtime.MutationPolicyRegistryFunc(func(key string) runtime.MutationPolicy {
		if key == policyKey {
			return policy
		}
		return nil
	})).WithMutationPolicyApprovalProvider(runtime.MutationPolicyApprovalProviderFunc(
		func(candidate runtime.MutationPolicyIdentity) *runtime.MutationPolicyApproval {
			if candidate == identity {
				return &runtime.MutationPolicyApproval{Policy: candidate, ApprovedBy: "customer-review-owner", ApprovedAt: time.Unix(1, 0)}
			}
			return nil
		})).WithAppAuditEventSink(audit)

	email, channel := "approved-"+label+"@example.com", "Email-"+label
	approved, err := prepareCustomer(context, email, channel)
	if err != nil {
		return err
	}
	if _, err := approved.AuditAs("Approve customer onboarding and its contact preference").Save(context); err != nil {
		return err
	}

	blocked, err := prepareCustomer(context, blockedEmail, "Rejected-"+label)
	if err != nil {
		return err
	}
	_, err = blocked.AuditAs("Review the blocked customer onboarding submission").Save(context)
	var policyError *runtime.RuntimeError
	if !errors.As(err, &policyError) || policyError.Type != "MutationPolicy" ||
		!strings.Contains(policyError.Error(), denialCode) {
		return fmt.Errorf("expected %s, got %v", denialCode, err)
	}
	if len(policy.reviewed) != 2 || policy.reviewed[0] != 2 || policy.reviewed[1] != 2 {
		return fmt.Errorf("each submission must reach policy review as a two-operation graph")
	}
	for _, event := range audit.events {
		snapshot := event.MutationGovernance
		if snapshot == nil || snapshot.Policy == nil || *snapshot.Policy != identity ||
			snapshot.Approval != runtime.MutationPolicyApprovalApproved || len(snapshot.Operations) != 2 {
			return fmt.Errorf("onboarding audit is missing its approved graph identity")
		}
	}
	if len(audit.events) != 2 {
		return fmt.Errorf("expected only the approved profile and contact audit events")
	}

	found, err := profileByEmail(context, email)
	if err != nil {
		return err
	}
	if found == nil || found.Id() != approved.Id() || found.Email() != email {
		return fmt.Errorf("the approved profile was not persisted")
	}
	contact, err := domain.Q.CustomerContacts().WithCustomerIs(found.Id()).WithChannelIs(channel).Limit(2).
		Comment("Read the approved contact preference").Purpose("Confirm it was saved with its customer").ExecuteForOne(context)
	if err != nil {
		return err
	}
	if contact == nil || contact.Channel() != channel {
		return fmt.Errorf("the approved customer contact was not persisted")
	}
	absent, err := profileByEmail(context, blockedEmail)
	if err != nil {
		return err
	}
	deniedContact, err := domain.Q.CustomerContacts().WithChannelIs("Rejected-" + label).Limit(2).
		Comment("Look for a contact belonging to the denied submission").
		Purpose("Verify policy rejection left both business tables unchanged").ExecuteForOne(context)
	if err != nil {
		return err
	}
	if absent != nil || deniedContact != nil {
		return fmt.Errorf("the denied customer graph partially persisted")
	}
	fmt.Print("CUSTOMER_APPROVAL=")
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"round": label, "customerId": found.Id(), "email": email, "contactId": contact.Id(), "channel": channel,
		"reviewedOperations": policy.reviewed, "approvedAudits": len(audit.events), "denial": denialCode,
	})
}

func prepareCustomer(context *runtime.UserContext, email, channel string) (*customer.Customer, error) {
	profile, err := domain.Q.Customers().Comment("Initialize the customer profile").
		Purpose("Prepare one governed customer onboarding submission").NewEntity(context)
	if err != nil {
		return nil, err
	}
	profile.UpdatePlatformId(1).UpdateEmail(email)
	contact, err := domain.Q.CustomerContacts().Comment("Initialize the contact preference").
		Purpose("Include the consent data in the same customer submission").NewEntity(context)
	if err != nil {
		return nil, err
	}
	contact.UpdateCustomerId(profile.Id()).UpdateChannel(channel)
	profile.CustomerContactList().Add(contact)
	return profile, nil
}

func profileByEmail(context *runtime.UserContext, email string) (*customer.Customer, error) {
	return domain.Q.Customers().WithEmailIs(email).Limit(2).Comment("Read a customer profile by email").
		Purpose("Check what was persisted after business approval").ExecuteForOne(context)
}

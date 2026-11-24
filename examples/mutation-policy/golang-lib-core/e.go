package lib

import (
	"customer-onboarding-service-core-workspace/lib/platform"
	"customer-onboarding-service-core-workspace/lib/customer"
	"customer-onboarding-service-core-workspace/lib/customer_contact"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) Customer(value *customer.Customer) *customer.CustomerExpression {
	return customer.NewCustomerExpression(value)
}

func (expressionFacade) CustomerContact(value *customer_contact.CustomerContact) *customer_contact.CustomerContactExpression {
	return customer_contact.NewCustomerContactExpression(value)
}

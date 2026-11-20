package lib

import (
	"customer-onboarding-service-core-workspace/lib/platform"
	"customer-onboarding-service-core-workspace/lib/customer"
	"customer-onboarding-service-core-workspace/lib/customer_contact"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) Customers() *customer.CustomerRequest {
	return customer.NewCustomerRequest()
}

func (q *QType) CustomersMinimal() *customer.CustomerRequest {
	return customer.NewCustomerMinimalRequest()
}

func (q *QType) CustomerContacts() *customer_contact.CustomerContactRequest {
	return customer_contact.NewCustomerContactRequest()
}

func (q *QType) CustomerContactsMinimal() *customer_contact.CustomerContactRequest {
	return customer_contact.NewCustomerContactMinimalRequest()
}


func (q *QType) CustomerPlatform(entity *customer.Customer) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}

func (q *QType) CustomerContactCustomer(entity *customer_contact.CustomerContact) (*customer.Customer, bool) {
	value, ok := entity.RelationEntity("customerEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer.Customer)
	return typed, ok
}
package lib

import (
	"merchant-order-privacy-service-core-workspace/lib/merchant"
	"merchant-order-privacy-service-core-workspace/lib/customer_profile"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Merchants() *merchant.MerchantRequest {
	return merchant.NewMerchantRequest()
}

func (q *QType) MerchantsMinimal() *merchant.MerchantRequest {
	return merchant.NewMerchantMinimalRequest()
}

func (q *QType) CustomerProfiles() *customer_profile.CustomerProfileRequest {
	return customer_profile.NewCustomerProfileRequest()
}

func (q *QType) CustomerProfilesMinimal() *customer_profile.CustomerProfileRequest {
	return customer_profile.NewCustomerProfileMinimalRequest()
}

func (q *QType) CustomerOrders() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderRequest()
}

func (q *QType) CustomerOrdersMinimal() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderMinimalRequest()
}


func (q *QType) CustomerProfileMerchant(entity *customer_profile.CustomerProfile) (*merchant.Merchant, bool) {
	value, ok := entity.RelationEntity("merchantEntity")
	if !ok { return nil, false }
	typed, ok := value.(*merchant.Merchant)
	return typed, ok
}

func (q *QType) CustomerOrderMerchant(entity *customer_order.CustomerOrder) (*merchant.Merchant, bool) {
	value, ok := entity.RelationEntity("merchantEntity")
	if !ok { return nil, false }
	typed, ok := value.(*merchant.Merchant)
	return typed, ok
}

func (q *QType) CustomerOrderCustomerProfile(entity *customer_order.CustomerOrder) (*customer_profile.CustomerProfile, bool) {
	value, ok := entity.RelationEntity("customerProfileEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_profile.CustomerProfile)
	return typed, ok
}
package lib

import (
	"merchant-order-privacy-service-core-workspace/lib/merchant"
	"merchant-order-privacy-service-core-workspace/lib/customer_profile"
	"merchant-order-privacy-service-core-workspace/lib/customer_order"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Merchant(value *merchant.Merchant) *merchant.MerchantExpression {
	return merchant.NewMerchantExpression(value)
}

func (expressionFacade) CustomerProfile(value *customer_profile.CustomerProfile) *customer_profile.CustomerProfileExpression {
	return customer_profile.NewCustomerProfileExpression(value)
}

func (expressionFacade) CustomerOrder(value *customer_order.CustomerOrder) *customer_order.CustomerOrderExpression {
	return customer_order.NewCustomerOrderExpression(value)
}

package lib

import (
	"order-management-service-core-workspace/lib/commerce_platform"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) CommercePlatform(value *commerce_platform.CommercePlatform) *commerce_platform.CommercePlatformExpression {
	return commerce_platform.NewCommercePlatformExpression(value)
}

func (expressionFacade) Customer(value *customer.Customer) *customer.CustomerExpression {
	return customer.NewCustomerExpression(value)
}

func (expressionFacade) OrderStatus(value *order_status.OrderStatus) *order_status.OrderStatusExpression {
	return order_status.NewOrderStatusExpression(value)
}

func (expressionFacade) CustomerOrder(value *customer_order.CustomerOrder) *customer_order.CustomerOrderExpression {
	return customer_order.NewCustomerOrderExpression(value)
}

func (expressionFacade) Product(value *product.Product) *product.ProductExpression {
	return product.NewProductExpression(value)
}

func (expressionFacade) OrderLine(value *order_line.OrderLine) *order_line.OrderLineExpression {
	return order_line.NewOrderLineExpression(value)
}

func (expressionFacade) OrderSearchPreset(value *order_search_preset.OrderSearchPreset) *order_search_preset.OrderSearchPresetExpression {
	return order_search_preset.NewOrderSearchPresetExpression(value)
}
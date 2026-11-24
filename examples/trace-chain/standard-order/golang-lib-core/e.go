package lib

import (
	"request-tree-order-service-core-workspace/lib/platform"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
	"request-tree-order-service-core-workspace/lib/customer_order"
	"request-tree-order-service-core-workspace/lib/order_item"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) OrderStatus(value *order_status.OrderStatus) *order_status.OrderStatusExpression {
	return order_status.NewOrderStatusExpression(value)
}

func (expressionFacade) Customer(value *customer.Customer) *customer.CustomerExpression {
	return customer.NewCustomerExpression(value)
}

func (expressionFacade) CustomerOrder(value *customer_order.CustomerOrder) *customer_order.CustomerOrderExpression {
	return customer_order.NewCustomerOrderExpression(value)
}

func (expressionFacade) OrderItem(value *order_item.OrderItem) *order_item.OrderItemExpression {
	return order_item.NewOrderItemExpression(value)
}

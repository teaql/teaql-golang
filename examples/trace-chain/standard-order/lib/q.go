package lib

import (
	"request-tree-order-service-core-workspace/lib/platform"
	"request-tree-order-service-core-workspace/lib/order_status"
	"request-tree-order-service-core-workspace/lib/customer"
	"request-tree-order-service-core-workspace/lib/customer_order"
	"request-tree-order-service-core-workspace/lib/order_item"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) OrderStatuses() *order_status.OrderStatusRequest {
	return order_status.NewOrderStatusRequest()
}

func (q *QType) OrderStatusesMinimal() *order_status.OrderStatusRequest {
	return order_status.NewOrderStatusMinimalRequest()
}

func (q *QType) Customers() *customer.CustomerRequest {
	return customer.NewCustomerRequest()
}

func (q *QType) CustomersMinimal() *customer.CustomerRequest {
	return customer.NewCustomerMinimalRequest()
}

func (q *QType) CustomerOrders() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderRequest()
}

func (q *QType) CustomerOrdersMinimal() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderMinimalRequest()
}

func (q *QType) OrderItems() *order_item.OrderItemRequest {
	return order_item.NewOrderItemRequest()
}

func (q *QType) OrderItemsMinimal() *order_item.OrderItemRequest {
	return order_item.NewOrderItemMinimalRequest()
}


func (q *QType) OrderStatusPlatform(entity *order_status.OrderStatus) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}

func (q *QType) CustomerPlatform(entity *customer.Customer) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}

func (q *QType) CustomerOrderCustomer(entity *customer_order.CustomerOrder) (*customer.Customer, bool) {
	value, ok := entity.RelationEntity("customerEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer.Customer)
	return typed, ok
}

func (q *QType) CustomerOrderStatus(entity *customer_order.CustomerOrder) (*order_status.OrderStatus, bool) {
	value, ok := entity.RelationEntity("statusEntity")
	if !ok { return nil, false }
	typed, ok := value.(*order_status.OrderStatus)
	return typed, ok
}

func (q *QType) OrderItemCustomerOrder(entity *order_item.OrderItem) (*customer_order.CustomerOrder, bool) {
	value, ok := entity.RelationEntity("customerOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_order.CustomerOrder)
	return typed, ok
}
package lib

import (
	"trace-chain-service-core-workspace/lib/platform"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/payment_attempt"
	"trace-chain-service-core-workspace/lib/shipment"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
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

func (q *QType) Payments() *payment.PaymentRequest {
	return payment.NewPaymentRequest()
}

func (q *QType) PaymentsMinimal() *payment.PaymentRequest {
	return payment.NewPaymentMinimalRequest()
}

func (q *QType) PaymentAttempts() *payment_attempt.PaymentAttemptRequest {
	return payment_attempt.NewPaymentAttemptRequest()
}

func (q *QType) PaymentAttemptsMinimal() *payment_attempt.PaymentAttemptRequest {
	return payment_attempt.NewPaymentAttemptMinimalRequest()
}

func (q *QType) Shipments() *shipment.ShipmentRequest {
	return shipment.NewShipmentRequest()
}

func (q *QType) ShipmentsMinimal() *shipment.ShipmentRequest {
	return shipment.NewShipmentMinimalRequest()
}


func (q *QType) CustomerOrderPlatform(entity *customer_order.CustomerOrder) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}

func (q *QType) OrderItemCustomerOrder(entity *order_item.OrderItem) (*customer_order.CustomerOrder, bool) {
	value, ok := entity.RelationEntity("customerOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_order.CustomerOrder)
	return typed, ok
}

func (q *QType) PaymentCustomerOrder(entity *payment.Payment) (*customer_order.CustomerOrder, bool) {
	value, ok := entity.RelationEntity("customerOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_order.CustomerOrder)
	return typed, ok
}

func (q *QType) PaymentAttemptPayment(entity *payment_attempt.PaymentAttempt) (*payment.Payment, bool) {
	value, ok := entity.RelationEntity("paymentEntity")
	if !ok { return nil, false }
	typed, ok := value.(*payment.Payment)
	return typed, ok
}

func (q *QType) ShipmentCustomerOrder(entity *shipment.Shipment) (*customer_order.CustomerOrder, bool) {
	value, ok := entity.RelationEntity("customerOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_order.CustomerOrder)
	return typed, ok
}
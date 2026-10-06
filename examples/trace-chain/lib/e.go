package lib

import (
	"trace-chain-service-core-workspace/lib/platform"
	"trace-chain-service-core-workspace/lib/customer_order"
	"trace-chain-service-core-workspace/lib/order_item"
	"trace-chain-service-core-workspace/lib/payment"
	"trace-chain-service-core-workspace/lib/payment_attempt"
	"trace-chain-service-core-workspace/lib/shipment"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) CustomerOrder(value *customer_order.CustomerOrder) *customer_order.CustomerOrderExpression {
	return customer_order.NewCustomerOrderExpression(value)
}

func (expressionFacade) OrderItem(value *order_item.OrderItem) *order_item.OrderItemExpression {
	return order_item.NewOrderItemExpression(value)
}

func (expressionFacade) Payment(value *payment.Payment) *payment.PaymentExpression {
	return payment.NewPaymentExpression(value)
}

func (expressionFacade) PaymentAttempt(value *payment_attempt.PaymentAttempt) *payment_attempt.PaymentAttemptExpression {
	return payment_attempt.NewPaymentAttemptExpression(value)
}

func (expressionFacade) Shipment(value *shipment.Shipment) *shipment.ShipmentExpression {
	return shipment.NewShipmentExpression(value)
}

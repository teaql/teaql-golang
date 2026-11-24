package lib

import (
	"order-attachment-service-core-workspace/lib/commerce_platform"
	"order-attachment-service-core-workspace/lib/customer_order"
	"order-attachment-service-core-workspace/lib/order_attachment"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) CommercePlatform(value *commerce_platform.CommercePlatform) *commerce_platform.CommercePlatformExpression {
	return commerce_platform.NewCommercePlatformExpression(value)
}

func (expressionFacade) CustomerOrder(value *customer_order.CustomerOrder) *customer_order.CustomerOrderExpression {
	return customer_order.NewCustomerOrderExpression(value)
}

func (expressionFacade) OrderAttachment(value *order_attachment.OrderAttachment) *order_attachment.OrderAttachmentExpression {
	return order_attachment.NewOrderAttachmentExpression(value)
}

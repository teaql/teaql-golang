package lib

import (
	"order-attachment-service-core-workspace/lib/commerce_platform"
	"order-attachment-service-core-workspace/lib/customer_order"
	"order-attachment-service-core-workspace/lib/order_attachment"
)

type QType struct {}
var Q = &QType{}

func (q *QType) CommercePlatforms() *commerce_platform.CommercePlatformRequest {
	return commerce_platform.NewCommercePlatformRequest()
}

func (q *QType) CommercePlatformsMinimal() *commerce_platform.CommercePlatformRequest {
	return commerce_platform.NewCommercePlatformMinimalRequest()
}

func (q *QType) CustomerOrders() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderRequest()
}

func (q *QType) CustomerOrdersMinimal() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderMinimalRequest()
}

func (q *QType) OrderAttachments() *order_attachment.OrderAttachmentRequest {
	return order_attachment.NewOrderAttachmentRequest()
}

func (q *QType) OrderAttachmentsMinimal() *order_attachment.OrderAttachmentRequest {
	return order_attachment.NewOrderAttachmentMinimalRequest()
}


func (q *QType) CustomerOrderCommercePlatform(entity *customer_order.CustomerOrder) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}

func (q *QType) OrderAttachmentCustomerOrder(entity *order_attachment.OrderAttachment) (*customer_order.CustomerOrder, bool) {
	value, ok := entity.RelationEntity("customerOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_order.CustomerOrder)
	return typed, ok
}
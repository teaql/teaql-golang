package lib

import (
	"purchase-order-numbering-core-workspace/lib/platform"
	"purchase-order-numbering-core-workspace/lib/purchase_order"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) PurchaseOrders() *purchase_order.PurchaseOrderRequest {
	return purchase_order.NewPurchaseOrderRequest()
}

func (q *QType) PurchaseOrdersMinimal() *purchase_order.PurchaseOrderRequest {
	return purchase_order.NewPurchaseOrderMinimalRequest()
}

func (q *QType) PurchaseOrderByOrderNumber(value purchase_order.OrderNumber) *purchase_order.PurchaseOrderRequest {
	return q.PurchaseOrders().WithOrderNumberIs(value).Limit(2)
}

func (q *QType) PurchaseOrderPlatform(entity *purchase_order.PurchaseOrder) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}
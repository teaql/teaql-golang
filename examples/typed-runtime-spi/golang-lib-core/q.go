package lib

import (
	"order-service-routing-core-workspace/lib/platform"
	"order-service-routing-core-workspace/lib/purchase_order"
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


func (q *QType) PurchaseOrderPlatform(entity *purchase_order.PurchaseOrder) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}
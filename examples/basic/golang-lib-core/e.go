package lib

import (
	"first-order-service-core-workspace/lib/platform"
	"first-order-service-core-workspace/lib/purchase_order"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) PurchaseOrder(value *purchase_order.PurchaseOrder) *purchase_order.PurchaseOrderExpression {
	return purchase_order.NewPurchaseOrderExpression(value)
}

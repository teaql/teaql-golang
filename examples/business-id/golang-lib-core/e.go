package lib

import (
	"purchase-order-numbering-core-workspace/lib/platform"
	"purchase-order-numbering-core-workspace/lib/purchase_order"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) PurchaseOrder(value *purchase_order.PurchaseOrder) *purchase_order.PurchaseOrderExpression {
	return purchase_order.NewPurchaseOrderExpression(value)
}

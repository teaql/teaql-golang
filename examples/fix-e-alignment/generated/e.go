package lib

import (
	"checker-fix-advanced-core-workspace/lib/tenant"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"checker-fix-advanced-core-workspace/lib/order_item"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Tenant(value *tenant.Tenant) *tenant.TenantExpression {
	return tenant.NewTenantExpression(value)
}

func (expressionFacade) PurchaseOrder(value *purchase_order.PurchaseOrder) *purchase_order.PurchaseOrderExpression {
	return purchase_order.NewPurchaseOrderExpression(value)
}

func (expressionFacade) OrderItem(value *order_item.OrderItem) *order_item.OrderItemExpression {
	return order_item.NewOrderItemExpression(value)
}

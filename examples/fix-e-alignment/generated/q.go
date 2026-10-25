package lib

import (
	"checker-fix-advanced-core-workspace/lib/tenant"
	"checker-fix-advanced-core-workspace/lib/purchase_order"
	"checker-fix-advanced-core-workspace/lib/order_item"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Tenants() *tenant.TenantRequest {
	return tenant.NewTenantRequest()
}

func (q *QType) TenantsMinimal() *tenant.TenantRequest {
	return tenant.NewTenantMinimalRequest()
}

func (q *QType) PurchaseOrders() *purchase_order.PurchaseOrderRequest {
	return purchase_order.NewPurchaseOrderRequest()
}

func (q *QType) PurchaseOrdersMinimal() *purchase_order.PurchaseOrderRequest {
	return purchase_order.NewPurchaseOrderMinimalRequest()
}

func (q *QType) OrderItems() *order_item.OrderItemRequest {
	return order_item.NewOrderItemRequest()
}

func (q *QType) OrderItemsMinimal() *order_item.OrderItemRequest {
	return order_item.NewOrderItemMinimalRequest()
}


func (q *QType) PurchaseOrderTenant(entity *purchase_order.PurchaseOrder) (*tenant.Tenant, bool) {
	value, ok := entity.RelationEntity("tenantEntity")
	if !ok { return nil, false }
	typed, ok := value.(*tenant.Tenant)
	return typed, ok
}

func (q *QType) OrderItemPurchaseOrder(entity *order_item.OrderItem) (*purchase_order.PurchaseOrder, bool) {
	value, ok := entity.RelationEntity("purchaseOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*purchase_order.PurchaseOrder)
	return typed, ok
}

func (q *QType) OrderItemTenant(entity *order_item.OrderItem) (*tenant.Tenant, bool) {
	value, ok := entity.RelationEntity("tenantEntity")
	if !ok { return nil, false }
	typed, ok := value.(*tenant.Tenant)
	return typed, ok
}
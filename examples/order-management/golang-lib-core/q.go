package lib

import (
	"order-management-service-core-workspace/lib/commerce_platform"
	"order-management-service-core-workspace/lib/customer"
	"order-management-service-core-workspace/lib/order_status"
	"order-management-service-core-workspace/lib/customer_order"
	"order-management-service-core-workspace/lib/product"
	"order-management-service-core-workspace/lib/order_line"
	"order-management-service-core-workspace/lib/order_search_preset"
)

type QType struct {}
var Q = &QType{}

func (q *QType) CommercePlatforms() *commerce_platform.CommercePlatformRequest {
	return commerce_platform.NewCommercePlatformRequest()
}

func (q *QType) CommercePlatformsMinimal() *commerce_platform.CommercePlatformRequest {
	return commerce_platform.NewCommercePlatformMinimalRequest()
}

func (q *QType) Customers() *customer.CustomerRequest {
	return customer.NewCustomerRequest()
}

func (q *QType) CustomersMinimal() *customer.CustomerRequest {
	return customer.NewCustomerMinimalRequest()
}

func (q *QType) OrderStatuses() *order_status.OrderStatusRequest {
	return order_status.NewOrderStatusRequest()
}

func (q *QType) OrderStatusesMinimal() *order_status.OrderStatusRequest {
	return order_status.NewOrderStatusMinimalRequest()
}

func (q *QType) CustomerOrders() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderRequest()
}

func (q *QType) CustomerOrdersMinimal() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderMinimalRequest()
}

func (q *QType) Products() *product.ProductRequest {
	return product.NewProductRequest()
}

func (q *QType) ProductsMinimal() *product.ProductRequest {
	return product.NewProductMinimalRequest()
}

func (q *QType) OrderLines() *order_line.OrderLineRequest {
	return order_line.NewOrderLineRequest()
}

func (q *QType) OrderLinesMinimal() *order_line.OrderLineRequest {
	return order_line.NewOrderLineMinimalRequest()
}

func (q *QType) OrderSearchPresets() *order_search_preset.OrderSearchPresetRequest {
	return order_search_preset.NewOrderSearchPresetRequest()
}

func (q *QType) OrderSearchPresetsMinimal() *order_search_preset.OrderSearchPresetRequest {
	return order_search_preset.NewOrderSearchPresetMinimalRequest()
}

func (q *QType) CustomerCommercePlatform(entity *customer.Customer) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}

func (q *QType) OrderStatusCommercePlatform(entity *order_status.OrderStatus) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}

func (q *QType) CustomerOrderStatus(entity *customer_order.CustomerOrder) (*order_status.OrderStatus, bool) {
	value, ok := entity.RelationEntity("statusEntity")
	if !ok { return nil, false }
	typed, ok := value.(*order_status.OrderStatus)
	return typed, ok
}

func (q *QType) CustomerOrderCustomer(entity *customer_order.CustomerOrder) (*customer.Customer, bool) {
	value, ok := entity.RelationEntity("customerEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer.Customer)
	return typed, ok
}

func (q *QType) CustomerOrderCommercePlatform(entity *customer_order.CustomerOrder) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}

func (q *QType) ProductCommercePlatform(entity *product.Product) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}

func (q *QType) OrderLineCustomerOrder(entity *order_line.OrderLine) (*customer_order.CustomerOrder, bool) {
	value, ok := entity.RelationEntity("customerOrderEntity")
	if !ok { return nil, false }
	typed, ok := value.(*customer_order.CustomerOrder)
	return typed, ok
}

func (q *QType) OrderLineProduct(entity *order_line.OrderLine) (*product.Product, bool) {
	value, ok := entity.RelationEntity("productEntity")
	if !ok { return nil, false }
	typed, ok := value.(*product.Product)
	return typed, ok
}

func (q *QType) OrderLineCommercePlatform(entity *order_line.OrderLine) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}

func (q *QType) OrderSearchPresetCommercePlatform(entity *order_search_preset.OrderSearchPreset) (*commerce_platform.CommercePlatform, bool) {
	value, ok := entity.RelationEntity("commercePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*commerce_platform.CommercePlatform)
	return typed, ok
}
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
	return commerce_platform.NewCommercePlatformRequest()
}

func (q *QType) Customers() *customer.CustomerRequest {
	return customer.NewCustomerRequest()
}

func (q *QType) CustomersMinimal() *customer.CustomerRequest {
	return customer.NewCustomerRequest()
}

func (q *QType) OrderStatuses() *order_status.OrderStatusRequest {
	return order_status.NewOrderStatusRequest()
}

func (q *QType) OrderStatusesMinimal() *order_status.OrderStatusRequest {
	return order_status.NewOrderStatusRequest()
}

func (q *QType) CustomerOrders() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderRequest()
}

func (q *QType) CustomerOrdersMinimal() *customer_order.CustomerOrderRequest {
	return customer_order.NewCustomerOrderRequest()
}

func (q *QType) Products() *product.ProductRequest {
	return product.NewProductRequest()
}

func (q *QType) ProductsMinimal() *product.ProductRequest {
	return product.NewProductRequest()
}

func (q *QType) OrderLines() *order_line.OrderLineRequest {
	return order_line.NewOrderLineRequest()
}

func (q *QType) OrderLinesMinimal() *order_line.OrderLineRequest {
	return order_line.NewOrderLineRequest()
}

func (q *QType) OrderSearchPresets() *order_search_preset.OrderSearchPresetRequest {
	return order_search_preset.NewOrderSearchPresetRequest()
}

func (q *QType) OrderSearchPresetsMinimal() *order_search_preset.OrderSearchPresetRequest {
	return order_search_preset.NewOrderSearchPresetRequest()
}
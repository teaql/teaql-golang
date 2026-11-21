package lib

import (
	"supplier-onboarding-service-core-workspace/lib/warehouse"
	"supplier-onboarding-service-core-workspace/lib/supplier"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Warehouses() *warehouse.WarehouseRequest {
	return warehouse.NewWarehouseRequest()
}

func (q *QType) WarehousesMinimal() *warehouse.WarehouseRequest {
	return warehouse.NewWarehouseMinimalRequest()
}

func (q *QType) Suppliers() *supplier.SupplierRequest {
	return supplier.NewSupplierRequest()
}

func (q *QType) SuppliersMinimal() *supplier.SupplierRequest {
	return supplier.NewSupplierMinimalRequest()
}


func (q *QType) SupplierWarehouse(entity *supplier.Supplier) (*warehouse.Warehouse, bool) {
	value, ok := entity.RelationEntity("warehouseEntity")
	if !ok { return nil, false }
	typed, ok := value.(*warehouse.Warehouse)
	return typed, ok
}
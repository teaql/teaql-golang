package lib

import (
	"supplier-onboarding-service-core-workspace/lib/warehouse"
	"supplier-onboarding-service-core-workspace/lib/supplier"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Warehouse(value *warehouse.Warehouse) *warehouse.WarehouseExpression {
	return warehouse.NewWarehouseExpression(value)
}

func (expressionFacade) Supplier(value *supplier.Supplier) *supplier.SupplierExpression {
	return supplier.NewSupplierExpression(value)
}

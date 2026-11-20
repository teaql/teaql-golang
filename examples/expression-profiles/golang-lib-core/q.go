package lib

import (
	"dispatch-document-core-workspace/lib/distribution_center"
	"dispatch-document-core-workspace/lib/dispatch_order"
)

type QType struct {}
var Q = &QType{}

func (q *QType) DistributionCenters() *distribution_center.DistributionCenterRequest {
	return distribution_center.NewDistributionCenterRequest()
}

func (q *QType) DistributionCentersMinimal() *distribution_center.DistributionCenterRequest {
	return distribution_center.NewDistributionCenterMinimalRequest()
}

func (q *QType) DispatchOrders() *dispatch_order.DispatchOrderRequest {
	return dispatch_order.NewDispatchOrderRequest()
}

func (q *QType) DispatchOrdersMinimal() *dispatch_order.DispatchOrderRequest {
	return dispatch_order.NewDispatchOrderMinimalRequest()
}


func (q *QType) DispatchOrderDistributionCenter(entity *dispatch_order.DispatchOrder) (*distribution_center.DistributionCenter, bool) {
	value, ok := entity.RelationEntity("distributionCenterEntity")
	if !ok { return nil, false }
	typed, ok := value.(*distribution_center.DistributionCenter)
	return typed, ok
}
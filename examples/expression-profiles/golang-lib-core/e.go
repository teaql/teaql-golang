package lib

import (
	"dispatch-document-core-workspace/lib/distribution_center"
	"dispatch-document-core-workspace/lib/dispatch_order"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) DistributionCenter(value *distribution_center.DistributionCenter) *distribution_center.DistributionCenterExpression {
	return distribution_center.NewDistributionCenterExpression(value)
}

func (expressionFacade) DispatchOrder(value *dispatch_order.DispatchOrder) *dispatch_order.DispatchOrderExpression {
	return dispatch_order.NewDispatchOrderExpression(value)
}

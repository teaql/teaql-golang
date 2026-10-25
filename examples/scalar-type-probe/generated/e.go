package lib

import (
	"scalar-type-conformance-service-core-workspace/lib/type_platform"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) TypePlatform(value *type_platform.TypePlatform) *type_platform.TypePlatformExpression {
	return type_platform.NewTypePlatformExpression(value)
}

func (expressionFacade) ScalarTypeProbe(value *scalar_type_probe.ScalarTypeProbe) *scalar_type_probe.ScalarTypeProbeExpression {
	return scalar_type_probe.NewScalarTypeProbeExpression(value)
}

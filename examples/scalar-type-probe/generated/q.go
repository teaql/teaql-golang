package lib

import (
	"scalar-type-conformance-service-core-workspace/lib/type_platform"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
)

type QType struct {}
var Q = &QType{}

func (q *QType) TypePlatforms() *type_platform.TypePlatformRequest {
	return type_platform.NewTypePlatformRequest()
}

func (q *QType) TypePlatformsMinimal() *type_platform.TypePlatformRequest {
	return type_platform.NewTypePlatformMinimalRequest()
}

func (q *QType) ScalarTypeProbes() *scalar_type_probe.ScalarTypeProbeRequest {
	return scalar_type_probe.NewScalarTypeProbeRequest()
}

func (q *QType) ScalarTypeProbesMinimal() *scalar_type_probe.ScalarTypeProbeRequest {
	return scalar_type_probe.NewScalarTypeProbeMinimalRequest()
}


func (q *QType) ScalarTypeProbeTypePlatform(entity *scalar_type_probe.ScalarTypeProbe) (*type_platform.TypePlatform, bool) {
	value, ok := entity.RelationEntity("typePlatformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*type_platform.TypePlatform)
	return typed, ok
}
package lib

import (
	"school-management-service-core-workspace/lib/platform"
	"school-management-service-core-workspace/lib/school"
	"school-management-service-core-workspace/lib/school_type"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) SchoolType(value *school_type.SchoolType) *school_type.SchoolTypeExpression {
	return school_type.NewSchoolTypeExpression(value)
}

func (expressionFacade) School(value *school.School) *school.SchoolExpression {
	return school.NewSchoolExpression(value)
}

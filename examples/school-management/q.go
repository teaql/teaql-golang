package lib

import (
	"school-management-service-core-workspace/lib/platform"
	"school-management-service-core-workspace/lib/school_type"
	"school-management-service-core-workspace/lib/school"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) SchoolTypes() *school_type.SchoolTypeRequest {
	return school_type.NewSchoolTypeRequest()
}

func (q *QType) SchoolTypesMinimal() *school_type.SchoolTypeRequest {
	return school_type.NewSchoolTypeMinimalRequest()
}

func (q *QType) Schools() *school.SchoolRequest {
	return school.NewSchoolRequest()
}

func (q *QType) SchoolsMinimal() *school.SchoolRequest {
	return school.NewSchoolMinimalRequest()
}
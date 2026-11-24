package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"school-management-service-core-workspace/lib/internal/generatedjson"
	"school-management-service-core-workspace/lib/platform"
	"school-management-service-core-workspace/lib/school_type"
	"school-management-service-core-workspace/lib/school"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "School Type": return school_type.TeaqlSchoolTypeFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "School": return school.TeaqlSchoolFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// PlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*platform.Platform, error) {
	record, err := generatedjson.Model.Parse("Platform", payload, profile...)
	if err != nil { return nil, err }
	return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// SchoolTypeFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func SchoolTypeFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*school_type.SchoolType, error) {
	record, err := generatedjson.Model.Parse("School Type", payload, profile...)
	if err != nil { return nil, err }
	return school_type.TeaqlSchoolTypeFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// SchoolFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func SchoolFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*school.School, error) {
	record, err := generatedjson.Model.Parse("School", payload, profile...)
	if err != nil { return nil, err }
	return school.TeaqlSchoolFromJSONRecord(record, teaqlEntityFromJSONRecord)
}
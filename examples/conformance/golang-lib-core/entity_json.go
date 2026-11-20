package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"runtime-example-conformance-service-core-workspace/lib/internal/generatedjson"
	"runtime-example-conformance-service-core-workspace/lib/platform"
	"runtime-example-conformance-service-core-workspace/lib/work_item"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Work Item": return work_item.TeaqlWorkItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
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

// WorkItemFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func WorkItemFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*work_item.WorkItem, error) {
	record, err := generatedjson.Model.Parse("Work Item", payload, profile...)
	if err != nil { return nil, err }
	return work_item.TeaqlWorkItemFromJSONRecord(record, teaqlEntityFromJSONRecord)
}
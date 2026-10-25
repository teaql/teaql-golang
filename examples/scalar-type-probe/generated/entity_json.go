package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"scalar-type-conformance-service-core-workspace/lib/internal/generatedjson"
	"scalar-type-conformance-service-core-workspace/lib/type_platform"
	"scalar-type-conformance-service-core-workspace/lib/scalar_type_probe"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Type Platform": return type_platform.TeaqlTypePlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Scalar Type Probe": return scalar_type_probe.TeaqlScalarTypeProbeFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// TypePlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func TypePlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*type_platform.TypePlatform, error) {
	record, err := generatedjson.Model.Parse("Type Platform", payload, profile...)
	if err != nil { return nil, err }
	return type_platform.TeaqlTypePlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// ScalarTypeProbeFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func ScalarTypeProbeFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*scalar_type_probe.ScalarTypeProbe, error) {
	record, err := generatedjson.Model.Parse("Scalar Type Probe", payload, profile...)
	if err != nil { return nil, err }
	return scalar_type_probe.TeaqlScalarTypeProbeFromJSONRecord(record, teaqlEntityFromJSONRecord)
}
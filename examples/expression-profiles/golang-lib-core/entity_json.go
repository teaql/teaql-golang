package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"dispatch-document-core-workspace/lib/internal/generatedjson"
	"dispatch-document-core-workspace/lib/distribution_center"
	"dispatch-document-core-workspace/lib/dispatch_order"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "distribution_center": return distribution_center.TeaqlDistributionCenterFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "dispatch_order": return dispatch_order.TeaqlDispatchOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// DistributionCenterFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func DistributionCenterFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*distribution_center.DistributionCenter, error) {
	record, err := generatedjson.Model.Parse("distribution_center", payload, profile...)
	if err != nil { return nil, err }
	return distribution_center.TeaqlDistributionCenterFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// DispatchOrderFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func DispatchOrderFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*dispatch_order.DispatchOrder, error) {
	record, err := generatedjson.Model.Parse("dispatch_order", payload, profile...)
	if err != nil { return nil, err }
	return dispatch_order.TeaqlDispatchOrderFromJSONRecord(record, teaqlEntityFromJSONRecord)
}
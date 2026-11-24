package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"natural-message-service-core-workspace/lib/internal/generatedjson"
	"natural-message-service-core-workspace/lib/catalog"
	"natural-message-service-core-workspace/lib/entry"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "catalog": return catalog.TeaqlCatalogFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "entry": return entry.TeaqlEntryFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// CatalogFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func CatalogFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*catalog.Catalog, error) {
	record, err := generatedjson.Model.Parse("catalog", payload, profile...)
	if err != nil { return nil, err }
	return catalog.TeaqlCatalogFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// EntryFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func EntryFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*entry.Entry, error) {
	record, err := generatedjson.Model.Parse("entry", payload, profile...)
	if err != nil { return nil, err }
	return entry.TeaqlEntryFromJSONRecord(record, teaqlEntityFromJSONRecord)
}
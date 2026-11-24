package lib

import (
	"fmt"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	"freight-quote-settlement-core-workspace/lib/internal/generatedjson"
	"freight-quote-settlement-core-workspace/lib/platform"
	"freight-quote-settlement-core-workspace/lib/freight_quote"
)

// Concrete factories belong to this facade, not the passive metadata package.
// Every parse creates independent clean mutation ownership; no Context or I/O.
func teaqlEntityFromJSONRecord(entityType string, record core.Record) (core.Entity, error) {
	switch entityType {
	case "Dispatch Platform": return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
	case "Freight Quote": return freight_quote.TeaqlFreightQuoteFromJSONRecord(record, teaqlEntityFromJSONRecord)
	default: return nil, fmt.Errorf("unknown generated entity JSON type")
	}
}

// PlatformFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func PlatformFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*platform.Platform, error) {
	record, err := generatedjson.Model.Parse("Dispatch Platform", payload, profile...)
	if err != nil { return nil, err }
	return platform.TeaqlPlatformFromJSONRecord(record, teaqlEntityFromJSONRecord)
}

// FreightQuoteFromJSON decodes a selected, loaded snapshot, not a
// create command or authorization decision. Save still requires audit/checker.
func FreightQuoteFromJSON(payload []byte, profile ...runtime.JsonFieldNamingProfile) (*freight_quote.FreightQuote, error) {
	record, err := generatedjson.Model.Parse("Freight Quote", payload, profile...)
	if err != nil { return nil, err }
	return freight_quote.TeaqlFreightQuoteFromJSONRecord(record, teaqlEntityFromJSONRecord)
}
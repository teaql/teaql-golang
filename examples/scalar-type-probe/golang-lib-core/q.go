package lib

import (
	"freight-quote-settlement-core-workspace/lib/platform"
	"freight-quote-settlement-core-workspace/lib/freight_quote"
)

type QType struct {}
var Q = &QType{}

func (q *QType) Platforms() *platform.PlatformRequest {
	return platform.NewPlatformRequest()
}

func (q *QType) PlatformsMinimal() *platform.PlatformRequest {
	return platform.NewPlatformMinimalRequest()
}

func (q *QType) FreightQuotes() *freight_quote.FreightQuoteRequest {
	return freight_quote.NewFreightQuoteRequest()
}

func (q *QType) FreightQuotesMinimal() *freight_quote.FreightQuoteRequest {
	return freight_quote.NewFreightQuoteMinimalRequest()
}


func (q *QType) FreightQuotePlatform(entity *freight_quote.FreightQuote) (*platform.Platform, bool) {
	value, ok := entity.RelationEntity("platformEntity")
	if !ok { return nil, false }
	typed, ok := value.(*platform.Platform)
	return typed, ok
}
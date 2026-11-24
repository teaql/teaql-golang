package lib

import (
	"freight-quote-settlement-core-workspace/lib/platform"
	"freight-quote-settlement-core-workspace/lib/freight_quote"
)

type expressionFacade struct{}

var E expressionFacade

func (expressionFacade) Platform(value *platform.Platform) *platform.PlatformExpression {
	return platform.NewPlatformExpression(value)
}

func (expressionFacade) FreightQuote(value *freight_quote.FreightQuote) *freight_quote.FreightQuoteExpression {
	return freight_quote.NewFreightQuoteExpression(value)
}

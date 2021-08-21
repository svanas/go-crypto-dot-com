package crypto

type Symbol struct {
	Symbol           string `json:"instrument_name"`
	QuoteCurrency    string `json:"quote_currency"`
	BaseCurrency     string `json:"base_currency"`
	PriceDecimals    int    `json:"price_decimals"`
	QuantityDecimals int    `json:"quantity_decimals"`
}

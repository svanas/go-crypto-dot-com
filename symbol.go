package crypto

type Symbol struct {
	Symbol          string `json:"symbol"`
	CountCoin       string `json:"count_coin"`
	AmountPrecision int    `json:"amount_precision"`
	BaseCoin        string `json:"base_coin"`
	PricePrecision  int    `json:"price_precision"`
}

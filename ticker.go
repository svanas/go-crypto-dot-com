package crypto

type Ticker struct {
	Symbol string  `json:"i,omitempty"` // instrument name, e.g. BTC_USDT, ETH_CRO, etc.
	Last   float64 `json:"a,string"`    // the price of the latest trade, null if there weren't any trades
	Volume float64 `json:"v,string"`    // the total 24h traded volume
	High   float64 `json:"h,string"`    // price of the 24h highest trade
	Low    float64 `json:"l,string"`    // price of the 24h lowest trade, null if there weren't any trades
}

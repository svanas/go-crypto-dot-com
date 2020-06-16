package crypto

type Ticker struct {
	Symbol string  `json:"symbol,omitempty"`
	High   float64 `json:"high,string"`
	Vol    float64 `json:"vol,string"`
	Last   float64 `json:"last,string"`
	Low    float64 `json:"low,string"`
	Buy    float64 `json:"buy,string"`
	Sell   float64 `json:"sell,string"`
}

type Tickers struct {
	Date   int      `json:"date"`
	Ticker []Ticker `json:"ticker"`
}

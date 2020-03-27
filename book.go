package crypto

type (
	BookEntry []float64
)

func (be *BookEntry) Price() float64 {
	return (*be)[0]
}

func (be *BookEntry) Size() float64 {
	return (*be)[1]
}

type OrderBook struct {
	Tick struct {
		Asks []BookEntry `json:"asks"`
		Bids []BookEntry `json:"bids"`
	} `json:"tick"`
}

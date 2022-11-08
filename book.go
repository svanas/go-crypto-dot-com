package crypto

import "strconv"

type (
	BookEntry []string
)

func (be *BookEntry) Price() float64 {
	out, _ := strconv.ParseFloat((*be)[0], 64)
	return out
}

func (be *BookEntry) Size() float64 {
	out, _ := strconv.ParseFloat((*be)[1], 64)
	return out
}

type OrderBook struct {
	Bids []BookEntry `json:"bids"`
	Asks []BookEntry `json:"asks"`
}

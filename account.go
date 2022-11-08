package crypto

type Account struct {
	Balance   float64 `json:"balance"`   // total balance
	Available float64 `json:"available"` // available balance (e.g. not in orders, or locked, etc.)
	Order     float64 `json:"order"`     // balance locked in orders
	Stake     float64 `json:"stake"`     // balance locked for staking (typically only used for CRO)
	Currency  string  `json:"currency"`  // e.g. CRO
}

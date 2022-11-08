package crypto

import "time"

type Trade struct {
	Side        OrderSide `json:"side"`            // BUY or SELL
	Symbol      string    `json:"instrument_name"` // e.g. ETH_CRO, BTC_USDT
	Fee         float64   `json:"fee"`             // trade fee
	TradeId     string    `json:"trade_id"`        // trade ID
	CreatedAt   int64     `json:"create_time"`     // trade creation time
	Price       float64   `json:"traded_price"`    // executed trade price
	Quantity    float64   `json:"traded_quantity"` // executed trade quantity
	FeeCurrency string    `json:"fee_currency"`    // currency used for the fees (e.g. CRO)
	OrderId     string    `json:"order_id"`
}

func (trade *Trade) GetCreatedAt() time.Time {
	if trade.CreatedAt > 0 {
		return time.Unix(trade.CreatedAt/1000, 0)
	}
	return time.Time{}
}

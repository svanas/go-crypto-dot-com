package crypto

import "time"

type Trade struct {
	Id        int64   `json:"id,string"`
	Volume    float64 `json:"volume,string"`
	Side      string  `json:"side"`
	FeeCoin   string  `json:"feeCoin"`
	Price     float64 `json:"price,string"`
	Fee       float64 `json:"fee,string"`
	CTime     int64   `json:"ctime"`
	DealPrice float64 `json:"deal_price,string"`
	Type      string  `json:"type"`
	Symbol    string  `json:"symbol"`
}

func (trade *Trade) GetSide() OrderSide {
	for os := range OrderSideString {
		if os.String() == trade.Side {
			return os
		}
	}
	return ORDER_SIDE_UNKNOWN
}

func (trade *Trade) GetCreatedAt() time.Time {
	if trade.CTime > 0 {
		return time.Unix(trade.CTime/1000, 0)
	}
	return time.Time{}
}

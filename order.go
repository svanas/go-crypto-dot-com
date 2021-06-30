package crypto

import (
	"strings"
	"time"
)

type OrderSide int

const (
	ORDER_SIDE_UNKNOWN OrderSide = iota
	BUY
	SELL
)

var OrderSideString = map[OrderSide]string{
	ORDER_SIDE_UNKNOWN: "",
	BUY:                "BUY",
	SELL:               "SELL",
}

func (os *OrderSide) String() string {
	return OrderSideString[*os]
}

type OrderType int

const (
	ORDER_TYPE_UNKNOWN OrderType = iota
	LIMIT
	MARKET
)

var OrderTypeString = map[OrderType]string{
	ORDER_TYPE_UNKNOWN: "",
	LIMIT:              "1",
	MARKET:             "2",
}

func (ot *OrderType) String() string {
	return OrderTypeString[*ot]
}

const (
	ORDER_STATUS_INIT           = 0 // initial order
	ORDER_STATUS_NEW            = 1 // new order, unfinished business enters the market
	ORDER_STATUS_FILLED         = 2 // full deal
	ORDER_STATUS_PART_FILLED    = 3 // partial transaction
	ORDER_STATUS_CANCELED       = 4 // order cancelled
	ORDER_STATUS_PENDING_CANCEL = 5 // order will be cancelled
	ORDER_STATUS_EXPIRED        = 6 // abnormal order
)

type Order struct {
	Id           int64       `json:"id,string"`
	Side         string      `json:"side"`
	TotalPrice   float64     `json:"total_price,string,omitempty"`
	Fee          float64     `json:"fee,string,omitempty"`
	CreatedAt    int64       `json:"created_at,omitempty"`
	UpdatedAt    int64       `json:"updated_at,omitempty"`
	DealPrice    float64     `json:"deal_price,string,omitempty"`
	AvgPrice     float64     `json:"avg_price,string,omitempty"`
	CountCoin    string      `json:"countCoin,omitempty"`
	Source       int         `json:"source,omitempty"`
	Type         interface{} `json:"type,omitempty"`
	SideMsg      string      `json:"side_msg,omitempty"`
	Volume       float64     `json:"volume,string,omitempty"`
	Price        float64     `json:"price,string,omitempty"`
	StatusMsg    string      `json:"status_msg,omitempty"`
	DealVolume   float64     `json:"deal_volume,string,omitempty"`
	FeeCoin      string      `json:"fee_coin,omitempty"`
	RemainVolume float64     `json:"remain_volume,string,omitempty"`
	BaseCoin     string      `json:"baseCoin,omitempty"`
	Status       int         `json:"status,omitempty"`
}

func (order *Order) GetSide() OrderSide {
	for os := range OrderSideString {
		if os.String() == order.Side {
			return os
		}
	}
	return ORDER_SIDE_UNKNOWN
}

func (order *Order) GetType() OrderType {
	switch order.Type {
	case 1, "1":
		return LIMIT
	case 2, "2":
		return MARKET
	default:
		return ORDER_TYPE_UNKNOWN
	}
}

func (order *Order) GetCreatedAt() time.Time {
	if order.CreatedAt > 0 {
		return time.Unix(order.CreatedAt/1000, 0)
	}
	return time.Time{}
}

func (order *Order) GetUpdatedAt() time.Time {
	if order.UpdatedAt > 0 {
		return time.Unix(order.UpdatedAt/1000, 0)
	}
	return time.Time{}
}

func (order *Order) GetSymbol() string {
	return strings.ToUpper(order.BaseCoin) + "_" + strings.ToUpper(order.CountCoin)
}

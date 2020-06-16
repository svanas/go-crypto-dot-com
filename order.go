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
	TotalPrice   float64     `json:"total_price,string"`
	Fee          float64     `json:"fee,string"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
	DealPrice    float64     `json:"deal_price,string"`
	AvgPrice     float64     `json:"avg_price,string"`
	CountCoin    string      `json:"countCoin"`
	Source       int         `json:"source"`
	Type         interface{} `json:"type"`
	SideMsg      string      `json:"side_msg"`
	Volume       float64     `json:"volume,string"`
	Price        float64     `json:"price,string"`
	StatusMsg    string      `json:"status_msg"`
	DealVolume   float64     `json:"deal_volume,string"`
	FeeCoin      string      `json:"fee_coin"`
	RemainVolume float64     `json:"remain_volume,string"`
	BaseCoin     string      `json:"baseCoin"`
	Status       int         `json:"status"`
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

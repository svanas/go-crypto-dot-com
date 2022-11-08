package crypto

import (
	"time"
)

type OrderSide string

const (
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"
)

type OrderType string

const (
	LIMIT             OrderType = "LIMIT"
	MARKET            OrderType = "MARKET"
	STOP_LOSS         OrderType = "STOP_LOSS"
	STOP_LIMIT        OrderType = "STOP_LIMIT"
	TAKE_PROFIT       OrderType = "TAKE_PROFIT"
	TAKE_PROFIT_LIMIT OrderType = "TAKE_PROFIT_LIMIT"
)

type TimeInForce string

const (
	GOOD_TILL_CANCEL    TimeInForce = "GOOD_TILL_CANCEL"
	FILL_OR_KILL        TimeInForce = "FILL_OR_KILL"
	IMMEDIATE_OR_CANCEL TimeInForce = "IMMEDIATE_OR_CANCEL"
)

type OrderStatus string

const (
	ORDER_STATUS_ACTIVE   OrderStatus = "ACTIVE"
	ORDER_STATUS_CANCELED OrderStatus = "CANCELED"
	ORDER_STATUS_FILLED   OrderStatus = "FILLED"
	ORDER_STATUS_REJECTED OrderStatus = "REJECTED"
	ORDER_STATUS_EXPIRED  OrderStatus = "EXPIRED"
)

type Order struct {
	Status    OrderStatus `json:"status"`           // ACTIVE, CANCELED, FILLED, REJECTED or EXPIRED
	Reason    interface{} `json:"reason,omitempty"` // reason -- only for REJECTED orders
	Side      OrderSide   `json:"side"`             // BUY or SELL
	Price     float64     `json:"price,omitempty"`
	Quantity  float64     `json:"quantity"`
	OrderId   string      `json:"order_id"`
	CreatedAt int64       `json:"create_time"`
	UpdatedAt int64       `json:"update_time"`
	Type      OrderType   `json:"type"`
	Symbol    string      `json:"instrument_name"`
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

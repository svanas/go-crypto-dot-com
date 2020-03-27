package crypto

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

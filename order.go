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
	Id           int     `json:"id"`
	Side         string  `json:"side"`
	TotalPrice   float64 `json:"total_price,string"`
	Fee          float64 `json:"fee,string"`
	CreatedAt    int     `json:"created_at"`
	UpdatedAt    int     `json:"updated_at"`
	DealPrice    float64 `json:"deal_price,string"`
	AvgPrice     float64 `json:"avg_price,string"`
	CountCoin    string  `json:"countCoin"`
	Source       int     `json:"source"`
	Type         int     `json:"type"`
	SideMsg      string  `json:"side_msg"`
	Volume       float64 `json:"volume,string"`
	Price        float64 `json:"price,string"`
	SourceMsg    string  `json:"source_msg"`
	StatusMsg    string  `json:"status_msg"`
	DealVolume   float64 `json:"deal_volume,string"`
	FeeCoin      string  `json:"fee_coin"`
	RemainVolume float64 `json:"remain_volume,string"`
	BaseCoin     string  `json:"baseCoin"`
	Status       int     `json:"status"`
}

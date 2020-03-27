package crypto

type Account struct {
	TotalAsset float64 `json:"total_asset,string"`
	CoinList   []struct {
		Normal       float64     `json:"normal,string"`
		Locked       float64     `json:"locked,string"`
		BtcValuation interface{} `json:"btcValuation"`
		Coin         string      `json:"coin"`
	} `json:"coin_list"`
}

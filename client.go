package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const endpoint = "https://api.crypto.com/v2/"

type rateLimit int

const (
	RATE_LIMIT_NORMAL rateLimit = iota
	RATE_LIMIT_COOL_DOWN
)

var RequestsPerSecond = map[rateLimit]float64{
	RATE_LIMIT_NORMAL:    100,           // 100 req/second (default)
	RATE_LIMIT_COOL_DOWN: 0.01666666667, // 1 req/minute
}

var (
	cooldown    bool
	lastRequest time.Time
)

var (
	BeforeRequest    func(method, path string, rps float64) error = nil
	AfterRequest     func()                                       = nil
	OnRateLimitError func(method, path string) error              = nil
)

func init() {
	BeforeRequest = func(method, path string, rps float64) error {
		elapsed := time.Since(lastRequest)
		if cooldown {
			cooldown = false
			rps = RequestsPerSecond[RATE_LIMIT_COOL_DOWN]
		} else if rps == 0 {
			rps = RequestsPerSecond[RATE_LIMIT_NORMAL]
		}
		if elapsed.Seconds() < (float64(1) / rps) {
			time.Sleep(time.Duration((float64(time.Second) / rps) - float64(elapsed)))
		}
		return nil
	}
	AfterRequest = func() {
		lastRequest = time.Now()
	}
	OnRateLimitError = func(method, path string) error {
		cooldown = true
		return nil
	}
}

type Client struct {
	URL        string
	Key        string
	Secret     string
	httpClient *http.Client
}

func New(apiKey, apiSecret string) *Client {
	return &Client{
		URL:    endpoint,
		Key:    apiKey,
		Secret: apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type Request struct {
	Id     int                    `json:"id"`
	Method string                 `json:"method"`
	ApiKey string                 `json:"api_key"`
	Params map[string]interface{} `json:"params"`
	Sig    string                 `json:"sig"`
	Nonce  int64                  `json:"nonce"`
}

type Response struct {
	Code   interface{}     `json:"code"`
	Result json.RawMessage `json:"result"`
}

func (client *Client) get(path string, params *url.Values) (json.RawMessage, error) {
	// parse the root URL
	endpoint, err := url.Parse(client.URL)
	if err != nil {
		return nil, err
	}

	// set the endpoint for this request
	endpoint.Path += path
	if params != nil {
		endpoint.RawQuery = params.Encode()
	}

	var data []byte
	for {
		var code int
		code, data, err = func() (int, []byte, error) {
			// satisfy the rate limiter
			if err := BeforeRequest("GET", path, RequestsPerSecond[RATE_LIMIT_NORMAL]); err != nil {
				return 0, nil, err
			}
			defer func() {
				AfterRequest()
			}()

			response, err := client.httpClient.Get(endpoint.String())
			if err != nil {
				return 0, nil, err
			}
			defer response.Body.Close()

			// are we exceeding the rate limits?
			if response.StatusCode == http.StatusTooManyRequests {
				if err := OnRateLimitError("GET", path); err != nil {
					return response.StatusCode, nil, err
				}
			}

			// read the body of the response into a byte array
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return response.StatusCode, nil, err
			}

			// is this an error?
			status := make(map[string]interface{})
			if json.Unmarshal(body, &status) == nil {
				if code, ok := status["code"]; ok {
					if code != float64(0) {
						msg := func() string {
							if det, ok := status["details"]; ok {
								return fmt.Sprintf("%v", det)
							} else if msg, ok := status["message"]; ok {
								return fmt.Sprintf("%v", msg)
							} else {
								return fmt.Sprintf("%v", code)
							}
						}()
						return response.StatusCode, nil, func() error {
							if params == nil {
								return fmt.Errorf("GET %s %s", path, msg)
							} else {
								return fmt.Errorf("GET %s?%s %s", path, params.Encode(), msg)
							}
						}()
					}
				}
			}

			if response.StatusCode < 200 || response.StatusCode >= 300 {
				return response.StatusCode, nil, func() error {
					if params == nil {
						return fmt.Errorf("GET %s %s", path, response.Status)
					} else {
						return fmt.Errorf("GET %s?%s %s", path, params.Encode(), response.Status)
					}
				}()
			}

			var output Response
			if err := json.Unmarshal(body, &output); err != nil {
				return response.StatusCode, nil, err
			}

			return response.StatusCode, output.Result, nil
		}()

		if code != http.StatusTooManyRequests {
			break
		}
	}

	return data, err
}

func params(symbol string, page int) map[string]interface{} {
	output := make(map[string]interface{})
	if symbol != "" {
		output["instrument_name"] = symbol
	}
	if page > 0 {
		output["page"] = page
	}
	return output
}

func (client *Client) post(path string, params map[string]interface{}, rps float64) ([]byte, error) {
	// create the endpoint for this request
	endpoint, err := url.Parse(client.URL)
	if err != nil {
		return nil, err
	}
	endpoint.Path += path

	var data []byte
	for {
		var code int
		code, data, err = func() (int, []byte, error) {
			// satisfy the rate limiter
			if err := BeforeRequest("POST", path, rps); err != nil {
				return 0, nil, err
			}
			defer func() {
				AfterRequest()
			}()

			nonce := time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)

			// generate signature
			var sig strings.Builder
			sig.WriteString(path)       // method
			sig.WriteString("0")        // id
			sig.WriteString(client.Key) // api_key
			keys := make([]string, 0, len(params))
			for key := range params {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				value := params[key]
				if value != nil {
					sig.WriteString(key)
					sig.WriteString(func(v interface{}) string {
						if i, ok := v.(int); ok {
							return strconv.Itoa(i)
						}
						if i64, ok := v.(int64); ok {
							return strconv.FormatInt(i64, 10)
						}
						if f64, ok := v.(float64); ok {
							return strconv.FormatFloat(f64, 'f', -1, 64)
						}
						return fmt.Sprintf("%v", v)
					}(value))
				}
			}
			sig.WriteString(strconv.FormatInt(nonce, 10))
			mac := hmac.New(sha256.New, []byte(client.Secret))
			mac.Write([]byte(sig.String()))

			payload, err := json.Marshal(Request{
				Id:     0,
				Method: path,
				ApiKey: client.Key,
				Params: params,
				Sig:    hex.EncodeToString(mac.Sum(nil)),
				Nonce:  nonce,
			})
			if err != nil {
				return 0, nil, err
			}

			// create the request
			request, err := http.NewRequest("POST", endpoint.String(), strings.NewReader(string(payload)))
			if err != nil {
				return 0, nil, err
			}
			request.Header.Add("Content-Type", "application/json")

			// submit the http request
			response, err := client.httpClient.Do(request)
			if err != nil {
				return 0, nil, err
			}
			defer response.Body.Close()

			// are we exceeding the rate limits?
			if response.StatusCode == http.StatusTooManyRequests {
				if err = OnRateLimitError("POST", path); err != nil {
					return response.StatusCode, nil, err
				}
			}

			// read the body of the response into a byte array
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return response.StatusCode, nil, err
			}

			// is this an error?
			status := make(map[string]interface{})
			if json.Unmarshal(body, &status) == nil {
				if code, ok := status["code"]; ok {
					if code != float64(0) {
						msg := func() string {
							if det, ok := status["details"]; ok {
								return fmt.Sprintf("%v", det)
							} else if msg, ok := status["message"]; ok {
								return fmt.Sprintf("%v", msg)
							} else {
								return fmt.Sprintf("%v", code)
							}
						}()
						return response.StatusCode, nil, fmt.Errorf("POST %s %s", path, msg)
					}
				}
			}

			if response.StatusCode < 200 || response.StatusCode >= 300 {
				return response.StatusCode, nil, fmt.Errorf("POST %s %s", path, response.Status)
			}

			// unmarshal the response body
			var output Response
			if err = json.Unmarshal(body, &output); err != nil {
				return response.StatusCode, nil, err
			}

			return response.StatusCode, output.Result, nil
		}()

		if code != http.StatusTooManyRequests {
			break
		}
	}

	return data, err
}

func (client *Client) Symbols() ([]Symbol, error) {
	raw, err := client.get("public/get-instruments", nil)
	if err != nil {
		return nil, err
	}
	type Result struct {
		Instruments []Symbol `json:"instruments"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Instruments, nil
}

func (client *Client) Tickers() ([]Ticker, error) {
	raw, err := client.get("public/get-ticker", nil)
	if err != nil {
		return nil, err
	}
	type Result struct {
		Data []Ticker `json:"data"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (client *Client) Ticker(symbol string) (*Ticker, error) {
	params := url.Values{}
	params.Add("instrument_name", symbol)
	raw, err := client.get("public/get-ticker", &params)
	if err != nil {
		return nil, err
	}
	type Result struct {
		Data []Ticker `json:"data"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("%s does not exist", symbol)
	}
	return &result.Data[0], nil
}

func (client *Client) OrderBook(symbol string) (*OrderBook, error) {
	params := url.Values{}
	params.Add("instrument_name", symbol)
	raw, err := client.get("public/get-book", &params)
	if err != nil {
		return nil, err
	}
	type Result struct {
		Data []OrderBook `json:"data"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("%s does not exist", symbol)
	}
	return &result.Data[0], nil
}

func (client *Client) Accounts() ([]Account, error) {
	raw, err := client.post("private/get-account-summary", nil, 30)
	if err != nil {
		return nil, err
	}
	type Result struct {
		Accounts []Account `json:"accounts"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Accounts, nil
}

func (client *Client) Account(asset string) (*Account, error) {
	params := make(map[string]interface{})
	params["currency"] = asset
	raw, err := client.post("private/get-account-summary", params, 30)
	if err != nil {
		return nil, err
	}
	type Result struct {
		Accounts []Account `json:"accounts"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if len(result.Accounts) == 0 {
		return nil, fmt.Errorf("%s does not exist", asset)
	}
	return &result.Accounts[0], nil
}

func (client *Client) CreateOrder(symbol string, side OrderSide, kind OrderType, quantity, price float64) (*string, error) { // -> (order_id, error)
	params := make(map[string]interface{})
	params["instrument_name"] = symbol
	params["side"] = side
	params["type"] = kind
	params["quantity"] = quantity
	if kind == LIMIT || kind == STOP_LIMIT {
		params["price"] = price
	}
	raw, err := client.post("private/create-order", params, 150)
	if err != nil {
		return nil, err
	}
	type Result struct {
		OrderId string `json:"order_id"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	order, err := client.GetOrder(symbol, result.OrderId)
	if err != nil {
		return &result.OrderId, err
	}
	if order.Status == ORDER_STATUS_REJECTED {
		return &result.OrderId, fmt.Errorf("order rejected. reason: %v", order.Reason)
	}
	if order.Status == ORDER_STATUS_EXPIRED {
		var (
			base  = strings.Split(symbol, "/")[0]
			quote = strings.Split(symbol, "/")[1]
		)
		return &result.OrderId, fmt.Errorf("cannot %v %s unit(s) of %s at %s %s. your available balance is %s %s",
			side, strconv.FormatFloat(quantity, 'f', -1, 64), base, quote,
			strconv.FormatFloat(func() float64 {
				if kind == MARKET {
					ticker, err := client.Ticker(symbol)
					if err == nil {
						return ticker.Last
					}
				}
				return price
			}(), 'f', -1, 64), func() string {
				if side == SELL {
					return base
				}
				return quote
			}(), strconv.FormatFloat(func() float64 {
				account, err := client.Account(func() string {
					if side == SELL {
						return base
					}
					return quote
				}())
				if err == nil {
					return account.Available
				}
				return 0
			}(), 'f', -1, 64))
	}
	return &result.OrderId, nil
}

func (client *Client) GetOrder(symbol, orderId string) (*Order, error) {
	params := make(map[string]interface{})
	params["instrument_name"] = symbol
	params["order_id"] = orderId
	raw, err := client.post("private/get-order-detail", params, 300)
	if err != nil {
		return nil, err
	}
	type Result struct {
		OrderInfo Order `json:"order_info"`
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result.OrderInfo, nil
}

func (client *Client) CancelOrder(symbol, orderId string) error {
	params := make(map[string]interface{})
	params["instrument_name"] = symbol
	params["order_id"] = orderId
	_, err := client.post("private/cancel-order", params, 150)
	return err
}

func (client *Client) OpenOrders(symbol string) ([]Order, error) {
	call := func(params map[string]interface{}) (int, []Order, error) {
		raw, err := client.post("private/get-open-orders", params, 30)
		if err != nil {
			return 0, nil, err
		}
		type Result struct {
			Count     int     `json:"count"`
			OrderList []Order `json:"order_list"`
		}
		var result Result
		if err := json.Unmarshal(raw, &result); err != nil {
			return 0, nil, err
		}
		return result.Count, result.OrderList, nil
	}

	var (
		page   int = 0
		result []Order
	)

	count, orders, err := call(params(symbol, page))
	if err != nil {
		return nil, err
	}
	result = append(result, orders...)

	for len(result) < count {
		page++
		_, orders, err := call(params(symbol, page))
		if err != nil {
			return nil, err
		}
		result = append(result, orders...)
	}

	return result, nil
}

func (client *Client) MyTrades(symbol string) ([]Trade, error) {
	call := func(params map[string]interface{}) (int, []Trade, error) {
		raw, err := client.post("private/get-trades", params, 1)
		if err != nil {
			return 0, nil, err
		}
		type Result struct {
			Count     int     `json:"count"`
			TradeList []Trade `json:"trade_list"`
		}
		var result Result
		if err := json.Unmarshal(raw, &result); err != nil {
			return 0, nil, err
		}
		return result.Count, result.TradeList, nil
	}

	var (
		page   int = 0
		result []Trade
	)

	count, trades, err := call(params(symbol, page))
	if err != nil {
		return nil, err
	}
	result = append(result, trades...)

	for len(result) < count {
		page++
		_, trades, err := call(params(symbol, page))
		if err != nil {
			return nil, err
		}
		result = append(result, trades...)
	}

	return result, nil
}

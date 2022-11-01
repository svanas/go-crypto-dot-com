package crypto

import (
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

const (
	pageSize = 50
	endpoint = "https://api.crypto.com"
)

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
	BeforeRequest    func(method, path string, params *url.Values, rps float64) error = nil
	AfterRequest     func()                                                           = nil
	OnRateLimitError func(method, path string) error                                  = nil
)

func init() {
	BeforeRequest = func(method, path string, params *url.Values, rps float64) error {
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

type Response1 struct {
	Code interface{}     `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type Response2 struct {
	Code   interface{}     `json:"code"`
	Result json.RawMessage `json:"result"`
}

func params(symbol string, page, pageSize int) url.Values {
	output := url.Values{}
	output.Set("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	if page > 0 {
		output.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		output.Set("pageSize", strconv.Itoa(pageSize))
	}
	return output
}

func (client *Client) get_v1(path string, params *url.Values) (json.RawMessage, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("GET", path, params, RequestsPerSecond[RATE_LIMIT_NORMAL]); err != nil {
		return nil, err
	}
	defer func() {
		AfterRequest()
	}()

	// parse the root URL
	var endpoint *url.URL
	if endpoint, err = url.Parse(client.URL); err != nil {
		return nil, err
	}

	// set the endpoint for this request
	endpoint.Path += path
	if params != nil {
		endpoint.RawQuery = params.Encode()
	}

	var resp *http.Response
	if resp, err = client.httpClient.Get(endpoint.String()); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// are we exceeding the rate limits?
	if resp.StatusCode == http.StatusTooManyRequests {
		if err = OnRateLimitError("GET", path); err != nil {
			return nil, err
		}
	}

	// read the body of the response into a byte array
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	// is this an error?
	status := make(map[string]interface{})
	if json.Unmarshal(body, &status) == nil {
		if msg, ok := status["msg"]; ok {
			if msg != "suc" {
				return nil, fmt.Errorf("%v", msg)
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, func() error {
			if params == nil {
				return fmt.Errorf("GET %s %s", resp.Status, path)
			} else {
				return fmt.Errorf("GET %s %s?%s", resp.Status, path, params.Encode())
			}
		}()
	}

	var output Response1
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Data, nil
}

func (client *Client) get_v2(path string, params *url.Values) (json.RawMessage, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("GET", path, params, RequestsPerSecond[RATE_LIMIT_NORMAL]); err != nil {
		return nil, err
	}
	defer func() {
		AfterRequest()
	}()

	// parse the root URL
	var endpoint *url.URL
	if endpoint, err = url.Parse(client.URL); err != nil {
		return nil, err
	}

	// set the endpoint for this request
	endpoint.Path += path
	if params != nil {
		endpoint.RawQuery = params.Encode()
	}

	var resp *http.Response
	if resp, err = client.httpClient.Get(endpoint.String()); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// are we exceeding the rate limits?
	if resp.StatusCode == http.StatusTooManyRequests {
		if err = OnRateLimitError("GET", path); err != nil {
			return nil, err
		}
	}

	// read the body of the response into a byte array
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
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
				return nil, func() error {
					if params == nil {
						return fmt.Errorf("GET %s %s", msg, path)
					} else {
						return fmt.Errorf("GET %s %s?%s", msg, path, params.Encode())
					}
				}()
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, func() error {
			if params == nil {
				return fmt.Errorf("GET %s %s", resp.Status, path)
			} else {
				return fmt.Errorf("GET %s %s?%s", resp.Status, path, params.Encode())
			}
		}()
	}

	var output Response2
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Result, nil
}

func (client *Client) post(path string, params url.Values, rps float64) ([]byte, error) {
	// create the endpoint for this request
	endpoint, err := url.Parse(client.URL)
	if err != nil {
		return nil, err
	}
	endpoint.Path += path

	var (
		code int
		data []byte
	)
	for {
		code, data, err = func(params url.Values) (int, []byte, error) {
			// satisfy the rate limiter
			if err = BeforeRequest("POST", path, &params, rps); err != nil {
				return 0, nil, err
			}
			defer func() {
				AfterRequest()
			}()

			// add API key & time to params
			params.Set("api_key", client.Key)
			time := time.Now().UnixNano() / int64(time.Millisecond/time.Nanosecond)
			params.Set("time", strconv.FormatInt(time, 10))

			// add signature to params
			keys := make([]string, 0, len(params))
			for key := range params {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			var buf strings.Builder
			for _, key := range keys {
				value := params.Get(key)
				if value != "" {
					buf.WriteString(key)
					buf.WriteString(value)
				}
			}
			buf.WriteString(client.Secret)
			hash := sha256.New()
			hash.Write([]byte(buf.String()))
			params.Set("sign", hex.EncodeToString(hash.Sum(nil)))

			// create the request
			var req *http.Request
			if req, err = http.NewRequest("POST", endpoint.String(), strings.NewReader(params.Encode())); err != nil {
				return 0, nil, err
			}
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			// submit the http request
			var resp *http.Response
			if resp, err = client.httpClient.Do(req); err != nil {
				return 0, nil, err
			}
			defer resp.Body.Close()

			// are we exceeding the rate limits?
			if resp.StatusCode == http.StatusTooManyRequests {
				if err = OnRateLimitError("POST", path); err != nil {
					return resp.StatusCode, nil, err
				}
			}

			// read the body of the response into a byte array
			var body []byte
			if body, err = ioutil.ReadAll(resp.Body); err != nil {
				return resp.StatusCode, nil, err
			}

			// is this an error?
			status := make(map[string]interface{})
			if json.Unmarshal(body, &status) == nil {
				if msg, ok := status["msg"]; ok {
					if msg != "suc" {
						return resp.StatusCode, nil, fmt.Errorf("%v", msg)
					}
				}
			}

			if resp.StatusCode != http.StatusOK {
				return resp.StatusCode, nil, func() error {
					if len(params) == 0 {
						return fmt.Errorf("POST %s %s", resp.Status, path)
					} else {
						return fmt.Errorf("POST %s %s?%s", resp.Status, path, params.Encode())
					}
				}()
			}

			// unmarshal the response body
			var result Response1
			if err = json.Unmarshal(body, &result); err != nil {
				return resp.StatusCode, nil, err
			}

			return resp.StatusCode, result.Data, nil
		}(func() url.Values {
			copied := url.Values{}
			for key, value := range params {
				copied[key] = value
			}
			return copied
		}())

		if code != http.StatusTooManyRequests {
			break
		}
	}

	return data, err
}

func (client *Client) Symbols() ([]Symbol, error) {
	var (
		err error
		raw json.RawMessage
	)
	if raw, err = client.get_v2("/v2/public/get-instruments", nil); err != nil {
		return nil, err
	}
	type Result struct {
		Instruments []Symbol `json:"instruments"`
	}
	var result Result
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Instruments, nil
}

func (client *Client) Tickers() (*Tickers, error) {
	var (
		err  error
		data json.RawMessage
	)
	if data, err = client.get_v1("/v1/ticker", nil); err != nil {
		return nil, err
	}
	var output Tickers
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (client *Client) Ticker(symbol string) (*Ticker, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Add("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	if data, err = client.get_v1("/v1/ticker", &params); err != nil {
		return nil, err
	}
	var output Ticker
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (client *Client) OrderBook(symbol string) (*OrderBook, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Add("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	if data, err = client.get_v1("/v1/depth", &params); err != nil {
		return nil, err
	}
	var output OrderBook
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (client *Client) Account() (*Account, error) {
	var (
		err  error
		data json.RawMessage
	)
	if data, err = client.post("/v1/account", url.Values{}, 1); err != nil {
		return nil, err
	}
	var output Account
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (client *Client) CreateOrder(symbol string, side OrderSide, kind OrderType, quantity, price float64) (int64, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Set("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	params.Set("side", side.String())
	params.Set("type", kind.String())
	params.Set("volume", strconv.FormatFloat(quantity, 'f', -1, 64))
	if kind != MARKET {
		params.Add("price", strconv.FormatFloat(price, 'f', -1, 64))
	}
	if data, err = client.post("/v1/order", params, 5); err != nil {
		return 0, err
	}
	type Response struct {
		OrderId int64 `json:"order_id,string"`
	}
	var (
		resp  Response
		order *Order
	)
	if err = json.Unmarshal(data, &resp); err != nil {
		return 0, err
	}
	if order, err = client.GetOrder(symbol, resp.OrderId); err != nil {
		return resp.OrderId, err
	}
	if order.Status == ORDER_STATUS_EXPIRED {
		//lint:ignore ST1005 error strings should not be capitalized
		return resp.OrderId, fmt.Errorf("Cannot %s %s unit(s) of %s at %s %s. Your available balance is %s %s.", func() string {
			if side == SELL {
				return "sell"
			} else {
				return "buy"
			}
		}(), strconv.FormatFloat(quantity, 'f', -1, 64), order.BaseCoin, order.CountCoin, strconv.FormatFloat(func() float64 {
			if kind == MARKET {
				ticker, err := client.Ticker(symbol)
				if err == nil {
					return ticker.Last
				}
			}
			return price
		}(), 'f', -1, 64), order.CountCoin, strconv.FormatFloat(func() float64 {
			account, err := client.Account()
			if err == nil {
				for _, coin := range account.CoinList {
					if strings.EqualFold(coin.Coin, order.CountCoin) {
						return coin.Normal
					}
				}
			}
			return 0
		}(), 'f', -1, 64))
	}
	return resp.OrderId, nil
}

func (client *Client) GetOrder(symbol string, orderId int64) (*Order, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Set("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	params.Set("order_id", strconv.FormatInt(orderId, 10))
	if data, err = client.post("/v1/showOrder", params, 10); err != nil {
		return nil, err
	}
	type Output struct {
		OrderInfo Order `json:"orderInfo"`
	}
	var output Output
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output.OrderInfo, nil
}

func (client *Client) CancelOrder(symbol string, orderId int64) error {
	params := url.Values{}
	params.Set("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	params.Set("order_id", strconv.FormatInt(orderId, 10))
	if _, err := client.post("/v1/orders/cancel", params, 5); err != nil {
		return err
	}
	return nil
}

func (client *Client) OpenOrders(symbol string) ([]Order, error) {
	call := func(params url.Values) (int, []Order, error) {
		var (
			err  error
			data json.RawMessage
		)
		if data, err = client.post("/v1/openOrders", params, 1); err != nil {
			return 0, nil, err
		}
		type Output struct {
			Count      int     `json:"count"`
			ResultList []Order `json:"resultList"`
		}
		var output Output
		if err = json.Unmarshal(data, &output); err != nil {
			return 0, nil, err
		}
		return output.Count, output.ResultList, nil
	}

	var (
		page   int = 0
		result []Order
	)

	count, orders, err := call(params(symbol, page, pageSize))
	if err != nil {
		return nil, err
	}
	result = append(result, orders...)

	for len(result) < count {
		page++
		_, orders, err := call(params(symbol, page, pageSize))
		if err != nil {
			return nil, err
		}
		result = append(result, orders...)
	}

	return result, nil
}

func (client *Client) MyTrades(symbol string) ([]Trade, error) {
	call := func(params url.Values) (int, []Trade, error) {
		var (
			err  error
			data json.RawMessage
		)
		if data, err = client.post("/v1/myTrades", params, 1); err != nil {
			return 0, nil, err
		}
		type Output struct {
			Count      int     `json:"count"`
			ResultList []Trade `json:"resultList"`
		}
		var output Output
		if err = json.Unmarshal(data, &output); err != nil {
			return 0, nil, err
		}
		return output.Count, output.ResultList, nil
	}

	var (
		page   int = 0
		result []Trade
	)

	count, trades, err := call(params(symbol, page, pageSize))
	if err != nil {
		return nil, err
	}
	result = append(result, trades...)

	for len(result) < count {
		page++
		_, trades, err := call(params(symbol, page, pageSize))
		if err != nil {
			return nil, err
		}
		result = append(result, trades...)
	}

	for i := range result {
		result[i].Symbol = symbol
	}

	return result, nil
}

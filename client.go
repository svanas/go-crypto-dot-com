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
	RATE_LIMIT_NORMAL:    1,          // 1 req/second (default)
	RATE_LIMIT_COOL_DOWN: 0.01666666, // 1 req/minute
}

var (
	cooldown    bool
	lastRequest time.Time
)

func getRequestsPerSecond() float64 {
	if cooldown {
		cooldown = false
		return RequestsPerSecond[RATE_LIMIT_COOL_DOWN]
	}
	return RequestsPerSecond[RATE_LIMIT_NORMAL]
}

var (
	BeforeRequest    func(method, path string, params *url.Values) error = nil
	AfterRequest     func()                                              = nil
	OnRateLimitError func(method, path string) error                     = nil
)

func init() {
	BeforeRequest = func(method, path string, params *url.Values) error {
		elapsed := time.Since(lastRequest)
		rps := getRequestsPerSecond()
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

func (client *Client) get(path string, params *url.Values) (json.RawMessage, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("GET", path, params); err != nil {
		return nil, err
	}
	defer func() {
		AfterRequest()
	}()

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

	// read the body of the response into a byte array
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	// are we exceeding the rate limits?
	if resp.StatusCode == http.StatusTooManyRequests {
		if err = OnRateLimitError("GET", path); err != nil {
			return nil, err
		}
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
		return nil, fmt.Errorf("GET %s %s", resp.Status, path)
	}

	var output Response1
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Data, nil
}

func (client *Client) get2(path string, params *url.Values) (json.RawMessage, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("GET", path, params); err != nil {
		return nil, err
	}
	defer func() {
		AfterRequest()
	}()

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

	// read the body of the response into a byte array
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	// are we exceeding the rate limits?
	if resp.StatusCode == http.StatusTooManyRequests {
		if err = OnRateLimitError("GET", path); err != nil {
			return nil, err
		}
	}

	// is this an error?
	status := make(map[string]interface{})
	if json.Unmarshal(body, &status) == nil {
		if code, ok := status["code"]; ok {
			if code != float64(0) {
				return nil, fmt.Errorf("GET error code %v %s", code, path)
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s %s", resp.Status, path)
	}

	var output Response2
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Result, nil
}

func (client *Client) post(path string, params url.Values) ([]byte, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("POST", path, &params); err != nil {
		return nil, err
	}
	defer func() {
		AfterRequest()
	}()

	var endpoint *url.URL
	if endpoint, err = url.Parse(client.URL); err != nil {
		return nil, err
	}

	// set the endpoint for this request
	endpoint.Path += path

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

	// encode the url.Values into the request body
	var input *strings.Reader
	input = strings.NewReader(params.Encode())

	// create the request
	var req *http.Request
	if req, err = http.NewRequest("POST", endpoint.String(), input); err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// submit the http request
	var resp *http.Response
	if resp, err = client.httpClient.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read the body of the response into a byte array
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	// are we exceeding the rate limits?
	if resp.StatusCode == http.StatusTooManyRequests {
		if err = OnRateLimitError("POST", path); err != nil {
			return nil, err
		}
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
		return nil, fmt.Errorf("POST %s %s", resp.Status, path)
	}

	var output Response1
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Data, nil
}

func (client *Client) Symbols() ([]Symbol, error) {
	var (
		err error
		raw json.RawMessage
	)
	if raw, err = client.get2("/v2/public/get-instruments", nil); err != nil {
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
	if data, err = client.get("/v1/ticker", nil); err != nil {
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
	if data, err = client.get("/v1/ticker", &params); err != nil {
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
	if data, err = client.get("/v1/depth", &params); err != nil {
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
	if data, err = client.post("/v1/account", url.Values{}); err != nil {
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
	if data, err = client.post("/v1/order", params); err != nil {
		return 0, err
	}
	type Output struct {
		OrderId int64 `json:"order_id,string"`
	}
	var output Output
	if err = json.Unmarshal(data, &output); err != nil {
		return 0, err
	}
	return output.OrderId, nil
}

func (client *Client) GetOrder(symbol string, orderId int64) (*Order, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Set("symbol", strings.ToLower(strings.Replace(symbol, "_", "", -1)))
	params.Set("order_id", strconv.FormatInt(orderId, 10))
	if data, err = client.post("/v1/showOrder", params); err != nil {
		return nil, err
	}
	type Output struct {
		OrderInfo Order `json:"order_info"`
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
	if _, err := client.post("/v1/orders/cancel", params); err != nil {
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
		if data, err = client.post("/v1/openOrders", params); err != nil {
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
		if data, err = client.post("/v1/myTrades", params); err != nil {
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

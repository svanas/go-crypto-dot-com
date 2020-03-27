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

const endpoint = "https://api.crypto.com"

var (
	lastRequest       time.Time
	RequestsPerSecond float64                         = 10
	BeforeRequest     func(method, path string) error = nil
	AfterRequest      func()                          = nil
)

func init() {
	BeforeRequest = func(method, path string) error {
		elapsed := time.Since(lastRequest)
		if elapsed.Seconds() < (float64(1) / RequestsPerSecond) {
			time.Sleep(time.Duration((float64(time.Second) / RequestsPerSecond) - float64(elapsed)))
		}
		return nil
	}
	AfterRequest = func() {
		lastRequest = time.Now()
	}
}

type Client struct {
	URL    string
	Key    string
	Secret string
}

func New(apiKey, apiSecret string) *Client {
	return &Client{
		URL:    endpoint,
		Key:    apiKey,
		Secret: apiSecret,
	}
}

type Response struct {
	Code int             `json:"code,string"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (resp *Response) success() bool {
	return resp.Code == 0 && resp.Msg == "suc"
}

func (client *Client) get(path string, params *url.Values) (json.RawMessage, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("GET", path); err != nil {
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
	if resp, err = http.Get(endpoint.String()); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
		return nil, fmt.Errorf("GET %s %s", resp.Status, path)
	}

	var output Response
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Data, nil
}

func (client *Client) post(path string, params url.Values) ([]byte, error) {
	var err error

	// satisfy the rate limiter
	if err = BeforeRequest("POST", path); err != nil {
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
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
		return nil, fmt.Errorf("POST %s %s", resp.Status, path)
	}

	var output Response
	if err = json.Unmarshal(body, &output); err != nil {
		return nil, err
	}

	return output.Data, nil
}

func (client *Client) Symbols() ([]Symbol, error) {
	var (
		err  error
		data json.RawMessage
	)
	if data, err = client.get("/v1/symbols", nil); err != nil {
		return nil, err
	}
	var output []Symbol
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return output, nil
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
	params.Add("symbol", symbol)
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
	params.Add("symbol", symbol)
	params.Add("type", "step0")
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

func (client *Client) CreateOrder(symbol string, side OrderSide, kind OrderType, quantity, price float64) (int, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Set("symbol", symbol)
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
		OrderId int `json:"order_id"`
	}
	var output Output
	if err = json.Unmarshal(data, &output); err != nil {
		return 0, err
	}
	return output.OrderId, nil
}

func (client *Client) GetOrder(symbol string, orderId int) (*Order, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("order_id", strconv.Itoa(orderId))
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

func (client *Client) CancelOrder(symbol string, orderId int) error {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("order_id", strconv.Itoa(orderId))
	if _, err := client.post("/v1/orders/cancel", params); err != nil {
		return err
	}
	return nil
}

func (client *Client) OpenedOrders(symbol string) ([]Order, error) {
	var (
		err  error
		data json.RawMessage
	)
	params := url.Values{}
	params.Set("symbol", symbol)
	if data, err = client.post("/v1/openOrders", params); err != nil {
		return nil, err
	}
	type Output struct {
		Count      int     `json:"count"`
		ResultList []Order `json:"resultList"`
	}
	var output Output
	if err = json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return output.ResultList, nil
}

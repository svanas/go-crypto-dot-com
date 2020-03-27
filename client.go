package crypto

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

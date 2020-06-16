package crypto

import (
	"testing"
)

func TestSymbols(t *testing.T) {
	client := New("", "")

	symbols, err := client.Symbols()
	if err != nil {
		t.Errorf("Symbols() failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Symbols() returned an empty response")
	}

	t.Logf("%+v", symbols)
}

func TestTickers(t *testing.T) {
	client := New("", "")

	tickers, err := client.Tickers()
	if err != nil {
		t.Errorf("Tickers() failed: %v", err)
	}

	if len(tickers.Ticker) == 0 {
		t.Error("Tickers() returned an empty response")
	}

	t.Logf("%+v", tickers)
}

func TestTicker(t *testing.T) {
	client := New("", "")

	ticker, err := client.Ticker("ETH_BTC")
	if err != nil {
		t.Errorf("Ticker(\"ETH_BTC\") failed: %v", err)
	}

	t.Logf("%+v", ticker)
}

func TestOrderBook(t *testing.T) {
	client := New("", "")

	book, err := client.OrderBook("ETH_BTC")
	if err != nil {
		t.Errorf("OrderBook(\"ETH_BTC\") failed: %v", err)
	}

	t.Logf("%+v", book)
}

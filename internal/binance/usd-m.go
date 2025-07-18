package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const url = "https://fapi.binance.com"

type BinanceUsdM struct{}

type ExchangeInfo struct {
	Symbols []Symbol `json:"symbols"`
}

type Symbol struct {
	Symbol string `json:"symbol"`
	Status string `json:"status"`
}

func (b *BinanceUsdM) GetSymbols() ([]Symbol, error) {
	httpClient := &http.Client{}
	resp, err := httpClient.Get(url + "/fapi/v1/exchangeInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var exchangeInfo ExchangeInfo
	err = json.Unmarshal(content, &exchangeInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return exchangeInfo.Symbols, nil
}

type Kline struct {
	OpenTime            int64
	Open                string
	High                string
	Low                 string
	Close               string
	Volume              string
	CloseTime           int64
	QuoteAssetVolume    string
	NumberOfTrades      int64
	TakerBuyBaseVolume  string
	TakerBuyQuoteVolume string
	Ignore              string
}

func (b *BinanceUsdM) KLines(symbol, interval string, limit int) ([]Kline, error) {
	httpClient := &http.Client{}
	resp, err := httpClient.Get(url + fmt.Sprintf("/fapi/v1/klines?symbol=%s&limit=%d&interval=%s", symbol, limit, interval))
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var klines [][]interface{}
	err = json.Unmarshal(content, &klines)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var result []Kline
	for _, kline := range klines {
		if len(kline) < 12 {
			continue
		}

		result = append(result, Kline{
			OpenTime:            int64(kline[0].(float64)),
			Open:                kline[1].(string),
			High:                kline[2].(string),
			Low:                 kline[3].(string),
			Close:               kline[4].(string),
			Volume:              kline[5].(string),
			CloseTime:           int64(kline[6].(float64)),
			QuoteAssetVolume:    kline[7].(string),
			NumberOfTrades:      int64(kline[8].(float64)),
			TakerBuyBaseVolume:  kline[9].(string),
			TakerBuyQuoteVolume: kline[10].(string),
			Ignore:              kline[11].(string),
		})
	}

	return result, nil
}

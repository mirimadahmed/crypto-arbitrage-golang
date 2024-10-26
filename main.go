package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shopspring/decimal"
)

type ExchangePrice struct {
	Symbol   string
	BidPrice decimal.Decimal
	AskPrice decimal.Decimal
}

type BybitInstrumentsInfo struct {
	Result struct {
		List []struct {
			Symbol    string `json:"symbol"`
			BaseCoin  string `json:"baseCoin"`
			QuoteCoin string `json:"quoteCoin"`
			Status    string `json:"status"`
		} `json:"list"`
	} `json:"result"`
}

type BybitTickers struct {
	Result struct {
		List []struct {
			Symbol    string `json:"symbol"`
			Bid1Price string `json:"bid1Price"`
			Ask1Price string `json:"ask1Price"`
		} `json:"list"`
	} `json:"result"`
}

type BinanceTicker struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
}

const minProfitPercentage = 0.01 // Minimum 2% profit
const transactionFee = 0.001     // 0.1% transaction fee per exchange

func main() {
	bybitPairs, err := getBybitPairs()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Retrieved %d pairs from Bybit", len(bybitPairs))

	binancePairs, err := getBinancePairs()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Retrieved %d pairs from Binance", len(binancePairs))

	findArbitrageBetweenExchanges(bybitPairs, binancePairs)
}

func getBybitPairs() (map[string]ExchangePrice, error) {
	instrumentsInfo, err := getBybitInstrumentsInfo()
	if err != nil {
		return nil, err
	}

	tickers, err := getBybitTickers()
	if err != nil {
		return nil, err
	}

	// Create a map of active trading pairs
	activePairs := make(map[string]bool)
	for _, instrument := range instrumentsInfo.Result.List {
		if instrument.Status == "Trading" {
			activePairs[instrument.Symbol] = true
		}
	}

	pairs := make(map[string]ExchangePrice)
	for _, ticker := range tickers.Result.List {
		if !activePairs[ticker.Symbol] {
			continue
		}
		bidPrice, err := decimal.NewFromString(ticker.Bid1Price)
		if err != nil || bidPrice.IsZero() {
			continue
		}
		askPrice, err := decimal.NewFromString(ticker.Ask1Price)
		if err != nil || askPrice.IsZero() {
			continue
		}
		pairs[ticker.Symbol] = ExchangePrice{
			Symbol:   ticker.Symbol,
			BidPrice: bidPrice,
			AskPrice: askPrice,
		}
	}

	return pairs, nil
}

func getBinancePairs() (map[string]ExchangePrice, error) {
	apiURL := "https://api.binance.com/api/v3/ticker/bookTicker"
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching Binance tickers: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading Binance response: %v", err)
	}

	var tickers []BinanceTicker
	err = json.Unmarshal(body, &tickers)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling Binance tickers: %v", err)
	}

	pairs := make(map[string]ExchangePrice)
	for _, ticker := range tickers {
		bidPrice, err := decimal.NewFromString(ticker.BidPrice)
		if err != nil || bidPrice.IsZero() {
			continue
		}
		askPrice, err := decimal.NewFromString(ticker.AskPrice)
		if err != nil || askPrice.IsZero() {
			continue
		}
		pairs[ticker.Symbol] = ExchangePrice{
			Symbol:   ticker.Symbol,
			BidPrice: bidPrice,
			AskPrice: askPrice,
		}
	}

	return pairs, nil
}

func getBybitInstrumentsInfo() (BybitInstrumentsInfo, error) {
	apiURL := "https://api.bybit.com/v5/market/instruments-info?category=spot"
	resp, err := http.Get(apiURL)
	if err != nil {
		return BybitInstrumentsInfo{}, fmt.Errorf("error fetching Bybit instruments info: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BybitInstrumentsInfo{}, fmt.Errorf("error reading Bybit response: %v", err)
	}

	var instrumentsInfo BybitInstrumentsInfo
	err = json.Unmarshal(body, &instrumentsInfo)
	if err != nil {
		return BybitInstrumentsInfo{}, fmt.Errorf("error unmarshalling Bybit instruments info: %v", err)
	}

	return instrumentsInfo, nil
}

func getBybitTickers() (BybitTickers, error) {
	apiURL := "https://api.bybit.com/v5/market/tickers?category=spot"
	resp, err := http.Get(apiURL)
	if err != nil {
		return BybitTickers{}, fmt.Errorf("error fetching Bybit tickers: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return BybitTickers{}, fmt.Errorf("error reading Bybit response: %v", err)
	}

	var tickers BybitTickers
	err = json.Unmarshal(body, &tickers)
	if err != nil {
		return BybitTickers{}, fmt.Errorf("error unmarshalling Bybit tickers: %v", err)
	}

	return tickers, nil
}

func findArbitrageBetweenExchanges(bybitPairs, binancePairs map[string]ExchangePrice) {
	log.Printf("Comparing %d Bybit pairs with %d Binance pairs", len(bybitPairs), len(binancePairs))

	opportunitiesFound := 0
	pairsCompared := 0

	for symbol, bybitPrice := range bybitPairs {
		binancePrice, exists := binancePairs[symbol]
		if !exists {
			continue
		}

		pairsCompared++

		// Check for zero prices
		if bybitPrice.AskPrice.IsZero() || bybitPrice.BidPrice.IsZero() ||
			binancePrice.AskPrice.IsZero() || binancePrice.BidPrice.IsZero() {
			continue
		}

		// Check Bybit buy, Binance sell
		bybitBuyPrice := bybitPrice.AskPrice.Mul(decimal.NewFromFloat(1 + transactionFee))
		binanceSellPrice := binancePrice.BidPrice.Mul(decimal.NewFromFloat(1 - transactionFee))

		if bybitBuyPrice.IsPositive() {
			profitPercentage := binanceSellPrice.Sub(bybitBuyPrice).Div(bybitBuyPrice)

			if profitPercentage.GreaterThanOrEqual(decimal.NewFromFloat(minProfitPercentage)) {
				fmt.Printf("Arbitrage opportunity found for %s:\n", symbol)
				fmt.Printf("  Buy from Bybit at %s\n", bybitBuyPrice.StringFixed(8))
				fmt.Printf("  Sell on Binance at %s\n", binanceSellPrice.StringFixed(8))
				fmt.Printf("  Profit percentage: %s%%\n\n", profitPercentage.Mul(decimal.NewFromInt(100)).StringFixed(2))
				opportunitiesFound++
			}
		}

		// Check Binance buy, Bybit sell
		binanceBuyPrice := binancePrice.AskPrice.Mul(decimal.NewFromFloat(1 + transactionFee))
		bybitSellPrice := bybitPrice.BidPrice.Mul(decimal.NewFromFloat(1 - transactionFee))

		if binanceBuyPrice.IsPositive() {
			profitPercentage := bybitSellPrice.Sub(binanceBuyPrice).Div(binanceBuyPrice)

			if profitPercentage.GreaterThanOrEqual(decimal.NewFromFloat(minProfitPercentage)) {
				fmt.Printf("Arbitrage opportunity found for %s:\n", symbol)
				fmt.Printf("  Buy from Binance at %s\n", binanceBuyPrice.StringFixed(8))
				fmt.Printf("  Sell on Bybit at %s\n", bybitSellPrice.StringFixed(8))
				fmt.Printf("  Profit percentage: %s%%\n\n", profitPercentage.Mul(decimal.NewFromInt(100)).StringFixed(2))
				opportunitiesFound++
			}
		}
	}

	log.Printf("Compared %d pairs", pairsCompared)
	log.Printf("Found %d arbitrage opportunities", opportunitiesFound)

	if opportunitiesFound == 0 {
		log.Println("No arbitrage opportunities found meeting the 2% profit threshold.")
		// Print a few sample comparisons for debugging
		count := 0
		for symbol, bybitPrice := range bybitPairs {
			if binancePrice, exists := binancePairs[symbol]; exists {
				fmt.Printf("Sample comparison for %s:\n", symbol)
				fmt.Printf("  Bybit  - Bid: %s, Ask: %s\n", bybitPrice.BidPrice.StringFixed(8), bybitPrice.AskPrice.StringFixed(8))
				fmt.Printf("  Binance - Bid: %s, Ask: %s\n", binancePrice.BidPrice.StringFixed(8), binancePrice.AskPrice.StringFixed(8))
				count++
				if count >= 5 {
					break
				}
			}
		}
	}
}

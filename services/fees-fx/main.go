package main

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type MarketFee struct {
	BuyBps  int     `yaml:"buy_bps"`
	SellBps int     `yaml:"sell_bps"`
	Fixed   float64 `yaml:"fixed"`
}
type FeeConfig struct {
	Markets map[string]MarketFee `yaml:"markets"`
}

var cfg *FeeConfig

type NetSpreadParams struct {
	BuyPrice  float64
	SellPrice float64
	BuyFee    MarketFee
	SellFee   MarketFee
	FXRate    float64
}

func computeNetSpread(p NetSpreadParams) (netBuy, netSell, spread float64) {
	netBuy = p.BuyPrice*(1+float64(p.BuyFee.BuyBps)/10000.0) + p.BuyFee.Fixed
	sellPrice := p.SellPrice * p.FXRate
	netSell = sellPrice*(1-float64(p.SellFee.SellBps)/10000.0) - p.SellFee.Fixed
	spread = (netSell - netBuy) / netBuy * 10000
	return
}

func main() {
	path := os.Getenv("FEES_CONFIG")
	if path == "" {
		log.Fatal("FEES_CONFIG is not set")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read fees config: %v", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("cannot parse fees config: %v", err)
	}

	http.HandleFunc("/compute", computeHandler)

	addr := os.Getenv("FEESFX_ADDR")
	if addr == "" {
		addr = ":8090"
	}
	log.Printf("fees-fx listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func computeHandler(w http.ResponseWriter, r *http.Request) {
	buyPrice, _ := strconv.ParseFloat(r.URL.Query().Get("buy_price"), 64)
	sellPrice, _ := strconv.ParseFloat(r.URL.Query().Get("sell_price"), 64)
	buyMarket := r.URL.Query().Get("buy_market")
	sellMarket := r.URL.Query().Get("sell_market")
	fxRateStr := r.URL.Query().Get("fx_rate")
	fxRate := 1.0
	if fxRateStr != "" {
		fxRate, _ = strconv.ParseFloat(fxRateStr, 64)
	}

	params := NetSpreadParams{
		BuyPrice:  buyPrice,
		SellPrice: sellPrice,
		BuyFee:    cfg.Markets[buyMarket],
		SellFee:   cfg.Markets[sellMarket],
		FXRate:    fxRate,
	}

	netBuy, netSell, spread := computeNetSpread(params)

	resp := map[string]any{
		"net_buy":    netBuy,
		"net_sell":   netSell,
		"spread_bps": spread,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

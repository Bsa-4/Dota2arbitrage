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

	buyFee := cfg.Markets[buyMarket]
	sellFee := cfg.Markets[sellMarket]

	netBuy := buyPrice*(1+float64(buyFee.BuyBps)/10000.0) + buyFee.Fixed
	netSell := sellPrice*(1-float64(sellFee.SellBps)/10000.0) - sellFee.Fixed

	spread := (netSell - netBuy) / netBuy * 10000 // в bps

	resp := map[string]any{
		"net_buy":    netBuy,
		"net_sell":   netSell,
		"spread_bps": spread,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

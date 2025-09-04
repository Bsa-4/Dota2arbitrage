package main

import (
	"math"
	"testing"
)

func TestComputeNetSpread(t *testing.T) {
	tests := []struct {
		name    string
		params  NetSpreadParams
		netBuy  float64
		netSell float64
		spread  float64
	}{
		{
			name: "no fees",
			params: NetSpreadParams{
				BuyPrice:  100,
				SellPrice: 110,
				FXRate:    1,
			},
			netBuy:  100,
			netSell: 110,
			spread:  1000,
		},
		{
			name: "with fees",
			params: NetSpreadParams{
				BuyPrice:  100,
				SellPrice: 110,
				BuyFee:    MarketFee{BuyBps: 10, Fixed: 1},
				SellFee:   MarketFee{SellBps: 20, Fixed: 2},
				FXRate:    1,
			},
			netBuy:  101.1,
			netSell: 107.78,
			spread:  660.7319485657771,
		},
		{
			name: "fx conversion",
			params: NetSpreadParams{
				BuyPrice:  100,
				SellPrice: 100,
				BuyFee:    MarketFee{BuyBps: 100, Fixed: 1},
				SellFee:   MarketFee{SellBps: 50, Fixed: 2},
				FXRate:    1.1,
			},
			netBuy:  102,
			netSell: 107.45,
			spread:  534.3137254901977,
		},
	}

	const eps = 1e-6
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb, ns, sp := computeNetSpread(tt.params)
			if math.Abs(nb-tt.netBuy) > eps {
				t.Fatalf("net buy got %f want %f", nb, tt.netBuy)
			}
			if math.Abs(ns-tt.netSell) > eps {
				t.Fatalf("net sell got %f want %f", ns, tt.netSell)
			}
			if math.Abs(sp-tt.spread) > eps {
				t.Fatalf("spread got %f want %f", sp, tt.spread)
			}
		})
	}
}

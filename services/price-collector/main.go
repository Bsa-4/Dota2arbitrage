package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type PriceEvent struct {
	Market   string    `json:"market"`             // "lisskins" | "buff" | "market_dota2"
	ItemKey  string    `json:"item_key"`           // нормализованное имя предмета
	Price    float64   `json:"price"`              // цена в валюте источника
	Currency string    `json:"currency"`           // "USD"/"CNY"/"RUB"...
	Ts       time.Time `json:"ts"`                 // время измерения
	Meta     any       `json:"meta,omitempty"`     // опционально: сырой payload
}

func main() {
	ctx := context.Background()

	brokers := getenv("KAFKA_BROKERS", "kafka:9092")
	topic := getenv("PRICES_TOPIC", "prices.raw")
	pollMs := atoi(getenv("POLL_MS", "1500"))
	sources := strings.Split(getenv("SOURCES", "lisskins,buff,market_dota2"), ",")

	w := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // ключом будет item_key
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	defer w.Close()

	log.Printf("collector starting: brokers=%s topic=%s sources=%v pollMs=%d", brokers, topic, sources, pollMs)

	t := time.NewTicker(time.Duration(pollMs) * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			var events []PriceEvent

			// Заглушки: здесь будут реальные вызовы API
			if contains(sources, "lisskins") {
				evs, _ := fetchLisSkins(ctx)
				events = append(events, evs...)
			}
			if contains(sources, "buff") {
				evs, _ := fetchBuff(ctx)
				events = append(events, evs...)
			}
			if contains(sources, "market_dota2") {
				evs, _ := fetchMarketDota2(ctx)
				events = append(events, evs...)
			}

			if len(events) == 0 {
				continue
			}

			msgs := make([]kafka.Message, 0, len(events))
			for _, e := range events {
				b, _ := json.Marshal(e)
				msgs = append(msgs, kafka.Message{
					Key:   []byte(e.ItemKey), // ключ партиционирования
					Value: b,
					Time:  time.Now(),
					Headers: []kafka.Header{
						{Key: "market", Value: []byte(e.Market)},
						{Key: "currency", Value: []byte(e.Currency)},
					},
				})
			}

			if err := w.WriteMessages(ctx, msgs...); err != nil {
				log.Printf("kafka write error: %v", err)
			} else {
				log.Printf("sent %d price events", len(msgs))
			}
		}
	}
}

// ----- Ниже заглушки источников (позже заменим на реальные HTTP-клиенты) -----

var demoItems = []string{
	"Unusual Voidhammer",          // примеры
	"Inscribed Demon Eater",
	"Dragonclaw Hook",
}

func fetchLisSkins(ctx context.Context) ([]PriceEvent, error) {
	// TODO: заменить на вызовы LisSkins_API (Python sidecar либо прямой HTTP)
	// для MVP эмулируем цены
	out := make([]PriceEvent, 0, len(demoItems))
	for _, it := range demoItems {
		out = append(out, PriceEvent{
			Market:   "lisskins",
			ItemKey:  it,
			Price:    50 + rand.Float64()*20,
			Currency: "USD",
			Ts:       time.Now(),
		})
	}
	return out, nil
}

func fetchBuff(ctx context.Context) ([]PriceEvent, error) {
	// TODO: заменить на buff.163.com (Python обёртка/клиент)
	out := make([]PriceEvent, 0, len(demoItems))
	for _, it := range demoItems {
		out = append(out, PriceEvent{
			Market:   "buff",
			ItemKey:  it,
			Price:    320 + rand.Float64()*50,
			Currency: "CNY",
			Ts:       time.Now(),
		})
	}
	return out, nil
}

func fetchMarketDota2(ctx context.Context) ([]PriceEvent, error) {
	// TODO: заменить на market.dota2.net HTTP клиент
	out := make([]PriceEvent, 0, len(demoItems))
	for _, it := range demoItems {
		out = append(out, PriceEvent{
			Market:   "market_dota2",
			ItemKey:  it,
			Price:    4100 + rand.Float64()*900,
			Currency: "RUB",
			Ts:       time.Now(),
		})
	}
	return out, nil
}

// ----- утилиты -----

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if strings.TrimSpace(x) == s {
			return true
		}
	}
	return false
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// для будущих реальных вызовов
var ErrRateLimited = errors.New("rate limited")

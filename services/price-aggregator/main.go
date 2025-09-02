package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type PriceEvent struct {
	Market   string    `json:"market"`
	ItemKey  string    `json:"item_key"`
	Price    float64   `json:"price"`
	Currency string    `json:"currency"`
	Ts       time.Time `json:"ts"`
}

type Quote struct {
	BestBid  float64   `json:"best_bid"`
	BestAsk  float64   `json:"best_ask"`
	Ema      float64   `json:"ema"`
	Currency string    `json:"currency"`
	Ts       time.Time `json:"ts"`
}

var (
	emaMap   = make(map[string]float64) // item_key → EMA
	alpha    = 0.2                      // сглаживание
	redisCli *redis.Client
	chConn   clickhouse.Conn
)

func main() {
	ctx := context.Background()

	// Redis
	redisCli = redis.NewClient(&redis.Options{
		Addr: getenv("REDIS_ADDR", "redis:6379"),
	})
	defer redisCli.Close()

	// ClickHouse
	// заменяем то место, где было: clickhouse.Open(&clickhouse.Options{ Addr: ... })
	ch, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{getenv("CLICKHOUSE_ADDR", "clickhouse:9000")},
		Auth: clickhouse.Auth{
			Database: getenv("CLICKHOUSE_DB", "arbch"),      // <-- твоя БД
			Username: getenv("CLICKHOUSE_USER", "chuser"),   // <-- твой пользователь
			Password: getenv("CLICKHOUSE_PASSWORD", "chpass"),// <-- твой пароль
		},
		// Можно добавить: Protocol: clickhouse.Native,  // по умолчанию Native на 9000
	})
	if err != nil {
		log.Fatalf("clickhouse open: %v", err)
	}
	chConn = ch
	if err := ensureTable(ctx, chConn); err != nil {
		log.Fatalf("clickhouse ensure table: %v", err)
	}
	defer chConn.Close()

	// Kafka consumer
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(getenv("KAFKA_BROKERS", "kafka:9092"), ","),
		Topic:   getenv("PRICES_TOPIC", "prices.raw"),
		GroupID: "price-aggregator",
	})
	defer r.Close()

	log.Println("aggregator started")
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("kafka read err: %v", err)
			continue
		}
		var ev PriceEvent
		if err := json.Unmarshal(m.Value, &ev); err != nil {
			log.Printf("json err: %v", err)
			continue
		}
		process(ctx, ev)
	}
}

func process(ctx context.Context, ev PriceEvent) {
	// EMA
	key := ev.ItemKey
	prev := emaMap[key]
	if prev == 0 {
		prev = ev.Price
	}
	ema := alpha*ev.Price + (1-alpha)*prev
	emaMap[key] = ema

	q := Quote{
		BestBid:  ev.Price, // упрощённо: пока берём одно значение
		BestAsk:  ev.Price,
		Ema:      ema,
		Currency: ev.Currency,
		Ts:       ev.Ts,
	}

	// Redis
	b, _ := json.Marshal(q)
	if err := redisCli.Set(ctx, "quotes:"+key, b, 5*time.Minute).Err(); err != nil {
		log.Printf("redis err: %v", err)
	}

	// ClickHouse
	if err := chConn.Exec(ctx, "INSERT INTO quotes (item_key,market,price,currency,ts) VALUES (?,?,?,?,?)",
		ev.ItemKey, ev.Market, ev.Price, ev.Currency, ev.Ts); err != nil {
		log.Printf("clickhouse err: %v", err)
	}
}

func ensureTable(ctx context.Context, conn clickhouse.Conn) error {
    // Таблица создаётся в выбранной DB (из Auth.Database)
    ddl := `
CREATE TABLE IF NOT EXISTS quotes
(
  item_key String,
  market String,
  price Float64,
  currency String,
  ts DateTime
) ENGINE = MergeTree()
ORDER BY (item_key, ts)
`
    return conn.Exec(ctx, ddl)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

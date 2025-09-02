package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"cs2arb/gateway/store"
)

var redisCli *redis.Client

func main() {
	ctx := context.Background()

	// --- Redis (строго из ENV) ---
	redisAddr := mustEnv("REDIS_ADDR") // например: redis:6379
	redisCli = redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisCli.Close()

	// --- Postgres (строго из ENV) ---
	dsn := mustEnv("DB_DSN") // например: postgres://user:pass@postgres:5432/db?sslmode=disable
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()
	q := store.New(pool)

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Markets
	r.Get("/api/markets", func(w http.ResponseWriter, _ *http.Request) {
		rows, err := q.ListMarkets(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows, http.StatusOK)
	})

	// Bots
	r.Get("/api/bots", func(w http.ResponseWriter, _ *http.Request) {
		rows, err := q.ListBots(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows, http.StatusOK)
	})

	// Quotes (из Redis)
	r.Get("/api/quotes/{item_key}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "item_key")
		val, err := redisCli.Get(ctx, "quotes:"+key).Result()
		if err == redis.Nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(val))
	})

	addr := mustEnv("GATEWAY_ADDR") // например: :8088
	log.Printf("gateway listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is not set", key)
	}
	return v
}

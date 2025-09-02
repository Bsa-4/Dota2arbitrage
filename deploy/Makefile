COMPOSE = docker compose
DC_FILE = -f docker-compose.yml

.PHONY: up down restart ps logs clean proto

up:
	$(COMPOSE) $(DC_FILE) up -d

down:
	$(COMPOSE) $(DC_FILE) down

restart:
	$(COMPOSE) $(DC_FILE) down
	$(COMPOSE) $(DC_FILE) up -d

ps:
	$(COMPOSE) $(DC_FILE) ps

logs:
	$(COMPOSE) $(DC_FILE) logs -f --tail=200

clean:
	$(COMPOSE) $(DC_FILE) down -v

proto:
	@echo "Place your .proto files under ../proto and run buf/protoc here."
	@echo "Example:"
	@echo "  buf generate ../proto"

# Абсолютные пути (Windows-friendly)
ROOT_DIR := $(abspath ..)
MIGRATIONS_DIR := $(abspath ../db/migrations)

# DSN для подключения ИЗ контейнера (hostname "postgres" из docker-compose)
DB_DSN_IN_DOCKER ?= postgres://arb:arbpass@postgres:5432/arbdb?sslmode=disable

.PHONY: migrate-up migrate-down sqlc sqlc-docker up-gateway down-gateway logs-gateway

migrate-up:
	@echo "Running migrate from $(MIGRATIONS_DIR)"
	docker run --rm \
		--network=cs2-net \
		-v "$(MIGRATIONS_DIR):/migrations:ro" \
		migrate/migrate:4 \
		-source=file:///migrations \
		-database "$(DB_DSN_IN_DOCKER)" up

migrate-down:
	docker run --rm \
		--network=cs2-net \
		-v "$(MIGRATIONS_DIR):/migrations:ro" \
		migrate/migrate:4 \
		-source=file:///migrations \
		-database "$(DB_DSN_IN_DOCKER)" down

# <<< Главное: sqlc в Docker >>>
# Абсолютный путь к корню репо (Windows-friendly)
ROOT_DIR := $(abspath ..)

.PHONY: sqlc sqlc-docker

sqlc-docker:
	@echo "Generating sqlc code via Docker (no local install needed)"
	docker run --rm \
		-v "$(ROOT_DIR):/work" \
		sqlc/sqlc:1.27.0 \
		-f /work/sqlc.yaml generate

sqlc: sqlc-docker

up-gateway:
	docker compose up -d gateway

down-gateway:
	docker compose down gateway

logs-gateway:
	docker compose logs -f gateway

.PHONY: up-collector logs-collector

up-collector:
	docker compose up -d price-collector

logs-collector:
	docker compose logs -f price-collector

.PHONY: up-aggregator logs-aggregator

up-aggregator:
	docker compose up -d price-aggregator

logs-aggregator:
	docker compose logs -f price-aggregator
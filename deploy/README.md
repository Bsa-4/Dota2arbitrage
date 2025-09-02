# Dev stack
Setup env: Copy-Item deploy/.env.example deploy/.env -Force
Start stack: docker compose -f deploy/docker-compose.yml up -d
Show services: docker compose -f deploy/docker-compose.yml ps
Stop stack: docker compose -f deploy/docker-compose.yml down
Clean volumes: docker compose -f deploy/docker-compose.yml down -v


## 3) Проверка UI/портов
- RabbitMQ: http://localhost:15672 (guest/guest)
- Kafka UI: http://localhost:8080
- Temporal UI: http://localhost:8233
- ClickHouse: http://localhost:8123

## 4) Остановка
make down

## 5) Полная очистка
make clean

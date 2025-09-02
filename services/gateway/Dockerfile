# Build
FROM golang:1.22 AS build
WORKDIR /src

# Копируем весь репозиторий (у нас есть сгенерированный store)
COPY . .

# Сборка
WORKDIR /src/services/gateway
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/gateway .

# Runtime (минимальный образ)
FROM gcr.io/distroless/base-debian12
ENV GATEWAY_ADDR=:8088
# Подключение к Postgres через compose hostname "postgres"
ENV DB_DSN=postgres://arb:arbpass@postgres:5432/arbdb?sslmode=disable
COPY --from=build /bin/gateway /bin/gateway
EXPOSE 8088
USER nonroot:nonroot
ENTRYPOINT ["/bin/gateway"]

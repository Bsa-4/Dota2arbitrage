FROM golang:1.22 AS build
WORKDIR /src
COPY . .
WORKDIR /src/services/price-aggregator
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/aggregator .

FROM gcr.io/distroless/base-debian12
ENV KAFKA_BROKERS=kafka:9092
ENV PRICES_TOPIC=prices.raw
ENV REDIS_ADDR=redis:6379
ENV CLICKHOUSE_DSN=clickhouse://default:@clickhouse:9000/default
COPY --from=build /bin/aggregator /bin/aggregator
ENTRYPOINT ["/bin/aggregator"]

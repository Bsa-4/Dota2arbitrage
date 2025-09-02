# Build
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
WORKDIR /src/services/price-collector
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/collector .

# Runtime
FROM gcr.io/distroless/base-debian12
ENV KAFKA_BROKERS=kafka:9092
ENV PRICES_TOPIC=prices.raw
ENV POLL_MS=1500
ENV SOURCES=lisskins,buff,market_dota2
COPY --from=build /bin/collector /bin/collector
ENTRYPOINT ["/bin/collector"]

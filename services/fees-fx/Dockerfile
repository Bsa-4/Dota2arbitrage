FROM golang:1.22 AS build
WORKDIR /src
COPY . .
WORKDIR /src/services/fees-fx
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/feesfx .

FROM gcr.io/distroless/base-debian12
ENV FEESFX_ADDR=:8090
COPY --from=build /bin/feesfx /bin/feesfx
ENTRYPOINT ["/bin/feesfx"]

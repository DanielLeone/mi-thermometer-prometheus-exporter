FROM golang:1.22.3 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /mi-thermometer-prometheus-exporter

FROM scratch
COPY --from=builder /mi-thermometer-prometheus-exporter /mi-thermometer-prometheus-exporter
EXPOSE 9000
ENTRYPOINT ["/mi-thermometer-prometheus-exporter"]

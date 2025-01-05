FROM golang:1.21 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

RUN GOOS=linux go build -ldflags="-s -w" -o /mi-thermometer-prometheus-exporter

EXPOSE 9000

CMD ["/mi-thermometer-prometheus-exporter"]

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /order-service ./cmd

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir -p /app/cache_data

COPY --from=builder /order-service /app/order-service

COPY static /app/static

WORKDIR /app

EXPOSE 8080

CMD ["./order-service"]

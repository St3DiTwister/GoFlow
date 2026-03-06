FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/worker-service ./cmd/worker/main.go

FROM alpine:latest
WORKDIR /

COPY --from=builder /app/worker-service /worker-service

ENTRYPOINT ["/worker-service"]
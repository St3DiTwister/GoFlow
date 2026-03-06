FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/admin-service ./cmd/admin/main.go

FROM alpine:latest
WORKDIR /

COPY --from=builder /app/admin-service /admin-service

ENTRYPOINT ["/admin-service"]
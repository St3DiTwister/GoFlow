package main

import (
	"GoFlow/internal/handler"
	"GoFlow/internal/kafka"
	"GoFlow/internal/limiter"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	ctx := context.Background()
	rateLimiter := limiter.NewIPRateLimiter(ctx, 5.0, 20)

	godotenv.Load()
	kafkaAddr := os.Getenv("KAFKA_BROKER")
	if kafkaAddr == "" {
		kafkaAddr = "kafka:29092"
	}

	prod := kafka.NewProducer(kafkaAddr, "telemetry")
	defer prod.Close()

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	eventHandler := &handler.EventHandler{
		Producer: prod,
	}
	r.POST("/track", handler.RateLimitMiddleware(rateLimiter), eventHandler.Track)

	r.Run(":8080")
}

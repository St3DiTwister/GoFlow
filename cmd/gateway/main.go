package main

import (
	"GoFlow/internal/handler"
	"GoFlow/internal/kafka"
	"GoFlow/internal/limiter"
	"context"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()
	rateLimiter := limiter.NewIPRateLimiter(ctx, 5.0, 20)

	prod := kafka.NewProducer("localhost:9092", "telemetry")
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

package main

import (
	"GoFlow/internal/handler"
	"GoFlow/internal/limiter"
	"context"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()
	rateLimiter := limiter.NewIPRateLimiter(ctx, 5.0, 20)
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	eventHandler := &handler.EventHandler{}
	r.POST("/track", handler.RateLimitMiddleware(rateLimiter), eventHandler.Track)

	r.Run(":8080")
}

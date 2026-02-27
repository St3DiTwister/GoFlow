package handler

import (
	"GoFlow/internal/limiter"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RateLimitMiddleware(rateLimiter *limiter.IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		lim := rateLimiter.GetLimiter(ip)
		if !lim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// Set example variable
		c.Set("example", "12345")

		// before request

		c.Next()

		// after request
		latency := time.Since(t)
		status := c.Writer.Status()

		log.Printf("Path: %s | Status: %d | Latency: %v", c.Request.URL.Path, status, latency)
	}
}

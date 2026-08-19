package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that adds CORS headers.
// In production, CORS is handled by the CDN/tunnel.
func CORS() gin.HandlerFunc {
	devMode := os.Getenv("DIVOENE_DEV_MODE") == "true"

	return func(c *gin.Context) {
		if devMode {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Actor-UID, X-Actor-Roles, X-Actor-Name")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}

		c.Next()
	}
}

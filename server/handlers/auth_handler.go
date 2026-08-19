package handlers

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers the /api/auth routes.
func RegisterAuthRoutes(rg *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	rg.POST("/login", authMiddleware.LoginHandler)
	rg.GET("/refresh_token", authMiddleware.RefreshHandler)
}

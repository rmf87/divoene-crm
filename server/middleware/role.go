package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
)

// RequireRole returns middleware that allows only the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, exists := c.Get("actor")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no actor in context"})
			return
		}
		a, ok := actor.(*domain.Actor)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid actor type"})
			return
		}
		for _, r := range roles {
			if a.HasRole(r) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// ManagerOnly allows only manager role.
func ManagerOnly() gin.HandlerFunc {
	return RequireRole("manager")
}

// SellerOrManager allows seller and manager roles.
func SellerOrManager() gin.HandlerFunc {
	return RequireRole("seller", "manager")
}

// GuideOrManager allows guide and manager roles.
func GuideOrManager() gin.HandlerFunc {
	return RequireRole("guide", "manager")
}

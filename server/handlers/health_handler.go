package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// HealthHandler provides the health check endpoint.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// GetHealth returns the service health status.
func (h *HealthHandler) GetHealth(c *gin.Context) {
	status := "ok"
	dbOK := true
	if err := h.db.Ping(); err != nil {
		status = "degraded"
		dbOK = false
	}
	c.JSON(200, gin.H{
		"status":   status,
		"database": dbOK,
	})
}

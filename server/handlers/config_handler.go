package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// RegisterConfigRoutes registers manager-only config routes.
// reloadFn is called after a config value changes so clients can re-read their settings.
func RegisterConfigRoutes(rg *gin.RouterGroup, svc *services.ConfigService, reloadFn func(key string)) {
	h := &configHandler{svc: svc, reloadFn: reloadFn}
	rg.GET("", h.listConfigs)
	rg.GET("/:key", h.getConfig)
	rg.PUT("/:key", h.updateConfig)
}

type configHandler struct {
	svc      *services.ConfigService
	reloadFn func(key string)
}

func (h *configHandler) listConfigs(c *gin.Context) {
	settings, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if settings == nil {
		settings = []domain.ConfigSetting{}
	}
	c.JSON(http.StatusOK, settings)
}

func (h *configHandler) getConfig(c *gin.Context) {
	key := c.Param("key")
	setting, err := h.svc.GetByKey(c.Request.Context(), key)
	if err != nil {
		if strings.Contains(err.Error(), "config not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Mask secret values
	if setting.ValueType == "secret" && setting.Value != "" {
		setting.MaskedValue = maskValue(setting.Value)
		setting.Value = ""
	}
	c.JSON(http.StatusOK, setting)
}

type updateConfigReq struct {
	Value string `json:"value"`
}

func (h *configHandler) updateConfig(c *gin.Context) {
	key := c.Param("key")
	var req updateConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	updatedBy := ""
	actor, ok := getActor(c)
	if ok {
		updatedBy = actor.Name
	}

	setting, err := h.svc.Upsert(c.Request.Context(), key, req.Value, updatedBy)
	if err != nil {
		if strings.Contains(err.Error(), "config not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Notify clients to reload their config
	if h.reloadFn != nil {
		h.reloadFn(key)
	}

	c.JSON(http.StatusOK, setting)
}

// maskValue shows first 3 and last 2 characters of s.
func maskValue(s string) string {
	if len(s) <= 5 {
		return strings.Repeat("*", len(s))
	}
	return s[:3] + strings.Repeat("*", len(s)-5) + s[len(s)-2:]
}

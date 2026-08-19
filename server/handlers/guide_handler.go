package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// RegisterGuideRoutes registers protected /api/guides routes.
func RegisterGuideRoutes(rg *gin.RouterGroup, svc *services.VisitService) {
	h := &guideHandler{svc: svc}
	rg.GET("", h.list)
	rg.POST("", h.create)
	rg.GET("/:id", h.get)
	rg.PATCH("/:id", h.update)
	rg.GET("/:id/availability", h.getAvailability)
	rg.PATCH("/:id/availability", h.updateAvailability)
}

type guideHandler struct {
	svc *services.VisitService
}

func (h *guideHandler) list(c *gin.Context) {
	guides, err := h.svc.ListGuides(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, guides)
}

func (h *guideHandler) create(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	guide := &domain.Guide{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Active:           true,
		MaxPerSlot:       3,
		WeeklySchedule:   map[string][]string{},
		UnavailableDates: []domain.DateRange{},
	}

	if err := h.svc.CreateGuide(c.Request.Context(), guide); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, guide)
}

func (h *guideHandler) get(c *gin.Context) {
	id := c.Param("id")
	guide, err := h.svc.GetGuide(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, guide)
}

func (h *guideHandler) update(c *gin.Context) {
	id := c.Param("id")
	var guide domain.Guide
	if err := c.ShouldBindJSON(&guide); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if err := h.svc.UpdateGuide(c.Request.Context(), id, &guide); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *guideHandler) getAvailability(c *gin.Context) {
	id := c.Param("id")
	guide, err := h.svc.GetGuide(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"weekly_schedule":   guide.WeeklySchedule,
		"unavailable_dates": guide.UnavailableDates,
		"max_per_slot":      guide.MaxPerSlot,
	})
}

func (h *guideHandler) updateAvailability(c *gin.Context) {
	id := c.Param("id")
	var req domain.UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if err := h.svc.UpdateGuideAvailability(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

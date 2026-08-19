package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// RegisterVisitRoutes registers protected /api/visits routes.
func RegisterVisitRoutes(rg *gin.RouterGroup, svc *services.VisitService) {
	h := &visitHandler{svc: svc}
	rg.POST("", h.create)
	rg.GET("", h.list)
	rg.GET("/slots", h.listSlots)
	rg.GET("/:id", h.get)
	rg.PATCH("/:id", h.update)
}

type visitHandler struct {
	svc *services.VisitService
}

func (h *visitHandler) create(c *gin.Context) {
	actor, _ := getActor(c)
	var req domain.CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	visit, err := h.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		code := http.StatusInternalServerError
		errMsg := err.Error()
		if isValidationError(errMsg) {
			code = http.StatusBadRequest
		}
		c.JSON(code, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusCreated, visit)
}

func (h *visitHandler) list(c *gin.Context) {
	actor, _ := getActor(c)
	guideID := c.Query("guide_id")
	leadID := c.Query("lead_id")

	var visits []*domain.Visit
	var err error

	switch {
	case guideID != "":
		visits, err = h.svc.ListByGuide(c.Request.Context(), guideID)
	case leadID != "":
		visits, err = h.svc.ListByLead(c.Request.Context(), leadID)
	default:
		visits, err = h.svc.ListAll(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visits)
	_ = actor // role-based filtering already in service
}

func (h *visitHandler) get(c *gin.Context) {
	actor, _ := getActor(c)
	id := c.Param("id")
	visit, err := h.svc.Get(c.Request.Context(), actor, id)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrada") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visit)
}

func (h *visitHandler) update(c *gin.Context) {
	actor, _ := getActor(c)
	id := c.Param("id")

	var req domain.UpdateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	visit, err := h.svc.Update(c.Request.Context(), actor, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visit)
}

func (h *visitHandler) listSlots(c *gin.Context) {
	date := c.Query("date")
	week := c.Query("week")

	var slots []*domain.GuideSlot
	var err error

	if week != "" {
		year, weekNum, parseErr := parseISOWeek(week)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "formato de semana inválido (use YYYY-Www)"})
			return
		}
		slots, err = h.svc.ListSlotsForWeek(c.Request.Context(), year, weekNum)
	} else if date != "" {
		slots, err = h.svc.ListSlots(c.Request.Context(), date)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parâmetro 'date' ou 'week' é obrigatório"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, slots)
}

func isValidationError(msg string) bool {
	return strings.Contains(msg, "obrigatório") || strings.Contains(msg, "inválido") ||
		strings.Contains(msg, "não pode") || strings.Contains(msg, "já registrado")
}

func parseISOWeek(weekStr string) (int, int, error) {
	// Accept "2026-W26" format
	weekStr = strings.TrimPrefix(weekStr, "W")
	parts := strings.SplitN(weekStr, "-W", 2)
	if len(parts) == 2 {
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		week, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return year, week, nil
	}
	// Try "2026-26" format
	parts = strings.SplitN(weekStr, "-", 2)
	if len(parts) == 2 {
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		week, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
		return year, week, nil
	}
	return 0, 0, nil
}

package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// RegisterLeadRoutes registers public + protected lead routes.
func RegisterLeadRoutes(publicRG *gin.RouterGroup, protectedRG *gin.RouterGroup, svc *services.LeadService) {
	h := &leadHandler{svc: svc}
	if publicRG != nil {
		publicRG.POST("/leads", h.createLead)
	}
	if protectedRG != nil {
		protectedRG.GET("", h.listLeads)
		protectedRG.GET("/:id", h.getLead)
		protectedRG.PATCH("/:id", h.updateStage)
		protectedRG.PATCH("/:id/validation", h.updateValidation)
		protectedRG.POST("/:id/notes", h.addNote)
	}
}

type leadHandler struct {
	svc *services.LeadService
}

func (h *leadHandler) createLead(c *gin.Context) {
	var req domain.CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.WhatsApp = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, req.WhatsApp)
	req.Product = strings.TrimSpace(req.Product)
	req.Source = strings.TrimSpace(req.Source)

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nome é obrigatório"})
		return
	}
	if len(req.WhatsApp) < 10 || len(req.WhatsApp) > 13 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "whatsapp inválido"})
		return
	}
	if req.Product == "" || !domain.IsValidProduct(req.Product) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "produto inválido"})
		return
	}
	if req.Source == "" {
		req.Source = "site"
	}

	lead, err := h.svc.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, lead)
}

func (h *leadHandler) addNote(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id é obrigatório"})
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	createdBy := "system"
	if actor, ok := getActor(c); ok {
		createdBy = actor.Name
	}

	lead, err := h.svc.AddNote(c.Request.Context(), id, req.Text, createdBy)
	if err != nil {
		if err.Error() == "nota vazia" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lead)
}

func (h *leadHandler) listLeads(c *gin.Context) {
	leads, err := h.svc.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, leads)
}

func (h *leadHandler) getLead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id é obrigatório"})
		return
	}

	lead, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lead)
}

func (h *leadHandler) updateStage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id é obrigatório"})
		return
	}

	var req domain.UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if req.Stage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stage é obrigatório"})
		return
	}
	if req.ChangedBy == "" {
		req.ChangedBy = "system"
	}

	lead, err := h.svc.UpdateStage(c.Request.Context(), id, req.Stage, req.ChangedBy)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "transição inválida") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lead)
}

func (h *leadHandler) updateValidation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id é obrigatório"})
		return
	}
	// Remove trailing /validation if present
	id = strings.TrimSuffix(id, "/validation")

	var req domain.UpdateValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if err := h.svc.UpdateValidation(c.Request.Context(), id, req.Event, req.ContactPerson, req.AddOns); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	lead, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lead)
}

package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// RegisterPaymentRoutes registers protected /api/payments routes.
func RegisterPaymentRoutes(rg *gin.RouterGroup, svc *services.PaymentService) {
	h := &paymentHandler{svc: svc}
	rg.POST("", h.create)
	rg.GET("/:id", h.get)
}

type paymentHandler struct {
	svc *services.PaymentService
}

func (h *paymentHandler) create(c *gin.Context) {
	actor, _ := getActor(c)

	var req domain.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	payment, err := h.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		code := http.StatusInternalServerError
		if isValidationError(err.Error()) {
			code = http.StatusBadRequest
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, payment)
}

func (h *paymentHandler) get(c *gin.Context) {
	actor, _ := getActor(c)
	id := c.Param("id")

	payment, err := h.svc.Get(c.Request.Context(), actor, id)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payment)
}

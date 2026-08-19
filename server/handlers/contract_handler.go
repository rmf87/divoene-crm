package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
)

// RegisterContractRoutes registers protected /api/contracts routes.
func RegisterContractRoutes(rg *gin.RouterGroup, svc *services.ContractService) {
	h := &contractHandler{svc: svc}
	rg.POST("", h.create)
	rg.GET("/:id", h.get)
}

type contractHandler struct {
	svc *services.ContractService
}

func getActor(c *gin.Context) (domain.Actor, bool) {
	if v, ok := c.Get("actor"); ok {
		if a, ok := v.(*domain.Actor); ok {
			return *a, true
		}
	}
	return domain.Actor{}, false
}

func (h *contractHandler) create(c *gin.Context) {
	actor, _ := getActor(c)

	var req domain.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	contract, err := h.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		code := http.StatusInternalServerError
		if isValidationError(err.Error()) {
			code = http.StatusBadRequest
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, contract)
}

func (h *contractHandler) get(c *gin.Context) {
	actor, _ := getActor(c)
	id := c.Param("id")

	contract, err := h.svc.Get(c.Request.Context(), actor, id)
	if err != nil {
		if strings.Contains(err.Error(), "não encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contract)
}

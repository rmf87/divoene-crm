package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/services"
)

// UserHandler exposes manager-only user CRUD routes.
type UserHandler struct {
	svc *services.UserService
}

// NewUserHandler creates a UserHandler backed by the user service.
func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterUserRoutes registers manager-only user CRUD routes.
func RegisterUserRoutes(rg *gin.RouterGroup, h *UserHandler) {
	rg.GET("", h.listUsers)
	rg.POST("", h.createUser)
	rg.PATCH("/:id", h.updateUser)
	rg.DELETE("/:id", h.deleteUser)
}

type createUserReq struct {
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
}

func (h *UserHandler) listUsers(c *gin.Context) {
	users, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) createUser(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	user, err := h.svc.Create(c.Request.Context(), req.Email, req.Name, req.Password, req.Roles)
	if err != nil {
		if err.Error() == "invalid roles: at least one of associate, seller, guide, manager" ||
			err.Error() == "email, name, password required" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

type updateUserReq struct {
	Name     *string   `json:"name,omitempty"`
	Roles    *[]string `json:"roles,omitempty"`
	Password *string   `json:"password,omitempty"`
	Active   *bool     `json:"active,omitempty"`
}

func (h *UserHandler) updateUser(c *gin.Context) {
	id := c.Param("id")
	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	user, err := h.svc.Update(c.Request.Context(), id, services.UpdateUserRequest{
		Name:     req.Name,
		Roles:    req.Roles,
		Password: req.Password,
		Active:   req.Active,
	})
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invalid roles: at least one of associate, seller, guide, manager" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) deleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Deactivate(c.Request.Context(), id); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

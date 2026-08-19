package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/internal/core/domain"
)

func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("actor", &domain.Actor{UID: "mgr-1", Roles: []string{"manager"}})

	RequireRole("manager")(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_DeniesNonMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("actor", &domain.Actor{UID: "seller-1", Roles: []string{"seller"}})

	RequireRole("manager")(c)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_MissingActorReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No actor set
	RequireRole("manager")(c)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestManagerOnly_AllowsManager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("actor", &domain.Actor{UID: "mgr-1", Roles: []string{"manager"}})

	ManagerOnly()(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestManagerOnly_DeniesSeller(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("actor", &domain.Actor{UID: "seller-1", Roles: []string{"seller"}})

	ManagerOnly()(c)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireRole_MultipleRolesGrantsAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	actor := &domain.Actor{UID: "dual-1", Roles: []string{"seller", "manager"}}
	c.Set("actor", actor)

	ManagerOnly()(c)
	if w.Code != http.StatusOK {
		t.Errorf("manager gate: expected 200, got %d", w.Code)
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set("actor", actor)
	GuideOrManager()(c2)
	if w2.Code != http.StatusOK {
		t.Errorf("guide gate: expected 200, got %d", w2.Code)
	}
}

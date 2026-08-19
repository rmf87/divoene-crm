package domain

import (
	"context"
	"testing"
)

func TestIsValidProduct(t *testing.T) {
	tests := []struct {
		name    string
		product string
		want    bool
	}{
		{"valid ensaio", "ensaio_fotografico", true},
		{"valid eventos", "locacao_eventos", true},
		{"valid corporativo", "corporativo", true},
		{"valid casamentos", "casamentos", true},
		{"valid buffet", "buffet_infantil", true},
		{"valid passeios", "passeios_escolares", true},
		{"empty", "", false},
		{"unknown", "produto_invalido", false},
		{"mixed case", "Ensaio_Fotografico", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidProduct(tt.product); got != tt.want {
				t.Errorf("IsValidProduct(%q) = %v, want %v", tt.product, got, tt.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{"associate", "associate", true},
		{"seller", "seller", true},
		{"guide", "guide", true},
		{"manager", "manager", true},
		{"empty", "", false},
		{"admin", "admin", false},
		{"unknown", "some_role", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRole(tt.role); got != tt.want {
				t.Errorf("IsValidRole(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestIsValidRoles(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{"single", []string{"seller"}, true},
		{"multiple", []string{"seller", "manager"}, true},
		{"all", []string{"associate", "seller", "guide", "manager"}, true},
		{"empty", nil, false},
		{"invalid one", []string{"seller", "admin"}, false},
		{"invalid only", []string{"admin"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRoles(tt.roles); got != tt.want {
				t.Errorf("IsValidRoles(%v) = %v, want %v", tt.roles, got, tt.want)
			}
		})
	}
}

func TestHasRole(t *testing.T) {
	a := Actor{Roles: []string{"seller", "manager"}}
	if !a.HasRole("manager") {
		t.Error("expected HasRole(manager) true")
	}
	if a.HasRole("guide") {
		t.Error("expected HasRole(guide) false")
	}
	u := User{Roles: []string{"seller", "manager"}}
	if !u.HasRole("manager") {
		t.Error("expected User.HasRole(manager) true")
	}
	if u.HasRole("guide") {
		t.Error("expected User.HasRole(guide) false")
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		want    bool
	}{
		{"lead to validated", "lead", "validated", true},
		{"validated to visit_scheduled", "validated", "visit_scheduled", true},
		{"visit_scheduled to visit_done", "visit_scheduled", "visit_done", true},
		{"visit_done to contract", "visit_done", "contract", true},
		{"contract to paid", "contract", "paid", true},
		{"paid to booked", "paid", "booked", true},
		{"booked to completed", "booked", "completed", true},
		{"lead to cancelled", "lead", "cancelled", true},
		{"completed to paid", "completed", "paid", false},
		{"paid to lead", "paid", "lead", false},
		{"skip stage", "lead", "contract", false},
		{"same stage", "lead", "lead", false},
		{"empty from", "", "lead", false},
		{"empty to", "lead", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransition(tt.from, tt.to); got != tt.want {
				t.Errorf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestFromCancelled(t *testing.T) {
	if got := CanTransition("cancelled", "lead"); got != false {
		t.Errorf("cancelled should be terminal")
	}
	if got := CanTransition("cancelled", "completed"); got != false {
		t.Errorf("cancelled should be terminal")
	}
}

func TestActorWithContext(t *testing.T) {
	ctx := WithActor(context.Background(), Actor{UID: "test-uid", Roles: []string{"manager"}, Name: "Test"})
	actor, ok := ActorFromContext(ctx)
	if !ok {
		t.Error("expected actor in context")
	}
	if actor.UID != "test-uid" {
		t.Errorf("got uid %q, want %q", actor.UID, "test-uid")
	}
	if !actor.HasRole("manager") {
		t.Errorf("got roles %v, want manager", actor.Roles)
	}
}

func TestActorFromContextMissing(t *testing.T) {
	_, ok := ActorFromContext(context.Background())
	if ok {
		t.Error("expected false for context without actor")
	}
}

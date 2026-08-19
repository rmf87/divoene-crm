package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rmf87/divoene/internal/core/domain"
)

type stubUserRepo struct {
	users map[string]*domain.User
	order []string
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: map[string]*domain.User{}}
}

func (r *stubUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("usuário não encontrado")
}

func (r *stubUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, errors.New("usuário não encontrado")
	}
	return u, nil
}

func (r *stubUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, err := r.FindByEmail(ctx, user.Email); err == nil {
		return errors.New("UNIQUE constraint failed: users.email")
	}
	r.users[user.ID] = user
	r.order = append(r.order, user.ID)
	return nil
}

func (r *stubUserRepo) Update(ctx context.Context, user *domain.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return errors.New("usuário não encontrado")
	}
	r.users[user.ID] = user
	return nil
}

func (r *stubUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	var out []*domain.User
	for _, id := range r.order {
		out = append(out, r.users[id])
	}
	return out, nil
}

func TestUserServiceCreate(t *testing.T) {
	svc := NewUserService(newStubUserRepo())

	u, err := svc.Create(context.Background(), "seller@divoene.com.br", "Vendedor", "secret", []string{"seller"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == "" {
		t.Error("expected generated id")
	}
	if u.PasswordHash == "" || u.PasswordHash == "secret" {
		t.Error("password must be hashed")
	}
	if !u.Active {
		t.Error("new user should be active")
	}
}

func TestUserServiceCreateRejectsInvalid(t *testing.T) {
	svc := NewUserService(newStubUserRepo())

	cases := []struct {
		name     string
		email    string
		roles    []string
		wantErr  string
	}{
		{"empty email", "", []string{"seller"}, "email, name, password required"},
		{"no roles", "a@b.com", nil, "invalid roles"},
		{"invalid role", "a@b.com", []string{"ceo"}, "invalid roles"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.email, "N", "pw", tc.roles)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestUserServiceUpdate(t *testing.T) {
	svc := NewUserService(newStubUserRepo())
	u, err := svc.Create(context.Background(), "a@b.com", "A", "pw", []string{"seller"})
	if err != nil {
		t.Fatal(err)
	}

	name := "A2"
	active := false
	updated, err := svc.Update(context.Background(), u.ID, UpdateUserRequest{Name: &name, Active: &active})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "A2" || updated.Active {
		t.Errorf("update did not apply: %+v", updated)
	}

	roles := []string{"manager"}
	if _, err := svc.Update(context.Background(), u.ID, UpdateUserRequest{Roles: &roles}); err != nil {
		t.Fatalf("update roles: %v", err)
	}
	if !updated.HasRole("manager") {
		t.Error("roles not applied")
	}
}

func TestUserServiceDeactivate(t *testing.T) {
	svc := NewUserService(newStubUserRepo())
	u, err := svc.Create(context.Background(), "a@b.com", "A", "pw", []string{"seller"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Deactivate(context.Background(), u.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	got, _ := svc.repo.FindByID(context.Background(), u.ID)
	if got.Active {
		t.Error("user should be inactive")
	}
}

func TestUserServiceNotFound(t *testing.T) {
	svc := NewUserService(newStubUserRepo())
	if _, err := svc.Update(context.Background(), "nope", UpdateUserRequest{}); err == nil || err.Error() != "user not found" {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := svc.Deactivate(context.Background(), "nope"); err == nil || err.Error() != "user not found" {
		t.Fatalf("expected not found, got %v", err)
	}
}

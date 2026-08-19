package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rmf87/divoene/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestUserRepository_Create(t *testing.T) {
	repo := setupUserDB(t)
	ctx := context.Background()

	user := &domain.User{
		ID:     uuid.New().String(),
		Email:  "test@example.com",
		Name:   "Test User",
		Roles:  []string{"seller"},
		Active: true,
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.FindByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got.Name != "Test User" || !got.HasRole("seller") {
		t.Errorf("got %+v, want name=Test User role=seller", got)
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	repo := setupUserDB(t)
	ctx := context.Background()

	user := &domain.User{
		ID: uuid.New().String(), Email: "dup@example.com",
		Name: "First", Roles: []string{"seller"}, Active: true,
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	_ = repo.Create(ctx, user)

	dup := &domain.User{
		ID: uuid.New().String(), Email: "dup@example.com",
		Name: "Second", Roles: []string{"manager"}, Active: true,
	}
	dup.PasswordHash = string(hash)
	if err := repo.Create(ctx, dup); err == nil {
		t.Error("expected error for duplicate email")
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	repo := setupUserDB(t)
	_, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserRepository_List(t *testing.T) {
	repo := setupUserDB(t)
	ctx := context.Background()

	users := []*domain.User{
		{ID: "u1", Email: "a@test.com", Name: "A", Roles: []string{"seller"}, Active: true},
		{ID: "u2", Email: "b@test.com", Name: "B", Roles: []string{"manager"}, Active: true},
	}
	for _, u := range users {
		hash, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
		u.PasswordHash = string(hash)
		_ = repo.Create(ctx, u)
	}

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("got %d users, want 2", len(all))
	}
}

func TestUserRepository_Update(t *testing.T) {
	repo := setupUserDB(t)
	ctx := context.Background()

	user := &domain.User{
		ID: "update-test", Email: "update@test.com",
		Name: "Old Name", Roles: []string{"seller"}, Active: true,
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	_ = repo.Create(ctx, user)

	user.Name = "New Name"
	user.Roles = []string{"manager"}
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByEmail(ctx, "update@test.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got.Name != "New Name" || !got.HasRole("manager") {
		t.Errorf("update failed: got %+v", got)
	}
}

func TestUserRepository_MultipleRoles(t *testing.T) {
	repo := setupUserDB(t)
	ctx := context.Background()

	user := &domain.User{
		ID: uuid.New().String(), Email: "multi@test.com",
		Name: "Multi", Roles: []string{"seller", "manager"}, Active: true,
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.FindByEmail(ctx, "multi@test.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if !got.HasRole("seller") || !got.HasRole("manager") {
		t.Errorf("got roles %v, want seller+manager", got.Roles)
	}

	got.Roles = []string{"guide"}
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}
	refetched, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if len(refetched.Roles) != 1 || !refetched.HasRole("guide") {
		t.Errorf("after update got roles %v, want only guide", refetched.Roles)
	}
}

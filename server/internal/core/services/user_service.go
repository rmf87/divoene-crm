package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rmf87/divoene/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements the business rules for system users.
type UserService struct {
	repo domain.UserRepository
}

// NewUserService wires a UserRepository into a UserService.
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

// Create validates roles, hashes the password and persists a new user.
func (s *UserService) Create(ctx context.Context, email, name, password string, roles []string) (*domain.User, error) {
	if email == "" || name == "" || password == "" {
		return nil, errors.New("email, name, password required")
	}
	if !domain.IsValidRoles(roles) {
		return nil, errors.New("invalid roles: at least one of associate, seller, guide, manager")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("hash failed")
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		Roles:        roles,
		PasswordHash: string(hash),
		Active:       true,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserRequest carries the optional fields an admin may change.
type UpdateUserRequest struct {
	Name     *string
	Roles    *[]string
	Password *string
	Active   *bool
}

// Update applies the requested changes to an existing user.
func (s *UserService) Update(ctx context.Context, id string, req UpdateUserRequest) (*domain.User, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Roles != nil {
		if !domain.IsValidRoles(*req.Roles) {
			return nil, errors.New("invalid roles: at least one of associate, seller, guide, manager")
		}
		existing.Roles = *req.Roles
	}
	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("hash failed")
		}
		existing.PasswordHash = string(hash)
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Deactivate soft-deletes a user (sets Active to false).
func (s *UserService) Deactivate(ctx context.Context, id string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}
	existing.Active = false
	return s.repo.Update(ctx, existing)
}

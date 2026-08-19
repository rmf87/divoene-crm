package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rmf87/divoene/internal/core/domain"
)

// UserRepository implements domain.UserRepository using SQLite.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `id, email, password_hash, name, active, created_at`

func (r *UserRepository) scanUser(row *sql.Row) (*domain.User, error) {
	u := &domain.User{}
	var active int
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &active, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuário não encontrado")
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.Active = active == 1
	return u, nil
}

func (r *UserRepository) loadRoles(ctx context.Context, u *domain.User) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT role FROM user_roles WHERE user_id = ? ORDER BY role`, u.ID)
	if err != nil {
		return fmt.Errorf("load roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return fmt.Errorf("scan role: %w", err)
		}
		u.Roles = append(u.Roles, role)
	}
	return rows.Err()
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = ?`, email))
	if err != nil {
		return nil, err
	}
	if err := r.loadRoles(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := r.scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	if err := r.loadRoles(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, name, active)
		 VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.PasswordHash, user.Name, boolToInt(user.Active)); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	if err := replaceRolesTx(ctx, tx, user.ID, user.Roles); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit user: %w", err)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+userColumns+` FROM users ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		var active int
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &active, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		u.Active = active == 1
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, u := range users {
		if err := r.loadRoles(ctx, u); err != nil {
			return nil, err
		}
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET email = ?, name = ?, password_hash = ?, active = ? WHERE id = ?`,
		user.Email, user.Name, user.PasswordHash, boolToInt(user.Active), user.ID); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if err := replaceRolesTx(ctx, tx, user.ID, user.Roles); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit user: %w", err)
	}
	return nil
}

func replaceRolesTx(ctx context.Context, tx *sql.Tx, userID string, roles []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("clear roles: %w", err)
	}
	for _, role := range roles {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO user_roles (user_id, role) VALUES (?, ?)`, userID, role); err != nil {
			return fmt.Errorf("insert role %s: %w", role, err)
		}
	}
	return nil
}

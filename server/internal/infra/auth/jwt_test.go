package auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rmf87/divoene/internal/infra/database"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.MigrateUp(db); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}
	return db
}

func TestSetupJWTAuth(t *testing.T) {
	db := setupAuthDB(t)
	userRepo := database.NewUserRepository(db)

	mw, err := SetupJWTAuth(db, userRepo, "test-secret")
	if err != nil {
		t.Fatalf("SetupJWTAuth: %v", err)
	}
	if mw == nil {
		t.Fatal("middleware is nil")
	}
}

func TestSeedUsers(t *testing.T) {
	db := setupAuthDB(t)
	credentials := AdminCredentials{Email: "admin@test.local", Password: "test-password"}

	if err := SeedUsers(db, credentials); err != nil {
		t.Fatalf("SeedUsers: %v", err)
	}

	userRepo := database.NewUserRepository(db)
	for _, email := range []string{credentials.Email} {
		user, err := userRepo.FindByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("FindByEmail(%s): %v", email, err)
		}
		if !user.Active {
			t.Errorf("%s: not active", email)
		}
	}
	// Seed again — must not rotate an existing password.
	if err := SeedUsers(db, credentials); err != nil {
		t.Fatalf("SeedUsers repeat: %v", err)
	}
}

func TestAccountLockout(t *testing.T) {
	db := setupAuthDB(t)

	if err := SeedUsers(db, AdminCredentials{Email: "admin@test.local", Password: "test-password"}); err != nil {
		t.Fatalf("SeedUsers: %v", err)
	}

	email := "admin@test.local"
	if isLockedOut(email) {
		t.Error("should not be locked out initially")
	}

	// 5 failed attempts
	for i := 0; i < 5; i++ {
		recordFailedAttempt(email)
	}

	if !isLockedOut(email) {
		t.Error("should be locked out after 5 attempts")
	}

	// Clear and verify
	clearFailedAttempts(email)
	if isLockedOut(email) {
		t.Error("should not be locked out after clear")
	}
}

func TestEnsureAdminUsesVaultHashAndDoesNotRotate(t *testing.T) {
	db := setupAuthDB(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("vault-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	credentials := AdminCredentials{
		Email:        "admin@vault.test",
		Password:     "vault-password",
		PasswordHash: string(hash),
	}
	if err := EnsureAdmin(db, credentials); err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}
	userRepo := database.NewUserRepository(db)
	user, err := userRepo.FindByEmail(context.Background(), credentials.Email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if user.PasswordHash != credentials.PasswordHash {
		t.Fatal("admin hash was not copied from vault")
	}
	if err := EnsureAdmin(db, AdminCredentials{Email: credentials.Email, Password: "different"}); err != nil {
		t.Fatalf("EnsureAdmin repeat: %v", err)
	}
	user, err = userRepo.FindByEmail(context.Background(), credentials.Email)
	if err != nil {
		t.Fatalf("FindByEmail repeat: %v", err)
	}
	if user.PasswordHash != credentials.PasswordHash {
		t.Fatal("existing admin hash was rotated")
	}
}

func TestEnsureAdminRequiresCredentialsForMissingUser(t *testing.T) {
	if err := EnsureAdmin(setupAuthDB(t), AdminCredentials{}); err == nil {
		t.Fatal("expected missing admin credentials error")
	}
}

func TestEnsureAdminRejectsInvalidEmail(t *testing.T) {
	invalid := []string{
		"not-an-email",
		"missing-at-sign",
		"@no-local.com",
		" space@example.com",
	}
	for _, email := range invalid {
		db := setupAuthDB(t)
		err := EnsureAdmin(db, AdminCredentials{Email: email, Password: "secret"})
		if err == nil {
			t.Errorf("expected error for invalid email %q", email)
		}
	}
}

func TestEnsureAdminAcceptsValidEmail(t *testing.T) {
	db := setupAuthDB(t)
	err := EnsureAdmin(db, AdminCredentials{Email: "admin@valid.test", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error for valid email: %v", err)
	}
}

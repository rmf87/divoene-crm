package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/infra/database"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxFailedAttempts = 5
	lockoutDuration   = 15 * time.Minute
)

// loginTrack tracks failed login attempts per email for account lockout.
type loginTrack struct {
	failedAttempts int
	lockedUntil    time.Time
}

var (
	loginTracker      = map[string]*loginTrack{}
	loginTrackerMutex sync.Mutex
)

// recordFailedAttempt increments the failed counter for an email.
// If maxFailedAttempts is reached, sets lockedUntil to now + lockoutDuration.
func recordFailedAttempt(email string) {
	loginTrackerMutex.Lock()
	defer loginTrackerMutex.Unlock()
	t, ok := loginTracker[email]
	if !ok {
		t = &loginTrack{}
		loginTracker[email] = t
	}
	t.failedAttempts++
	if t.failedAttempts >= maxFailedAttempts {
		t.lockedUntil = time.Now().Add(lockoutDuration)
	}
}

// isLockedOut returns true if the email is currently in lockout.
func isLockedOut(email string) bool {
	loginTrackerMutex.Lock()
	defer loginTrackerMutex.Unlock()
	t, ok := loginTracker[email]
	if !ok {
		return false
	}
	if t.lockedUntil.IsZero() {
		return false
	}
	if time.Now().After(t.lockedUntil) {
		delete(loginTracker, email) // lockout expired, reset
		return false
	}
	return true
}

// clearFailedAttempts resets the counter on successful login.
func clearFailedAttempts(email string) {
	loginTrackerMutex.Lock()
	defer loginTrackerMutex.Unlock()
	delete(loginTracker, email)
}

// SetupJWTAuth creates the JWT middleware with account lockout.
func SetupJWTAuth(db *sql.DB, userRepo *database.UserRepository, secret string) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "divoene",
		Key:         []byte(secret),
		Timeout:     24 * time.Hour,
		MaxRefresh:  7 * 24 * time.Hour,
		IdentityKey: "uid",
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*domain.User); ok {
				return jwt.MapClaims{
					"uid":   v.ID,
					"roles": v.Roles,
					"name":  v.Name,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			actor := &domain.Actor{
				UID:  claims["uid"].(string),
				Name: claims["name"].(string),
			}
			if roles, ok := claims["roles"].([]interface{}); ok {
				for _, r := range roles {
					if s, ok := r.(string); ok {
						actor.Roles = append(actor.Roles, s)
					}
				}
			}
			// Backward compatibility with single-role tokens.
			if len(actor.Roles) == 0 {
				if role, ok := claims["role"].(string); ok && role != "" {
					actor.Roles = []string{role}
				}
			}
			c.Set("actor", actor)
			return actor
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var login domain.LoginRequest
			if err := c.ShouldBindJSON(&login); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}
			// Check lockout before attempting auth
			if isLockedOut(login.Email) {
				return nil, jwt.ErrFailedAuthentication
			}
			user, err := userRepo.FindByEmail(c.Request.Context(), login.Email)
			if err != nil {
				recordFailedAttempt(login.Email)
				return nil, jwt.ErrFailedAuthentication
			}
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(login.Password)); err != nil {
				recordFailedAttempt(login.Email)
				return nil, jwt.ErrFailedAuthentication
			}
			if !user.Active {
				return nil, jwt.ErrFailedAuthentication
			}
			clearFailedAttempts(login.Email)
			return user, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{"error": "unauthorized"})
		},
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
	})
}

// AdminCredentials are supplied by a test fixture or the deploy secret
// resolver. They must never be populated with repository defaults.
type AdminCredentials struct {
	Email        string
	Password     string
	PasswordHash string
}

// EnsureAdmin creates the manager account only when it does not exist. It
// deliberately does not rotate an existing password on every boot.
func EnsureAdmin(db *sql.DB, credentials AdminCredentials) error {
	if credentials.Email == "" || (credentials.Password == "" && credentials.PasswordHash == "") {
		return fmt.Errorf("admin credentials required")
	}
	if strings.ContainsAny(credentials.Email, " \t\n\r") || !strings.Contains(credentials.Email, "@") || strings.HasPrefix(credentials.Email, "@") || strings.HasSuffix(credentials.Email, "@") {
		return fmt.Errorf("admin email must be a valid address, got %q", credentials.Email)
	}
	if credentials.PasswordHash != "" {
		if _, err := bcrypt.Cost([]byte(credentials.PasswordHash)); err != nil {
			return fmt.Errorf("invalid admin password hash: %w", err)
		}
		if credentials.Password != "" && bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(credentials.Password)) != nil {
			return fmt.Errorf("admin password does not match password hash")
		}
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash admin password: %w", err)
		}
		credentials.PasswordHash = string(hash)
	}

	userRepo := database.NewUserRepository(db)
	ctx := context.Background()
	if existing, err := userRepo.FindByEmail(ctx, credentials.Email); err == nil && existing != nil {
		log.Printf("[startup] admin bootstrap: skipped (exists) email=%s", credentials.Email)
		return nil
	}

	if err := userRepo.Create(ctx, &domain.User{
		ID:           uuid.New().String(),
		Email:        credentials.Email,
		PasswordHash: credentials.PasswordHash,
		Name:         "Gerente",
		Roles:        []string{"manager"},
		Active:       true,
	}); err != nil {
		return err
	}
	log.Printf("[startup] admin bootstrap: created email=%s", credentials.Email)
	return nil
}

// SeedUsers creates the admin account for test and fixture databases only.
func SeedUsers(db *sql.DB, credentials AdminCredentials) error {
	return EnsureAdmin(db, credentials)
}

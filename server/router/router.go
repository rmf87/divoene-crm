package router

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/rmf87/divoene/handlers"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/core/services"
	"github.com/rmf87/divoene/internal/infra/clients"
	"github.com/rmf87/divoene/middleware"
)

// Config holds all dependencies for the router.
type Config struct {
	DB              *sql.DB
	LeadService     *services.LeadService
	ContractService *services.ContractService
	PaymentService  *services.PaymentService
	VisitService    *services.VisitService
	ConfigService   *services.ConfigService
	UserService     *services.UserService
	BackupService   *services.BackupService
	ChatService     *services.ChatService
	JWTMiddleware   *jwt.GinJWTMiddleware
	ClicksignClient *clients.ClicksignClient
	OpenPixClient   *clients.OpenPixClient
	WhatsAppClient  *clients.WhatsAppClient
}

// SetupRouter creates and configures the Gin engine with role-based routes.
func SetupRouter(cfg *Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	// Health — always public
	healthHandler := handlers.NewHealthHandler(cfg.DB)
	r.GET("/health", healthHandler.GetHealth)

	// ── Public routes (no auth) ──
	public := r.Group("/api")
	public.POST("/auth/login", cfg.JWTMiddleware.LoginHandler)
	public.GET("/auth/refresh_token", cfg.JWTMiddleware.RefreshHandler)

	// Lead creation from site form — public
	public.POST("/leads", publicLeadForm(cfg.LeadService))

	// Webhooks — public, validated by HMAC
	handlers.RegisterWebhookRoutes(public, cfg.ContractService, cfg.PaymentService,
		cfg.ClicksignClient, cfg.OpenPixClient)

	// WhatsApp — webhook public (signature validated)
	handlers.RegisterWhatsAppWebhook(public, cfg.ChatService, cfg.WhatsAppClient)

	// ── Protected routes (JWT required) ──
	protected := r.Group("/api")
	protected.Use(cfg.JWTMiddleware.MiddlewareFunc())

	// Leads — seller or manager (associates and guides cannot access)
	leads := protected.Group("/leads")
	leads.Use(middleware.SellerOrManager())
	handlers.RegisterLeadRoutes(nil, leads, cfg.LeadService)
	handlers.RegisterWhatsAppMessages(leads, cfg.ChatService)

	// Contracts — seller or manager
	contracts := protected.Group("/contracts")
	contracts.Use(middleware.SellerOrManager())
	handlers.RegisterContractRoutes(contracts, cfg.ContractService)

	// Payments — seller or manager
	payments := protected.Group("/payments")
	payments.Use(middleware.SellerOrManager())
	handlers.RegisterPaymentRoutes(payments, cfg.PaymentService)

	// Visits — seller/manager for write, guide/manager for read
	visits := protected.Group("/visits")
	visits.Use(middleware.GuideOrManager())
	handlers.RegisterVisitRoutes(visits, cfg.VisitService)

	// Guides — read for any authenticated, write for manager only
	guidesRead := protected.Group("/guides")
	handlers.RegisterGuideRoutes(guidesRead, cfg.VisitService)

	// Users — manager only
	users := protected.Group("/users")
	users.Use(middleware.ManagerOnly())
	handlers.RegisterUserRoutes(users, handlers.NewUserHandler(cfg.UserService))

	// Config — manager only
	config := protected.Group("/config")
	config.Use(middleware.ManagerOnly())
	handlers.RegisterConfigRoutes(config, cfg.ConfigService, func(key string) {
		ctx := context.Background()
		if strings.HasPrefix(key, "clicksign_") {
			cfg.ClicksignClient.Reload(ctx, cfg.ConfigService)
		}
		if strings.HasPrefix(key, "openpix_") {
			cfg.OpenPixClient.Reload(ctx, cfg.ConfigService)
		}
		if strings.HasPrefix(key, "whatsapp_") {
			cfg.WhatsAppClient.Reload(ctx, cfg.ConfigService)
		}
	})

	// Backup — manager only
	backupGroup := protected.Group("/backup")
	backupGroup.Use(middleware.ManagerOnly())
	handlers.RegisterBackupRoutes(backupGroup, handlers.NewBackupHandler(cfg.BackupService))

	return r
}

func publicLeadForm(svc *services.LeadService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req domain.CreateLeadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.WhatsApp = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' { return r }; return -1
		}, req.WhatsApp)
		req.Product = strings.TrimSpace(req.Product)
		req.Source = strings.TrimSpace(req.Source)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nome é obrigatório"})
			return
		}
		if len(req.WhatsApp) < 10 || len(req.WhatsApp) > 13 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "whatsapp inválido"})
			return
		}
		if req.Product == "" || !domain.IsValidProduct(req.Product) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "produto inválido"})
			return
		}
		if req.Source == "" {
			req.Source = "site"
		}
		lead, err := svc.Create(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, lead)
	}
}


package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rmf87/divoene/internal/core/services"
	"github.com/rmf87/divoene/internal/infra/auth"
	"github.com/rmf87/divoene/internal/infra/clients"
	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/rmf87/divoene/internal/infra/scheduler"
	"github.com/rmf87/divoene/router"
)

func main() {
	// Read config from env vars
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "/data/divoene.sqlite3")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")

	// External API config (empty = mock mode)
	clicksignAPIKey := os.Getenv("CLICKSIGN_API_KEY")
	clicksignBaseURL := os.Getenv("CLICKSIGN_BASE_URL")
	clicksignWebhookSecret := os.Getenv("CLICKSIGN_WEBHOOK_SECRET")
	clicksignTemplateKey := os.Getenv("CLICKSIGN_TEMPLATE_KEY")
	openpixAppID := os.Getenv("OPENPIX_APP_ID")
	openpixBaseURL := os.Getenv("OPENPIX_BASE_URL")

	log.Printf("[startup] port=%s db=%s", port, dbPath)
	log.Printf("[startup] clicksign=mock:%t openpix=mock:%t",
		clicksignAPIKey == "", openpixAppID == "")

	// 1. Initialize SQLite database
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatalf("[startup] database: %v", err)
	}
	defer db.Close()

	// 2. Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("[startup] migrations: %v", err)
	}
	log.Printf("[startup] migrations complete")
	adminEmail := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminPasswordHash := os.Getenv("ADMIN_PASSWORD_HASH")
	adminRepo := database.NewUserRepository(db)
	if _, err := adminRepo.FindByEmail(context.Background(), adminEmail); err != nil {
		if err := auth.EnsureAdmin(db, auth.AdminCredentials{Email: adminEmail, Password: adminPassword, PasswordHash: adminPasswordHash}); err != nil {
			log.Fatalf("[startup] admin bootstrap: %v", err)
		}
	}

	// 3. Initialize repositories
	leadRepo := database.NewLeadRepository(db)
	contractRepo := database.NewContractRepository(db)
	paymentRepo := database.NewPaymentRepository(db)
	visitRepo := database.NewVisitRepository(db)
	userRepo := database.NewUserRepository(db)

	// 4. Initialize config repository and service (DB + env-var fallback)
	configRepo := database.NewConfigRepository(db, jwtSecret)
	configSvc := services.NewConfigService(configRepo, map[string]string{
		"clicksign_api_key":        "CLICKSIGN_API_KEY",
		"clicksign_base_url":       "CLICKSIGN_BASE_URL",
		"clicksign_webhook_secret": "CLICKSIGN_WEBHOOK_SECRET",
		"clicksign_template_key":   "CLICKSIGN_TEMPLATE_KEY",
		"openpix_app_id":           "OPENPIX_APP_ID",
		"openpix_base_url":         "OPENPIX_BASE_URL",
		"s3_endpoint":              "S3_ENDPOINT",
		"s3_bucket":                "S3_BUCKET",
		"s3_access_key":            "S3_ACCESS_KEY",
		"s3_secret_key":            "S3_SECRET_KEY",
	})

	// 5. Initialize API clients (empty env vars = mock mode; Reload sets DB overrides)
	clicksignClient := clients.NewClicksignClient(clicksignBaseURL, clicksignAPIKey, clicksignWebhookSecret, clicksignTemplateKey)
	clicksignClient.Reload(context.Background(), configSvc)

	openpixClient := clients.NewOpenPixClient(openpixBaseURL, openpixAppID)
	openpixClient.Reload(context.Background(), configSvc)

	// 6. Initialize services
	leadSvc := services.NewLeadService(leadRepo)
	contractSvc := services.NewContractService(contractRepo, clicksignClient)
	paymentSvc := services.NewPaymentService(paymentRepo, openpixClient)
	visitSvc := services.NewVisitService(visitRepo)

	// 7. Initialize JWT auth
	jwtMiddleware, err := auth.SetupJWTAuth(db, userRepo, jwtSecret)
	if err != nil {
		log.Fatalf("[startup] jwt: %v", err)
	}

	// 8. Start background scheduler
	sched := scheduler.NewScheduler(contractSvc, paymentSvc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Start(ctx)

	// 9. Setup Gin router
	cfg := &router.Config{
		DB:              db,
		LeadService:     leadSvc,
		ContractService: contractSvc,
		PaymentService:  paymentSvc,
		VisitService:    visitSvc,
		ConfigService:   configSvc,
		JWTMiddleware:   jwtMiddleware,
		ClicksignClient: clicksignClient,
		OpenPixClient:   openpixClient,
	}
	engine := router.SetupRouter(cfg)

	// 10. Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[startup] listen: %v", err)
		}
	}()
	log.Printf("[startup] server listening on :%s", port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("[shutdown] received signal, shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[shutdown] forced: %v", err)
	}
	log.Printf("[shutdown] complete")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

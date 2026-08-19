package commands

import (
	"context"
	"fmt"
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
	"github.com/rmf87/divoene/internal/infra/storage"
	"github.com/rmf87/divoene/router"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	RunE:  runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	port := getEnv("PORT", "8080")
	dbPath := resolveDBPath()
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")

	log.Printf("[startup] port=%s db=%s", port, dbPath)

	db, err := database.NewDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		return err
	}
	adminEmail := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminPasswordHash := os.Getenv("ADMIN_PASSWORD_HASH")
	userRepo := database.NewUserRepository(db)
	if _, err := userRepo.FindByEmail(context.Background(), adminEmail); err != nil {
		if err := auth.EnsureAdmin(db, auth.AdminCredentials{Email: adminEmail, Password: adminPassword, PasswordHash: adminPasswordHash}); err != nil {
			return fmt.Errorf("admin bootstrap: %w", err)
		}
	}

	leadRepo := database.NewLeadRepository(db)
	contractRepo := database.NewContractRepository(db)
	paymentRepo := database.NewPaymentRepository(db)
	visitRepo := database.NewVisitRepository(db)

	// Initialize config repository and service (DB + env-var fallback)
	configRepo := database.NewConfigRepository(db, jwtSecret)
	configSvc := services.NewConfigService(configRepo, map[string]string{
		"clicksign_api_key":             "CLICKSIGN_API_KEY",
		"clicksign_base_url":            "CLICKSIGN_BASE_URL",
		"clicksign_webhook_secret":      "CLICKSIGN_WEBHOOK_SECRET",
		"clicksign_template_key":        "CLICKSIGN_TEMPLATE_KEY",
		"openpix_app_id":                "OPENPIX_APP_ID",
		"openpix_base_url":              "OPENPIX_BASE_URL",
		"whatsapp_token":                "WHATSAPP_TOKEN",
		"whatsapp_phone_number_id":      "WHATSAPP_PHONE_NUMBER_ID",
		"whatsapp_app_secret":           "WHATSAPP_APP_SECRET",
		"whatsapp_webhook_verify_token": "WHATSAPP_WEBHOOK_VERIFY_TOKEN",
		"whatsapp_base_url":             "WHATSAPP_BASE_URL",
	})

	clicksignClient := clients.NewClicksignClient(
		os.Getenv("CLICKSIGN_BASE_URL"), os.Getenv("CLICKSIGN_API_KEY"),
		os.Getenv("CLICKSIGN_WEBHOOK_SECRET"), os.Getenv("CLICKSIGN_TEMPLATE_KEY"))
	clicksignClient.Reload(context.Background(), configSvc)

	openpixClient := clients.NewOpenPixClient(os.Getenv("OPENPIX_BASE_URL"), os.Getenv("OPENPIX_APP_ID"))
	openpixClient.Reload(context.Background(), configSvc)

	whatsappClient := clients.NewWhatsAppClient(
		os.Getenv("WHATSAPP_BASE_URL"), os.Getenv("WHATSAPP_TOKEN"),
		os.Getenv("WHATSAPP_PHONE_NUMBER_ID"), os.Getenv("WHATSAPP_APP_SECRET"),
		os.Getenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN"))
	whatsappClient.Reload(context.Background(), configSvc)

	leadSvc := services.NewLeadService(leadRepo)
	contractSvc := services.NewContractService(contractRepo, clicksignClient)
	paymentSvc := services.NewPaymentService(paymentRepo, openpixClient)
	visitSvc := services.NewVisitService(visitRepo)
	userSvc := services.NewUserService(userRepo)
	chatSvc := services.NewChatService(database.NewChatMessageRepository(db), leadRepo, whatsappClient,
		os.Getenv("WHATSAPP_MOCK_AUTOREPLY") == "1")
	backupSvc := services.NewBackupService(
		storage.NewSnapshotStore(db, dbPath), dbPath,
		func() {
			time.Sleep(500 * time.Millisecond)
			syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		},
	)

	jwtMiddleware, err := auth.SetupJWTAuth(db, userRepo, jwtSecret)
	if err != nil {
		return err
	}

	sched := scheduler.NewScheduler(contractSvc, paymentSvc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Start(ctx)

	cfg := &router.Config{
		DB:              db,
		LeadService:     leadSvc,
		ContractService: contractSvc,
		PaymentService:  paymentSvc,
		VisitService:    visitSvc,
		ConfigService:   configSvc,
		UserService:     userSvc,
		BackupService:   backupSvc,
		ChatService:     chatSvc,
		JWTMiddleware:   jwtMiddleware,
		ClicksignClient: clicksignClient,
		OpenPixClient:   openpixClient,
		WhatsAppClient:  whatsappClient,
	}
	engine := router.SetupRouter(cfg)

	srv := &http.Server{Addr: ":" + port, Handler: engine}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[startup] listen: %v", err)
		}
	}()
	log.Printf("[startup] listening on :%s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("[shutdown] shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	return srv.Shutdown(shutdownCtx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

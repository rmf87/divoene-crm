package commands

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/infra/auth"
	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/spf13/cobra"
)

var seedDemo bool

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed database with admin user (and demo leads with --demo)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := resolveDBPath()
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
		if err := auth.SeedUsers(db, auth.AdminCredentials{Email: adminEmail, Password: adminPassword, PasswordHash: adminPasswordHash}); err != nil {
			return err
		}
		fmt.Printf("Seed: admin fixture ready (%s)\n", adminEmail)

		if seedDemo {
			if err := seedDemoLeads(db); err != nil {
				return err
			}
		}
		return nil
	},
}

// seedDemoLeads inserts the demo/mock leads (ids L001–L009) used by the admin
// pipeline in dev. Same ids as packages/admin/src/hooks/usePipeline.ts so the
// WhatsApp chat and stage mutations work against real DB rows.
func seedDemoLeads(db *sql.DB) error {
	repo := database.NewLeadRepository(db)
	ctx := context.Background()

	demo := []domain.Lead{
		{ID: "L001", Name: "Maria Silva", WhatsApp: "5511999990001", Product: "buffet_infantil", DesiredDate: "2026-07-15", Source: "instagram", Stage: "lead"},
		{ID: "L002", Name: "João Fotógrafo", WhatsApp: "5511999990002", Product: "ensaio_fotografico", DesiredDate: "2026-07-10", Source: "google", Stage: "validated"},
		{ID: "L003", Name: "Empresa ABC Ltda", WhatsApp: "5511999990003", Product: "corporativo", DesiredDate: "2026-08-01", Source: "linkedin", Stage: "visit_scheduled"},
		{ID: "L004", Name: "Família Oliveira", WhatsApp: "5511999990004", Product: "casamentos", DesiredDate: "2026-09-20", Source: "indicacao", Stage: "visit_done"},
		{ID: "L005", Name: "Pedro Aniversário", WhatsApp: "5511999990005", Product: "locacao_eventos", DesiredDate: "2026-07-22", Source: "facebook", Stage: "contract"},
		{ID: "L006", Name: "Ana Festas", WhatsApp: "5511999990006", Product: "buffet_infantil", DesiredDate: "2026-07-08", Source: "instagram", Stage: "paid"},
		{ID: "L007", Name: "Confraria do Chopp", WhatsApp: "5511999990007", Product: "locacao_eventos", DesiredDate: "2026-06-28", Source: "indicacao", Stage: "booked"},
		{ID: "L008", Name: "Escola ABC Kids", WhatsApp: "5511999990008", Product: "passeios_escolares", DesiredDate: "2026-05-30", Source: "google", Stage: "completed"},
		{ID: "L009", Name: "Carlos Desistente", WhatsApp: "5511999990009", Product: "ensaio_fotografico", DesiredDate: "2026-07-01", Source: "facebook", Stage: "cancelled"},
	}

	count := 0
	for _, l := range demo {
		if _, err := repo.Get(ctx, l.ID); err == nil {
			continue // already present
		}
		if err := repo.Store(ctx, &l); err != nil {
			return fmt.Errorf("seed demo lead %s: %w", l.ID, err)
		}
		count++
	}
	fmt.Printf("Seed: %d demo leads criados (L001–L009)\n", count)
	return nil
}

func init() {
	seedCmd.Flags().BoolVar(&seedDemo, "demo", false, "seed demo leads (dev/mock)")
	rootCmd.AddCommand(seedCmd)
}

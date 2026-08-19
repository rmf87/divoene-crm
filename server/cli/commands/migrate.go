package commands

import (
	"fmt"
	"os"

	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := resolveDBPath()
		db, err := database.NewDB(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := database.MigrateUp(db); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Migrations applied successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := resolveDBPath()
		db, err := database.NewDB(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := database.MigrateDown(db); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Last migration rolled back")
		return nil
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := resolveDBPath()
		db, err := database.NewDB(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		return database.MigrateStatus(db)
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}

func resolveDBPath() string {
	if dbPath != "" {
		return dbPath
	}
	if p := os.Getenv("DB_PATH"); p != "" {
		return p
	}
	return "/data/divoene.sqlite3"
}

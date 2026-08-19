package commands

import (
	"fmt"

	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/rmf87/divoene/internal/infra/storage"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a local database snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := resolveDBPath()
		db, err := database.NewDB(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		path, err := storage.CreateSnapshot(db)
		if err != nil {
			return fmt.Errorf("backup: %w", err)
		}
		fmt.Printf("Backup created: %s\n", path)
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database from a local backup file",
	RunE: func(cmd *cobra.Command, args []string) error {
		destPath := resolveDBPath()
		if override, _ := cmd.Flags().GetString("dest"); override != "" {
			destPath = override
		}

		name, _ := cmd.Flags().GetString("from")
		if name == "" {
			return fmt.Errorf("restore: --from is required (backup filename)")
		}

		if err := storage.RestoreSnapshot(name, destPath); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		fmt.Printf("Backup restored to %s\n", destPath)
		return nil
	},
}

func init() {
	restoreCmd.Flags().String("from", "", "backup filename in /data/backups/")
	restoreCmd.Flags().String("dest", "", "destination path (overrides DB_PATH)")
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}
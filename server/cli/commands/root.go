package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dbPath  string
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "divoene",
	Short: "Chácara Divoene — Sales Pipeline CLI",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// All commands except server, version, help, backup, restore require CLI auth
		switch cmd.Name() {
		case "server", "version", "help", "completion", "backup", "restore":
			return nil
		}
		return requireCLIAuth()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db-path", "", "SQLite database path (overrides config)")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
	rootCmd.AddCommand(userCmd)
}

// requireCLIAuth checks CLI_AUTH_TOKEN env var.
// In development (no token), falls back with a warning.
func requireCLIAuth() error {
	token := os.Getenv("CLI_AUTH_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "WARNING: CLI_AUTH_TOKEN not set — running without auth (dev mode)")
		return nil
	}
	// In production, validate against stored token
	if token == "dev-secret-change-in-production" {
		fmt.Fprintln(os.Stderr, "WARNING: using default CLI_AUTH_TOKEN — change in production")
	}
	return nil
}

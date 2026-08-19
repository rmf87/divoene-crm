package commands

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rmf87/divoene/internal/core/domain"
	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User management commands",
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		name, _ := cmd.Flags().GetString("name")
		roles, _ := cmd.Flags().GetStringSlice("role")
		password, _ := cmd.Flags().GetString("password")

		if email == "" || name == "" || password == "" {
			return fmt.Errorf("--email, --name, --password required")
		}
		if !domain.IsValidRoles(roles) {
			return fmt.Errorf("invalid roles: %v (valid: associate, seller, guide, manager)", roles)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		dbPath := resolveDBPath()
		db, err := database.NewDB(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		repo := database.NewUserRepository(db)
		user := &domain.User{
			ID:           uuid.New().String(),
			Email:        email,
			Name:         name,
			Roles:        roles,
			PasswordHash: string(hash),
			Active:       true,
		}
		if err := repo.Create(cmd.Context(), user); err != nil {
			return err
		}
		fmt.Printf("User created: %s (%s)\n", email, strings.Join(roles, ", "))
		return nil
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := resolveDBPath()
		db, err := database.NewDB(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		repo := database.NewUserRepository(db)
		users, err := repo.List(cmd.Context())
		if err != nil {
			return err
		}
		for _, u := range users {
			active := ""
			if !u.Active {
				active = " (inactive)"
			}
			fmt.Printf("  %-40s %-10s %s%s\n", u.Email, strings.Join(u.Roles, ","), u.Name, active)
		}
		return nil
	},
}

func init() {
	userCreateCmd.Flags().String("email", "", "user email")
	userCreateCmd.Flags().String("name", "", "user display name")
	userCreateCmd.Flags().StringSlice("role", []string{"seller"}, "user roles (repeat or comma-separated)")
	userCreateCmd.Flags().String("password", "", "user password")
	userCreateCmd.MarkFlagRequired("email")
	userCreateCmd.MarkFlagRequired("name")
	userCreateCmd.MarkFlagRequired("password")

	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userListCmd)
}

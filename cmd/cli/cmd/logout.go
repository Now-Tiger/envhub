package cmd

import (
	"fmt"

	"github.com/Now-Tiger/envhub/internal/cli/config"
	"github.com/spf13/cobra"
)

// logoutCmd returns the logout command
func logoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logout",
		Short:   "Logout and clear credentials",
		Long:    "Logout and clear stored authentication credentials",
		Example: `  envhub logout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Get()

			if err := cfg.ClearToken(); err != nil {
				return fmt.Errorf("failed to clear credentials: %w", err)
			}

			printSuccess("Logged out successfully")
			return nil
		},
	}

	return cmd
}

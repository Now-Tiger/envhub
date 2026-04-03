package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Now-Tiger/envhub/internal/cli/client"
	"github.com/Now-Tiger/envhub/internal/cli/config"
	"github.com/Now-Tiger/envhub/internal/cli/style"
	"github.com/spf13/cobra"
)

// whoamiCmd returns the whoami command
func whoamiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "whoami",
		Short:   "Show current user",
		Long:    "Display information about the currently authenticated user",
		Example: `  envhub whoami`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			cfg := config.Get()

			token, err := cfg.TokenKey()
			if err != nil {
				printError("not logged in")
				fmt.Println(style.Dim.Render("Hint: Set your API token:"))
				fmt.Println(style.Dim.Render("  • Export: export ENVUB_TOKEN=your_token"))
				fmt.Println(style.Dim.Render("  • Login:  envhub login"))
				return err
			}

			apiClient := client.New(cfg.APIBaseURL, token)
			resp, err := apiClient.Get(ctx, "/api/v1/auth/me")
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
			}

			var user map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			fmt.Println(style.Header.Render("Current User:"))
			fmt.Printf("  %s %s\n", style.ListBullet, user["email"])
			if name, ok := user["name"].(string); ok && name != "" {
				fmt.Printf("  %s %s\n", style.ListBullet, name)
			}

			return nil
		},
	}

	return cmd
}

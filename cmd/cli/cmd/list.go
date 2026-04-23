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

// listCmd returns the list command
func listCmd() *cobra.Command {
	var (
		org    string
		format string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Long:  "List all projects accessible to the current user",
		Example: `  envhub list
  envhub list --org myorg
  envhub list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			cfg := config.Get()

			token, err := cfg.TokenKey()
			if err != nil {
				printError("authentication required")
				fmt.Println(style.Dim.Render("Hint: Set your API token:"))
				fmt.Println(style.Dim.Render("  • Export: export ENVUB_TOKEN=your_token"))
				fmt.Println(style.Dim.Render("  • Login:  envhub login"))
				return err
			}

			apiClient := client.New(cfg.APIBaseURL, token)

			path := "/api/v1/projects"
			if org != "" {
				path = "/api/v1/organizations/" + org + "/projects"
			}

			resp, err := apiClient.Get(ctx, path)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close() //nolint:errcheck

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			projects, ok := result["projects"].([]interface{})
			if !ok || len(projects) == 0 {
				printInfo("No projects found")
				return nil
			}

			if format == "json" {
				output, err := json.MarshalIndent(projects, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}
				fmt.Println(string(output))
			} else {
				fmt.Println(style.Header.Render("Projects:"))
				for _, p := range projects {
					if proj, ok := p.(map[string]interface{}); ok {
						name := proj["name"]
						fmt.Printf("  %s %s\n", style.ListBullet, name)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "filter by organization")
	cmd.Flags().StringVarP(&format, "output", "o", "table", "output format: table, json")

	return cmd
}

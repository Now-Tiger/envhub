package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Now-Tiger/envhub/internal/cli/client"
	"github.com/Now-Tiger/envhub/internal/cli/config"
	"github.com/Now-Tiger/envhub/internal/cli/style"
	"github.com/Now-Tiger/envhub/internal/cli/util"
	"github.com/spf13/cobra"
)

// pullCmd returns the pull command
func pullCmd() *cobra.Command {
	var (
		project     string
		environment string
		output      string
		format      string
	)

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull secrets from EnvHub",
		Long:  "Pull secrets from EnvHub for a specific project and environment",
		Example: `  envhub pull --project myapp --env production
  envhub pull -p myapp -e staging --output .env
  envhub pull -p myapp -e prod --format json`,
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

			if project == "" {
				printError("project name required")
				fmt.Println(style.Dim.Render("Usage: envhub pull --project <name>"))
				return fmt.Errorf("project name required")
			}

			if environment == "" {
				environment = "development"
			}

			fmt.Println(style.Info.Render("Pulling secrets from " + style.Code.Render(project+"/"+environment) + "..."))

			apiClient := client.New(cfg.APIBaseURL, token)
			resp, err := apiClient.Get(ctx, fmt.Sprintf("/api/v1/cli/secrets/%s/%s", project, environment))
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			secrets, ok := result["secrets"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("no secrets found")
			}

			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				util.WriteEnvFile(f, secrets)
				printSuccess("Secrets written to " + output)
			} else {
				util.WriteEnvFile(os.Stdout, secrets)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "project name (required)")
	cmd.Flags().StringVarP(&environment, "env", "e", "development", "environment (default: development)")
	cmd.Flags().StringVar(&output, "output", "", "output file (.env)")
	cmd.Flags().StringVar(&format, "format", "env", "output format: env, json, yaml, docker")

	_ = cmd.MarkFlagRequired("project")

	return cmd
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Now-Tiger/envhub/internal/cli/client"
	"github.com/Now-Tiger/envhub/internal/cli/config"
	"github.com/Now-Tiger/envhub/internal/cli/style"
	"github.com/spf13/cobra"
)

// pushCmd returns the push command
func pushCmd() *cobra.Command {
	var (
		project     string
		environment string
		input       string
		format      string
	)

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push secrets to EnvHub",
		Long:  "Push secrets to EnvHub for a specific project and environment",
		Example: `  envhub push --project myapp --env production --input .env
  envhub push -p myapp -e staging --input secrets.json`,
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
				fmt.Println(style.Dim.Render("Usage: envhub push --project <name>"))
				return fmt.Errorf("project name required")
			}

			if environment == "" {
				environment = "development"
			}

			if input == "" {
				printError("input file required")
				fmt.Println(style.Dim.Render("Usage: envhub push --input <file>"))
				return fmt.Errorf("input file required")
			}

			// Read input file
			data, err := os.ReadFile(input)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}

			// Parse secrets based on format
			var secrets map[string]string
			switch format {
			case "env":
				secrets, err = parseEnvFile(string(data))
			case "json":
				err = json.Unmarshal(data, &secrets)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
			if err != nil {
				return fmt.Errorf("failed to parse input: %w", err)
			}

			fmt.Println(style.Info.Render("Pushing secrets to " + style.Code.Render(project+"/"+environment) + "..."))

			apiClient := client.New(cfg.APIBaseURL, token)
			payload := map[string]interface{}{
				"secrets": secrets,
			}
			resp, err := apiClient.Post(ctx, fmt.Sprintf("/api/v1/cli/secrets/%s/%s", project, environment), payload)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 && resp.StatusCode != 201 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
			}

			printSuccess("Secrets pushed successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "project name (required)")
	cmd.Flags().StringVarP(&environment, "env", "e", "development", "environment (default: development)")
	cmd.Flags().StringVar(&input, "input", "", "input file (required)")
	cmd.Flags().StringVar(&format, "format", "env", "input format: env, json")

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("input")

	return cmd
}

// parseEnvFile parses a .env file content into a map
func parseEnvFile(content string) (map[string]string, error) {
	secrets := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		secrets[key] = value
	}

	return secrets, nil
}

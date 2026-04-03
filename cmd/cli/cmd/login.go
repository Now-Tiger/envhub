package cmd

import (
	"fmt"

	"github.com/Now-Tiger/envhub/internal/cli/style"
	"github.com/spf13/cobra"
)

// loginCmd returns the login command
func loginCmd() *cobra.Command {
	var loginBrowser bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to EnvHub",
		Long:  "Login to EnvHub to authenticate and store credentials",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(style.Header.Render("Login to EnvHub"))
			fmt.Println()
			fmt.Println(style.Info.Render("1. Start the API server:"))
			fmt.Println(style.Dim.Render("   docker-compose up -d"))
			fmt.Println()
			fmt.Println(style.Info.Render("2. Open browser:"))
			fmt.Println(style.Dim.Render("   http://localhost:8080"))
			fmt.Println()
			fmt.Println(style.Info.Render("3. Register and create a project"))
			fmt.Println()
			fmt.Println(style.Info.Render("4. Get a CLI token:"))
			fmt.Println(style.Dim.Render("   - Login via web UI"))
			fmt.Println(style.Dim.Render("   - Use CLI login endpoint or create API token"))
			fmt.Println()
			fmt.Println(style.Info.Render("5. Set the token:"))
			fmt.Println(style.Dim.Render("   export ENVUB_TOKEN=<your-token>"))
			fmt.Println()
			fmt.Println(style.Info.Render("Quick start:"))
			fmt.Println(style.Dim.Render(`  export ENVUB_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \`))
			fmt.Println(style.Dim.Render(`    -H "Content-Type: application/json" \`))
			fmt.Println(style.Dim.Render(`    -d '{"email":"you@example.com","password":"password"}' | jq -r '.token')`))

			if loginBrowser {
				fmt.Println()
				fmt.Println(style.Info.Render("Opening browser..."))
				// Would open browser here
			}
		},
	}

	cmd.Flags().BoolVar(&loginBrowser, "browser", false, "open browser for login")

	return cmd
}

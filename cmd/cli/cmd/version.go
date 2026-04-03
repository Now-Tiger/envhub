package cmd

import (
	"fmt"

	"github.com/Now-Tiger/envhub/internal/cli/banner"
	"github.com/Now-Tiger/envhub/internal/cli/style"
	"github.com/spf13/cobra"
)

// versionCmd returns the version command
func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Show version information",
		Long:    "Display version information about EnvHub CLI",
		Example: `  envhub version`,
		Run: func(cmd *cobra.Command, args []string) {
			version := cmd.Version
			fmt.Println(banner.EnvHubBanner)
			fmt.Println()
			fmt.Println(style.Header.Render("Version: ") + style.Code.Render(version))
			fmt.Println(style.Dim.Render("Environment: " + style.Code.Render("production")))
		},
	}

	return cmd
}

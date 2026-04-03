package cmd

import (
	"fmt"
	"os"

	"github.com/Now-Tiger/envhub/internal/cli/banner"
	"github.com/Now-Tiger/envhub/internal/cli/config"
	"github.com/Now-Tiger/envhub/internal/cli/style"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Global flags
	cfgFile    string
	verbose    bool
	noColor    bool
	outputJSON bool
)

// NewRootCmd returns the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "envhub",
		Short: "EnvHub CLI - Environment Variable Management",
		Long:  banner.EnvHubBanner + "\n\n" + "EnvHub CLI - Environment Variable Management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg := config.Get()
			if err := cfg.Load(cfgFile); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Apply global flags
			cfg.SetVerbose(verbose)
			cfg.SetNoColor(noColor)

			return nil
		},
		SilenceUsage: true,
	}

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.envhub/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&outputJSON, "output", "o", false, "output in JSON format")

	// Bind viper keys
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Add subcommands
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(pullCmd())
	rootCmd.AddCommand(pushCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(whoamiCmd())
	rootCmd.AddCommand(logoutCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

// Execute runs the root command
func Execute(version string) error {
	rootCmd := NewRootCmd()
	rootCmd.Version = version
	return rootCmd.Execute()
}

// printError prints a styled error message
func printError(msg string) {
	fmt.Fprintln(os.Stderr, style.Error.Render("✗ ")+msg)
}

// printSuccess prints a styled success message
func printSuccess(msg string) {
	fmt.Println(style.Success.Render("✓ ") + msg)
}

// printInfo prints a styled info message
func printInfo(msg string) {
	fmt.Println(style.Info.Render("ℹ ") + msg)
}

// printWarning prints a styled warning message
func printWarning(msg string) {
	fmt.Println(style.Warning.Render("⚠ ") + msg)
}

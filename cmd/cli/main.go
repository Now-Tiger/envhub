package main

import (
	"os"

	"github.com/Now-Tiger/envhub/cmd/cli/cmd"
	"github.com/Now-Tiger/envhub/internal/cli/config"
)

// version is set via ldflags during build
var version = "dev"

func main() {
	// Initialize config before executing commands
	config.Init()

	if err := cmd.Execute(version); err != nil {
		if err != config.ErrNoToken {
			os.Exit(1)
		}
		os.Exit(0)
	}
}

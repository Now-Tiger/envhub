package banner

import (
	"github.com/charmbracelet/lipgloss"
)

// EnvHubBanner is the ASCII art banner for EnvHub CLI
var EnvHubBanner = lipgloss.NewStyle().
	Foreground(lipgloss.Color("86")).
	Bold(true).
	Render(`
██████╗ ██╗██████╗ ███████╗██╗     ██╗███╗   ██╗███████╗
██╔══██╗██║██╔══██╗██╔════╝██║     ██║████╗  ██║██╔════╝
██████╔╝██║██████╔╝█████╗  ██║     ██║██╔██╗ ██║█████╗  
██╔══██╗██║██╔══██╗██╔══╝  ██║     ██║██║╚██╗██║██╔══╝  
██║  ██║██║██║  ██║███████╗███████╗██║██║ ╚████║███████╗
╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝
`)

// CompactHeader returns a compact 1-line header
func CompactHeader() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Render("▶ EnvHub")
}

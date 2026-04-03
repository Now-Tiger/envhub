package style

import "github.com/charmbracelet/lipgloss"

// Global styles for CLI output
var (
	// Header is used for main headers
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	// Subheader is used for subheaders
	Subheader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75"))

	// Success is used for success messages
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	// Error is used for error messages
	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	// Warning is used for warning messages
	Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	// Info is used for info messages
	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("75"))

	// Dim is used for dimmed/hint text
	Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color("242"))

	// Code is used for code/values
	Code = lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Background(lipgloss.Color("236"))
)

// Table styles
var (
	// TableHeader is used for table headers
	TableHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	// TableRow is used for table rows
	TableRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	// TableAltRow is used for alternating table rows
	TableAltRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))
)

// List styles
var (
	// ListBullet is the bullet character for lists
	ListBullet = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Render("•")

	// ListBulletDim is a dimmed bullet
	ListBulletDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			Render("•")
)

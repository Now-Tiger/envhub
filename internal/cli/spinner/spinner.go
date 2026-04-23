package spinner

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

// Spinner provides lazy-loaded spinner functionality
type Spinner struct {
	initialized bool
	spinnerType string
}

// New creates a new Spinner instance (lazy loader)
func New() *Spinner {
	return &Spinner{
		initialized: false,
		spinnerType: "line",
	}
}

// Start begins the spinner animation (lazy initialization)
func (s *Spinner) Start() {
	// This is a placeholder - actual spinner would be initialized on first use
	// to avoid loading bubbles library at startup
	s.initialized = true
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.initialized = false
}

// Render prints a simple spinner to the given writer
func (s *Spinner) Render(w io.Writer, message string) {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for i := 0; i < len(spinners); i++ {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
		fmt.Fprintf(w, "\r%s %s", style.Render(spinners[i]), message) //nolint:errcheck
	}
}

// NewLoadingStyle returns a lipgloss style for loading states
func NewLoadingStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))
}

// NewSuccessStyle returns a lipgloss style for success
func NewSuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))
}

// NewErrorStyle returns a lipgloss style for errors
func NewErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))
}

// Package style provides centralized, reusable terminal styles for infra-cli.
// All colors and formatting are defined here so individual commands stay clean
// and styling remains consistent across the entire CLI.
package style

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Base color palette used across all styles.
var (
	colorGreen  = lipgloss.Color("#04B575")
	colorRed    = lipgloss.Color("#FF4C4C")
	colorYellow = lipgloss.Color("#FFCC00")
	colorCyan   = lipgloss.Color("#00D7FF")
	colorGray   = lipgloss.Color("#6C6C6C")
)

// Reusable styles for different output contexts.
var (
	Success = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	Info    = lipgloss.NewStyle().Foreground(colorCyan)
	Subtle  = lipgloss.NewStyle().Foreground(colorGray)
	Header  = lipgloss.NewStyle().
		Foreground(colorCyan).
		Bold(true).
		Underline(true)
)

// PrintSuccess writes a green, bold message to stdout.
func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(Success.Render(msg))
}

// PrintError writes a red, bold message to stderr.
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, Error.Render(msg))
}

// PrintWarning writes a yellow, bold message to stderr.
func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, Warning.Render(msg))
}

// PrintInfo writes a cyan message to stdout.
func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(Info.Render(msg))
}

// PrintSubtle writes a gray, dimmed message to stdout.
func PrintSubtle(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(Subtle.Render(msg))
}

// PrintHeader writes an underlined, bold cyan header to stdout.
func PrintHeader(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(Header.Render(msg))
}

// Divider returns a styled horizontal rule for section separation.
func Divider() string {
	return Subtle.Render("──────────────────────────────────────────────────")
}

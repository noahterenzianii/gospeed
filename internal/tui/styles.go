package tui

import "github.com/charmbracelet/lipgloss"

// Colors from the mockup palette
var (
	cWhite  = lipgloss.Color("#f0f0f0")
	cCyan   = lipgloss.Color("#56d8ff")
	cMuted  = lipgloss.Color("#555555")
	cScreen = lipgloss.Color("#c9c9c9")
)

// Shared lipgloss styles
var (
	mutedStyle = lipgloss.NewStyle().Foreground(cMuted)
	asciiStyle = lipgloss.NewStyle().Foreground(cScreen)
)

var asciiArt = []string{
	`  ██████╗  ██████╗ ███████╗██████╗ ███████╗███████╗██████╗ `,
	` ██╔════╝ ██╔═══██╗██╔════╝██╔══██╗██╔════╝██╔════╝██╔══██╗`,
	` ██║  ███╗██║   ██║███████╗██████╔╝█████╗  █████╗  ██║  ██║`,
	` ██║   ██║██║   ██║╚════██║██╔═══╝ ██╔══╝  ██╔══╝  ██║  ██║`,
	` ╚██████╔╝╚██████╔╝███████║██║     ███████╗███████╗██████╔╝`,
	`  ╚═════╝  ╚═════╝ ╚══════╝╚═╝     ╚══════╝╚══════╝╚═════╝ `,
}

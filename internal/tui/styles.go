package tui

import "github.com/charmbracelet/lipgloss"

// Colors from the mockup palette
var (
	cWhite  = lipgloss.Color("#f0f0f0")
	cCyan   = lipgloss.Color("#56d8ff")
	cGreen  = lipgloss.Color("#3dd68c")
	cYellow = lipgloss.Color("#ffd166")
	cRed    = lipgloss.Color("#ff6b6b")
	cPurple = lipgloss.Color("#c792ea")
	cMuted  = lipgloss.Color("#555555")
	cDim    = lipgloss.Color("#444444")
	cScreen = lipgloss.Color("#c9c9c9")
)

// Shared lipgloss styles
var (
	mutedStyle = lipgloss.NewStyle().Foreground(cMuted)
	dimStyle   = lipgloss.NewStyle().Foreground(cDim)
)

const labelWidth = 12

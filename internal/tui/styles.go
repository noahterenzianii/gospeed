package tui

import "github.com/charmbracelet/lipgloss"

// Accent colors — adaptive to background for maximum contrast
var (
	accentCyan   = lipgloss.AdaptiveColor{Light: "#0077aa", Dark: "#00d4ff"}
	accentGreen  = lipgloss.AdaptiveColor{Light: "#00884a", Dark: "#00e676"}
	accentYellow = lipgloss.AdaptiveColor{Light: "#996600", Dark: "#ffd600"}
	accentRed    = lipgloss.AdaptiveColor{Light: "#cc0033", Dark: "#ff5252"}
	accentPurple = lipgloss.AdaptiveColor{Light: "#7744aa", Dark: "#b388ff"}
	accentOrange = lipgloss.AdaptiveColor{Light: "#b35800", Dark: "#ffab40"}
)

// Text colors — adaptive to background for maximum contrast
var (
	textPrimary   = lipgloss.AdaptiveColor{Light: "#1a1a2e", Dark: "#e8e8e8"}
	textSecondary = lipgloss.AdaptiveColor{Light: "#44445a", Dark: "#b0b1c8"}
	textMuted     = lipgloss.AdaptiveColor{Light: "#66667a", Dark: "#8888a0"}
	textDim       = lipgloss.AdaptiveColor{Light: "#9999aa", Dark: "#666680"}
)

var (
	mutedStyle = lipgloss.NewStyle().Foreground(textMuted)
	dimStyle   = lipgloss.NewStyle().Foreground(textDim)
)

const labelWidth = 10

func addField(label, value string, color lipgloss.TerminalColor) string {
	labelS := lipgloss.NewStyle().Foreground(textSecondary).Width(labelWidth).Render(label)
	valueS := lipgloss.NewStyle().Foreground(color).Render(value)
	return "  " + labelS + valueS
}

func sectionHeader(title string, accent lipgloss.TerminalColor) string {
	return lipgloss.NewStyle().Foreground(accent).Bold(true).Render("  " + title)
}

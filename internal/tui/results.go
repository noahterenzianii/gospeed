package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) resultsView() string {
	if m.latency == nil || m.download == nil || m.upload == nil {
		return ""
	}

	var rows []string
	rows = append(rows, sectionHeader("results", accentPurple))

	dl := lipgloss.NewStyle().Foreground(accentCyan).Render(
		fmt.Sprintf("↓ %s Mbps", fmtSpeed(m.download.Speed)))
	ul := lipgloss.NewStyle().Foreground(accentYellow).Render(
		fmt.Sprintf("↑ %s Mbps", fmtSpeed(m.upload.Speed)))
	rows = append(rows, fmt.Sprintf("  %s  %s", dl, ul))

	ping := lipgloss.NewStyle().Foreground(accentGreen).Render(
		fmtDuration(m.latency.Ping))
	jit := lipgloss.NewStyle().Foreground(accentYellow).Render(
		fmtDuration(m.latency.Jitter))
	rows = append(rows, fmt.Sprintf("  %s %s  ·  %s %s",
		mutedStyle.Render("ping"), ping,
		mutedStyle.Render("jitter"), jit))

	totalDur := m.download.Elapsed + m.upload.Elapsed
	rows = append(rows, addField("duration", fmt.Sprintf("%.1f s", totalDur.Seconds()), textMuted))

	return strings.Join(rows, "\n")
}

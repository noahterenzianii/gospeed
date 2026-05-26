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

	dl := lipgloss.NewStyle().Foreground(cCyan).Render(
		fmt.Sprintf("↓ %.0f Mbps", m.download.Speed),
	)
	ul := lipgloss.NewStyle().Foreground(cYellow).Render(
		fmt.Sprintf("↑ %.0f Mbps", m.upload.Speed),
	)
	rows = append(rows, fmt.Sprintf("  %s  %s", dl, ul))

	ping := lipgloss.NewStyle().Foreground(cGreen).Render(
		"◈ " + fmtDuration(m.latency.Ping),
	)
	jitter := lipgloss.NewStyle().Foreground(cYellow).Render(
		fmtDuration(m.latency.Jitter),
	)
	rows = append(rows, fmt.Sprintf("  %s %s  •  %s %s",
		mutedStyle.Render("ping"), ping,
		mutedStyle.Render("jitter"), jitter,
	))

	totalDur := m.download.Elapsed + m.upload.Elapsed
	rows = addField(rows, "durata", fmt.Sprintf("%.1f s", totalDur.Seconds()), cMuted)

	return strings.Join(rows, "\n")
}

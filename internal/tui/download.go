package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatBytes converts bytes to a human-readable string (KB, MB, GB).
func formatBytes(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	case b < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(b)/(1024*1024*1024))
	}
}

// downloadView renders the current speed, progress percentage, and sparkline.
func downloadView(d *DownloadState) string {
	var rows []string

	pct := d.Elapsed.Seconds() / downloadDuration.Seconds()
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}

	arrow := lipgloss.NewStyle().Foreground(cCyan).Render("↓")
	val := lipgloss.NewStyle().Foreground(cCyan).Bold(true).Render(fmt.Sprintf("%.0f", d.Speed))
	pctS := lipgloss.NewStyle().Foreground(cCyan).Render(fmt.Sprintf("%.0f%%", pct*100))
	rows = append(rows, fmt.Sprintf("  %s %s%s  %s  %s", arrow, val, mutedStyle.Render(" Mbps"), dimStyle.Render("download"), pctS))

	rows = addField(rows, "received", formatBytes(int64(d.Speed*d.Elapsed.Seconds()*mbpsToBytes)), cMuted)
	rows = addField(rows, "elapsed", fmt.Sprintf("%.1f s", d.Elapsed.Seconds()), cMuted)

	if len(d.Samples) > 0 {
		rows = addField(rows, "samples", renderSparkline(d.Samples), cCyan)
	}

	return strings.Join(rows, "\n")
}

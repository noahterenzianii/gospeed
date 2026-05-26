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

// transferView renders the current speed, progress percentage, and sparkline.
func transferView(s *TransferState, dir direction) string {
	var arrowStr, labelStr, byteLabel string
	var color lipgloss.Color

	switch dir {
	case dirDownload:
		arrowStr = "↓"
		color = cCyan
		labelStr = "download"
		byteLabel = "received"
	case dirUpload:
		arrowStr = "↑"
		color = cYellow
		labelStr = "upload"
		byteLabel = "sent"
	}

	var rows []string

	pct := s.Elapsed.Seconds() / transferDuration.Seconds()
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}

	arrow := lipgloss.NewStyle().Foreground(color).Render(arrowStr)
	val := lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("%.0f", s.Speed))
	pctS := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%.0f%%", pct*100))
	rows = append(rows, fmt.Sprintf("  %s %s%s  %s  %s", arrow, val, mutedStyle.Render(" Mbps"), dimStyle.Render(labelStr), pctS))

	rows = addField(rows, byteLabel, formatBytes(int64(s.Speed*s.Elapsed.Seconds()*mbpsToBytes)), cMuted)
	rows = addField(rows, "elapsed", fmt.Sprintf("%.1f s", s.Elapsed.Seconds()), cMuted)

	if len(s.Samples) > 0 {
		rows = addField(rows, "samples", renderSparkline(s.Samples), color)
	}

	return strings.Join(rows, "\n")
}

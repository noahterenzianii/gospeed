package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

func fmtDuration(d time.Duration) string {
	ms := float64(d) / float64(time.Millisecond)
	if ms < 1 {
		return fmt.Sprintf("%.1f ms", ms)
	}
	return fmt.Sprintf("%d ms", int(ms))
}

func pingView(latency *endpoints.Latency) string {
	var rows []string
	rows = append(rows, sectionHeader("latency", accentGreen))
	rows = append(rows, fmt.Sprintf("  %s  %s  ·  %s  %s",
		mutedStyle.Render("ping"),
		lipgloss.NewStyle().Foreground(accentGreen).Bold(true).Render(fmtDuration(latency.Ping)),
		mutedStyle.Render("jitter"),
		lipgloss.NewStyle().Foreground(accentYellow).Render(fmtDuration(latency.Jitter))))
	if len(latency.Samples) > 0 {
		f64 := make([]float64, len(latency.Samples))
		for i, d := range latency.Samples {
			f64[i] = float64(d)
		}
		rows = append(rows, "  "+lipgloss.NewStyle().Foreground(accentGreen).Render(renderSparkline(f64)))
	}
	return strings.Join(rows, "\n")
}

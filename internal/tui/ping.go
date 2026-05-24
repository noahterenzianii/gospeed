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

	latLabel := labelStyle.Render("latency")
	latVal := lipgloss.NewStyle().Foreground(cGreen).Bold(true).Render(
		fmtDuration(latency.Ping),
	)
	rows = append(rows, fmt.Sprintf("  %s%s", latLabel, latVal))

	jitLabel := labelStyle.Render("jitter")
	jitVal := lipgloss.NewStyle().Foreground(cYellow).Render(
		fmtDuration(latency.Jitter),
	)
	rows = append(rows, fmt.Sprintf("  %s%s", jitLabel, jitVal))

	if len(latency.Samples) > 0 {
		sampleLabel := labelStyle.Render("samples")
		f64 := make([]float64, len(latency.Samples))
		for i, d := range latency.Samples {
			f64[i] = float64(d)
		}
		spark := lipgloss.NewStyle().Foreground(cGreen).Render(
			renderSparkline(f64),
		)
		rows = append(rows, fmt.Sprintf("  %s%s", sampleLabel, spark))
	}

	return strings.Join(rows, "\n")
}

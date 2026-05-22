package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

const sparkWidth = 30

func downsample(samples []time.Duration) []time.Duration {
	if len(samples) <= sparkWidth {
		return samples
	}
	bucketSize := len(samples) / sparkWidth
	downsampled := make([]time.Duration, sparkWidth)
	for i := range sparkWidth {
		start := i * bucketSize
		bucket := samples[start : start+bucketSize]
		var total time.Duration
		for _, s := range bucket {
			total += s
		}
		downsampled[i] = total / time.Duration(bucketSize)
	}
	return downsampled
}

func renderSparkline(samples []time.Duration) string {
	samples = downsample(samples)
	if len(samples) == 0 {
		return ""
	}

	min, max := samples[0], samples[0]
	for _, s := range samples[1:] {
		if s < min {
			min = s
		}
		if s > max {
			max = s
		}
	}

	bars := []rune("▁▂▃▄▅▆▇█")
	var sb strings.Builder
	for _, s := range samples {
		idx := 3
		if max > min {
			idx = int((s - min) * 7 / (max - min))
			if idx > 7 {
				idx = 7
			} else if idx < 0 {
				idx = 0
			}
		}
		sb.WriteRune(bars[idx])
	}
	return sb.String()
}

func fmtDuration(d time.Duration) string {
	ms := float64(d) / float64(time.Millisecond)
	if ms < 1 {
		return fmt.Sprintf("%.1f ms", ms)
	}
	return fmt.Sprintf("%d ms", int(ms))
}

func pingView(latency *endpoints.Latency) string {
	var rows []string

	rows = append(rows, dimStyle.Render("  ping"))

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
		spark := lipgloss.NewStyle().Foreground(cGreen).Render(
			renderSparkline(latency.Samples),
		)
		rows = append(rows, fmt.Sprintf("  %s%s", sampleLabel, spark))
	}

	return strings.Join(rows, "\n")
}

package tui

import (
	"strings"
)

const (
	sparkWidth   = 30
	sampleWindow = 100
)

func renderSparkline(samples []float64) string {
	if len(samples) == 0 {
		return ""
	}

	ds := samples
	if len(samples) > sampleWindow {
		ds = samples[len(samples)-sampleWindow:]
	}

	if len(ds) > sparkWidth {
		n := len(ds)
		out := make([]float64, sparkWidth)
		for i := range sparkWidth {
			start := i * n / sparkWidth
			end := (i + 1) * n / sparkWidth
			bucket := ds[start:end]
			if len(bucket) == 0 {
				continue
			}
			var total float64
			for _, s := range bucket {
				total += s
			}
			out[i] = total / float64(len(bucket))
		}
		ds = out
	}

	min, max := ds[0], ds[0]
	for _, s := range ds[1:] {
		if s < min {
			min = s
		}
		if s > max {
			max = s
		}
	}

	bars := []rune("▁▂▃▄▅▆▇█")
	var sb strings.Builder
	for _, s := range ds {
		idx := 7
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

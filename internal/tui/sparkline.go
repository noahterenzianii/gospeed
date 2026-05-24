package tui

import (
	"strings"
)

const sparkWidth = 30

// renderSparkline downsamples data to sparkWidth buckets and renders a unicode bar chart.
func renderSparkline(samples []float64) string {
	if len(samples) == 0 {
		return ""
	}

	ds := samples
	if len(samples) > sparkWidth {
		bucketSize := len(samples) / sparkWidth
		ds = make([]float64, sparkWidth)
		for i := range sparkWidth {
			start := i * bucketSize
			bucket := samples[start : start+bucketSize]
			var total float64
			for _, s := range bucket {
				total += s
			}
			ds[i] = total / float64(bucketSize)
		}
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



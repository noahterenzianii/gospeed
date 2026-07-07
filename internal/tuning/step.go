package tuning

import (
	"fmt"
	"sync"
	"time"
)

func runMeasurement(
	url string,
	streams, bufSize int,
	duration time.Duration,
	measure MeasureFunc,
) (float64, float64, error) {
	var mu sync.Mutex
	var samples []float64

	_, err := measure(url, duration, streams, bufSize, func(mbps float64) {
		mu.Lock()
		samples = append(samples, mbps)
		mu.Unlock()
	})

	mu.Lock()
	defer mu.Unlock()

	if len(samples) == 0 {
		return 0, 0, fmt.Errorf("no samples collected: %w", err)
	}

	start := len(samples) / 5
	stable := samples[start:]
	if len(stable) == 0 {
		stable = samples
	}

	avg := mean(stable)
	return avg, variance(stable, avg), nil
}

func mean(samples []float64) float64 {
	var sum float64
	for _, v := range samples {
		sum += v
	}
	return sum / float64(len(samples))
}

func variance(samples []float64, mean float64) float64 {
	if len(samples) < 2 {
		return 0
	}
	var sum float64
	for _, v := range samples {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(samples)-1)
}

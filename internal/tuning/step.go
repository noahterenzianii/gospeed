package tuning

import (
	"fmt"
	"math"
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

	finalMbps, err := measure(url, duration, streams, bufSize, func(mbps float64) {
		mu.Lock()
		samples = append(samples, mbps)
		mu.Unlock()
	})

	mu.Lock()
	defer mu.Unlock()

	if len(samples) == 0 {
		return 0, 0, fmt.Errorf("no samples collected: %w", err)
	}
	if finalMbps <= 0 {
		return 0, 0, fmt.Errorf("zero final throughput: %w", err)
	}

	start := len(samples) * 40 / 100
	if start >= len(samples) {
		start = len(samples) / 2
	}
	stable := samples[start:]

	v := variance(stable, finalMbps)
	return finalMbps, v, nil
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

type measuredResult struct {
	throughput float64
	variance   float64
}

func measureWithQuality(
	url string,
	streams, bufSize int,
	duration time.Duration,
	measure MeasureFunc,
) (measuredResult, error) {
	result, err := tryMeasure(url, streams, bufSize, duration, measure)
	if err != nil {
		return result, err
	}

	if result.variance > 0 && result.throughput > 0 {
		cv := math.Sqrt(result.variance) / result.throughput
		if cv > 0.30 {
			result2, err2 := tryMeasure(url, streams, bufSize, duration, measure)
			if err2 == nil && result2.throughput > result.throughput {
				result = result2
			}
		}
	}

	return result, nil
}

func tryMeasure(
	url string,
	streams, bufSize int,
	duration time.Duration,
	measure MeasureFunc,
) (measuredResult, error) {
	throughput, v, err := runMeasurement(url, streams, bufSize, duration, measure)
	if err != nil {
		return measuredResult{}, err
	}
	if throughput <= 0 {
		return measuredResult{}, fmt.Errorf("zero throughput for streams=%d buf=%d", streams, bufSize)
	}
	return measuredResult{throughput: throughput, variance: v}, nil
}

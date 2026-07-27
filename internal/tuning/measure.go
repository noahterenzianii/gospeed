package tuning

import (
	"fmt"
	"math"
	"time"
)

func measureWithWarmup(
	url string,
	streams, bufSize int,
	duration time.Duration,
	measure MeasureFunc,
	warmupPercent int,
) (float64, float64, error) {
	var samples []float64

	finalMbps, err := measure(url, duration, streams, bufSize, func(mbps float64) {
		samples = append(samples, mbps)
	})

	if len(samples) == 0 {
		if err != nil {
			return 0, 0, fmt.Errorf("no samples collected: %w", err)
		}
		return 0, 0, fmt.Errorf("no samples collected: zero throughput")
	}

	start := len(samples) * warmupPercent / 100
	if start >= len(samples) {
		start = len(samples) / 2
	}
	stable := samples[start:]

	var sum float64
	for _, v := range stable {
		sum += v
	}
	mean := sum / float64(len(stable))

	if mean <= 0 {
		if finalMbps > 0 {
			mean = finalMbps
		} else {
			return 0, 0, fmt.Errorf("zero throughput")
		}
	}

	v := variance(stable, mean)
	return mean, v, nil
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
	opts Options,
) (measuredResult, error) {
	warmup := opts.warmupPercent()

	result, err := tryMeasure(url, streams, bufSize, duration, measure, warmup)
	if err != nil {
		return result, err
	}

	maxCV := opts.maxCV()
	maxRetries := opts.maxRetries()

	if result.throughput > 0 && result.variance > 0 {
		cv := math.Sqrt(result.variance) / result.throughput
		if cv > maxCV {
			bestResult := result
			bestCV := cv
			for i := 0; i < maxRetries; i++ {
				r, err2 := tryMeasure(url, streams, bufSize, duration, measure, warmup)
				if err2 != nil {
					continue
				}
				cv2 := math.Sqrt(r.variance) / r.throughput
				if cv2 < bestCV {
					bestResult = r
					bestCV = cv2
				}
				if cv2 <= maxCV {
					break
				}
			}
			result = bestResult
		}
	}

	return result, nil
}

func tryMeasure(
	url string,
	streams, bufSize int,
	duration time.Duration,
	measure MeasureFunc,
	warmupPercent int,
) (measuredResult, error) {
	throughput, v, err := measureWithWarmup(url, streams, bufSize, duration, measure, warmupPercent)
	if err != nil {
		return measuredResult{}, err
	}
	if throughput <= 0 {
		return measuredResult{}, fmt.Errorf("zero throughput for streams=%d buf=%d", streams, bufSize)
	}
	return measuredResult{throughput: throughput, variance: v}, nil
}

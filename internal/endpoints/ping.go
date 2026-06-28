package endpoints

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

func pingServer(url string, timeout time.Duration) (time.Duration, error) {
	client := &http.Client{Timeout: timeout}
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("ping %s: %w", url, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return time.Since(start), nil
}

func MeasureLatency(url string, count int, timeout time.Duration) (Latency, error) {
	if count < 2 {
		return Latency{}, fmt.Errorf("count must be >= 2")
	}

	// warmup
	if _, err := pingServer(url, timeout); err != nil {
		return Latency{}, fmt.Errorf("warmup ping: %w", err)
	}

	samples := make([]time.Duration, 0, count)
	for range count {
		lat, err := pingServer(url, timeout)
		if err != nil {
			return Latency{}, fmt.Errorf("ping sample: %w", err)
		}
		samples = append(samples, lat)
	}

	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	slices.Sort(sorted)
	ping := sorted[count/2]

	var sumDiff time.Duration
	for i := 1; i < len(samples); i++ {
		diff := samples[i] - samples[i-1]
		if diff < 0 {
			diff = -diff
		}
		sumDiff += diff
	}
	jitter := sumDiff / time.Duration(len(samples)-1)
	return Latency{Ping: ping, Jitter: jitter, Samples: samples}, nil
}

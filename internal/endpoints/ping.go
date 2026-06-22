package endpoints

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

func PingServer(url string, timeout time.Duration) (time.Duration, error) {
	client := &http.Client{Timeout: timeout}
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
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
	if _, err := PingServer(url, timeout); err != nil {
		return Latency{}, err
	}

	samples := make([]time.Duration, 0, count)
	for range count {
		lat, err := PingServer(url, timeout)
		if err != nil {
			return Latency{}, err
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

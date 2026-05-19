package endpoints

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

var client = &http.Client{Timeout: 2 * time.Second}

func PingServer(url string) (time.Duration, error) {
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return time.Since(start), nil
}

func MeasureLatency(url string, count int) (Latency, error) {
	if count < 2 {
		return Latency{}, fmt.Errorf("count must e >= 2")
	}

	// warmup
	if _, err := PingServer(url); err != nil {
		return Latency{}, err
	}

	samples := make([]float64, 0, count)
	for range count {
		lat, err := PingServer(url)
		if err != nil {
			return Latency{}, err
		}
		samples = append(samples, float64(lat))
	}

	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)
	ping := time.Duration(sorted[count/2])

	var sumDiff float64
	for i := 1; i < len(samples); i++ {
		diff := samples[i] - samples[i-1]
		if diff < 0 {
			diff = -diff
		}
		sumDiff += diff
	}
	jitter := time.Duration(sumDiff / float64(len(samples)-1))
	return Latency{ping, jitter}, nil
}

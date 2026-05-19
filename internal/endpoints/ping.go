package endpoints

import (
	"io"
	"net/http"
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

func MeasureJitter(url string, count int) (time.Duration, error) {
	var prev time.Duration
	var jitter float64
	for i := range count {
		lat, err := PingServer(url)
		if err != nil {
			return 0, err
		}
		if i > 0 {
			diff := float64(lat - prev)
			if diff < 0 {
				diff = -diff
			}
			jitter += (diff - jitter) / 16
		}
		prev = lat
	}
	return time.Duration(jitter), nil
}

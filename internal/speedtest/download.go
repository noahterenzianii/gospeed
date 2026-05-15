package speedtest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func MeasureDownload(url string, duration time.Duration, streams int) (float64, error) {
	if streams < 1 {
		streams = 1
	}

	// Size the connection pool to match the number of parallel streams.
	// Client timeout is slightly longer than the test to let requests finish cleanly.
	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        streams * 2,
			MaxIdleConnsPerHost: streams * 2,
			IdleConnTimeout:     15 * time.Second,
		},
		Timeout: duration + 2*time.Second,
	}

	// All goroutines share this context; when it expires, downloads stop.
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	start := time.Now()
	var totalBytes atomic.Int64
	var completedRuns atomic.Int64
	var wg sync.WaitGroup

	for workerID := 0; workerID < streams; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			buf := make([]byte, 256*1024) // reused across requests to avoid GC pressure
			run := 0
			for {
				if ctx.Err() != nil {
					return
				}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, withCacheBuster(url, id, run), nil)
				if err != nil {
					return
				}
				resp, err := client.Do(req)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					continue
				}
				if resp.StatusCode != http.StatusOK {
					resp.Body.Close()
					if ctx.Err() != nil {
						return
					}
					continue
				}
				n, err := io.CopyBuffer(io.Discard, resp.Body, buf)
				resp.Body.Close()
				if n > 0 {
					totalBytes.Add(n)
					completedRuns.Add(1)
				}
				if err != nil && ctx.Err() == nil {
					continue
				}
				run++
			}
		}(workerID)
	}
	wg.Wait()
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0, fmt.Errorf("invalid elapsed time")
	}
	if completedRuns.Load() == 0 {
		return 0, fmt.Errorf("no successful download request")
	}
	mbps := (float64(totalBytes.Load()) * 8) / elapsed / 1_000_000
	return mbps, nil
}

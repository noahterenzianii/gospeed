package speedtest

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func MeasureUpload(url string, duration time.Duration, streams int, onProgress ProgressFunc) (float64, error) {
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

	// All goroutines share this context; when it expires, uploads stop.
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Pre-generate a random data buffer so each goroutine sends unpredictable payloads.
	// Random data prevents proxies or the server from compressing the stream, which would
	chunkSize := 256 * 1024
	payload := make([]byte, chunkSize)
	if _, err := rand.Read(payload); err != nil {
		return 0, fmt.Errorf("failed to generate random data: %w", err)
	}

	start := time.Now()
	var totalBytes atomic.Int64
	var completedRuns atomic.Int64
	var wg sync.WaitGroup

	for workerID := 0; workerID < streams; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				// Each iteration gets a fresh reader so the buffer can be reused safely.
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
				if err != nil {
					return
				}
				req.Header.Set("Content-Type", "application/octet-stream")
				resp, err := client.Do(req)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					totalBytes.Add(int64(chunkSize))
					completedRuns.Add(1)
				}
				if ctx.Err() != nil {
					return
				}
			}
		}(workerID)
	}
	_, stopProgress := startProgress(&totalBytes, start, onProgress)
	defer stopProgress()
	wg.Wait()
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0, fmt.Errorf("invalid elapsed time")
	}
	if completedRuns.Load() == 0 {
		return 0, fmt.Errorf("no successful upload request")
	}
	mbps := (float64(totalBytes.Load()) * 8) / elapsed / 1_000_000
	return mbps, nil
}

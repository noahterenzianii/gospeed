package speedtest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func runMeasurement(duration time.Duration,
	streams int,
	bufSize int,
	onProgress ProgressFunc,
	direction string,
	worker func(ctx context.Context,
		client *http.Client,
		id int, totalBytes,
		completedRuns *atomic.Int64),
) (float64, error) {

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
			//upload and download callback
			worker(ctx, &client, id, &totalBytes, &completedRuns)
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
		return 0, fmt.Errorf("no successful %s request", direction)
	}
	mbps := (float64(totalBytes.Load()) * 8) / elapsed / 1_000_000
	return mbps, nil

}

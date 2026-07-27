package speedtest

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type ProgressFunc func(currentMbps float64)

func startProgress(totalBytes *atomic.Int64, start time.Time, onProgress ProgressFunc) func() {
	progressCtx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(start).Seconds()
				if elapsed > 0 && onProgress != nil {
					current := (float64(totalBytes.Load()) * 8) / elapsed / 1_000_000
					onProgress(current)
				}
			case <-progressCtx.Done():
				return
			}
		}
	}()
	return func() {
		cancel()
		wg.Wait()
	}
}

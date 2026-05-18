package speedtest

import (
	"context"
	"sync/atomic"
	"time"
)

type ProgressFunc func(currentMbps float64)

func startProgress(totalBytes *atomic.Int64, start time.Time, onProgress ProgressFunc) (context.Context, context.CancelFunc) {
	progressCtx, stopProgress := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
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
	return progressCtx, stopProgress
}

package speedtest

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

func MeasureDownload(url string, duration time.Duration, streams int, onProgress ProgressFunc) (float64, error) {
	//callback
	return runMeasurement(duration, streams, onProgress, "download",
		func(ctx context.Context, client *http.Client, id int, totalBytes, completedRuns *atomic.Int64) {
			buf := make([]byte, bufferSize)
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
		})
}

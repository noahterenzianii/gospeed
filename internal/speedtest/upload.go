package speedtest

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

func MeasureUpload(url string, duration time.Duration, streams int, bufSize int, onProgress ProgressFunc) (float64, error) {
	payload := make([]byte, bufSize)
	if _, err := rand.Read(payload); err != nil {
		return 0, fmt.Errorf("failed to generate random data: %w", err)
	}
	return runMeasurement(duration, streams, onProgress, "upload",
		func(ctx context.Context, client *http.Client, id int, totalBytes, completedRuns *atomic.Int64) {
			for {
				if ctx.Err() != nil {
					return
				}
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
					totalBytes.Add(int64(bufSize))
					completedRuns.Add(1)
				}
				if ctx.Err() != nil {
					return
				}

			}
		})
}

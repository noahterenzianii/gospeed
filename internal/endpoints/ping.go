package endpoints

import (
	"io"
	"net/http"
	"time"
)

func PingServer(url string) (time.Duration, error) {
	client := http.Client{Timeout: 2 * time.Second}
	var best time.Duration

	for i := range 3 {
		start := time.Now()
		resp, err := client.Get(url)
		if err != nil {
			return 0, err
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		lat := time.Since(start)
		if i == 0 || lat < best {
			best = lat
		}
	}
	return best, nil
}

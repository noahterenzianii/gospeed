package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	ServerURL string        `json:"server"`
	DlURL     string        `json:"dlURL"`
	UlURL     string        `json:"ulURL"`
	PingURL   string        `json:"pingURL"`
	Latency   time.Duration `json:"-"`
}

func fetchServers(url string) ([]Server, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: %d", resp.StatusCode)
	}
	var server []Server
	err = json.NewDecoder(resp.Body).Decode(&server)
	return server, nil
}

func pingServer(url string) (time.Duration, error) {
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
		if i == 0 || time.Since(start) < best {
			best = time.Since(start)
		}
	}
	return best, nil
}

func findBestServer(servers []Server) *Server {
	sem := make(chan struct{}, 20) // Create a channel to handle 20 concurrent requests
	var wg sync.WaitGroup          // Track the number of active connections
	for i := range servers {
		wg.Add(1)
		sem <- struct{}{} // Occupy a slot in the semaphore: struct{} is used as a zero-memory token
		go func(server *Server) {
			defer wg.Done()          // Decrement counter when the request is finished
			defer func() { <-sem }() // Release the slot back to the semaphore
			pingUrl := strings.TrimRight(server.ServerURL, "/") + "/" + strings.TrimLeft(server.PingURL, "/")
			lat, err := pingServer(pingUrl)
			if err != nil {
				// Assign maximum latency if the server is unreachable
				server.Latency = time.Duration(math.MaxInt64)
				return
			}
			server.Latency = lat
		}(&servers[i])
	}
	wg.Wait()

	// Sort the slice to find the server with the lowest latency
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Latency < servers[j].Latency
	})

	// If no servers exist or all failed (max latency), return nil
	if len(servers) == 0 || servers[0].Latency == time.Duration(math.MaxInt64) {
		return nil
	}
	return &servers[0]
}

// make sure that every respone isnt cached
func withCacheBuster(url string, workerID int, run int) string {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + "n=" + strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + strconv.Itoa(workerID) + "_" + strconv.Itoa(run)
}

func measureDownload(url string, duration time.Duration, streams int) (float64, error) {
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

func main() {
	url := "https://librespeed.org/backend-servers/servers.php"
	servers, err := fetchServers(url)
	if err != nil {
		panic(err)
	}
	best := findBestServer(servers)
	if best == nil {
		fmt.Println("No server reachable")
		return
	}
	fmt.Printf("Best Server: %v\n", best.Latency)
	downloadURL := strings.TrimRight(best.ServerURL, "/") + "/" + strings.TrimLeft(best.DlURL, "/")
	downloadSpeed, err := measureDownload(downloadURL, 10*time.Second, 4)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Download speed: %.2f Mbit/s\n", downloadSpeed)
}

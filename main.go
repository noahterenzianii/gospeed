package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
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
			return 0, err // fallisce solo se tutti falliscono
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

			lat, err := pingServer(strings.TrimRight(server.ServerURL, "/") + "/" + strings.TrimLeft(server.PingURL, "/"))
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
	fmt.Printf("Best Server: %v", best.Latency)
}

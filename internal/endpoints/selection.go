package endpoints

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

func FindBestServer(servers []Server) *Server {
	if len(servers) == 0 {
		return nil
	}

	sem := make(chan struct{}, 20) // Create a channel to handle 20 concurrent requests
	var wg sync.WaitGroup          // Track the number of active connections

	for i := range servers {
		wg.Add(1)
		sem <- struct{}{} // Occupy a slot in the semaphore: struct{} is used as a zero-memory token

		go func(s *Server) {
			defer wg.Done()          // Decrement counter when the request is finished
			defer func() { <-sem }() // Release the slot back to the semaphore

			pingURL := strings.TrimRight(s.ServerURL, "/") + "/" + strings.TrimLeft(s.PingURL, "/")
			success := 0
			var best time.Duration
			for _ = range 3 {
				lat, err := PingServer(pingURL)
				if err != nil {
					continue
				}
				if success == 0 || lat < best {
					best = lat
				}
				success++
			}
			if success == 0 {
				s.Latency = time.Duration(math.MaxInt64)
			} else {
				s.Latency = best
			}
		}(&servers[i])
	}

	wg.Wait()

	// Sort the slice to find the server with the lowest latency
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Latency < servers[j].Latency
	})

	// If all failed (max latency), return nil
	if servers[0].Latency == time.Duration(math.MaxInt64) {
		return nil
	}
	return &servers[0]
}

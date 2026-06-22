package endpoints

import (
	"math"
	"sort"
	"sync"
	"time"
)

func FindBestServer(servers []Server, concurrency, attempts int, timeout time.Duration) *Server {
	if len(servers) == 0 {
		return nil
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := range servers {
		wg.Add(1)
		sem <- struct{}{}

		go func(s *Server) {
			defer wg.Done()
			defer func() { <-sem }()

			pingURL := s.URL(s.PingURL)
			success := 0
			var best time.Duration
			for _ = range attempts {
				lat, err := PingServer(pingURL, timeout)
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

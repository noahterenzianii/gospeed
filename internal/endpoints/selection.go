package endpoints

import (
	"sync"
	"time"
)

func FindBestServer(servers []Server, concurrency, attempts int, timeout time.Duration) *Server {
	if len(servers) == 0 {
		return nil
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var best *Server

	for i := range servers {
		wg.Add(1)
		sem <- struct{}{}

		go func(s *Server) {
			defer wg.Done()
			defer func() { <-sem }()

			pingURL := s.URL(s.PingURL)
			success := 0
			var bestLatency time.Duration
			for range attempts {
				lat, err := pingServer(pingURL, timeout)
				if err != nil {
					continue
				}
				if success == 0 || lat < bestLatency {
					bestLatency = lat
				}
				success++
			}
			if success == 0 {
				return
			}

			mu.Lock()
			if best == nil || bestLatency < best.Latency {
				s.Latency = bestLatency
				best = s
			}
			mu.Unlock()
		}(&servers[i])
	}

	wg.Wait()

	return best
}

package endpoints

import (
	"strings"
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

func (s *Server) URL(path string) string {
	return strings.TrimRight(s.ServerURL, "/") + "/" + strings.TrimLeft(path, "/")
}

type ClientInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

func (c *ClientInfo) LocationString() string {
	parts := make([]string, 0, 3)
	if c.City != "" {
		parts = append(parts, c.City)
	}
	if c.Region != "" {
		parts = append(parts, c.Region)
	}
	if c.Country != "" {
		parts = append(parts, c.Country)
	}
	return strings.Join(parts, ", ")
}

type Latency struct {
	Ping   time.Duration
	Jitter time.Duration
}

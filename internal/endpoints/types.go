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

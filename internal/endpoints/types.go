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

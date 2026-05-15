package endpoints

import "time"

type Server struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	ServerURL string        `json:"server"`
	DlURL     string        `json:"dlURL"`
	UlURL     string        `json:"ulURL"`
	PingURL   string        `json:"pingURL"`
	Latency   time.Duration `json:"-"`
}

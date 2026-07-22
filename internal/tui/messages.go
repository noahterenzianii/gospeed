package tui

import (
	"time"

	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

type clientInfoMsg struct {
	info *endpoints.ClientInfo
}

type serverMsg struct {
	server *endpoints.Server
}

type latencyMsg struct {
	latency *endpoints.Latency
}

type transferProgressMsg struct {
	state TransferState
	done  bool
	dir   direction
}

type errMsg struct {
	err error
}

type tuningProgressMsg struct {
	stepLabel string
	value     float64
	streams   int
	bufSize   int
	phase     int
}

type tuningResultMsg struct {
	dlStreams    int
	dlBufferSize int
	dlBandwidth  float64
	ulStreams    int
	ulBufferSize int
	ulBandwidth  float64
	elapsed      time.Duration
	err          error
}

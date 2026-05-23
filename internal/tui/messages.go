package tui

import "github.com/noahterenzianii/gospeed/internal/endpoints"

type clientInfoMsg struct {
	gen  int
	info *endpoints.ClientInfo
}

type serverMsg struct {
	gen    int
	server *endpoints.Server
}

type latencyMsg struct {
	gen     int
	latency *endpoints.Latency
}

type errMsg struct {
	gen int
	err error
}

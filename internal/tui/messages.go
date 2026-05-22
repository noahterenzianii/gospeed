package tui

import "github.com/noahterenzianii/gospeed/internal/endpoints"

type clientInfoMsg struct {
	info *endpoints.ClientInfo
}

type serverMsg struct {
	server *endpoints.Server
}

type latencyMsg struct {
	latency *endpoints.Latency
}

type errMsg struct {
	err error
}

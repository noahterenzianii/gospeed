package tui

import "github.com/noahterenzianii/gospeed/internal/endpoints"

type clientInfoMsg struct {
	info *endpoints.ClientInfo
}

type errMsg struct {
	err error
}

package tui

import (
	"strings"

	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

func serverView(server *endpoints.Server) string {
	hostname := strings.TrimPrefix(server.ServerURL, "https://")
	var rows []string
	rows = append(rows, sectionHeader("server", accentPurple))
	rows = append(rows, addField("name", server.Name, textPrimary))
	rows = append(rows, addField("host", hostname, textSecondary))
	return strings.Join(rows, "\n")
}

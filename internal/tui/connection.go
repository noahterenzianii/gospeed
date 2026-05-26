package tui

import (
	"strings"

	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

func connectionView(info *endpoints.ClientInfo, server *endpoints.Server) string {
	var rows []string
	rows = append(rows, sectionHeader("connection", accentPurple))
	rows = append(rows, addField("ip", info.IP, accentCyan))
	rows = append(rows, addField("isp", info.Org, textPrimary))
	if info.Hostname != "" {
		rows = append(rows, addField("hostname", info.Hostname, textSecondary))
	}
	rows = append(rows, addField("location", info.LocationString(), textSecondary))
	if server != nil {
		hostname := strings.TrimPrefix(server.ServerURL, "https://")
		rows = append(rows, addField("server", server.Name, textPrimary))
		rows = append(rows, addField("host", hostname, textSecondary))
	}
	return strings.Join(rows, "\n")
}

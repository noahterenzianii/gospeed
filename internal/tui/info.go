package tui

import (
	"strings"

	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

func infoView(info *endpoints.ClientInfo) string {
	var rows []string
	rows = append(rows, sectionHeader("client", accentPurple))
	rows = append(rows, addField("ip", info.IP, accentPurple))
	rows = append(rows, addField("isp", info.Org, textPrimary))
	if info.Hostname != "" {
		rows = append(rows, addField("hostname", info.Hostname, textSecondary))
	}
	rows = append(rows, addField("location", info.LocationString(), textSecondary))
	return strings.Join(rows, "\n")
}

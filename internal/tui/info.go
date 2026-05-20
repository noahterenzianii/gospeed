package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

const labelWidth = 12

var labelStyle = lipgloss.NewStyle().Foreground(cMuted).Width(labelWidth)

func addField(rows []string, label, value string, color lipgloss.Color) []string {
	labelS := labelStyle.Render(label)
	valueS := lipgloss.NewStyle().Foreground(color).Render(value)
	return append(rows, fmt.Sprintf("  %s%s", labelS, valueS))
}

func infoView(info *endpoints.ClientInfo) string {
	var rows []string

	rows = addField(rows, "ip", info.IP, cCyan)
	rows = addField(rows, "isp", info.Org, cWhite)

	if info.Hostname != "" {
		rows = addField(rows, "hostname", info.Hostname, cWhite)
	}

	rows = addField(rows, "location", info.LocationString(), cScreen)

	if info.Timezone != "" {
		rows = addField(rows, "timezone", info.Timezone, cScreen)
	}

	return strings.Join(rows, "\n")
}

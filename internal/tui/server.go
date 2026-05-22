package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

func serverView(server *endpoints.Server) string {
	hostname := strings.TrimPrefix(server.ServerURL, "https://")
	var rows []string
	rows = append(rows, fmt.Sprintf("  %s%s",
		labelStyle.Render("server"),
		lipgloss.NewStyle().Foreground(cWhite).Render(server.Name),
	))
	rows = append(rows, fmt.Sprintf("  %s%s",
		labelStyle.Render("endpoint"),
		lipgloss.NewStyle().Foreground(cMuted).Render(hostname),
	))
	return strings.Join(rows, "\n")
}

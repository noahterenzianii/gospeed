package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var asciiArt = []string{
	`  ██████╗  ██████╗ ███████╗██████╗ ███████╗███████╗██████╗ `,
	` ██╔════╝ ██╔═══██╗██╔════╝██╔══██╗██╔════╝██╔════╝██╔══██╗`,
	` ██║  ███╗██║   ██║███████╗██████╔╝█████╗  █████╗  ██║  ██║`,
	` ██║   ██║██║   ██║╚════██║██╔═══╝ ██╔══╝  ██╔══╝  ██║  ██║`,
	` ╚██████╔╝╚██████╔╝███████║██║     ███████╗███████╗██████╔╝`,
	`  ╚═════╝  ╚═════╝ ╚══════╝╚═╝     ╚══════╝╚══════╝╚═════╝ `,
}

func asciiView() string {
	var lines []string
	for _, line := range asciiArt {
		lines = append(lines, lipgloss.NewStyle().Foreground(accentCyan).Render(line))
	}
	return strings.Join(lines, "\n")
}

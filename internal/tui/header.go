package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var asciiStyle = lipgloss.NewStyle().Foreground(cScreen)

var asciiArt = []string{
	``,
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
		lines = append(lines, asciiStyle.Render(line))
	}
	return strings.Join(lines, "\n")
}

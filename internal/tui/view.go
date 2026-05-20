package tui

import "strings"

func asciiView() string {
	var lines []string
	for _, line := range asciiArt {
		lines = append(lines, asciiStyle.Render(line))
	}
	return strings.Join(lines, "\n")
}

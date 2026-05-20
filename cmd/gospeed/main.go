package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/app"
	"github.com/noahterenzianii/gospeed/internal/tui"
)

func main() {
	cliMode := flag.Bool("cli", false, "run in CLI mode")
	flag.Parse()
	if *cliMode {
		if err := app.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	p := tea.NewProgram(tui.NewModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

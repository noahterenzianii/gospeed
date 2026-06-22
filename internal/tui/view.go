package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	header := asciiView()

	if m.showConfig {
		return fmt.Sprintf("%s\n\n%s\n\n%s", header, m.configView(), m.footerView())
	}

	if m.err != nil {
		return fmt.Sprintf("%s\n\n%s\n\n%s", header, errStyle(m.err.Error()), m.footerView())
	}

	screens := m.buildScreens()
	if len(screens) > 0 {
		return fmt.Sprintf("%s\n\n%s\n\n%s", header, strings.Join(screens, "\n\n"), m.footerView())
	}

	load := fmt.Sprintf("  %s %s",
		lipgloss.NewStyle().Foreground(accentCyan).Render(m.spinner.View()),
		lipgloss.NewStyle().Foreground(textSecondary).Render("initializing..."))
	return fmt.Sprintf("%s\n\n%s\n\n%s", header, load, m.footerView())
}

func (m Model) buildScreens() []string {
	if m.phase == phaseIdle {
		return []string{startView()}
	}
	var screens []string
	if m.clientInfo != nil {
		screens = append(screens, infoView(m.clientInfo))
	}
	if m.server != nil {
		screens = append(screens, serverView(m.server))
	}
	if s := m.pingScreen(); s != "" {
		screens = append(screens, s)
	}
	if s := m.transferView(); s != "" {
		screens = append(screens, s)
	}
	if m.phase == phaseDone {
		if s := m.resultsView(); s != "" {
			screens = append(screens, s)
		}
	}
	return screens
}

func (m Model) pingScreen() string {
	if m.latency != nil {
		return pingView(m.latency)
	}
	var status string
	switch m.phase {
	case phaseFetching:
		status = "fetching client info..."
	case phaseInfo:
		status = "selecting best server..."
	case phasePinging:
		status = "measuring network latency..."
	default:
		return ""
	}
	return fmt.Sprintf("  %s %s",
		lipgloss.NewStyle().Foreground(accentCyan).Render(m.spinner.View()),
		lipgloss.NewStyle().Foreground(textSecondary).Render(status))
}

func (m Model) configView() string {
	var rows []string
	rows = append(rows, sectionHeader("configuration", accentCyan))

	for i, f := range configFields {
		if f.isSection {
			rows = append(rows, "", lipgloss.NewStyle().Foreground(accentPurple).Bold(true).Render("  "+f.label))
			continue
		}
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(textSecondary)
		if i == m.configCursor {
			cursor = lipgloss.NewStyle().Foreground(accentGreen).Render("▸ ")
			style = lipgloss.NewStyle().Foreground(textPrimary)
		}
		label := style.Render(f.label)
		val := lipgloss.NewStyle().Foreground(accentCyan).Render(f.value(m.cfg))
		pad := strings.Repeat(" ", 24-len(f.label))
		rows = append(rows, fmt.Sprintf("%s%s%s%s", cursor, label, pad, val))
	}

	return strings.Join(rows, "\n")
}

func startView() string {
	return fmt.Sprintf("  %s",
		lipgloss.NewStyle().Foreground(textSecondary).Render("Press s to start the speed test"))
}

func (m Model) footerView() string {
	if m.showConfig {
		return mutedStyle.Render("  ↑/↓ navigate  •  +/- modify  •  r: reset  •  esc/c: close")
	}
	if m.phase == phaseIdle {
		return mutedStyle.Render("  s: start  •  c: config  •  q: quit")
	}
	if m.canRedo() {
		return mutedStyle.Render("  q: quit  •  r: redo  •  c: config")
	}
	return mutedStyle.Render("  q: quit")
}

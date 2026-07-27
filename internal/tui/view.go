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
	if m.phase == phaseIdle && !m.isTuningDone() {
		return []string{startView()}
	}
	if m.phase == phaseTuning {
		return []string{m.tuningView()}
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
	if m.isTuningDone() {
		screens = append(screens, m.tuningDoneView())
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
			if f.label == "advanced" {
				rows = append(rows, "", lipgloss.NewStyle().Foreground(accentYellow).Render("  ── advanced ──"))
				continue
			}
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

func (m Model) tuningView() string {
	if m.tuningCancel == nil && (m.tuningDlStreams > 0 || m.tuningUlStreams > 0) {
		return m.tuningDoneView()
	}

	label := m.tuningLabel
	switch label {
	case "tuning download...", "tuning upload...", "":
		label = "tuning..."
	}

	var sections []string
	if m.tuningDlStreams > 0 {
		sections = append(sections, sectionView(dirDownload,
			m.tuningDlStreams, m.tuningDlBufSize, 0, ""))
	}
	sections = append(sections, sectionView(m.tuningDir,
		0, 0, m.tuningValue, label))

	return strings.Join(sections, "\n\n")
}

func (m Model) tuningDoneView() string {
	var sections []string

	sections = append(sections, sectionHeader("tuning complete", accentGreen))

	if m.tuningDlStreams > 0 {
		dl := lipgloss.NewStyle().Foreground(accentCyan).Render(
			fmt.Sprintf("↓ %d streams · %d kb buffer", m.tuningDlStreams, m.tuningDlBufSize/1024))
		sections = append(sections, fmt.Sprintf("  %s", dl))
	}

	if m.tuningUlStreams > 0 {
		ul := lipgloss.NewStyle().Foreground(accentYellow).Render(
			fmt.Sprintf("↑ %d streams · %d kb buffer", m.tuningUlStreams, m.tuningUlBufSize/1024))
		sections = append(sections, fmt.Sprintf("  %s", ul))
	}

	if m.tuningElapsed > 0 {
		sections = append(sections, addField("elapsed",
			fmt.Sprintf("%.1f s", m.tuningElapsed.Seconds()), textMuted))
	}

	return strings.Join(sections, "\n")
}

func sectionView(dir direction, streams, bufSize int, value float64, label string) string {
	var title, arrow string
	var accent lipgloss.TerminalColor
	switch dir {
	case dirDownload:
		title, arrow = "download", "↓"
		accent = accentCyan
	case dirUpload:
		title, arrow = "upload", "↑"
		accent = accentYellow
	}

	prefix := lipgloss.NewStyle().Foreground(accent).Render(arrow)

	var line string
	if streams > 0 {
		paramS := fmt.Sprintf("%d streams · %d kb buffer", streams, bufSize/1024)
		line = fmt.Sprintf("  %s  %s",
			prefix, lipgloss.NewStyle().Foreground(accent).Render(paramS))
	} else if value > 0 {
		speedS := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(fmtSpeed(value))
		line = fmt.Sprintf("  %s  %s%s  %s",
			prefix, speedS, mutedStyle.Render(unitMbps),
			lipgloss.NewStyle().Foreground(textSecondary).Render(label))
	} else {
		line = fmt.Sprintf("  %s  %s",
			prefix, lipgloss.NewStyle().Foreground(textSecondary).Render(label))
	}

	return fmt.Sprintf("%s\n%s", sectionHeader(title, accent), line)
}

func startView() string {
	return fmt.Sprintf("  %s",
		lipgloss.NewStyle().Foreground(textSecondary).Render("Press s to start the speed test"))
}

func (m Model) footerView() string {
	if m.showConfig {
		return mutedStyle.Render("  ↑/↓ navigate  •  +/- modify  •  r: reset  •  esc/c: close")
	}
	if m.isTuningDone() {
		return mutedStyle.Render("  s: start  •  c: config  •  t: re-tune  •  q: quit  •  esc: dismiss")
	}
	if m.canStart() {
		return mutedStyle.Render("  s: start  •  c: config  •  t: tune  •  q: quit")
	}
	if m.phase == phaseTuning {
		return mutedStyle.Render("  q: cancel tuning")
	}
	return mutedStyle.Render("  q: quit")
}

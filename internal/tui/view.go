package tui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	header := asciiView()

	if m.err != nil {
		return fmt.Sprintf("%s\n\n  error: %v\n%s", header, m.err, m.footerView())
	}

	screens := m.buildScreens()
	if len(screens) > 0 {
		return fmt.Sprintf("%s\n\n%s\n%s", header, strings.Join(screens, "\n\n"), m.footerView())
	}

	return fmt.Sprintf("%s\n\n  %s\n%s", header, mutedStyle.Render("fetching client info..."), m.footerView())
}

// buildScreens collects all non-nil result views in order.
func (m Model) buildScreens() []string {
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
	if s := m.downloadScreen(); s != "" {
		screens = append(screens, s)
	}
	return screens
}

// pingScreen shows either latency results or a spinner for in-progress phases.
func (m Model) pingScreen() string {
	if m.latency != nil {
		return pingView(m.latency)
	}
	var status string
	switch m.phase {
	case phaseFetching:
		status = "fetching client info"
	case phaseInfo:
		status = "choosing server"
	case phasePinging:
		status = "pinging server"
	default:
		return ""
	}
	return fmt.Sprintf("  %s %s", dimStyle.Render(m.spinner.View()), status)
}

// downloadScreen shows the download speed view during/after measurement.
func (m Model) downloadScreen() string {
	if (m.phase == phaseDownloading || m.phase == phaseDownload) && m.download != nil {
		return downloadView(m.download)
	}
	return ""
}

func (m Model) footerView() string {
	if m.canRedo() {
		return mutedStyle.Render("\n  q: quit • r: redo")
	}
	return mutedStyle.Render("\n  q: quit")
}

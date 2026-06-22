package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func fmtSpeed(mbps float64) string {
	switch {
	case mbps < 0.01:
		return "—"
	case mbps < 0.1:
		return fmt.Sprintf("%.2f", mbps)
	case mbps < 1:
		return fmt.Sprintf("%.1f", mbps)
	default:
		return fmt.Sprintf("%.0f", mbps)
	}
}

func (m Model) transferView() string {
	if m.download == nil && m.upload == nil {
		return ""
	}

	var sections []string

	if m.download != nil {
		show := m.phase == phaseDownloading || m.phase == phaseUploading || m.phase == phaseDone
		if show {
			sections = append(sections, m.formatTransferContent(m.download, dirDownload))
		}
	}

	if m.upload != nil {
		show := m.phase == phaseUploading || m.phase == phaseDone
		if show {
			sections = append(sections, "", m.formatTransferContent(m.upload, dirUpload))
		}
	}

	if len(sections) == 0 {
		return ""
	}

	return strings.Join(sections, "\n")
}

func (m Model) formatTransferContent(s *TransferState, dir direction) string {
	var title, arrow string
	var accent lipgloss.TerminalColor

	switch dir {
	case dirDownload:
		title = "download"
		arrow = "↓"
		accent = accentCyan
	case dirUpload:
		title = "upload"
		arrow = "↑"
		accent = accentYellow
	}

	pct := s.Elapsed.Seconds() / m.cfg.TransferDuration.Seconds()
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}

	var rows []string
	rows = append(rows, sectionHeader(title, accent))

	arrowS := lipgloss.NewStyle().Foreground(accent).Render(arrow)
	speedS := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(fmtSpeed(s.Speed))
	pctS := lipgloss.NewStyle().Foreground(accent).Render(fmt.Sprintf("%.0f%%", pct*100))
	rows = append(rows, fmt.Sprintf("  %s %s%s  %s",
		arrowS, speedS, mutedStyle.Render(unitMbps), pctS))

	if len(s.Samples) > 0 {
		rows = append(rows, "  "+lipgloss.NewStyle().Foreground(accent).Render(renderSparkline(s.Samples)))
	}

	return strings.Join(rows, "\n")
}

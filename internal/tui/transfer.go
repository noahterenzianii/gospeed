package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func formatBytes(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	case b < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(b)/(1024*1024*1024))
	}
}

func (m Model) transferView() string {
	if m.download == nil && m.upload == nil {
		return ""
	}

	var sections []string

	if m.download != nil {
		show := m.phase == phaseDownloading || m.phase == phaseDownload ||
			m.phase == phaseUploading || m.phase == phaseUpload
		if show {
			sections = append(sections, formatTransferContent(m.download, dirDownload))
		}
	}

	if m.upload != nil {
		show := m.phase == phaseUploading || m.phase == phaseUpload
		if show {
			sections = append(sections, "", formatTransferContent(m.upload, dirUpload))
		}
	}

	if len(sections) == 0 {
		return ""
	}

	return strings.Join(sections, "\n")
}

func formatTransferContent(s *TransferState, dir direction) string {
	var title, arrow, byteLabel string
	var accent lipgloss.TerminalColor

	switch dir {
	case dirDownload:
		title = "download"
		arrow = "↓"
		byteLabel = "received"
		accent = accentCyan
	case dirUpload:
		title = "upload"
		arrow = "↑"
		byteLabel = "sent"
		accent = accentYellow
	}

	pct := s.Elapsed.Seconds() / transferDuration.Seconds()
	if pct > 1.0 {
		pct = 1.0
	}
	if pct < 0 {
		pct = 0
	}

	var rows []string
	rows = append(rows, sectionHeader(title, accent))

	arrowS := lipgloss.NewStyle().Foreground(accent).Render(arrow)
	speedS := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(fmt.Sprintf("%.0f", s.Speed))
	pctS := lipgloss.NewStyle().Foreground(accent).Render(fmt.Sprintf("%.0f%%", pct*100))
	rows = append(rows, fmt.Sprintf("  %s %s%s  %s",
		arrowS, speedS, mutedStyle.Render(" Mbps"), pctS))

	bytesStr := formatBytes(int64(s.Speed * s.Elapsed.Seconds() * mbpsToBytes))
	elapsedStr := fmt.Sprintf("%.1f s", s.Elapsed.Seconds())
	rows = append(rows, fmt.Sprintf("    %s %s  ·  %s %s",
		mutedStyle.Render(byteLabel),
		lipgloss.NewStyle().Foreground(textSecondary).Render(bytesStr),
		mutedStyle.Render("elapsed"),
		lipgloss.NewStyle().Foreground(textSecondary).Render(elapsedStr)))

	if len(s.Samples) > 0 {
		rows = append(rows, "  "+lipgloss.NewStyle().Foreground(accent).Render(renderSparkline(s.Samples)))
	}

	return strings.Join(rows, "\n")
}

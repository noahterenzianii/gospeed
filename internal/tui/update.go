package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case clientInfoMsg:
		return m.handleClientInfo(msg)
	case serverMsg:
		return m.handleServer(msg)
	case latencyMsg:
		return m.handleLatency(msg)
	case downloadProgressMsg:
		return m.handleDownloadProgress(msg)
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	case errMsg:
		return m.handleError(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		if !m.canRedo() {
			return m, nil
		}
		m.clientInfo = nil
		m.server = nil
		m.latency = nil
		m.download = nil
		m.downloadCh = nil
		m.err = nil
		m.phase = phaseFetching
		return m, tea.Batch(m.fetchClientInfo(), m.spinner.Tick)
	}
	return m, nil
}

func (m Model) handleClientInfo(msg clientInfoMsg) (tea.Model, tea.Cmd) {
	m.clientInfo = msg.info
	m.phase = phaseInfo
	return m, tea.Batch(m.fetchServers(), m.spinner.Tick)
}

func (m Model) handleServer(msg serverMsg) (tea.Model, tea.Cmd) {
	m.server = msg.server
	m.phase = phasePinging
	return m, tea.Batch(m.measureLatency(), m.spinner.Tick)
}

func (m Model) handleLatency(msg latencyMsg) (tea.Model, tea.Cmd) {
	m.latency = msg.latency
	m.phase = phaseDownloading
	cmd, ch := m.measureDownload()
	m.downloadCh = ch
	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m Model) handleDownloadProgress(msg downloadProgressMsg) (tea.Model, tea.Cmd) {
	m.download = &DownloadState{
		Speed:   msg.state.Speed,
		Samples: msg.state.Samples,
		Elapsed: msg.state.Elapsed,
	}
	if msg.done {
		m.phase = phaseDownload
		return m, nil
	}
	return m, listenDownload(m.downloadCh)
}

func (m Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	if m.loading() {
		return m, cmd
	}
	return m, nil
}

func (m Model) handleError(msg errMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	return m, nil
}

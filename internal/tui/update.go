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
	case transferProgressMsg:
		return m.handleTransferProgress(msg)
	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)
	case errMsg:
		return m.handleError(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showConfig {
		switch msg.String() {
		case "c", "esc":
			m.showConfig = false
		case "up", "k":
			m.configCursor = prevField(m.configCursor)
		case "down", "j":
			m.configCursor = nextField(m.configCursor)
		case "left", "-":
			m.applyConfigDelta(-1)
		case "right", "+", "=":
			m.applyConfigDelta(1)
		case "r":
			m.cfg = defaultConfig()
		}
		return m, nil
	}
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "c":
		if !m.canConfig() {
			return m, nil
		}
		m.showConfig = true
		m.configCursor = nextField(-1)
		return m, nil
	case "s":
		if !m.canStart() {
			return m, nil
		}
		return m.startTest()
	}
	return m, nil
}

func (m Model) startTest() (tea.Model, tea.Cmd) {
	m.clientInfo = nil
	m.server = nil
	m.latency = nil
	m.download = nil
	m.downloadCh = nil
	m.upload = nil
	m.uploadCh = nil
	m.err = nil
	m.phase = phaseFetching
	return m, tea.Batch(m.fetchClientInfo(), m.spinner.Tick)
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
	return m.startTransfer(dirDownload)
}

func (m Model) handleTransferProgress(msg transferProgressMsg) (tea.Model, tea.Cmd) {
	s := &TransferState{
		Speed:   msg.state.Speed,
		Samples: msg.state.Samples,
		Elapsed: msg.state.Elapsed,
	}
	switch msg.dir {
	case dirDownload:
		m.download = s
	case dirUpload:
		m.upload = s
	}

	if msg.done {
		switch msg.dir {
		case dirDownload:
			m.phase = phaseUploading
			return m.startTransfer(dirUpload)
		case dirUpload:
			m.phase = phaseDone
			return m, nil
		}
	}

	if msg.dir == dirUpload {
		m.phase = phaseUploading
	}
	return m, listenTransfer(m.chForDir(msg.dir))
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

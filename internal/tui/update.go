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

// handleKey processes keyboard input: q to quit, r to restart.
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
		m.upload = nil
		m.uploadCh = nil
		m.err = nil
		m.phase = phaseFetching
		return m, tea.Batch(m.fetchClientInfo(), m.spinner.Tick)
	case "s":
		if m.phase != phaseIdle {
			return m, nil
		}
		m.phase = phaseFetching
		return m, tea.Batch(m.fetchClientInfo(), m.spinner.Tick)
	}
	return m, nil
}

// handleClientInfo stores client info and moves to server selection.
func (m Model) handleClientInfo(msg clientInfoMsg) (tea.Model, tea.Cmd) {
	m.clientInfo = msg.info
	m.phase = phaseInfo
	return m, tea.Batch(m.fetchServers(), m.spinner.Tick)
}

// handleServer stores the chosen server and starts latency measurement.
func (m Model) handleServer(msg serverMsg) (tea.Model, tea.Cmd) {
	m.server = msg.server
	m.phase = phasePinging
	return m, tea.Batch(m.measureLatency(), m.spinner.Tick)
}

// handleLatency stores ping results and starts the download test.
func (m Model) handleLatency(msg latencyMsg) (tea.Model, tea.Cmd) {
	m.latency = msg.latency
	m.phase = phaseDownloading
	return m.startTransfer(dirDownload)
}

// handleTransferProgress updates speed/samples and checks for completion.
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
			m.phase = phaseUpload
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

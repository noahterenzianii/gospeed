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
	case tuningProgressMsg:
		return m.handleTuningProgress(msg)
	case tuningResultMsg:
		return m.handleTuningResult(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showConfig {
		return m.handleConfigKey(msg)
	}

	if m.isTuningDone() {
		return m.handleTuningDoneKey(msg)
	}

	if m.tuningCancel != nil && m.phase == phaseTuning {
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.tuningCancel()
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		if m.tuningCancel != nil {
			m.tuningCancel()
		}
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
	case "t":
		if !m.canTune() {
			return m, nil
		}
		return m.startTuningCmd()
	}
	return m, nil
}

func (m Model) isTuningDone() bool {
	return m.tuningCancel == nil && (m.tuningUlStreams > 0 || m.tuningDlStreams > 0)
}

func (m Model) handleTuningDoneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		m = m.dismissTuning()
		return m.startTest()
	case "c":
		m = m.dismissTuning()
		m.showConfig = true
		m.configCursor = nextField(-1)
		return m, nil
	case "t":
		m = m.dismissTuning()
		return m.startTuningCmd()
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m = m.dismissTuning()
		return m, nil
	}
	return m, nil
}

func (m Model) handleConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m Model) dismissTuning() Model {
	m.tuningLabel = ""
	m.tuningValue = 0
	m.tuningDlStreams = 0
	m.tuningDlBufSize = 0
	m.tuningDlRawBW = 0
	m.tuningUlStreams = 0
	m.tuningUlBufSize = 0
	m.tuningUlRawBW = 0
	m.tuningElapsed = 0
	m.tuningDir = 0
	m.phase = phaseIdle
	return m
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

func (m Model) handleTuningProgress(msg tuningProgressMsg) (tea.Model, tea.Cmd) {
	switch msg.stepLabel {
	case "tuning download...":
		m.tuningDir = dirDownload
	case "tuning upload...":
		m.tuningDlStreams = msg.streams
		m.tuningDlBufSize = msg.bufSize
		m.tuningDir = dirUpload
	}
	m.tuningLabel = msg.stepLabel
	m.tuningValue = msg.value
	return m, listenTransfer(m.tuningCh)
}

func (m Model) handleTuningResult(msg tuningResultMsg) (tea.Model, tea.Cmd) {
	if m.tuningCancel != nil {
		m.tuningCancel()
		m.tuningCancel = nil
	}
	m.tuningCh = nil

	// Always apply download results (may be partial if upload failed)
	if msg.dlStreams > 0 {
		m.cfg.DownloadStreams = msg.dlStreams
		m.cfg.DownloadBufferSize = msg.dlBufferSize
		m.tuningDlStreams = msg.dlStreams
		m.tuningDlBufSize = msg.dlBufferSize
		m.tuningDlRawBW = msg.dlBandwidth
	}
	if msg.ulStreams > 0 {
		m.cfg.UploadStreams = msg.ulStreams
		m.cfg.UploadBufferSize = msg.ulBufferSize
		m.tuningUlStreams = msg.ulStreams
		m.tuningUlBufSize = msg.ulBufferSize
		m.tuningUlRawBW = msg.ulBandwidth
	}

	if msg.err != nil {
		m.err = msg.err
		m.phase = phaseIdle
		return m, nil
	}

	m.tuningElapsed = msg.elapsed
	m.tuningLabel = ""
	m.tuningValue = 0
	return m, nil
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

package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

const (
	serverListURL    = "https://librespeed.org/backend-servers/servers.php"
	pingSamples      = 200
	transferDuration = 10 * time.Second
	transferStreams  = 4
	transferBufSize  = 100
)

// Phase constants represent the sequential test lifecycle.
type phase int

const (
	phaseFetching phase = iota
	phaseInfo
	phasePinging
	phasePing
	phaseDownloading
	phaseDownload
	phaseUploading
	phaseUpload
)

type direction int

const (
	dirDownload direction = iota
	dirUpload
)

// TransferState holds live transfer metrics for the TUI view.
type TransferState struct {
	Speed   float64
	Samples []float64
	Elapsed time.Duration
}

type Model struct {
	clientInfo *endpoints.ClientInfo
	server     *endpoints.Server
	latency    *endpoints.Latency
	download   *TransferState
	upload     *TransferState

	downloadCh chan tea.Msg
	uploadCh   chan tea.Msg

	phase   phase
	spinner spinner.Model
	err     error
}

func NewModel() Model {
	s := spinner.New()
	s.Style = dimStyle
	return Model{phase: phaseFetching, spinner: s}
}

func (m Model) canRedo() bool {
	return m.err != nil || m.phase == phasePing || m.phase == phaseDownload || m.phase == phaseUpload
}

func (m Model) loading() bool {
	return m.phase == phaseFetching || m.phase == phaseInfo || m.phase == phasePinging || m.phase == phaseDownloading || m.phase == phaseUploading
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchClientInfo(), m.spinner.Tick)
}

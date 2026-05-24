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
	downloadDuration = 10 * time.Second
	downloadStreams  = 4
	downloadBufSize  = 100
	mbpsToBytes      = 125_000 // 1 Mbps = 125,000 bytes/s
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
)

// DownloadState holds live download metrics for the TUI view.
type DownloadState struct {
	Speed   float64
	Samples []float64
	Elapsed time.Duration
}

type Model struct {
	clientInfo *endpoints.ClientInfo
	server     *endpoints.Server
	latency    *endpoints.Latency
	download   *DownloadState

	downloadCh chan tea.Msg

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
	return m.err != nil || m.phase == phasePing || m.phase == phaseDownload
}

func (m Model) loading() bool {
	return m.phase == phaseFetching || m.phase == phaseInfo || m.phase == phasePinging || m.phase == phaseDownloading
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchClientInfo(), m.spinner.Tick)
}

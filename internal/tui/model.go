package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

// Phase constants represent the sequential test lifecycle.
type phase int

const (
	phaseIdle phase = iota
	phaseFetching
	phaseInfo
	phasePinging
	phaseDownloading
	phaseUploading
	phaseDone
	phaseTuning
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

	phase        phase
	spinner      spinner.Model
	err          error
	showConfig   bool
	configCursor int
	cfg          *Config

	tuningLabel      string
	tuningValue      float64
	tuningCh         chan tea.Msg
	tuningCancel     context.CancelFunc
	tuningDir        direction // which direction is currently being tuned

	tuningDlStreams  int         // download result (streams)
	tuningDlBufSize  int         // download result (buffer)
	tuningDlRawBW    float64     // download measured bandwidth during tuning
	tuningUlStreams  int         // upload result (streams)
	tuningUlBufSize  int         // upload result (buffer)
	tuningUlRawBW    float64     // upload measured bandwidth during tuning
	tuningElapsed    time.Duration
}

func NewModel() Model {
	s := spinner.New()
	s.Style = dimStyle
	return Model{phase: phaseIdle, spinner: s, cfg: defaultConfig()}
}

func (m Model) canStart() bool {
	return m.phase == phaseIdle || m.err != nil || m.phase == phaseDone
}

func (m Model) canConfig() bool {
	return m.canStart()
}

func (m Model) loading() bool {
	if m.isTuningDone() {
		return false
	}
	return m.phase >= phaseFetching && m.phase <= phaseUploading || m.phase == phaseTuning
}

func (m Model) canTune() bool {
	return m.canStart()
}

func (m Model) Init() tea.Cmd {
	return nil
}

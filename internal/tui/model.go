package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

const (
	serverListURL = "https://librespeed.org/backend-servers/servers.php"
	pingSamples   = 200
)

type phase int

const (
	phaseFetching phase = iota
	phaseInfo
	phasePinging
	phasePing
)

type Model struct {
	clientInfo *endpoints.ClientInfo
	server     *endpoints.Server
	latency    *endpoints.Latency
	phase      phase
	spinner    spinner.Model
	err        error
}

func NewModel() Model {
	s := spinner.New()
	s.Style = dimStyle
	return Model{phase: phaseFetching, spinner: s}
}

func (m Model) Init() tea.Cmd {
	return m.fetchClientInfo()
}

// --- async commands ---

func (m Model) fetchClientInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := endpoints.FetchClientInfo()
		if err != nil {
			return errMsg{err}
		}
		return clientInfoMsg{info}
	}
}

func (m Model) fetchServers() tea.Cmd {
	return func() tea.Msg {
		servers, err := endpoints.FetchServers(serverListURL)
		if err != nil {
			return errMsg{err}
		}
		best := endpoints.FindBestServer(servers)
		if best == nil {
			return errMsg{fmt.Errorf("no reachable server")}
		}
		return serverMsg{best}
	}
}

func (m Model) measureLatency() tea.Cmd {
	return func() tea.Msg {
		pingURL := m.server.URL(m.server.PingURL)
		latency, err := endpoints.MeasureLatency(pingURL, pingSamples)
		if err != nil {
			return errMsg{err}
		}
		return latencyMsg{&latency}
	}
}

// --- update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil

	case clientInfoMsg:
		m.clientInfo = msg.info
		m.phase = phaseInfo
		return m, tea.Batch(m.fetchServers(), m.spinner.Tick)

	case serverMsg:
		m.server = msg.server
		m.phase = phasePinging
		return m, tea.Batch(m.measureLatency(), m.spinner.Tick)

	case latencyMsg:
		m.latency = msg.latency
		m.phase = phasePing
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.phase == phaseInfo || m.phase == phasePinging {
			return m, cmd
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// --- view ---

func (m Model) View() string {
	header := asciiView()

	if m.err != nil {
		return fmt.Sprintf("%s\n\n  error: %v\n%s", header, m.err, footerView())
	}

	screens := m.buildScreens()
	if len(screens) > 0 {
		return fmt.Sprintf("%s\n\n%s\n%s", header, strings.Join(screens, "\n\n"), footerView())
	}

	return fmt.Sprintf("%s\n\n  %s\n%s", header, mutedStyle.Render("fetching client info..."), footerView())
}

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
	return screens
}

func (m Model) pingScreen() string {
	if m.latency != nil {
		return pingView(m.latency)
	}
	if m.phase != phaseInfo && m.phase != phasePinging {
		return ""
	}
	status := "choosing server"
	if m.phase == phasePinging {
		status = "pinging server"
	}
	return fmt.Sprintf("  %s %s", dimStyle.Render(m.spinner.View()), status)
}

func footerView() string {
	return mutedStyle.Render("\n  q: quit")
}

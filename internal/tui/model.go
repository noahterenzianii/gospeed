package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
)

type Model struct {
	clientInfo *endpoints.ClientInfo
	err        error
	loading    bool
}

func NewModel() Model {
	// Start in loading state — Init() will fetch data immediately.
	return Model{loading: true}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		info, err := endpoints.FetchClientInfo()
		if err != nil {
			return errMsg{err}
		}
		return clientInfoMsg{info}
	}
}

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
		m.loading = false
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	header := asciiView()

	if m.loading {
		return fmt.Sprintf("%s\n\n  %s", header, mutedStyle.Render("fetching client info..."))
	}
	if m.err != nil {
		return fmt.Sprintf("%s\n\n  error: %v\n\n  press q to quit", header, m.err)
	}

	return fmt.Sprintf("%s\n\n%s", header, infoView(m.clientInfo))
}

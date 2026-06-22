package tui

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
	"github.com/noahterenzianii/gospeed/internal/speedtest"
)

type measureFunc func(url string, duration time.Duration, streams int, bufSize int, onProgress speedtest.ProgressFunc) (float64, error)

// fetchClientInfo retrieves the client's IP, ISP, and location.
func (m Model) fetchClientInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := endpoints.FetchClientInfo(m.cfg.ClientInfoTimeout)
		if err != nil {
			return errMsg{err}
		}
		return clientInfoMsg{info}
	}
}

// fetchServers fetches the server list and selects the best one.
func (m Model) fetchServers() tea.Cmd {
	return func() tea.Msg {
		servers, err := endpoints.FetchServers(serverListURL, m.cfg.ServerListTimeout)
		if err != nil {
			return errMsg{err}
		}
		best := endpoints.FindBestServer(servers, m.cfg.MaxConcurrentPings, m.cfg.PingAttempts, m.cfg.PingTimeout)
		if best == nil {
			return errMsg{fmt.Errorf("no reachable server")}
		}
		return serverMsg{best}
	}
}

// measureLatency runs ping samples against the selected server.
func (m Model) measureLatency() tea.Cmd {
	return func() tea.Msg {
		pingURL := m.server.URL(m.server.PingURL)
		latency, err := endpoints.MeasureLatency(pingURL, m.cfg.PingSamples, m.cfg.PingTimeout)
		if err != nil {
			return errMsg{err}
		}
		return latencyMsg{&latency}
	}
}

// startTransfer launches a transfer goroutine and returns the listen cmd.
func (m Model) startTransfer(dir direction) (tea.Model, tea.Cmd) {
	cmd, ch := m.measureTransfer(dir)
	switch dir {
	case dirDownload:
		m.downloadCh = ch
	case dirUpload:
		m.uploadCh = ch
	}
	return m, tea.Batch(cmd, m.spinner.Tick)
}

// measureTransfer starts a goroutine that streams transfer progress via a channel.
func (m Model) measureTransfer(dir direction) (tea.Cmd, chan tea.Msg) {
	var url string
	var fn measureFunc

	switch dir {
	case dirDownload:
		url = m.server.URL(m.server.DlURL)
		fn = speedtest.MeasureDownload
	case dirUpload:
		url = m.server.URL(m.server.UlURL)
		fn = speedtest.MeasureUpload
	}

	ch := make(chan tea.Msg, 100)
	go func() {
		var mu sync.Mutex
		var samples []float64
		start := time.Now()
		speed, err := fn(url, m.cfg.TransferDuration, m.cfg.TransferStreams, m.cfg.BufferSize, func(mbps float64) {
			elapsed := time.Since(start)
			mu.Lock()
			samples = append(samples, mbps)
			snapshot := make([]float64, len(samples))
			copy(snapshot, samples)
			mu.Unlock()
			select {
			case ch <- transferProgressMsg{TransferState{Speed: mbps, Samples: snapshot, Elapsed: elapsed}, false, dir}:
			default:
			}
		})
		if err != nil {
			ch <- errMsg{err}
		} else {
			mu.Lock()
			final := make([]float64, len(samples))
			copy(final, samples)
			mu.Unlock()
			ch <- transferProgressMsg{TransferState{Speed: speed, Samples: final, Elapsed: time.Since(start)}, true, dir}
		}
		close(ch)
	}()
	return listenTransfer(ch), ch
}

// listenTransfer wraps channel reads as a bubbletea command.
func listenTransfer(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// chForDir returns the active channel for the given direction.
func (m Model) chForDir(dir direction) chan tea.Msg {
	switch dir {
	case dirDownload:
		return m.downloadCh
	case dirUpload:
		return m.uploadCh
	}
	panic("tui: unknown direction")
}

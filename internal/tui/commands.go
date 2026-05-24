package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
	"github.com/noahterenzianii/gospeed/internal/speedtest"
)

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

func (m Model) measureDownload() (tea.Cmd, chan tea.Msg) {
	dlURL := m.server.URL(m.server.DlURL)
	ch := make(chan tea.Msg, downloadBufSize)
	go func() {
		var samples []float64
		start := time.Now()
		speed, err := speedtest.MeasureDownload(dlURL, downloadDuration, downloadStreams, func(mbps float64) {
			elapsed := time.Since(start)
			samples = append(samples, mbps)
			snapshot := make([]float64, len(samples))
			copy(snapshot, samples)
			select {
			case ch <- downloadProgressMsg{DownloadState{Speed: mbps, Samples: snapshot, Elapsed: elapsed}, false}:
			default:
			}
		})
		if err != nil {
			ch <- errMsg{err}
		} else {
			ch <- downloadProgressMsg{DownloadState{Speed: speed, Samples: samples, Elapsed: time.Since(start)}, true}
		}
		close(ch)
	}()
	return listenDownload(ch), ch
}

func listenDownload(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

package tui

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noahterenzianii/gospeed/internal/endpoints"
	"github.com/noahterenzianii/gospeed/internal/speedtest"
	"github.com/noahterenzianii/gospeed/internal/tuning"
)

const serverListURL = "https://librespeed.org/backend-servers/servers.php"

const (
	tuningLowThreshold  = 50   // Mbps
	tuningMidThreshold  = 200  // Mbps
	tuningStepDuration  = 3 * time.Second
	tuningMinBuffer     = 16 * 1024
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
	var streams int
	var bufSize int

	switch dir {
	case dirDownload:
		url = m.server.URL(m.server.DlURL)
		fn = speedtest.MeasureDownload
		streams = m.cfg.DownloadStreams
		bufSize = m.cfg.DownloadBufferSize
	case dirUpload:
		url = m.server.URL(m.server.UlURL)
		fn = speedtest.MeasureUpload
		streams = m.cfg.UploadStreams
		bufSize = m.cfg.UploadBufferSize
	}

	ch := make(chan tea.Msg, 100)
	go func() {
		var mu sync.Mutex
		var samples []float64
		start := time.Now()
		speed, err := fn(url, m.cfg.TransferDuration, streams, bufSize, func(mbps float64) {
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

func makeTuningMeasure(fn func(string, time.Duration, int, int, speedtest.ProgressFunc) (float64, error)) tuning.MeasureFunc {
	return func(url string, duration time.Duration, streams int, bufSize int, onProgress func(float64)) (float64, error) {
		return fn(url, duration, streams, bufSize, onProgress)
	}
}

func tuningProgressFunc(ch chan tea.Msg) func(tuning.State) {
	return func(st tuning.State) {
		ch <- tuningProgressMsg{
			stepLabel: st.StepLabel,
			value:     st.Current,
			streams:   st.Streams,
			bufSize:   st.BufSize,
			phase:     int(st.Phase),
		}
	}
}

// startTuningCmd starts the tuning process in a goroutine, streaming progress
// messages via a channel until the final tuningResultMsg is sent.
func (m Model) startTuningCmd() (tea.Model, tea.Cmd) {
	ch := make(chan tea.Msg, 100)
	m.tuningCh = ch
	m.tuningLabel = ""
	m.tuningValue = 0
	m.err = nil
	m.phase = phaseTuning

	ctx, cancel := context.WithCancel(context.Background())
	m.tuningCancel = cancel

	go func() {
		defer cancel()
		m.runTuning(ctx, ch)
	}()

	return m, tea.Batch(listenTransfer(ch), m.spinner.Tick)
}

func (m Model) runTuning(ctx context.Context, ch chan tea.Msg) {
	best, err := m.tuningSelectServer(ctx, ch)
	if err != nil {
		ch <- tuningResultMsg{err: fmt.Errorf("server selection: %w", err)}
		close(ch)
		return
	}

	progress := tuningProgressFunc(ch)

	ch <- tuningProgressMsg{stepLabel: "tuning download..."}
	dlResult, err := m.tuneDirection(ctx, best, best.DlURL, speedtest.MeasureDownload, false, progress)
	if err != nil {
		ch <- tuningResultMsg{err: fmt.Errorf("download tuning: %w", err)}
		close(ch)
		return
	}

	ch <- tuningProgressMsg{stepLabel: "tuning upload..."}
	ulResult, err := m.tuneDirection(ctx, best, best.UlURL, speedtest.MeasureUpload, true, progress)
	if err != nil {
		ch <- tuningResultMsg{
			dlStreams:    dlResult.Streams,
			dlBufferSize: dlResult.BufferSize,
			dlBandwidth:  dlResult.RawBandwidth,
			err:          fmt.Errorf("upload tuning: %w", err),
		}
		close(ch)
		return
	}

	ch <- tuningResultMsg{
		dlStreams:    dlResult.Streams,
		dlBufferSize: dlResult.BufferSize,
		dlBandwidth:  dlResult.RawBandwidth,
		ulStreams:    ulResult.Streams,
		ulBufferSize: ulResult.BufferSize,
		ulBandwidth:  ulResult.RawBandwidth,
		elapsed:      dlResult.Elapsed + ulResult.Elapsed,
	}
	close(ch)
}

func (m Model) tuningSelectServer(ctx context.Context, ch chan tea.Msg) (*endpoints.Server, error) {
	if m.server != nil {
		ch <- tuningProgressMsg{stepLabel: "server selected"}
		return m.server, nil
	}
	servers, err := endpoints.FetchServers(serverListURL, m.cfg.ServerListTimeout)
	if err != nil {
		return nil, err
	}
	best := endpoints.FindBestServer(servers, m.cfg.MaxConcurrentPings, m.cfg.PingAttempts, m.cfg.PingTimeout)
	if best == nil {
		return nil, fmt.Errorf("no reachable server")
	}
	ch <- tuningProgressMsg{stepLabel: "server selected"}
	return best, nil
}

func (m Model) tuneDirection(
	ctx context.Context,
	best *endpoints.Server,
	urlPath string,
	fn func(string, time.Duration, int, int, speedtest.ProgressFunc) (float64, error),
	isUpload bool,
	progress func(tuning.State),
) (tuning.Result, error) {
	opts := tuning.Options{
		MaxStreams:   m.cfg.MaxStreams,
		StepDuration: tuningStepDuration,
		LowThreshold: tuningLowThreshold,
		MidThreshold: tuningMidThreshold,
		MinBuffer:    tuningMinBuffer,
		MaxBuffer:    m.cfg.MaxBuffer,
		IsUpload:     isUpload,
	}
	return tuning.Tune(ctx, best.URL(urlPath), makeTuningMeasure(fn), progress, opts)
}

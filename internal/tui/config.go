package tui

import (
	"fmt"
	"time"
)

type Config struct {
	PingSamples        int
	TransferDuration   time.Duration
	DownloadStreams    int
	DownloadBufferSize int
	UploadStreams      int
	UploadBufferSize   int
	ClientInfoTimeout  time.Duration
	ServerListTimeout  time.Duration
	PingTimeout        time.Duration
	MaxConcurrentPings int
	PingAttempts       int
}

type configField struct {
	label     string
	isSection bool
	value     func(*Config) string
	apply     func(*Config, int)
}

var configFields = []configField{
	{label: "test", isSection: true},
	{
		label: "ping samples",
		value: func(c *Config) string { return fmt.Sprintf("%d", c.PingSamples) },
		apply: func(c *Config, d int) {
			c.PingSamples = clamp(c.PingSamples+d*10, 10, 1_000)
		},
	},
	{
		label: "transfer duration",
		value: func(c *Config) string {
			return fmt.Sprintf("%.0fs", c.TransferDuration.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.TransferDuration.Seconds() + float64(d*5)
			c.TransferDuration = time.Duration(clamp(int(v), 5, 120)) * time.Second
		},
	},
	{label: "download", isSection: true},
	{
		label: "download streams",
		value: func(c *Config) string { return fmt.Sprintf("%d", c.DownloadStreams) },
		apply: func(c *Config, d int) {
			c.DownloadStreams = clamp(c.DownloadStreams+d, 1, 64)
		},
	},
	{label: "upload", isSection: true},
	{
		label: "upload streams",
		value: func(c *Config) string { return fmt.Sprintf("%d", c.UploadStreams) },
		apply: func(c *Config, d int) {
			c.UploadStreams = clamp(c.UploadStreams+d, 1, 64)
		},
	},
	{label: "timeouts", isSection: true},
	{
		label: "client info timeout",
		value: func(c *Config) string {
			return fmt.Sprintf("%.0fs", c.ClientInfoTimeout.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.ClientInfoTimeout.Seconds() + float64(d)
			c.ClientInfoTimeout = time.Duration(clamp(int(v), 1, 30)) * time.Second
		},
	},
	{
		label: "server list timeout",
		value: func(c *Config) string {
			return fmt.Sprintf("%.0fs", c.ServerListTimeout.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.ServerListTimeout.Seconds() + float64(d)
			c.ServerListTimeout = time.Duration(clamp(int(v), 1, 30)) * time.Second
		},
	},
	{
		label: "ping timeout",
		value: func(c *Config) string {
			return fmt.Sprintf("%.0fs", c.PingTimeout.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.PingTimeout.Seconds() + float64(d)
			c.PingTimeout = time.Duration(clamp(int(v), 1, 15)) * time.Second
		},
	},
}

func clamp(val, low, high int) int {
	if val < low {
		return low
	}
	if val > high {
		return high
	}
	return val
}

func prevField(cur int) int {
	for i := cur - 1; i >= 0; i-- {
		if !configFields[i].isSection {
			return i
		}
	}
	return cur
}

func nextField(cur int) int {
	for i := cur + 1; i < len(configFields); i++ {
		if !configFields[i].isSection {
			return i
		}
	}
	return cur
}

func (m Model) applyConfigDelta(delta int) {
	f := configFields[m.configCursor]
	f.apply(m.cfg, delta)
}

func defaultConfig() *Config {
	return &Config{
		PingSamples:        200,
		TransferDuration:   15 * time.Second,
		DownloadStreams:    8,
		DownloadBufferSize: 1024 * 1024,
		UploadStreams:      4,
		UploadBufferSize:   256 * 1024,
		ClientInfoTimeout:  5 * time.Second,
		ServerListTimeout:  10 * time.Second,
		PingTimeout:        2 * time.Second,
		MaxConcurrentPings: 20,
		PingAttempts:       3,
	}
}

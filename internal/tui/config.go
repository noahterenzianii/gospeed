package tui

import (
	"fmt"
	"time"
)

type Config struct {
	PingSamples        int
	TransferDuration   time.Duration
	TransferStreams    int
	BufferSize         int
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
			c.PingSamples = clamp(c.PingSamples+d*10, 10, 1000)
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
	{
		label: "transfer streams",
		value: func(c *Config) string { return fmt.Sprintf("%d", c.TransferStreams) },
		apply: func(c *Config, d int) {
			c.TransferStreams = clamp(c.TransferStreams+d, 1, 32)
		},
	},
	{
		label: "buffer size",
		value: func(c *Config) string {
			return fmt.Sprintf("%d KB", c.BufferSize/1024)
		},
		apply: func(c *Config, d int) {
			c.BufferSize = clamp(c.BufferSize+d*65536, 65536, 4194304)
		},
	},
	{label: "timeouts", isSection: true},
	{
		label: "client info timeout",
		value: func(c *Config) string {
			return fmt.Sprintf("%.1fs", c.ClientInfoTimeout.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.ClientInfoTimeout.Seconds() + float64(d)
			c.ClientInfoTimeout = time.Duration(clamp(int(v), 1, 30)) * time.Second
		},
	},
	{
		label: "server list timeout",
		value: func(c *Config) string {
			return fmt.Sprintf("%.1fs", c.ServerListTimeout.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.ServerListTimeout.Seconds() + float64(d)
			c.ServerListTimeout = time.Duration(clamp(int(v), 1, 30)) * time.Second
		},
	},
	{
		label: "ping timeout",
		value: func(c *Config) string {
			return fmt.Sprintf("%.1fs", c.PingTimeout.Seconds())
		},
		apply: func(c *Config, d int) {
			v := c.PingTimeout.Seconds() + float64(d)
			c.PingTimeout = time.Duration(clamp(int(v), 1, 15)) * time.Second
		},
	},
	{label: "server selection", isSection: true},
	{
		label: "max concurrent pings",
		value: func(c *Config) string { return fmt.Sprintf("%d", c.MaxConcurrentPings) },
		apply: func(c *Config, d int) {
			c.MaxConcurrentPings = clamp(c.MaxConcurrentPings+d*5, 5, 100)
		},
	},
	{
		label: "ping attempts",
		value: func(c *Config) string { return fmt.Sprintf("%d", c.PingAttempts) },
		apply: func(c *Config, d int) {
			c.PingAttempts = clamp(c.PingAttempts+d, 1, 20)
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
		TransferStreams:    8,
		BufferSize:         1024 * 1024,
		ClientInfoTimeout:  5 * time.Second,
		ServerListTimeout:  10 * time.Second,
		PingTimeout:        2 * time.Second,
		MaxConcurrentPings: 20,
		PingAttempts:       3,
	}
}

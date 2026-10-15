package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

var (
	configDir  = "gospeed"
	configFile = "config.json"
)

type Config struct {
	PingSamples        int
	TransferDuration   time.Duration
	DownloadStreams    int
	DownloadBufferSize int
	UploadStreams      int
	UploadBufferSize   int
	MaxStreams         int
	MaxBuffer          int
	ClientInfoTimeout  time.Duration
	ServerListTimeout  time.Duration
	PingTimeout        time.Duration
	MaxConcurrentPings int
	PingAttempts       int
}

type fileConfig struct {
	PingSamples        int   `json:"ping_samples"`
	TransferDuration   int64 `json:"transfer_duration"`
	DownloadStreams    int   `json:"download_streams"`
	DownloadBufferSize int   `json:"download_buffer_size"`
	UploadStreams      int   `json:"upload_streams"`
	UploadBufferSize   int   `json:"upload_buffer_size"`
	MaxStreams         int   `json:"max_streams"`
	MaxBuffer          int   `json:"max_buffer"`
	ClientInfoTimeout  int64 `json:"client_info_timeout"`
	ServerListTimeout  int64 `json:"server_list_timeout"`
	PingTimeout        int64 `json:"ping_timeout"`
	MaxConcurrentPings int   `json:"max_concurrent_pings"`
	PingAttempts       int   `json:"ping_attempts"`
}

func DefaultConfig() *Config {
	return &Config{
		PingSamples:        200,
		TransferDuration:   15 * time.Second,
		DownloadStreams:    8,
		DownloadBufferSize: 1024 * 1024,
		UploadStreams:      4,
		UploadBufferSize:   256 * 1024,
		MaxStreams:         64,
		MaxBuffer:          4 * 1024 * 1024,
		ClientInfoTimeout:  5 * time.Second,
		ServerListTimeout:  10 * time.Second,
		PingTimeout:        2 * time.Second,
		MaxConcurrentPings: 20,
		PingAttempts:       3,
	}
}

func (c *Config) toFileConfig() *fileConfig {
	return &fileConfig{
		PingSamples:        c.PingSamples,
		TransferDuration:   int64(c.TransferDuration),
		DownloadStreams:    c.DownloadStreams,
		DownloadBufferSize: c.DownloadBufferSize,
		UploadStreams:      c.UploadStreams,
		UploadBufferSize:   c.UploadBufferSize,
		MaxStreams:         c.MaxStreams,
		MaxBuffer:          c.MaxBuffer,
		ClientInfoTimeout:  int64(c.ClientInfoTimeout),
		ServerListTimeout:  int64(c.ServerListTimeout),
		PingTimeout:        int64(c.PingTimeout),
		MaxConcurrentPings: c.MaxConcurrentPings,
		PingAttempts:       c.PingAttempts,
	}
}

func fromFileConfig(fc *fileConfig) *Config {
	return &Config{
		PingSamples:        fc.PingSamples,
		TransferDuration:   time.Duration(fc.TransferDuration),
		DownloadStreams:    fc.DownloadStreams,
		DownloadBufferSize: fc.DownloadBufferSize,
		UploadStreams:      fc.UploadStreams,
		UploadBufferSize:   fc.UploadBufferSize,
		MaxStreams:         fc.MaxStreams,
		MaxBuffer:          fc.MaxBuffer,
		ClientInfoTimeout:  time.Duration(fc.ClientInfoTimeout),
		ServerListTimeout:  time.Duration(fc.ServerListTimeout),
		PingTimeout:        time.Duration(fc.PingTimeout),
		MaxConcurrentPings: fc.MaxConcurrentPings,
		PingAttempts:       fc.PingAttempts,
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configDir, configFile), nil
}

func Load() *Config {
	path, err := configPath()
	if err != nil {
		return DefaultConfig()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return DefaultConfig()
	}

	if !fc.isValid() {
		return DefaultConfig()
	}

	return fromFileConfig(&fc)
}

func (fc *fileConfig) isValid() bool {
	if fc.PingSamples < 10 || fc.PingSamples > 1000 {
		return false
	}
	if fc.TransferDuration < 5*1e9 || fc.TransferDuration > 120*1e9 {
		return false
	}
	if fc.DownloadStreams < 1 || fc.DownloadStreams > fc.MaxStreams {
		return false
	}
	if fc.UploadStreams < 1 || fc.UploadStreams > fc.MaxStreams {
		return false
	}
	if fc.MaxStreams < 4 || fc.MaxStreams > 256 {
		return false
	}
	if fc.MaxBuffer < 256*1024 || fc.MaxBuffer > 64*1024*1024 {
		return false
	}
	if fc.DownloadBufferSize < 64*1024 || fc.DownloadBufferSize > fc.MaxBuffer {
		return false
	}
	if fc.UploadBufferSize < 64*1024 || fc.UploadBufferSize > fc.MaxBuffer {
		return false
	}
	if fc.ClientInfoTimeout < 1*1e9 || fc.ClientInfoTimeout > 30*1e9 {
		return false
	}
	if fc.ServerListTimeout < 1*1e9 || fc.ServerListTimeout > 30*1e9 {
		return false
	}
	if fc.PingTimeout < 1*1e9 || fc.PingTimeout > 15*1e9 {
		return false
	}
	if fc.MaxConcurrentPings < 5 || fc.MaxConcurrentPings > 100 {
		return false
	}
	if fc.PingAttempts < 1 || fc.PingAttempts > 20 {
		return false
	}

	return true
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fc := cfg.toFileConfig()
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}
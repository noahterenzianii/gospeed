package tuning

import "time"

type MeasureFunc func(url string, duration time.Duration, streams int, bufSize int, onProgress func(float64)) (float64, error)

type ProgressFunc func(State)

type Phase int

const (
	PhaseBandwidthEstimate Phase = iota
	PhaseStreamSearch
	PhaseBufferSearch
)

type State struct {
	Phase     Phase
	StepLabel string
	Current   float64
	Elapsed   time.Duration
}

type StepResult struct {
	Param      int
	Throughput float64
	Variance   float64
}

type Result struct {
	Streams      int
	BufferSize   int
	RawBandwidth float64
}

type Options struct {
	MaxStreams     int
	MaxBufferSize  int
	StepDuration   time.Duration
	DefaultStreams int
	DefaultBufSize int
	LowThreshold   float64
	MinBuffer      int
	MaxBuffer      int
	BufferStep     int
}

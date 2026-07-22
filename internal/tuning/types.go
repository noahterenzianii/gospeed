package tuning

import "time"

type MeasureFunc func(
	url string,
	duration time.Duration,
	streams int,
	bufSize int,
	onProgress func(float64),
) (float64, error)

type ProgressFunc func(State)

type SearchTarget int

const (
	SearchStreams SearchTarget = iota
	SearchBuffer
)

type Phase int

const (
	PhaseBandwidthProbe Phase = iota
	PhaseStreamSweep
	PhaseBufferSweep
	PhaseValidation
	PhaseComplete
)

type State struct {
	Phase     Phase
	StepLabel string
	Current   float64
	Elapsed   time.Duration
	Streams   int
	BufSize   int
}

type Result struct {
	Streams      int
	BufferSize   int
	RawBandwidth float64
	Elapsed      time.Duration
}

type Options struct {
	MaxStreams   int
	StepDuration time.Duration
	LowThreshold float64
	MidThreshold float64
	MinBuffer    int
	MaxBuffer    int
	IsUpload     bool
}

type Point struct {
	Param      int
	Throughput float64
}

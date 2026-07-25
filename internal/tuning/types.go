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

const (
	defaultSweepDegradeThreshold = 0.005
	defaultSweepBestThreshold    = 0.02
	defaultImprovementThreshold  = 0.01
	defaultTieThreshold          = 0.01
	defaultMaxCV                 = 0.30
	defaultMaxRetries            = 1
	defaultWarmupPercent         = 40
)

type Options struct {
	MaxStreams   int
	StepDuration time.Duration
	LowThreshold float64
	MidThreshold float64
	MinBuffer    int
	MaxBuffer    int
	IsUpload     bool

	SweepDegradeThreshold float64
	SweepBestThreshold    float64
	ImprovementThreshold  float64
	TieThreshold          float64
	MaxCV                 float64
	MaxRetries            int
	WarmupPercent         int
}

func (o Options) sweepDegradeThreshold() float64 {
	if o.SweepDegradeThreshold <= 0 {
		return defaultSweepDegradeThreshold
	}
	return o.SweepDegradeThreshold
}

func (o Options) sweepBestThreshold() float64 {
	if o.SweepBestThreshold <= 0 {
		return defaultSweepBestThreshold
	}
	return o.SweepBestThreshold
}

func (o Options) improvementThreshold() float64 {
	if o.ImprovementThreshold <= 0 {
		return defaultImprovementThreshold
	}
	return o.ImprovementThreshold
}

func (o Options) tieThreshold() float64 {
	if o.TieThreshold <= 0 {
		return defaultTieThreshold
	}
	return o.TieThreshold
}

func (o Options) maxCV() float64 {
	if o.MaxCV <= 0 {
		return defaultMaxCV
	}
	return o.MaxCV
}

func (o Options) maxRetries() int {
	if o.MaxRetries <= 0 {
		return defaultMaxRetries
	}
	return o.MaxRetries
}

func (o Options) warmupPercent() int {
	if o.WarmupPercent <= 0 || o.WarmupPercent >= 100 {
		return defaultWarmupPercent
	}
	return o.WarmupPercent
}

type Point struct {
	Param      int
	Throughput float64
}

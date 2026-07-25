package tuning

import (
	"context"
	"fmt"
)

type bandwidthTier int

const (
	tierHigh bandwidthTier = iota
	tierMid
	tierLow
)

func bwTier(rawBW float64, opts Options) bandwidthTier {
	if rawBW <= 0 {
		return tierLow
	}
	if rawBW < opts.LowThreshold {
		return tierLow
	}
	if rawBW < opts.MidThreshold {
		return tierMid
	}
	return tierHigh
}

func generateStreamCandidates(rawBW float64, opts Options) []int {
	maxStreams := opts.MaxStreams
	if opts.IsUpload && maxStreams > 32 {
		maxStreams = 32
	}
	switch bwTier(rawBW, opts) {
	case tierLow:
		maxStreams = 16
	case tierMid:
		if maxStreams > 32 {
			maxStreams = 32
		}
	}

	var candidates []int
	for n := 1; n <= maxStreams; {
		candidates = append(candidates, n)
		if n < 8 {
			n *= 2
		} else {
			n += n / 2
		}
	}
	return candidates
}

func generateBufferCandidates(rawBW float64, opts Options) []int {
	maxBuffer := opts.MaxBuffer
	switch bwTier(rawBW, opts) {
	case tierLow:
		maxBuffer = 512 * 1024
	case tierMid:
		maxBuffer = 1024 * 1024
	case tierHigh:
		if opts.IsUpload && maxBuffer > 1024*1024 {
			maxBuffer = 1024 * 1024
		}
	}

	base := 16 * 1024
	var candidates []int
	for v := base; v <= maxBuffer; v *= 2 {
		if v >= opts.MinBuffer {
			candidates = append(candidates, v)
		}
	}
	return candidates
}

func selectSweepBuffer(rawBW float64, opts Options) int {
	var buf int
	if opts.IsUpload {
		switch bwTier(rawBW, opts) {
		case tierLow:
			buf = 128 * 1024
		case tierMid:
			buf = 256 * 1024
		default:
			buf = 512 * 1024
		}
	} else {
		switch bwTier(rawBW, opts) {
		case tierLow:
			buf = 256 * 1024
		case tierMid:
			buf = 512 * 1024
		default:
			buf = 1024 * 1024
		}
	}
	if buf > opts.MaxBuffer {
		return opts.MaxBuffer
	}
	return buf
}

func sweepParameter(
	ctx context.Context,
	url string,
	candidates []int,
	fixed int,
	target SearchTarget,
	opts Options,
	measure MeasureFunc,
	onProgress ProgressFunc,
) ([]Point, error) {
	var points []Point
	var bestThroughput float64
	var nerr int
	sweepDegrade := opts.sweepDegradeThreshold()
	sweepBest := opts.sweepBestThreshold()

	for _, v := range candidates {
		if err := ctx.Err(); err != nil {
			return points, err
		}

		streams, bufSize := v, fixed
		label, display := "streams", v
		if target == SearchBuffer {
			streams, bufSize = fixed, v
			label, display = "kb buffer", v/1024
		}

		result, err := measureWithQuality(url, streams, bufSize, opts.StepDuration, measure, opts)
		if err != nil {
			nerr++
			continue
		}

		if result.throughput > bestThroughput {
			bestThroughput = result.throughput
		}

		phase := PhaseStreamSweep
		if target == SearchBuffer {
			phase = PhaseBufferSweep
		}

		onProgress(State{
			Phase:     phase,
			StepLabel: fmt.Sprintf("testing %d %s", display, label),
			Current:   result.throughput,
			Elapsed:   opts.StepDuration,
			Streams:   streams,
			BufSize:   bufSize,
		})

		points = append(points, Point{Param: v, Throughput: result.throughput})

		if len(points) >= 2 {
			prev := points[len(points)-2].Throughput
			if result.throughput < prev*(1-sweepDegrade) {
				break
			}
		}

		if bestThroughput > 0 && result.throughput < bestThroughput*(1-sweepBest) {
			break
		}
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("all sweep measurements failed (%d/%d failed)", nerr, len(candidates))
	}

	return points, nil
}

func findBest(points []Point, opts Options) (int, error) {
	if len(points) == 0 {
		return 0, fmt.Errorf("no measurement points to evaluate")
	}

	improvement := opts.improvementThreshold()
	tie := opts.tieThreshold()

	best := points[0]
	for _, p := range points[1:] {
		imp := p.Throughput > best.Throughput*(1+improvement)
		tieOk := p.Throughput >= best.Throughput*(1-tie) && p.Param > best.Param
		if imp || tieOk {
			best = p
		}
	}

	return best.Param, nil
}

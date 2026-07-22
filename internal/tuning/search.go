package tuning

import (
	"context"
	"fmt"
)

func generateStreamCandidates(rawBW float64, opts Options) []int {
	maxStreams := opts.MaxStreams
	if opts.IsUpload && maxStreams > 32 {
		maxStreams = 32
	}
	if rawBW > 0 && rawBW < opts.LowThreshold {
		maxStreams = 16
	} else if rawBW > 0 && rawBW < opts.MidThreshold {
		if maxStreams > 32 {
			maxStreams = 32
		}
	}

	base := []int{1, 2, 4, 8, 12, 16, 24, 32, 48, 64}
	var candidates []int
	for _, v := range base {
		if v <= maxStreams {
			candidates = append(candidates, v)
		}
	}
	return candidates
}

func generateBufferCandidates(rawBW float64, opts Options) []int {
	maxBuffer := opts.MaxBuffer
	if rawBW > 0 && rawBW < opts.LowThreshold {
		maxBuffer = 512 * 1024
	} else if rawBW > 0 && rawBW < opts.MidThreshold {
		maxBuffer = 1024 * 1024
	} else if opts.IsUpload && maxBuffer > 1024*1024 {
		maxBuffer = 1024 * 1024
	}

	base := []int{
		16 * 1024, 32 * 1024, 64 * 1024, 128 * 1024,
		256 * 1024, 512 * 1024, 1024 * 1024,
		2 * 1024 * 1024, 4 * 1024 * 1024,
	}
	var candidates []int
	for _, v := range base {
		if v > maxBuffer {
			continue
		}
		if v >= opts.MinBuffer {
			candidates = append(candidates, v)
		}
	}
	return candidates
}

func selectSweepBuffer(rawBW float64, opts Options) int {
	if opts.IsUpload {
		switch {
		case rawBW > 0 && rawBW < opts.LowThreshold:
			return 128 * 1024
		case rawBW > 0 && rawBW < opts.MidThreshold:
			return 256 * 1024
		default:
			return 512 * 1024
		}
	}
	switch {
	case rawBW > 0 && rawBW < opts.LowThreshold:
		return 256 * 1024
	case rawBW > 0 && rawBW < opts.MidThreshold:
		return 512 * 1024
	default:
		return 1024 * 1024
	}
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

		result, err := measureWithQuality(url, streams, bufSize, opts.StepDuration, measure)
		if err != nil {
			continue
		}

		if result.throughput > bestThroughput {
			bestThroughput = result.throughput
		}

		points = append(points, Point{Param: v, Throughput: result.throughput})

		if len(points) >= 2 {
			prev := points[len(points)-2].Throughput
			if result.throughput < prev*0.995 {
				break
			}
		}

		if bestThroughput > 0 && result.throughput < bestThroughput*0.98 {
			break
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
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("all sweep measurements failed")
	}

	return points, nil
}

func findBest(points []Point) int {
	if len(points) == 0 {
		return 0
	}

	best := points[0]
	for _, p := range points[1:] {
		imp := p.Throughput > best.Throughput*1.01
		tie := p.Throughput >= best.Throughput*0.99 && p.Param > best.Param
		if imp || tie {
			best = p
		}
	}

	return best.Param
}

package tuning

import (
	"context"
	"fmt"
)

func streamValues(opts Options, rawBandwidth float64) []int {
	var vals []int
	for v := 1; v <= opts.MaxStreams; v *= 2 {
		if rawBandwidth < opts.LowThreshold && v > 16 {
			break
		}
		vals = append(vals, v)
	}
	return vals
}

func bufferValues(opts Options, rawBandwidth float64) []int {
	var vals []int
	for v := opts.MinBuffer; v <= opts.MaxBuffer; v *= 2 {
		if rawBandwidth < opts.LowThreshold && v > 1024*1024 {
			break
		}
		vals = append(vals, v)
	}
	return vals
}

func hillClimb(ctx context.Context, url string, values []int, fixed int, isBuffer bool, opts Options, measure MeasureFunc, phase Phase, onProgress ProgressFunc) (StepResult, error) {
	if len(values) == 0 {
		return StepResult{}, fmt.Errorf("no values to test")
	}

	var best StepResult
	plateauCount := 0

	for _, v := range values {
		if err := ctx.Err(); err != nil {
			return best, err
		}

		streams, bufSize := fixed, v
		if isBuffer {
			streams, bufSize = v, fixed
		}

		throughput, variance, err := runMeasurement(url, streams, bufSize, opts.StepDuration, measure)
		if err != nil {
			continue
		}

		sr := StepResult{Param: v, Throughput: throughput, Variance: variance}

		paramName := "streams"
		if isBuffer {
			paramName = "buffer"
		}

		onProgress(State{
			Phase:     phase,
			StepLabel: fmt.Sprintf("testing %d %s", v, paramName),
			Current:   throughput,
			Elapsed:   opts.StepDuration,
		})

		if best.Throughput <= 0 {
			best = sr
			continue
		}

		switch {
		case throughput > best.Throughput*1.05:
			best = sr
			plateauCount = 0
		case throughput < best.Throughput*0.95:
			return best, nil
		default:
			if throughput > best.Throughput {
				best = sr
			}
			plateauCount++
			if plateauCount >= 3 {
				return best, nil
			}
		}
	}

	if best.Throughput <= 0 {
		return StepResult{}, fmt.Errorf("all test steps failed")
	}

	return best, nil
}

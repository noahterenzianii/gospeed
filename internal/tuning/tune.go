package tuning

import (
	"context"
	"fmt"
	"time"
)

func Tune(
	ctx context.Context,
	url string,
	measure MeasureFunc,
	onProgress ProgressFunc,
	opts Options,
) (Result, error) {
	start := time.Now()
	var res Result

	if opts.MaxStreams <= 0 {
		opts.MaxStreams = 64
	}
	if opts.MinBuffer < 16*1024 {
		opts.MinBuffer = 16 * 1024
	}
	if opts.MaxBuffer < opts.MinBuffer {
		opts.MaxBuffer = opts.MinBuffer
	}
	if opts.StepDuration <= 0 {
		opts.StepDuration = 3 * time.Second
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}

	rawBW, err := probeBandwidth(ctx, url, measure, onProgress, opts)
	if err != nil {
		rawBW = 0
	}
	res.RawBandwidth = rawBW

	if err := ctx.Err(); err != nil {
		return res, err
	}

	sweepBuf := selectSweepBuffer(rawBW, opts)
	streamCandidates := generateStreamCandidates(rawBW, opts)

	streamPoints, err := sweepParameter(
		ctx, url, streamCandidates, sweepBuf,
		SearchStreams, opts, measure, onProgress,
	)
	if err != nil {
		return res, fmt.Errorf("stream sweep failed: %w", err)
	}
	bestStreams, err := findBest(streamPoints, opts)
	if err != nil {
		return res, fmt.Errorf("stream sweep: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}

	bufferCandidates := generateBufferCandidates(rawBW, opts)

	bufferPoints, err := sweepParameter(
		ctx, url, bufferCandidates, bestStreams,
		SearchBuffer, opts, measure, onProgress,
	)
	if err != nil {
		bestBuffer := opts.MinBuffer
		if len(bufferCandidates) > 0 {
			bestBuffer = bufferCandidates[0]
		}
		res.Streams = bestStreams
		res.BufferSize = bestBuffer
		res.Elapsed = time.Since(start)
		onProgress(State{
			Phase:     PhaseComplete,
			StepLabel: fmt.Sprintf("tuned (fallback): %d streams · %d kb buffer", bestStreams, bestBuffer/1024),
			Elapsed:   res.Elapsed,
			Streams:   bestStreams,
			BufSize:   bestBuffer,
		})
		return res, nil
	}
	bestBuffer, err := findBest(bufferPoints, opts)
	if err != nil {
		bestBuffer := opts.MinBuffer
		if len(bufferCandidates) > 0 {
			bestBuffer = bufferCandidates[0]
		}
		res.Streams = bestStreams
		res.BufferSize = bestBuffer
		res.Elapsed = time.Since(start)
		onProgress(State{
			Phase:     PhaseComplete,
			StepLabel: fmt.Sprintf("tuned (fallback): %d streams · %d kb buffer", bestStreams, bestBuffer/1024),
			Elapsed:   res.Elapsed,
			Streams:   bestStreams,
			BufSize:   bestBuffer,
		})
		return res, nil
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}

	if err := validateConfig(ctx, url, bestStreams, bestBuffer, opts, measure, onProgress); err != nil {
		onProgress(State{
			Phase:     PhaseValidation,
			StepLabel: fmt.Sprintf("proceeding with best guess (%d streams · %d kb buffer)", bestStreams, bestBuffer/1024),
		})
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}

	onProgress(State{
		Phase:     PhaseComplete,
		StepLabel: fmt.Sprintf("tuned: %d streams · %d kb buffer", bestStreams, bestBuffer/1024),
		Current:   0,
		Elapsed:   time.Since(start),
		Streams:   bestStreams,
		BufSize:   bestBuffer,
	})

	res.Streams = bestStreams
	res.BufferSize = bestBuffer
	res.Elapsed = time.Since(start)
	return res, nil
}

func probeBandwidth(
	ctx context.Context,
	url string,
	measure MeasureFunc,
	onProgress ProgressFunc,
	opts Options,
) (float64, error) {
	onProgress(State{
		Phase:     PhaseBandwidthProbe,
		StepLabel: "probing bandwidth...",
	})

	probeDuration := opts.StepDuration
	if probeDuration < time.Second {
		probeDuration = time.Second
	}

	var throughput float64
	var err error
	warmup := opts.warmupPercent()
	if opts.IsUpload {
		throughput, _, err = runMeasurement(url, 4, 256*1024, probeDuration, measure, warmup)
		if err != nil || throughput <= 0 {
			onProgress(State{
				Phase:     PhaseBandwidthProbe,
				StepLabel: "retrying probe with conservative parameters...",
			})
			throughput, _, err = runMeasurement(url, 2, 128*1024, probeDuration, measure, warmup)
		}
	} else {
		throughput, _, err = runMeasurement(url, 8, 512*1024, probeDuration, measure, warmup)
		if err != nil || throughput <= 0 {
			onProgress(State{
				Phase:     PhaseBandwidthProbe,
				StepLabel: "retrying probe with conservative parameters...",
			})
			throughput, _, err = runMeasurement(url, 4, 256*1024, probeDuration, measure, warmup)
		}
	}

	if throughput <= 0 {
		if err != nil {
			return 0, fmt.Errorf("bandwidth probe failed: %w", err)
		}
		return 0, fmt.Errorf("bandwidth probe failed: zero throughput")
	}

	onProgress(State{
		Phase:     PhaseBandwidthProbe,
		StepLabel: "probing bandwidth...",
		Current:   throughput,
		Elapsed:   probeDuration,
	})
	return throughput, nil
}

func validateConfig(
	ctx context.Context,
	url string,
	streams, bufSize int,
	opts Options,
	measure MeasureFunc,
	onProgress ProgressFunc,
) error {
	result, err := measureWithQuality(url, streams, bufSize, opts.StepDuration, measure, opts)
	if err != nil {
		onProgress(State{
			Phase:     PhaseValidation,
			StepLabel: fmt.Sprintf("validation failed for %d streams · %d kb buffer", streams, bufSize/1024),
		})
		return err
	}

	onProgress(State{
		Phase:     PhaseValidation,
		StepLabel: fmt.Sprintf("validating %d streams · %d kb buffer", streams, bufSize/1024),
		Current:   result.throughput,
		Elapsed:   opts.StepDuration,
		Streams:   streams,
		BufSize:   bufSize,
	})
	return nil
}

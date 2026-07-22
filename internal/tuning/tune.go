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

	if err := ctx.Err(); err != nil {
		return res, err
	}

	rawBW := probeBandwidth(ctx, url, measure, onProgress, opts)
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
	bestStreams := findBest(streamPoints)

	if err := ctx.Err(); err != nil {
		return res, err
	}

	bufferCandidates := generateBufferCandidates(rawBW, opts)

	bufferPoints, err := sweepParameter(
		ctx, url, bufferCandidates, bestStreams,
		SearchBuffer, opts, measure, onProgress,
	)
	if err != nil {
		bestBuffer := bufferCandidates[len(bufferCandidates)-1]
		res.Streams = bestStreams
		res.BufferSize = bestBuffer
		res.Elapsed = time.Since(start)
		return res, nil
	}
	bestBuffer := findBest(bufferPoints)

	if err := ctx.Err(); err != nil {
		return res, err
	}

	validateConfig(ctx, url, bestStreams, bestBuffer, opts, measure, onProgress)

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
) float64 {
	onProgress(State{
		Phase:     PhaseBandwidthProbe,
		StepLabel: "probing bandwidth...",
	})

	probeDuration := opts.StepDuration
	if probeDuration < time.Second {
		probeDuration = time.Second
	}

	var throughput float64
	var probeErr error
	if opts.IsUpload {
		throughput, _, probeErr = runMeasurement(url, 4, 256*1024, probeDuration, measure)
		if probeErr != nil || throughput <= 0 {
			throughput, _, probeErr = runMeasurement(url, 2, 128*1024, probeDuration, measure)
			if probeErr != nil || throughput <= 0 {
				throughput = 50
			}
		}
	} else {
		throughput, _, probeErr = runMeasurement(url, 8, 512*1024, probeDuration, measure)
		if probeErr != nil || throughput <= 0 {
			throughput, _, probeErr = runMeasurement(url, 4, 256*1024, probeDuration, measure)
			if probeErr != nil || throughput <= 0 {
				throughput = 100
			}
		}
	}

	onProgress(State{
		Phase:     PhaseBandwidthProbe,
		StepLabel: "probing bandwidth...",
		Current:   throughput,
		Elapsed:   probeDuration,
	})

	return throughput
}

func validateConfig(
	ctx context.Context,
	url string,
	streams, bufSize int,
	opts Options,
	measure MeasureFunc,
	onProgress ProgressFunc,
) {
	result, err := measureWithQuality(url, streams, bufSize, opts.StepDuration, measure)
	if err != nil {
		return
	}

	onProgress(State{
		Phase:     PhaseValidation,
		StepLabel: fmt.Sprintf("validating %d streams · %d kb buffer", streams, bufSize/1024),
		Current:   result.throughput,
		Elapsed:   opts.StepDuration,
		Streams:   streams,
		BufSize:   bufSize,
	})
}

package tuning

import (
	"context"
	"fmt"
)

func Tune(ctx context.Context, url string, measure MeasureFunc, onProgress ProgressFunc, opts Options) (Result, error) {
	var res Result

	if err := ctx.Err(); err != nil {
		return res, err
	}

	onProgress(State{Phase: PhaseBandwidthEstimate, StepLabel: "estimating bandwidth..."})
	rawBandwidth, _, err := runMeasurement(url, opts.DefaultStreams, opts.DefaultBufSize, opts.StepDuration, measure)
	if err != nil {
		return res, fmt.Errorf("bandwidth estimate failed: %w", err)
	}

	res.RawBandwidth = rawBandwidth
	onProgress(State{Phase: PhaseBandwidthEstimate, StepLabel: "estimating bandwidth...", Current: rawBandwidth, Elapsed: opts.StepDuration})

	if err := ctx.Err(); err != nil {
		return res, err
	}

	streamVals := streamValues(opts, rawBandwidth)

	bestStream, err := hillClimb(ctx, url, streamVals, opts.DefaultBufSize, SearchStreams, opts, measure, PhaseStreamSearch, onProgress)
	if err != nil {
		return res, err
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}

	bufVals := bufferValues(opts, rawBandwidth)

	bestBuffer, err := hillClimb(ctx, url, bufVals, bestStream.Param, SearchBuffer, opts, measure, PhaseBufferSearch, onProgress)
	if err != nil {
		return res, err
	}

	res.Streams = bestStream.Param
	res.BufferSize = bestBuffer.Param
	return res, nil
}

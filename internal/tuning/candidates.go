package tuning

type bandwidthTier int

const (
	tierHigh bandwidthTier = iota
	tierMid
	tierLow
)

func bwTier(rawBW float64, lowThreshold, midThreshold float64) bandwidthTier {
	if rawBW <= 0 {
		return tierLow
	}
	if rawBW < lowThreshold {
		return tierLow
	}
	if rawBW < midThreshold {
		return tierMid
	}
	return tierHigh
}

func generateStreamCandidates(rawBW float64, opts Options) []int {
	maxStreams := opts.MaxStreams
	if opts.IsUpload && maxStreams > 32 {
		maxStreams = 32
	}
	switch bwTier(rawBW, opts.LowThreshold, opts.MidThreshold) {
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
	switch bwTier(rawBW, opts.LowThreshold, opts.MidThreshold) {
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
		switch bwTier(rawBW, opts.LowThreshold, opts.MidThreshold) {
		case tierLow:
			buf = 128 * 1024
		case tierMid:
			buf = 256 * 1024
		default:
			buf = 512 * 1024
		}
	} else {
		switch bwTier(rawBW, opts.LowThreshold, opts.MidThreshold) {
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

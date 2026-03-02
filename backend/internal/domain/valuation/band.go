package valuation

import "sort"

// ComputeBand calculates min, max, average, and median for a set of values.
// Returns ErrInsufficientData if values is empty.
// The input slice is not modified.
func ComputeBand(values []float64) (*BandStats, error) {
	if len(values) == 0 {
		return nil, ErrInsufficientData
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	lo := sorted[0]
	hi := sorted[len(sorted)-1]

	var sum float64
	for _, v := range sorted {
		sum += v
	}
	avg := sum / float64(len(sorted))

	med := median(sorted)

	return &BandStats{
		Min:    lo,
		Max:    hi,
		Avg:    avg,
		Median: med,
	}, nil
}

// median returns the median of a sorted slice.
func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

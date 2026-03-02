package valuation

import (
	"errors"
	"math"
	"testing"
)

type bandTestCase struct {
	name    string
	values  []float64
	want    *BandStats
	wantErr error
}

var bandTests = []bandTestCase{
	{
		name:    "empty slice",
		values:  []float64{},
		wantErr: ErrInsufficientData,
	},
	{
		name:    "nil slice",
		values:  nil,
		wantErr: ErrInsufficientData,
	},
	{
		name:   "single element",
		values: []float64{2.5},
		want:   &BandStats{Min: 2.5, Max: 2.5, Avg: 2.5, Median: 2.5},
	},
	{
		name:   "two elements",
		values: []float64{1.0, 3.0},
		want:   &BandStats{Min: 1.0, Max: 3.0, Avg: 2.0, Median: 2.0},
	},
	{
		name:   "odd count",
		values: []float64{1.0, 3.0, 5.0},
		want:   &BandStats{Min: 1.0, Max: 5.0, Avg: 3.0, Median: 3.0},
	},
	{
		name:   "even count",
		values: []float64{1.0, 2.0, 3.0, 4.0},
		want:   &BandStats{Min: 1.0, Max: 4.0, Avg: 2.5, Median: 2.5},
	},
	{
		name:   "unsorted input",
		values: []float64{5.0, 1.0, 3.0, 2.0, 4.0},
		want:   &BandStats{Min: 1.0, Max: 5.0, Avg: 3.0, Median: 3.0},
	},
	{
		name:   "PBV band realistic values",
		values: []float64{2.1, 2.5, 3.0, 2.8, 2.3},
		want:   &BandStats{Min: 2.1, Max: 3.0, Avg: 2.54, Median: 2.5},
	},
}

func TestComputeBand(t *testing.T) {
	for _, tc := range bandTests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ComputeBand(tc.values)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ComputeBand() error = %v, want %v", err, tc.wantErr)
				}
				if got != nil {
					t.Fatalf("ComputeBand() = %v, want nil on error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ComputeBand() unexpected error: %v", err)
			}
			assertBandStats(t, got, tc.want)
		})
	}
}

func TestComputeBandNoMutation(t *testing.T) {
	original := []float64{5.0, 1.0, 3.0, 2.0, 4.0}
	snapshot := make([]float64, len(original))
	copy(snapshot, original)

	_, err := ComputeBand(original)
	if err != nil {
		t.Fatalf("ComputeBand() unexpected error: %v", err)
	}

	for i := range original {
		if original[i] != snapshot[i] {
			t.Fatalf("ComputeBand mutated input: index %d changed from %v to %v", i, snapshot[i], original[i])
		}
	}
}

func assertBandStats(t *testing.T, got, want *BandStats) {
	t.Helper()
	const epsilon = 0.01
	if math.Abs(got.Min-want.Min) > epsilon {
		t.Errorf("Min = %v, want %v", got.Min, want.Min)
	}
	if math.Abs(got.Max-want.Max) > epsilon {
		t.Errorf("Max = %v, want %v", got.Max, want.Max)
	}
	if math.Abs(got.Avg-want.Avg) > epsilon {
		t.Errorf("Avg = %v, want %v", got.Avg, want.Avg)
	}
	if math.Abs(got.Median-want.Median) > epsilon {
		t.Errorf("Median = %v, want %v", got.Median, want.Median)
	}
}

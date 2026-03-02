package valuation

import (
	"errors"
	"math"
	"testing"
)

type grahamTestCase struct {
	name    string
	eps     float64
	bvps    float64
	want    float64
	wantErr error
}

var grahamTests = []grahamTestCase{
	{
		name: "BBCA typical values",
		eps:  284.0,
		bvps: 1696.0,
		want: math.Sqrt(22.5 * 284.0 * 1696.0),
	},
	{
		name: "BBRI typical values",
		eps:  302.0,
		bvps: 2141.0,
		want: math.Sqrt(22.5 * 302.0 * 2141.0),
	},
	{
		name: "TLKM typical values",
		eps:  233.0,
		bvps: 1530.0,
		want: math.Sqrt(22.5 * 233.0 * 1530.0),
	},
	{
		name: "small EPS and BVPS",
		eps:  1.0,
		bvps: 1.0,
		want: math.Sqrt(22.5),
	},
	{
		name:    "negative EPS",
		eps:     -10.0,
		bvps:    100.0,
		wantErr: ErrNegativeEPS,
	},
	{
		name:    "zero EPS",
		eps:     0.0,
		bvps:    100.0,
		wantErr: ErrNegativeEPS,
	},
	{
		name:    "negative BVPS",
		eps:     50.0,
		bvps:    -200.0,
		wantErr: ErrNegativeBVPS,
	},
	{
		name:    "zero BVPS",
		eps:     50.0,
		bvps:    0.0,
		wantErr: ErrNegativeBVPS,
	},
	{
		name:    "both negative",
		eps:     -10.0,
		bvps:    -200.0,
		wantErr: ErrNegativeEPS,
	},
}

func TestGrahamNumber(t *testing.T) {
	for _, tc := range grahamTests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GrahamNumber(tc.eps, tc.bvps)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("GrahamNumber(%v, %v) error = %v, want %v", tc.eps, tc.bvps, err, tc.wantErr)
				}
				if got != 0 {
					t.Fatalf("GrahamNumber(%v, %v) = %v, want 0 on error", tc.eps, tc.bvps, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("GrahamNumber(%v, %v) unexpected error: %v", tc.eps, tc.bvps, err)
			}
			if math.Abs(got-tc.want) > 0.01 {
				t.Fatalf("GrahamNumber(%v, %v) = %v, want %v", tc.eps, tc.bvps, got, tc.want)
			}
		})
	}
}

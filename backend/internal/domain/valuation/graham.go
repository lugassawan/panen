package valuation

import "math"

// GrahamNumber computes the intrinsic value of a stock using
// Benjamin Graham's formula: √(22.5 × EPS × BVPS).
// Returns ErrNegativeEPS if eps <= 0, ErrNegativeBVPS if bvps <= 0.
func GrahamNumber(eps, bvps float64) (float64, error) {
	if eps <= 0 {
		return 0, ErrNegativeEPS
	}
	if bvps <= 0 {
		return 0, ErrNegativeBVPS
	}
	return math.Sqrt(22.5 * eps * bvps), nil
}

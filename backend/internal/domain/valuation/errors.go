package valuation

import "errors"

var (
	ErrNegativeEPS      = errors.New("EPS must be positive for Graham Number")
	ErrNegativeBVPS     = errors.New("BVPS must be positive for Graham Number")
	ErrInvalidRisk      = errors.New("unknown risk profile")
	ErrInsufficientData = errors.New("insufficient data for valuation")
)

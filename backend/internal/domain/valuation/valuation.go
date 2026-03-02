package valuation

import "fmt"

// Evaluate computes valuation metrics for a stock based on its financial data
// and the investor's risk profile. It returns entry/exit prices and a verdict.
func Evaluate(input ValuationInput) (*ValuationResult, error) {
	margin, err := marginForRisk(input.RiskProfile)
	if err != nil {
		return nil, err
	}

	graham, grahamErr := GrahamNumber(input.EPS, input.BVPS)

	pbvBand, pbvErr := ComputeBand(input.HistPBV)
	perBand, perErr := ComputeBand(input.HistPER)

	intrinsic, err := intrinsicValue(input, graham, grahamErr, pbvBand, pbvErr, perBand, perErr)
	if err != nil {
		return nil, err
	}

	entry := intrinsic * (1 - margin/100)
	exit := exitTarget(input, graham, grahamErr, pbvBand, pbvErr, perBand, perErr)

	verdict := VerdictFair
	if input.Price <= entry {
		verdict = VerdictUndervalued
	} else if input.Price >= exit {
		verdict = VerdictOvervalued
	}

	result := &ValuationResult{
		Ticker:         input.Ticker,
		GrahamNumber:   graham,
		MarginOfSafety: margin,
		EntryPrice:     entry,
		ExitTarget:     exit,
		Verdict:        verdict,
	}
	if pbvErr == nil {
		result.PBVBand = pbvBand
	}
	if perErr == nil {
		result.PERBand = perBand
	}

	return result, nil
}

func marginForRisk(rp RiskProfile) (float64, error) {
	switch rp {
	case RiskConservative:
		return 50.0, nil
	case RiskModerate:
		return 25.0, nil
	case RiskAggressive:
		return 10.0, nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrInvalidRisk, rp)
	}
}

func intrinsicValue(
	input ValuationInput,
	graham float64, grahamErr error,
	pbvBand *BandStats, pbvErr error,
	perBand *BandStats, perErr error,
) (float64, error) {
	switch input.RiskProfile {
	case RiskConservative:
		return intrinsicConservative(graham, grahamErr)
	case RiskModerate:
		return intrinsicModerate(input, graham, grahamErr, pbvBand, pbvErr)
	case RiskAggressive:
		return intrinsicAggressive(input, pbvBand, pbvErr, perBand, perErr)
	default:
		return 0, fmt.Errorf("%w: %s", ErrInvalidRisk, input.RiskProfile)
	}
}

func intrinsicConservative(graham float64, grahamErr error) (float64, error) {
	if grahamErr != nil {
		return 0, fmt.Errorf("%w: %w", ErrInsufficientData, grahamErr)
	}
	return graham, nil
}

func intrinsicModerate(
	input ValuationInput, graham float64, grahamErr error,
	pbvBand *BandStats, pbvErr error,
) (float64, error) {
	hasGraham := grahamErr == nil
	hasPBV := pbvErr == nil

	if !hasGraham && !hasPBV {
		return 0, ErrInsufficientData
	}
	if !hasGraham {
		return pbvBand.Avg * input.BVPS, nil
	}
	if !hasPBV {
		return graham, nil
	}
	return (graham + pbvBand.Avg*input.BVPS) / 2, nil
}

func intrinsicAggressive(
	input ValuationInput,
	pbvBand *BandStats, pbvErr error,
	perBand *BandStats, perErr error,
) (float64, error) {
	hasPBV := pbvErr == nil
	hasPER := perErr == nil

	if !hasPBV && !hasPER {
		return 0, ErrInsufficientData
	}
	if !hasPBV {
		return perBand.Avg * input.EPS, nil
	}
	if !hasPER {
		return pbvBand.Avg * input.BVPS, nil
	}
	return (pbvBand.Avg*input.BVPS + perBand.Avg*input.EPS) / 2, nil
}

func exitTarget(
	input ValuationInput,
	graham float64, grahamErr error,
	pbvBand *BandStats, pbvErr error,
	perBand *BandStats, perErr error,
) float64 {
	var prices []float64

	if pbvErr == nil {
		prices = append(prices, pbvBand.Max*input.BVPS)
	}
	if perErr == nil {
		prices = append(prices, perBand.Max*input.EPS)
	}

	if len(prices) == 0 {
		if grahamErr == nil {
			return graham * 1.2
		}
		return 0
	}

	var sum float64
	for _, p := range prices {
		sum += p
	}
	return sum / float64(len(prices))
}

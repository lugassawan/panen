package crashplaybook

// EvaluateDiagnostic runs the 4-check falling knife diagnostic.
//   - marketCrashed: broad market is in crash/correction (auto-detected)
//   - companyBadNews: company-specific bad news (manual, nil = unknown)
//   - fundamentalsOK: fundamentals still healthy (manual, nil = unknown)
//   - belowEntry: price is below valuation entry target (auto-detected)
//
// Returns OPPORTUNITY if it's a broad-market-driven dip with healthy fundamentals,
// FALLING_KNIFE if company-specific issues exist, INCONCLUSIVE otherwise.
func EvaluateDiagnostic(marketCrashed bool, companyBadNews, fundamentalsOK *bool, belowEntry bool) DiagnosticSignal {
	if companyBadNews != nil && *companyBadNews {
		return SignalFallingKnife
	}

	if fundamentalsOK != nil && !*fundamentalsOK {
		return SignalFallingKnife
	}

	if companyBadNews == nil || fundamentalsOK == nil {
		return SignalInconclusive
	}

	if marketCrashed && belowEntry {
		return SignalOpportunity
	}

	if belowEntry && !marketCrashed {
		return SignalInconclusive
	}

	return SignalInconclusive
}

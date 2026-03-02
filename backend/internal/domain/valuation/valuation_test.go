package valuation

import (
	"errors"
	"math"
	"testing"
)

type evaluateTestCase struct {
	name           string
	input          ValuationInput
	wantErr        error
	wantGraham     float64
	wantMargin     float64
	wantEntry      float64
	wantExit       float64
	wantVerdict    Verdict
	wantPBVBandNil bool
	wantPERBandNil bool
}

var evaluateTests = []evaluateTestCase{
	{
		// BBCA Conservative: Graham only
		// Graham = √(22.5 × 284 × 1696) ≈ 3293.58
		// Margin = 50%, Entry = 3293.58 × 0.50 = 1646.79
		// Exit = Graham × 1.2 = 3952.30 (no bands used for Conservative)
		// Price 9575 >> Exit → OVERVALUED
		name: "BBCA conservative Graham only",
		input: ValuationInput{
			Ticker:      "BBCA",
			Price:       9575,
			EPS:         284,
			BVPS:        1696,
			PBV:         5.65,
			PER:         33.7,
			RiskProfile: RiskConservative,
		},
		wantGraham:     math.Sqrt(22.5 * 284 * 1696),
		wantMargin:     50.0,
		wantEntry:      math.Sqrt(22.5*284*1696) * 0.50,
		wantExit:       math.Sqrt(22.5*284*1696) * 1.2,
		wantVerdict:    VerdictOvervalued,
		wantPBVBandNil: true,
		wantPERBandNil: true,
	},
	{
		// BBRI Moderate: Graham + PBV band
		// Graham = √(22.5 × 302 × 2141) ≈ 3811.90
		// PBV band avg = (1.8+2.0+2.2+2.5+2.3)/5 = 2.16
		// PBV avg price = 2.16 × 2141 = 4624.56
		// Intrinsic = (3811.90 + 4624.56) / 2 = 4218.23
		// Margin = 25%, Entry = 4218.23 × 0.75 = 3163.67
		// PBV max = 2.5, upper PBV = 2.5 × 2141 = 5352.50
		// Exit = 5352.50 (only PBV band)
		// Price 4400 between entry and exit → FAIR
		name: "BBRI moderate Graham plus PBV band",
		input: ValuationInput{
			Ticker:      "BBRI",
			Price:       4400,
			EPS:         302,
			BVPS:        2141,
			PBV:         2.06,
			PER:         14.6,
			RiskProfile: RiskModerate,
			HistPBV:     []float64{1.8, 2.0, 2.2, 2.5, 2.3},
		},
		wantGraham:     math.Sqrt(22.5 * 302 * 2141),
		wantMargin:     25.0,
		wantEntry:      (math.Sqrt(22.5*302*2141) + 2.16*2141) / 2 * 0.75,
		wantExit:       2.5 * 2141,
		wantVerdict:    VerdictFair,
		wantPBVBandNil: false,
		wantPERBandNil: true,
	},
	{
		// TLKM Aggressive: PBV + PER bands
		// PBV band avg = (2.5+2.8+3.0+3.2+2.9)/5 = 2.88
		// PER band avg = (14+15+16+18+15)/5 = 15.6
		// PBV avg price = 2.88 × 1530 = 4406.40
		// PER avg price = 15.6 × 233 = 3634.80
		// Intrinsic = (4406.40 + 3634.80) / 2 = 4020.60
		// Margin = 10%, Entry = 4020.60 × 0.90 = 3618.54
		// PBV max = 3.2, upper PBV = 3.2 × 1530 = 4896.00
		// PER max = 18, upper PER = 18 × 233 = 4194.00
		// Exit = (4896 + 4194) / 2 = 4545.00
		// Price 3800 between entry and exit → FAIR
		name: "TLKM aggressive PBV plus PER bands",
		input: ValuationInput{
			Ticker:      "TLKM",
			Price:       3800,
			EPS:         233,
			BVPS:        1530,
			PBV:         2.48,
			PER:         16.3,
			RiskProfile: RiskAggressive,
			HistPBV:     []float64{2.5, 2.8, 3.0, 3.2, 2.9},
			HistPER:     []float64{14, 15, 16, 18, 15},
		},
		wantGraham:     math.Sqrt(22.5 * 233 * 1530),
		wantMargin:     10.0,
		wantEntry:      (2.88*1530 + 15.6*233) / 2 * 0.90,
		wantExit:       (3.2*1530 + 18*233) / 2,
		wantVerdict:    VerdictFair,
		wantPBVBandNil: false,
		wantPERBandNil: false,
	},
	{
		name: "invalid risk profile",
		input: ValuationInput{
			Ticker:      "TEST",
			Price:       1000,
			EPS:         100,
			BVPS:        500,
			RiskProfile: "INVALID",
		},
		wantErr: ErrInvalidRisk,
	},
	{
		name: "conservative negative EPS insufficient data",
		input: ValuationInput{
			Ticker:      "LOSS",
			Price:       500,
			EPS:         -10,
			BVPS:        200,
			RiskProfile: RiskConservative,
		},
		wantErr: ErrInsufficientData,
	},
	{
		name: "aggressive no bands insufficient data",
		input: ValuationInput{
			Ticker:      "NODATA",
			Price:       1000,
			EPS:         50,
			BVPS:        300,
			RiskProfile: RiskAggressive,
		},
		wantErr: ErrInsufficientData,
	},
	{
		name: "moderate Graham fails falls back to PBV",
		input: ValuationInput{
			Ticker:      "NEGEPS",
			Price:       800,
			EPS:         -5,
			BVPS:        400,
			RiskProfile: RiskModerate,
			HistPBV:     []float64{1.5, 2.0, 2.5},
		},
		// Graham fails, intrinsic = PBV avg × BVPS = 2.0 × 400 = 800
		// Entry = 800 × 0.75 = 600
		// Exit = PBV max × BVPS = 2.5 × 400 = 1000
		// Price 800 between entry and exit → FAIR
		wantGraham:     0,
		wantMargin:     25.0,
		wantEntry:      2.0 * 400 * 0.75,
		wantExit:       2.5 * 400,
		wantVerdict:    VerdictFair,
		wantPBVBandNil: false,
		wantPERBandNil: true,
	},
	{
		name: "moderate no Graham no PBV insufficient data",
		input: ValuationInput{
			Ticker:      "NONE",
			Price:       500,
			EPS:         -10,
			BVPS:        -100,
			RiskProfile: RiskModerate,
		},
		wantErr: ErrInsufficientData,
	},
	{
		// Undervalued verdict: price well below entry
		name: "undervalued verdict",
		input: ValuationInput{
			Ticker:      "CHEAP",
			Price:       100,
			EPS:         200,
			BVPS:        1000,
			RiskProfile: RiskConservative,
		},
		// Graham = √(22.5 × 200 × 1000) ≈ 2121.32
		// Entry = 2121.32 × 0.50 = 1060.66
		// Price 100 <= 1060.66 → UNDERVALUED
		wantGraham:     math.Sqrt(22.5 * 200 * 1000),
		wantMargin:     50.0,
		wantEntry:      math.Sqrt(22.5*200*1000) * 0.50,
		wantExit:       math.Sqrt(22.5*200*1000) * 1.2,
		wantVerdict:    VerdictUndervalued,
		wantPBVBandNil: true,
		wantPERBandNil: true,
	},
	{
		// Aggressive with only PBV band (no PER)
		name: "aggressive PBV only",
		input: ValuationInput{
			Ticker:      "PBVONLY",
			Price:       3000,
			EPS:         100,
			BVPS:        1000,
			RiskProfile: RiskAggressive,
			HistPBV:     []float64{2.0, 2.5, 3.0},
		},
		// Intrinsic = PBV avg × BVPS = 2.5 × 1000 = 2500
		// Entry = 2500 × 0.90 = 2250
		// Exit = PBV max × BVPS = 3.0 × 1000 = 3000
		// Price 3000 >= 3000 → OVERVALUED
		wantGraham:     math.Sqrt(22.5 * 100 * 1000),
		wantMargin:     10.0,
		wantEntry:      2.5 * 1000 * 0.90,
		wantExit:       3.0 * 1000,
		wantVerdict:    VerdictOvervalued,
		wantPBVBandNil: false,
		wantPERBandNil: true,
	},
}

func TestEvaluate(t *testing.T) {
	for _, tc := range evaluateTests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Evaluate(tc.input)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("Evaluate() expected error wrapping %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Evaluate() error = %v, want error wrapping %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Evaluate() unexpected error: %v", err)
			}
			assertResult(t, got, tc)
		})
	}
}

func assertResult(t *testing.T, got *ValuationResult, tc evaluateTestCase) {
	t.Helper()
	if got.Ticker != tc.input.Ticker {
		t.Errorf("Ticker = %v, want %v", got.Ticker, tc.input.Ticker)
	}
	assertFloat(t, "GrahamNumber", got.GrahamNumber, tc.wantGraham)
	assertFloat(t, "MarginOfSafety", got.MarginOfSafety, tc.wantMargin)
	assertFloat(t, "EntryPrice", got.EntryPrice, tc.wantEntry)
	assertFloat(t, "ExitTarget", got.ExitTarget, tc.wantExit)
	if got.Verdict != tc.wantVerdict {
		t.Errorf("Verdict = %v, want %v", got.Verdict, tc.wantVerdict)
	}
	assertNilness(t, "PBVBand", got.PBVBand == nil, tc.wantPBVBandNil)
	assertNilness(t, "PERBand", got.PERBand == nil, tc.wantPERBandNil)
}

func assertFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func assertNilness(t *testing.T, name string, gotNil, wantNil bool) {
	t.Helper()
	if wantNil && !gotNil {
		t.Errorf("%s should be nil", name)
	}
	if !wantNil && gotNil {
		t.Errorf("%s should not be nil", name)
	}
}

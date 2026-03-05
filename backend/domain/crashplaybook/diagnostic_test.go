package crashplaybook

import "testing"

func TestEvaluateDiagnostic(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }

	tests := []struct {
		name           string
		marketCrashed  bool
		companyBadNews *bool
		fundamentalsOK *bool
		belowEntry     bool
		want           DiagnosticSignal
	}{
		{
			name:           "opportunity - market crash, no bad news, good fundamentals, below entry",
			marketCrashed:  true,
			companyBadNews: boolPtr(false),
			fundamentalsOK: boolPtr(true),
			belowEntry:     true,
			want:           SignalOpportunity,
		},
		{
			name:           "falling knife - company bad news",
			marketCrashed:  true,
			companyBadNews: boolPtr(true),
			fundamentalsOK: boolPtr(true),
			belowEntry:     true,
			want:           SignalFallingKnife,
		},
		{
			name:           "falling knife - bad fundamentals",
			marketCrashed:  true,
			companyBadNews: boolPtr(false),
			fundamentalsOK: boolPtr(false),
			belowEntry:     true,
			want:           SignalFallingKnife,
		},
		{
			name:           "inconclusive - unknown company news",
			marketCrashed:  true,
			companyBadNews: nil,
			fundamentalsOK: boolPtr(true),
			belowEntry:     true,
			want:           SignalInconclusive,
		},
		{
			name:           "inconclusive - unknown fundamentals",
			marketCrashed:  true,
			companyBadNews: boolPtr(false),
			fundamentalsOK: nil,
			belowEntry:     true,
			want:           SignalInconclusive,
		},
		{
			name:           "inconclusive - both unknown",
			marketCrashed:  true,
			companyBadNews: nil,
			fundamentalsOK: nil,
			belowEntry:     true,
			want:           SignalInconclusive,
		},
		{
			name:           "inconclusive - no market crash, below entry",
			marketCrashed:  false,
			companyBadNews: boolPtr(false),
			fundamentalsOK: boolPtr(true),
			belowEntry:     true,
			want:           SignalInconclusive,
		},
		{
			name:           "inconclusive - market crash but above entry",
			marketCrashed:  true,
			companyBadNews: boolPtr(false),
			fundamentalsOK: boolPtr(true),
			belowEntry:     false,
			want:           SignalInconclusive,
		},
		{
			name:           "falling knife takes priority over opportunity",
			marketCrashed:  true,
			companyBadNews: boolPtr(true),
			fundamentalsOK: boolPtr(false),
			belowEntry:     true,
			want:           SignalFallingKnife,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateDiagnostic(tt.marketCrashed, tt.companyBadNews, tt.fundamentalsOK, tt.belowEntry)
			if got != tt.want {
				t.Errorf("EvaluateDiagnostic() = %v, want %v", got, tt.want)
			}
		})
	}
}

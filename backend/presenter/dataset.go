package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/domain/dataset"
)

// DatasetHandler handles financial dataset requests from the frontend.
type DatasetHandler struct {
	ctx  context.Context
	repo dataset.Repository
}

// Bind wires the handler to its dependencies.
func (h *DatasetHandler) Bind(ctx context.Context, repo dataset.Repository) {
	h.ctx = ctx
	h.repo = repo
}

// GetFinancials returns financial time series for a single ticker.
func (h *DatasetHandler) GetFinancials(ticker string) (*FinancialsResponse, error) {
	tf, err := h.repo.FetchFinancials(h.ctx, ticker)
	if err != nil {
		return nil, err
	}
	if tf == nil {
		return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
	}
	return &FinancialsResponse{
		Revenue:         convertTimeSeries(tf.Revenue),
		NetIncome:       convertTimeSeries(tf.NetIncome),
		TotalAssets:     convertTimeSeries(tf.TotalAssets),
		TotalEquity:     convertTimeSeries(tf.TotalEquity),
		TotalDebt:       convertTimeSeries(tf.TotalDebt),
		OperatingIncome: convertTimeSeries(tf.OperatingIncome),
	}, nil
}

// GetSectorMetrics returns sector-specific metrics for a ticker.
func (h *DatasetHandler) GetSectorMetrics(ticker string) (*SectorMetricsResponse, error) {
	metrics, err := h.repo.FetchSectorMetrics(h.ctx, ticker)
	if err != nil {
		return nil, err
	}
	if metrics == nil {
		return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
	}
	converted := make(map[string][]SectorMetricEntryDTO, len(metrics))
	for name, entries := range metrics {
		dtos := make([]SectorMetricEntryDTO, len(entries))
		for i, e := range entries {
			dtos[i] = SectorMetricEntryDTO{
				Year:    e.Year,
				Quarter: e.Quarter,
				Value:   e.Value,
				Source:  e.Source,
			}
		}
		converted[name] = dtos
	}
	return &SectorMetricsResponse{Metrics: converted}, nil
}

// GetSectorDefinitions returns metric definitions grouped by sector.
func (h *DatasetHandler) GetSectorDefinitions() (*SectorDefinitionsResponse, error) {
	defs, err := h.repo.FetchDefinitions(h.ctx)
	if err != nil {
		return nil, err
	}
	if defs == nil {
		return nil, nil //nolint:nilnil // nil signals "not configured" to the frontend
	}
	return &SectorDefinitionsResponse{Sectors: defs}, nil
}

// FinancialsResponse is the frontend-facing response for ticker financials.
type FinancialsResponse struct {
	Revenue         []TimeSeriesEntryDTO `json:"revenue"`
	NetIncome       []TimeSeriesEntryDTO `json:"netIncome"`
	TotalAssets     []TimeSeriesEntryDTO `json:"totalAssets"`
	TotalEquity     []TimeSeriesEntryDTO `json:"totalEquity"`
	TotalDebt       []TimeSeriesEntryDTO `json:"totalDebt"`
	OperatingIncome []TimeSeriesEntryDTO `json:"operatingIncome"`
}

// TimeSeriesEntryDTO is a single data point in a financial time series.
type TimeSeriesEntryDTO struct {
	Year    int     `json:"year"`
	Quarter int     `json:"quarter"`
	Value   float64 `json:"value"`
}

// SectorMetricsResponse is the frontend-facing response for sector metrics.
type SectorMetricsResponse struct {
	Metrics map[string][]SectorMetricEntryDTO `json:"metrics"`
}

// SectorMetricEntryDTO is a single sector-specific metric value.
type SectorMetricEntryDTO struct {
	Year    int     `json:"year"`
	Quarter int     `json:"quarter"`
	Value   float64 `json:"value"`
	Source  string  `json:"source"`
}

// SectorDefinitionsResponse is the frontend-facing response for metric definitions.
type SectorDefinitionsResponse struct {
	Sectors map[string][]string `json:"sectors"`
}

func convertTimeSeries(entries []dataset.TimeSeriesEntry) []TimeSeriesEntryDTO {
	if entries == nil {
		return nil
	}
	dtos := make([]TimeSeriesEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = TimeSeriesEntryDTO{
			Year:    e.Year,
			Quarter: e.Quarter,
			Value:   e.Value,
		}
	}
	return dtos
}

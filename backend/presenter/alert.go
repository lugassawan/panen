package presenter

import (
	"context"

	"github.com/lugassawan/panen/backend/usecase"
)

// AlertHandler handles alert-related requests from the frontend.
type AlertHandler struct {
	ctx    context.Context
	alerts *usecase.AlertService
}

// Bind wires the handler to its dependencies.
func (h *AlertHandler) Bind(ctx context.Context, alerts *usecase.AlertService) {
	h.ctx = ctx
	h.alerts = alerts
}

// GetActiveAlerts returns all active fundamental alerts.
func (h *AlertHandler) GetActiveAlerts() ([]FundamentalAlertResponse, error) {
	if h.alerts == nil {
		return nil, nil
	}
	alerts, err := h.alerts.GetActiveAlerts(h.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]FundamentalAlertResponse, len(alerts))
	for i, a := range alerts {
		result[i] = newFundamentalAlertResponse(a)
	}
	return result, nil
}

// GetAlertsByTicker returns all alerts for a given ticker.
func (h *AlertHandler) GetAlertsByTicker(ticker string) ([]FundamentalAlertResponse, error) {
	if h.alerts == nil {
		return nil, nil
	}
	alerts, err := h.alerts.GetAlertsByTicker(h.ctx, ticker)
	if err != nil {
		return nil, err
	}
	result := make([]FundamentalAlertResponse, len(alerts))
	for i, a := range alerts {
		result[i] = newFundamentalAlertResponse(a)
	}
	return result, nil
}

// AcknowledgeAlert marks an alert as acknowledged.
func (h *AlertHandler) AcknowledgeAlert(id string) error {
	if h.alerts == nil {
		return nil
	}
	return h.alerts.AcknowledgeAlert(h.ctx, id)
}

// GetAlertCount returns the number of active alerts.
func (h *AlertHandler) GetAlertCount() (int, error) {
	if h.alerts == nil {
		return 0, nil
	}
	return h.alerts.GetActiveCount(h.ctx)
}

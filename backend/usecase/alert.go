package usecase

import (
	"context"

	"github.com/lugassawan/panen/backend/domain/alert"
)

// AlertService handles querying and managing fundamental change alerts.
type AlertService struct {
	alerts alert.Repository
}

// NewAlertService creates a new AlertService.
func NewAlertService(alerts alert.Repository) *AlertService {
	return &AlertService{alerts: alerts}
}

// GetActiveAlerts returns all active alerts.
func (s *AlertService) GetActiveAlerts(ctx context.Context) ([]*alert.FundamentalAlert, error) {
	return s.alerts.GetActive(ctx)
}

// GetAlertsByTicker returns all alerts for a given ticker.
func (s *AlertService) GetAlertsByTicker(ctx context.Context, ticker string) ([]*alert.FundamentalAlert, error) {
	return s.alerts.GetByTicker(ctx, ticker)
}

// GetActiveByTicker returns active alerts for a given ticker.
func (s *AlertService) GetActiveByTicker(ctx context.Context, ticker string) ([]*alert.FundamentalAlert, error) {
	return s.alerts.GetActiveByTicker(ctx, ticker)
}

// AcknowledgeAlert marks an alert as acknowledged.
func (s *AlertService) AcknowledgeAlert(ctx context.Context, id string) error {
	return s.alerts.Acknowledge(ctx, id)
}

// GetActiveCount returns the number of active alerts.
func (s *AlertService) GetActiveCount(ctx context.Context) (int, error) {
	return s.alerts.CountActive(ctx)
}

// HasCriticalAlert returns true if the ticker has any active critical alerts.
func (s *AlertService) HasCriticalAlert(ctx context.Context, ticker string) (bool, error) {
	active, err := s.alerts.GetActiveByTicker(ctx, ticker)
	if err != nil {
		return false, err
	}
	for _, a := range active {
		if a.Severity == alert.SeverityCritical {
			return true, nil
		}
	}
	return false, nil
}

package provider

import (
	"context"
	"time"
)

// Status represents the health state of a data provider.
type Status string

const (
	StatusHealthy  Status = "healthy"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
	StatusUnknown  Status = "unknown"
)

// Info holds metadata and health status for a registered data provider.
type Info struct {
	Name      string
	Priority  int
	Status    Status
	LastCheck time.Time
	LastError string
	Enabled   bool
}

// Registry provides read and control operations over registered data providers.
type Registry interface {
	List() []Info
	SetEnabled(name string, enabled bool) bool
	HealthCheckAll(ctx context.Context)
}

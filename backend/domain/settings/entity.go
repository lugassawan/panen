package settings

import "time"

// RefreshSettings holds user preferences for background stock data refresh.
type RefreshSettings struct {
	AutoRefreshEnabled bool
	IntervalMinutes    int
	LastRefreshedAt    time.Time
}

// DefaultRefreshSettings returns sensible defaults for refresh settings:
// auto-refresh enabled with a 12-hour (720 min) interval and no prior refresh.
func DefaultRefreshSettings() *RefreshSettings {
	return &RefreshSettings{
		AutoRefreshEnabled: true,
		IntervalMinutes:    720,
	}
}

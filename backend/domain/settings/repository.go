package settings

import "context"

// Repository defines persistence operations for application settings.
type Repository interface {
	GetRefreshSettings(ctx context.Context) (*RefreshSettings, error)
	SaveRefreshSettings(ctx context.Context, s *RefreshSettings) error
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
}

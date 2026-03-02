package user

import (
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// Profile represents a top-level user of the application.
type Profile struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProfile creates a new Profile with generated ID and timestamps.
func NewProfile(name string) *Profile {
	now := time.Now().UTC()
	return &Profile{
		ID:        shared.NewID(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

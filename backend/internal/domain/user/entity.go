package user

import "time"

// Profile represents a top-level user of the application.
type Profile struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

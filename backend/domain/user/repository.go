package user

import "context"

// Repository defines persistence operations for user profiles.
type Repository interface {
	Create(ctx context.Context, profile *Profile) error
	GetByID(ctx context.Context, id string) (*Profile, error)
	List(ctx context.Context) ([]*Profile, error)
	Update(ctx context.Context, profile *Profile) error
	Delete(ctx context.Context, id string) error
}

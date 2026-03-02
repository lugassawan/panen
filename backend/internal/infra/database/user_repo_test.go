package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/internal/domain/shared"
	"github.com/lugassawan/panen/backend/internal/domain/user"
)

func newUserTestProfile(t *testing.T, repo *UserRepo, ctx context.Context) *user.Profile {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	p := &user.Profile{
		ID:        shared.NewID(),
		Name:      "Alice",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("create test profile: %v", err)
	}
	return p
}

func TestUserRepoCreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	p := newUserTestProfile(t, repo, ctx)

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("Name = %q, want %q", got.Name, "Alice")
	}
	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, p.CreatedAt)
	}
}

func TestUserRepoGetByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() error = %v, want ErrNotFound", err)
	}
}

func TestUserRepoList(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	newUserTestProfile(t, repo, ctx)

	profiles, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(profiles) < 1 {
		t.Error("List() returned empty slice")
	}
}

func TestUserRepoUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	p := newUserTestProfile(t, repo, ctx)
	p.Name = "Robert"
	p.UpdatedAt = p.UpdatedAt.Add(time.Hour)

	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "Robert" {
		t.Errorf("Name = %q, want %q", got.Name, "Robert")
	}
}

func TestUserRepoUpdateNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	p := &user.Profile{ID: "nonexistent", Name: "X", UpdatedAt: now}
	err := repo.Update(ctx, p)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestUserRepoDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	p := newUserTestProfile(t, repo, ctx)

	if err := repo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err := repo.GetByID(ctx, p.ID)
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetByID() after Delete error = %v, want ErrNotFound", err)
	}
}

func TestUserRepoDeleteNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepo(db)
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

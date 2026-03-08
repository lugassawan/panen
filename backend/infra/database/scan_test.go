package database

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

func scanUserProfile(scan func(dest ...any) error) (*user.Profile, error) {
	var p user.Profile
	var createdAt, updatedAt string
	if err := scan(&p.ID, &p.Name, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var err error
	if p.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if p.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func TestQueryRowHappyPath(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewUserRepo(db)
	p := newUserTestProfile(t, repo, ctx)

	got, err := QueryRow(ctx, db, userGetByID, scanUserProfile, p.ID)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("ID = %q, want %q", got.ID, p.ID)
	}
	if got.Name != p.Name {
		t.Errorf("Name = %q, want %q", got.Name, p.Name)
	}
	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, p.CreatedAt)
	}
}

func TestQueryRowNotFound(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := QueryRow(ctx, db, userGetByID, scanUserProfile, "nonexistent")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("QueryRow() error = %v, want ErrNotFound", err)
	}
}

func TestQueryRowScanError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewUserRepo(db)
	newUserTestProfile(t, repo, ctx)

	badScan := func(scan func(dest ...any) error) (string, error) {
		var s string
		// Scan into wrong number of fields to trigger error
		err := scan(&s)
		return s, err
	}

	_, err := QueryRow(ctx, db, userList, badScan)
	if err == nil {
		t.Error("QueryRow() with bad scan expected error, got nil")
	}
}

func TestQueryRowQueryError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := QueryRow(ctx, db, "SELECT * FROM nonexistent_table", scanUserProfile)
	if err == nil {
		t.Error("QueryRow() with bad query expected error, got nil")
	}
}

func TestQueryAllHappyPath(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewUserRepo(db)

	now := time.Now().UTC().Truncate(time.Second)
	for i := range 3 {
		p := &user.Profile{
			ID: shared.NewID(), Name: fmt.Sprintf("User%d", i),
			CreatedAt: now, UpdatedAt: now,
		}
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("create user %d: %v", i, err)
		}
	}

	got, err := QueryAll(ctx, db, userList, scanUserProfile)
	if err != nil {
		t.Fatalf("QueryAll() error = %v", err)
	}
	if len(got) != 3 {
		t.Errorf("QueryAll() returned %d items, want 3", len(got))
	}
}

func TestQueryAllEmptyResult(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	got, err := QueryAll(ctx, db, userList, scanUserProfile)
	if err != nil {
		t.Fatalf("QueryAll() error = %v", err)
	}
	if got != nil {
		t.Errorf("QueryAll() = %v, want nil", got)
	}
}

func TestQueryAllQueryError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := QueryAll(ctx, db, "SELECT * FROM nonexistent_table", scanUserProfile)
	if err == nil {
		t.Error("QueryAll() with bad query expected error, got nil")
	}
}

func TestQueryAllScanError(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewUserRepo(db)
	newUserTestProfile(t, repo, ctx)

	badScan := func(scan func(dest ...any) error) (string, error) {
		var s string
		err := scan(&s)
		return s, err
	}

	_, err := QueryAll(ctx, db, userList, badScan)
	if err == nil {
		t.Error("QueryAll() with bad scan expected error, got nil")
	}
}

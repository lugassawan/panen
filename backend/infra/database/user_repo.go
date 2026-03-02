package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/domain/user"
)

const (
	userInsert  = `INSERT INTO user_profiles (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	userGetByID = `SELECT id, name, created_at, updated_at FROM user_profiles WHERE id = ?`
	userList    = `SELECT id, name, created_at, updated_at FROM user_profiles ORDER BY created_at`
	userUpdate  = `UPDATE user_profiles SET name = ?, updated_at = ? WHERE id = ?`
	userDelete  = `DELETE FROM user_profiles WHERE id = ?`
)

// UserRepo implements user.Repository.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, p *user.Profile) error {
	_, err := r.db.ExecContext(ctx, userInsert,
		p.ID, p.Name, formatTime(p.CreatedAt), formatTime(p.UpdatedAt))
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*user.Profile, error) {
	var p user.Profile
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx, userGetByID, id).Scan(
		&p.ID, &p.Name, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if p.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if p.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *UserRepo) List(ctx context.Context) ([]*user.Profile, error) {
	rows, err := r.db.QueryContext(ctx, userList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*user.Profile
	for rows.Next() {
		var p user.Profile
		var createdAt, updatedAt string
		if err := rows.Scan(&p.ID, &p.Name, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if p.CreatedAt, err = parseTime(createdAt); err != nil {
			return nil, err
		}
		if p.UpdatedAt, err = parseTime(updatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, &p)
	}
	return profiles, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, p *user.Profile) error {
	res, err := r.db.ExecContext(ctx, userUpdate, p.Name, formatTime(p.UpdatedAt), p.ID)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, userDelete, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

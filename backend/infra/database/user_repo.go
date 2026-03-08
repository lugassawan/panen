package database

import (
	"context"
	"database/sql"

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
	return queryRow(ctx, r.db, userGetByID, scanUser, id)
}

func (r *UserRepo) List(ctx context.Context) ([]*user.Profile, error) {
	return queryAll(ctx, r.db, userList, scanUser)
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

func scanUser(scan func(dest ...any) error) (*user.Profile, error) {
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

package postgre

import (
	"context"
	"database/sql"
	"time"

	"github.com/YarKhan02/BlackBird/internal/domain/user"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	u.ID = uuid.New()
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (id, email, password_hash, is_verified, is_banned, failed_attempts, locked_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`, u.ID, u.Email, u.PasswordHash, u.IsVerified, u.IsBanned, u.FailedAttempts, u.LockedUntil).Scan(&u.CreatedAt, &u.UpdatedAt)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, is_verified, is_banned, failed_attempts, locked_until, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, is_verified, is_banned, failed_attempts, locked_until, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET email = $2, is_verified = $3, is_banned = $4, failed_attempts = $5, locked_until = $6, updated_at = NOW()
		WHERE id = $1
	`, u.ID, u.Email, u.IsVerified, u.IsBanned, u.FailedAttempts, u.LockedUntil)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1
	`, id, passwordHash)
	return err
}

func (r *UserRepository) UpdateFailedAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	var lock sql.NullTime
	if lockedUntil != nil {
		lock = sql.NullTime{Time: *lockedUntil, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET failed_attempts = $2, locked_until = $3, updated_at = NOW() WHERE id = $1
	`, id, attempts, lock)
	return err
}

func (r *UserRepository) Ban(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET is_banned = TRUE, updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *UserRepository) Unban(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET is_banned = FALSE, updated_at = NOW() WHERE id = $1
	`, id)
	return err
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (*user.User, error) {
	var u user.User
	var locked sql.NullTime
	if err := scanner.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.IsVerified,
		&u.IsBanned,
		&u.FailedAttempts,
		&locked,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if locked.Valid {
		u.LockedUntil = &locked.Time
	}
	return &u, nil
}

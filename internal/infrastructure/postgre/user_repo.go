package postgre

import (
	"context"
	"database/sql"
	_ "embed"
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

//go:embed sql/user_create.sql
var userCreateSQL string

//go:embed sql/user_find_by_id.sql
var userFindByIDSQL string

//go:embed sql/user_find_by_email.sql
var userFindByEmailSQL string

//go:embed sql/user_update.sql
var userUpdateSQL string

//go:embed sql/user_update_password.sql
var userUpdatePasswordSQL string

//go:embed sql/user_update_failed_attempts.sql
var userUpdateFailedAttemptsSQL string

//go:embed sql/user_ban.sql
var userBanSQL string

//go:embed sql/user_unban.sql
var userUnbanSQL string

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	u.ID = id
	err = r.db.QueryRowContext(ctx, userCreateSQL,
		u.ID,
		u.Email,
		u.PasswordHash,
		u.IsVerified,
		u.IsBanned,
		u.FailedAttempts,
		u.LockedUntil,
	).Scan(&u.CreatedAt, &u.UpdatedAt)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, userFindByIDSQL, id)
	return scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	row := r.db.QueryRowContext(ctx, userFindByEmailSQL, email)
	return scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	_, err := r.db.ExecContext(ctx, userUpdateSQL,
		u.ID,
		u.Email,
		u.IsVerified,
		u.IsBanned,
		u.FailedAttempts,
		u.LockedUntil,
	)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, userUpdatePasswordSQL, id, passwordHash)
	return err
}

func (r *UserRepository) UpdateFailedAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	var lock sql.NullTime
	if lockedUntil != nil {
		lock = sql.NullTime{Time: *lockedUntil, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, userUpdateFailedAttemptsSQL, id, attempts, lock)
	return err
}

func (r *UserRepository) Ban(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, userBanSQL, id)
	return err
}

func (r *UserRepository) Unban(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, userUnbanSQL, id)
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

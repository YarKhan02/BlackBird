package postgre

import (
	"context"
	"database/sql"

	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, rt *token.RefreshToken) error {
	rt.ID = uuid.New()
	var appID uuid.NullUUID
	if rt.AppID != nil {
		appID = uuid.NullUUID{UUID: *rt.AppID, Valid: true}
	}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, app_id, user_agent, ip_address, expires_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, FALSE)
		RETURNING created_at
	`, rt.ID, rt.UserID, rt.TokenHash, appID, rt.UserAgent, rt.IPAddress, rt.ExpiresAt).Scan(&rt.CreatedAt)
	return err
}

func (r *TokenRepository) FindByHash(ctx context.Context, hash string) (*token.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, app_id, user_agent, ip_address, expires_at, revoked, revoked_at, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`, hash)
	return scanRefreshToken(row)
}

func (r *TokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE user_id = $1 AND revoked = FALSE
	`, userID)
	return err
}

func (r *TokenRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*token.RefreshToken, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, token_hash, app_id, user_agent, ip_address, expires_at, revoked, revoked_at, created_at
		FROM refresh_tokens WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*token.RefreshToken
	for rows.Next() {
		rt, err := scanRefreshToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rt)
	}
	return out, rows.Err()
}

func (r *TokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanRefreshToken(scanner interface{ Scan(dest ...any) error }) (*token.RefreshToken, error) {
	var rt token.RefreshToken
	var appID uuid.NullUUID
	var revokedAt sql.NullTime
	if err := scanner.Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&appID,
		&rt.UserAgent,
		&rt.IPAddress,
		&rt.ExpiresAt,
		&rt.Revoked,
		&revokedAt,
		&rt.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if appID.Valid {
		rt.AppID = &appID.UUID
	}
	if revokedAt.Valid {
		rt.RevokedAt = &revokedAt.Time
	}

	return &rt, nil
}

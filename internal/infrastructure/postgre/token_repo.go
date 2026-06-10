package postgre

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

//go:embed sql/token_create.sql
var tokenCreateSQL string

//go:embed sql/token_find_by_hash.sql
var tokenFindByHashSQL string

//go:embed sql/token_revoke.sql
var tokenRevokeSQL string

//go:embed sql/token_revoke_all_for_user.sql
var tokenRevokeAllForUserSQL string

//go:embed sql/token_list_by_user.sql
var tokenListByUserSQL string

//go:embed sql/token_delete_expired.sql
var tokenDeleteExpiredSQL string

func (r *TokenRepository) Create(ctx context.Context, rt *token.RefreshToken) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	rt.ID = id
	appID := sql.NullString{
		String: rt.AppID,
		Valid:  rt.AppID != "",
	}
	err = r.db.QueryRowContext(ctx, tokenCreateSQL,
		rt.ID,
		rt.UserID,
		rt.TokenHash,
		appID,
		rt.UserAgent,
		rt.IPAddress,
		rt.ExpiresAt,
	).Scan(&rt.CreatedAt)
	return err
}

func (r *TokenRepository) FindByHash(ctx context.Context, hash string) (*token.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, tokenFindByHashSQL, hash)
	return scanRefreshToken(row)
}

func (r *TokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, tokenRevokeSQL, id)
	return err
}

func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, tokenRevokeAllForUserSQL, userID)
	return err
}

func (r *TokenRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*token.RefreshToken, error) {
	rows, err := r.db.QueryContext(ctx, tokenListByUserSQL, userID)
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
	res, err := r.db.ExecContext(ctx, tokenDeleteExpiredSQL)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func scanRefreshToken(scanner interface{ Scan(dest ...any) error }) (*token.RefreshToken, error) {
	var rt token.RefreshToken
	var appID sql.NullString
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
		rt.AppID = appID.String
	} else {
		rt.AppID = ""
	}
	if revokedAt.Valid {
		rt.RevokedAt = &revokedAt.Time
	}

	return &rt, nil
}

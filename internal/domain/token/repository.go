package token

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)
	DeleteExpired(ctx context.Context) (int64, error)
}
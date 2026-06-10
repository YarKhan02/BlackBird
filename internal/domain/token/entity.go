package token

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string     // SHA-256 of raw token — never store raw
	AppID     string	 // nil for system-level tokens
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	Revoked   bool
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (rt *RefreshToken) IsExpired() bool {
	return rt.ExpiresAt.Before(time.Now())
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && !rt.IsExpired()
}
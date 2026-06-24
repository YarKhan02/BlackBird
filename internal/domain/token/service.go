package token

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/YarKhan02/BlackBird/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

// Claims is the JWT payload. Used by both service (sign) and middleware (verify).
type Claims struct {
	UserID      string              `json:"sub"`
	ClientID	string				`json:"client_id"`
	Email       string              `json:"email"`
	GlobalRoles []string            `json:"global_roles"`
	AppRoles    map[string][]string `json:"apps,omitempty"`
	jwt.RegisteredClaims
}

type Service struct {
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	repo            Repository
	issuer          string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(
	privateKey *rsa.PrivateKey,
	repo Repository,
	issuer string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		privateKey:      privateKey,
		publicKey:       &privateKey.PublicKey,
		repo:            repo,
		issuer:          issuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// IssueAccessToken signs a short-lived JWT with user identity and roles.
func (s *Service) IssueAccessToken(u *user.User, clientID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      u.ID.String(),
		ClientID:	 clientID,
		Email:       u.Email,
		GlobalRoles: u.GlobalRoles,
		AppRoles:    u.AppRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(), // jti — used for blocklisting
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			Issuer:    s.issuer,
			Subject:   u.ID.String(),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = "key-v1" // key ID for rotation support
	return t.SignedString(s.privateKey)
}

// IssueRefreshToken generates an opaque random token, stores its SHA-256 hash.
func (s *Service) IssueRefreshToken(ctx context.Context, u *user.User, appID string, ip, ua string) (string, error) {
	raw := make([]byte, 64)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	tokenStr := hex.EncodeToString(raw)

	h := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(h[:])

	rt := &RefreshToken{
		UserID:    u.ID,
		TokenHash: tokenHash,
		AppID:     appID,
		IPAddress: ip,
		UserAgent: ua,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.repo.Create(ctx, rt); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}
	return tokenStr, nil
}

// RotateRefreshToken validates the old token, revokes it, and issues a new one.
// If the old token was already revoked (reuse attack), all user tokens are revoked.
func (s *Service) RotateRefreshToken(ctx context.Context, rawToken string) (*RefreshToken, string, error) {
	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	existing, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil || existing == nil {
		return nil, "", ErrInvalidRefreshToken
	}

	// Token reuse detected — revoke entire session family
	if existing.Revoked {
		s.repo.RevokeAllForUser(ctx, existing.UserID) //nolint:errcheck
		return nil, "", ErrInvalidRefreshToken
	}

	if existing.IsExpired() {
		return nil, "", ErrInvalidRefreshToken
	}

	// Revoke the consumed token
	if err := s.repo.Revoke(ctx, existing.ID); err != nil {
		return nil, "", err
	}

	// Issue replacement
	newRaw := make([]byte, 64)
	rand.Read(newRaw) //nolint:errcheck
	newStr := hex.EncodeToString(newRaw)
	newHash := sha256.Sum256([]byte(newStr))

	newRT := &RefreshToken{
		UserID:    existing.UserID,
		TokenHash: hex.EncodeToString(newHash[:]),
		AppID:     existing.AppID,
		IPAddress: existing.IPAddress,
		UserAgent: existing.UserAgent,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}
	if err := s.repo.Create(ctx, newRT); err != nil {
		return nil, "", err
	}

	return newRT, newStr, nil
}

// ValidateAccessToken parses and verifies the JWT, returning its claims.
func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// RevokeByRawToken revokes a refresh token given the raw (unhashed) value.
func (s *Service) RevokeByRawToken(ctx context.Context, rawToken string) error {
	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	rt, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil || rt == nil {
		return nil // already gone — treat as success
	}
	return s.repo.Revoke(ctx, rt.ID)
}

func (s *Service) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllForUser(ctx, userID)
}

func (s *Service) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) PublicKey() *rsa.PublicKey {
	return s.publicKey
}

// JWKS returns the public key set for apps to verify JWTs locally.
func (s *Service) JWKS() map[string]any {
	pub := s.publicKey
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())

	eBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(eBuf, uint32(pub.E))
	eBuf = bytes.TrimLeft(eBuf, "\x00")
	e := base64.RawURLEncoding.EncodeToString(eBuf)

	return map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": "key-v1",
			"n":   n,
			"e":   e,
		}},
	}
}

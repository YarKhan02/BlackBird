package app

import (
	"time"

	"github.com/google/uuid"
)

type App struct {
	ID               uuid.UUID
	ClientID         string
	ClientSecretHash string
	Name             string
	RedirectURIs     []string
	IsActive         bool
	CreatedAt        time.Time
}

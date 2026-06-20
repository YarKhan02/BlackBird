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
	Origin           string
	IsActive         bool
	CreatedAt        time.Time
}

type RegisteredApp struct {
	App          *App
	ClientSecret string
}

type AppFind struct {
	ClientID		string
	IsActive		bool
}

package app

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, app *App) error
	FindByName(ctx context.Context, name string) (*App, error)
	FindByID(ctx context.Context, id uuid.UUID) (*App, error)
	FindByClientID(ctx context.Context, clientID string) (*App, error)
	List(ctx context.Context) ([]*App, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
}

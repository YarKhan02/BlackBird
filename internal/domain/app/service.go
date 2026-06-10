package app

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/YarKhan02/BlackBird/internal/infrastructure/crypto"
	"github.com/google/uuid"
)

var (
	ErrAppNotFound        = errors.New("app not found")
	ErrAppIDTaken         = errors.New("app id already registered")
	ErrInvalidCredentials = errors.New("invalid client credentials")
	ErrAppInactive        = errors.New("app is inactive")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterApp(ctx context.Context, name string, redirectURI string) (*RegisteredApp, error) {
	existing, _ := s.repo.FindByName(ctx, name)
	if existing != nil {
		return nil, ErrAppIDTaken
	}

	clientID, err := crypto.GenerateClientID(name)
	if err != nil {
		return nil, err
	}

	clientSecret, err := crypto.GenerateClientSecret()
	if err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	app := &App{
		ClientID:         clientID,
		ClientSecretHash: string(hash),
		Name:             name,
		IsActive:         true,
	}
	if redirectURI != "" {
		app.RedirectURIs = []string{redirectURI}
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, err
	}

	return &RegisteredApp{App: app, ClientSecret: clientSecret}, nil
}

func (s *Service) List(ctx context.Context) ([]*App, error) {
	return s.repo.List(ctx)
}

func (s *Service) FindByClientID(ctx context.Context, clientID string) (*AppFind, error) {
	app, err := s.repo.FindByClientID(ctx, clientID)
	if err != nil || app == nil {
		return nil, ErrAppNotFound 
	}
	output := &AppFind{
		ClientID:			app.ClientID,
		IsActive:			app.IsActive,
	}
	return output, nil
}

func (s *Service) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.repo.Deactivate(ctx, id)
}

func (s *Service) Authenticate(ctx context.Context, clientID, clientSecret string) (*App, error) {
	app, err := s.repo.FindByClientID(ctx, clientID)
	if err != nil || app == nil {
		return nil, ErrInvalidCredentials
	}

	if !app.IsActive {
		return nil, ErrAppInactive
	}

	if bcrypt.CompareHashAndPassword([]byte(app.ClientSecretHash), []byte(clientSecret)) != nil {
		return nil, ErrInvalidCredentials
	}

	return app, nil
}

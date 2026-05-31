package app

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAppNotFound        = errors.New("app not found")
	ErrClientIDTaken      = errors.New("client id already registered")
	ErrInvalidCredentials = errors.New("invalid client credentials")
	ErrAppInactive        = errors.New("app is inactive")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterApp(ctx context.Context, clientID, clientSecret, name string, redirectURIs []string) (*App, error) {
	existing, _ := s.repo.FindByClientID(ctx, clientID)
	if existing != nil {
		return nil, ErrClientIDTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	app := &App{
		ClientID:         clientID,
		ClientSecretHash: string(hash),
		Name:             name,
		RedirectURIs:     redirectURIs,
		IsActive:         true,
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
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

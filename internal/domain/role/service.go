package role

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrRoleNotFound = errors.New("role not found")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListGlobalRoles(ctx context.Context) ([]GlobalRole, error) {
	return s.repo.ListGlobalRoles(ctx)
}

func (s *Service) AddGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	return s.repo.AddGlobalRole(ctx, userID, roleName)
}

func (s *Service) RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error {
	return s.repo.RemoveGlobalRole(ctx, userID, roleName)
}

func (s *Service) AddAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, roleName string) error {
	return s.repo.AddUserAppRole(ctx, userID, appID, roleName)
}

func (s *Service) RemoveAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, roleName string) error {
	return s.repo.RemoveUserAppRole(ctx, userID, appID, roleName)
}

func (s *Service) GetUserGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.repo.GetUserGlobalRoles(ctx, userID)
}

func (s *Service) GetAllUserAppRoles(ctx context.Context, userID uuid.UUID) (map[string][]string, error) {
	return s.repo.GetAllUserAppRoles(ctx, userID)
}

package role

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListGlobalRoles(ctx context.Context) ([]GlobalRole, error)
	AddGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error
	RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleName string) error
	GetUserGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	AddUserAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, role string) error
	RemoveUserAppRole(ctx context.Context, userID uuid.UUID, appID uuid.UUID, role string) error
	GetAllUserAppRoles(ctx context.Context, userID uuid.UUID) (map[string][]string, error)
}

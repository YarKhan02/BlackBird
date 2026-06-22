package user

import (
	"context"
	"errors"
	"time"
	
	"github.com/YarKhan02/BlackBird/internal/infrastructure/crypto"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrAccountBanned      = errors.New("account banned")
)

const (
	maxFailedAttempts = 5
	lockDuration      = 15 * time.Minute
)

// RoleLoader is a minimal interface so user.Service can load roles
// without depending on the full role.Repository.
type RoleLoader interface {
	AddGlobalRole(ctx context.Context, userID uuid.UUID, roleID string) error
	GetUserGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetAllUserAppRoles(ctx context.Context, userID uuid.UUID) (map[string][]string, error)
}

type Service struct {
	userRepo   Repository
	roleLoader RoleLoader
}

func NewService(userRepo Repository, roleLoader RoleLoader) *Service {
	return &Service{userRepo: userRepo, roleLoader: roleLoader}
}

func (s *Service) Register(ctx context.Context, email, password string, role string) (*User, error) {
	existing, _ := s.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &User{
		Email:        email,
		PasswordHash: hash,
		IsVerified:   true,
		GlobalRoles:  []string{role},
		AppRoles:     make(map[string][]string),
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	if err := s.roleLoader.AddGlobalRole(ctx, u.ID, role); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || u == nil {
		// constant-time even on not found to prevent timing attacks
		crypto.HashPassword(password) //nolint:errcheck
		return nil, ErrInvalidCredentials
	}

	if !u.IsVerified {
		return nil, ErrInvalidCredentials
	}

	if u.IsBanned {
		return nil, ErrAccountBanned
	}

	if u.IsLocked() {
		return nil, ErrAccountLocked
	}

	if !crypto.VerifyPassword(password, u.PasswordHash) {
		attempts := u.FailedAttempts + 1
		var lockedUntil *time.Time
		if attempts >= maxFailedAttempts {
			t := time.Now().Add(lockDuration)
			lockedUntil = &t
		}
		s.userRepo.UpdateFailedAttempts(ctx, u.ID, attempts, lockedUntil) //nolint:errcheck
		return nil, ErrInvalidCredentials
	}

	// globalRoles, err := s.roleLoader.GetUserGlobalRoles(ctx, u.ID)
	// if err != nil {
	// 	return nil, err
	// }

	// Require "super_admin" role for app login to prevent regular users from logging in via this endpoint.
	// hasSuperAdmin := false
	// for _, role := range globalRoles {
	// 	if role == "super_admin" {
	// 		hasSuperAdmin = true
	// 		break
	// 	}
	// }
	// if !hasSuperAdmin {
	// 	return nil, ErrInvalidCredentials
	// }

	// Reset failed attempts on success
	s.userRepo.UpdateFailedAttempts(ctx, u.ID, 0, nil) //nolint:errcheck

	// Load roles for JWT claims
	if err := s.hydrateRoles(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil || u == nil {
		return nil, ErrUserNotFound
	}
	if err := s.hydrateRoles(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || u == nil {
		return ErrUserNotFound
	}

	if !crypto.VerifyPassword(currentPassword, u.PasswordHash) {
		return ErrInvalidCredentials
	}

	newHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, newHash)
}

func (s *Service) BanUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Ban(ctx, userID)
}

func (s *Service) UnbanUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Unban(ctx, userID)
}

func (s *Service) hydrateRoles(ctx context.Context, u *User) error {
	globalRoles, err := s.roleLoader.GetUserGlobalRoles(ctx, u.ID)
	if err != nil {
		return err
	}
	if globalRoles == nil {
		globalRoles = []string{}
	}

	appRoles, err := s.roleLoader.GetAllUserAppRoles(ctx, u.ID)
	if err != nil {
		return err
	}
	if appRoles == nil {
		appRoles = make(map[string][]string)
	}

	u.GlobalRoles = globalRoles
	u.AppRoles = appRoles
	return nil
}

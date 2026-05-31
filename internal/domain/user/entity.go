package user

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID				uuid.UUID
	Email 			string
	PasswordHash 	string
	IsVerified 		bool
	IsBanned 		bool
	FailedAttempts 	int
	LockedUntil 	*time.Time
	GlobalRoles 	[]string
	AppRoles 		map[string][]string
	CreatedAt 		time.Time
	UpdatedAt 		time.Time
}

func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

func (u *User) HasGlobalRole(role string) bool {
    return slices.Contains(u.GlobalRoles, role)
}

func (u *User) HasAppRole(appClientID, role string) bool {
	roles, ok := u.AppRoles[appClientID]
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
package dto

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email")
	}
	if len(r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

type LoginRequest struct {
	Email    string     `json:"email"`
	Password string     `json:"password"`
	AppID    *uuid.UUID `json:"app_id,omitempty"`
}

func (r LoginRequest) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	if r.Email == "" || r.Password == "" {
		return fmt.Errorf("email and password are required")
	}
	return nil
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == "" || r.NewPassword == "" {
		return fmt.Errorf("current and new password are required")
	}
	if len(r.NewPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}
	return nil
}

type AssignGlobalRoleRequest struct {
	Role string `json:"role"`
}

type AssignAppRoleRequest struct {
	AppID uuid.UUID `json:"app_id"`
	Role  string    `json:"role"`
}

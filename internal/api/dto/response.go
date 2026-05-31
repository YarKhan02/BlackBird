package dto

import "github.com/google/uuid"

type ErrorResponse struct {
	Error string `json:"error"`
}

type UserResponse struct {
	ID          uuid.UUID           `json:"id"`
	Email       string              `json:"email"`
	GlobalRoles []string            `json:"global_roles,omitempty"`
	AppRoles    map[string][]string `json:"app_roles,omitempty"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type RolesResponse struct {
	GlobalRoles []string            `json:"global_roles"`
	AppRoles    map[string][]string `json:"app_roles"`
}

type GlobalRoleResponse struct {
	Name string `json:"name"`
}

package dto

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
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

type RegisterAppRequest struct {
	Name         string `json:"name"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	RedirectUI   string `json:"redirect_ui,omitempty"`
	RedirectURIs string `json:"redirect_uris,omitempty"`
}

var (
	ErrInvalidURL          = errors.New("invalid callback url")
	ErrHTTPSRequired       = errors.New("https required")
	ErrFragmentNotAllowed  = errors.New("fragments not allowed")
	ErrUserInfoNotAllowed  = errors.New("userinfo not allowed")
	ErrLocalhostNotAllowed = errors.New("localhost not allowed")
	ErrPrivateIPNotAllowed = errors.New("private ip not allowed")
	ErrHostNotAllowed      = errors.New("host not allowed")
)

func (r RegisterAppRequest) Validate() (string, error) {
	if strings.TrimSpace(r.Name) == "" {
		return "", errors.New("name is required")
	}

	rawURL := strings.TrimSpace(r.RedirectURI)
	if rawURL == "" {
		rawURL = strings.TrimSpace(r.RedirectUI)
	}
	if rawURL == "" {
		rawURL = strings.TrimSpace(r.RedirectURIs)
	}
	// redirect URI is optional for admin-registered apps
	if rawURL == "" {
		return "", nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	if !u.IsAbs() {
		return "", ErrInvalidURL
	}

	// OAuth callbacks should be HTTPS
	// if strings.ToLower(u.Scheme) != "https" {
	// 	return "", ErrHTTPSRequired
	// }

	// Prevent:
	// https://admin:password@example.com/callback
	if u.User != nil {
		return "", ErrUserInfoNotAllowed
	}

	// Fragments never reach the server anyway and can cause confusion
	if u.Fragment != "" {
		return "", ErrFragmentNotAllowed
	}

	host := strings.ToLower(u.Hostname())

	// Block localhost
	// if host == "localhost" {
	// 	return "", ErrLocalhostNotAllowed
	// }

	// Block IP addresses that are private/internal
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() ||
			ip.IsPrivate() ||
			ip.IsLinkLocalMulticast() ||
			ip.IsLinkLocalUnicast() {
			return "", ErrPrivateIPNotAllowed
		}
	}

	// Normalize
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove default port
	if strings.HasSuffix(u.Host, ":443") {
		u.Host = host
	}

	return u.String(), nil
}

type LoginRequest struct {
	Email    string     `json:"email"`
	Password string     `json:"password"`
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

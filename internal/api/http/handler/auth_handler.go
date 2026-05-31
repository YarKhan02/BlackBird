package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/YarKhan02/BlackBird/internal/api/dto"
	"github.com/YarKhan02/BlackBird/internal/domain/token"
	"github.com/YarKhan02/BlackBird/internal/domain/user"
)

type AuthHandler struct {
	userSvc  *user.Service
	tokenSvc *token.Service
}

func NewAuthHandler(userSvc *user.Service, tokenSvc *token.Service) *AuthHandler {
	return &AuthHandler{userSvc: userSvc, tokenSvc: tokenSvc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	u, err := h.userSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case user.ErrEmailTaken:
			writeError(w, http.StatusConflict, "email already registered")
		default:
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	writeJSON(w, http.StatusCreated, dto.UserResponse{ID: u.ID, Email: u.Email})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	u, err := h.userSvc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case user.ErrAccountLocked:
			writeError(w, http.StatusLocked, "account temporarily locked")
		case user.ErrAccountBanned:
			writeError(w, http.StatusForbidden, "account banned")
		default:
			// always same message — don't leak which field is wrong
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		}
		return
	}

	accessToken, err := h.tokenSvc.IssueAccessToken(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue access token")
		return
	}

	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}
	refreshToken, err := h.tokenSvc.IssueRefreshToken(r.Context(), u, req.AppID,
		ip, r.UserAgent())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue refresh token")
		return
	}

	// refresh token in HttpOnly cookie, access token in body
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh",
		MaxAge:   30 * 24 * 60 * 60,
	})

	writeJSON(w, http.StatusOK, dto.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   900,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	rt, newRawToken, err := h.tokenSvc.RotateRefreshToken(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	u, err := h.userSvc.FindByID(r.Context(), rt.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	accessToken, err := h.tokenSvc.IssueAccessToken(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue access token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRawToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh",
		MaxAge:   30 * 24 * 60 * 60,
	})

	writeJSON(w, http.StatusOK, dto.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   900,
	})
}

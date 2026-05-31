package handler

import (
	"encoding/json"
	"net/http"

	"github.com/YarKhan02/BlackBird/internal/api/dto"
	apimiddleware "github.com/YarKhan02/BlackBird/internal/api/http/middleware"
	"github.com/YarKhan02/BlackBird/internal/domain/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	userSvc *user.Service
}

func NewUserHandler(userSvc *user.Service) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := apimiddleware.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	u, err := h.userSvc.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, dto.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		GlobalRoles: u.GlobalRoles,
		AppRoles:    u.AppRoles,
	})
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := apimiddleware.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.userSvc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		switch err {
		case user.ErrInvalidCredentials:
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		case user.ErrUserNotFound:
			writeError(w, http.StatusNotFound, "user not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to change password")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	u, err := h.userSvc.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, dto.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		GlobalRoles: u.GlobalRoles,
		AppRoles:    u.AppRoles,
	})
}

func (h *UserHandler) Ban(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.userSvc.BanUser(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to ban user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *UserHandler) Unban(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.userSvc.UnbanUser(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unban user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

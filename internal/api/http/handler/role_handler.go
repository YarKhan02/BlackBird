package handler

import (
	"encoding/json"
	"net/http"

	"github.com/YarKhan02/BlackBird/internal/api/dto"
	"github.com/YarKhan02/BlackBird/internal/domain/role"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoleHandler struct {
	roleSvc *role.Service
}

func NewRoleHandler(roleSvc *role.Service) *RoleHandler {
	return &RoleHandler{roleSvc: roleSvc}
}

func (h *RoleHandler) ListGlobal(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleSvc.ListGlobalRoles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}

	resp := make([]dto.GlobalRoleResponse, 0, len(roles))
	for _, roleItem := range roles {
		resp = append(resp, dto.GlobalRoleResponse{Name: roleItem.Name})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *RoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	globalRoles, err := h.roleSvc.GetUserGlobalRoles(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load roles")
		return
	}

	appRoles, err := h.roleSvc.GetAllUserAppRoles(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load roles")
		return
	}

	writeJSON(w, http.StatusOK, dto.RolesResponse{
		GlobalRoles: globalRoles,
		AppRoles:    appRoles,
	})
}

func (h *RoleHandler) AddGlobalRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dto.AssignGlobalRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusUnprocessableEntity, "role is required")
		return
	}

	if err := h.roleSvc.AddGlobalRole(r.Context(), userID, req.Role); err != nil {
		if err == role.ErrRoleNotFound {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to assign role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RoleHandler) RemoveGlobalRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	roleName := chi.URLParam(r, "role")
	if roleName == "" {
		writeError(w, http.StatusUnprocessableEntity, "role is required")
		return
	}

	if err := h.roleSvc.RemoveGlobalRole(r.Context(), userID, roleName); err != nil {
		if err == role.ErrRoleNotFound {
			writeError(w, http.StatusNotFound, "role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RoleHandler) AddAppRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dto.AssignAppRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusUnprocessableEntity, "role is required")
		return
	}

	if err := h.roleSvc.AddAppRole(r.Context(), userID, req.AppID, req.Role); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to assign role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RoleHandler) RemoveAppRole(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	appIDStr := chi.URLParam(r, "appID")
	roleName := chi.URLParam(r, "role")
	if appIDStr == "" || roleName == "" {
		writeError(w, http.StatusUnprocessableEntity, "app id and role are required")
		return
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid app id")
		return
	}

	if err := h.roleSvc.RemoveAppRole(r.Context(), userID, appID, roleName); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseUserIDParam(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	return uuid.Parse(idStr)
}

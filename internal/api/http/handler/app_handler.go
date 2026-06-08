package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/YarKhan02/BlackBird/internal/api/dto"
	"github.com/YarKhan02/BlackBird/internal/domain/app"
)

type AppHandler struct {
	appSvc *app.Service
}

func NewAppHandler(appSvc *app.Service) *AppHandler {
	return &AppHandler{appSvc: appSvc}
}

func (h *AppHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	uri, err := req.Validate()

	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	registered, err := h.appSvc.RegisterApp(r.Context(), req.Name, uri)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register app")
		return
	}

	res := dto.RegisterAppResponse{
		ID:           registered.App.ID,
		Name:         registered.App.Name,
		ClientID:     registered.App.ClientID,
		ClientSecret: registered.ClientSecret,
		RedirectURIs: registered.App.RedirectURIs,
		CreatedAt:    registered.App.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeJSON(w, http.StatusCreated, res)
}

func (h *AppHandler) List(w http.ResponseWriter, r *http.Request) {
	apps, err := h.appSvc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list apps")
		return
	}

	res := make([]dto.AppResponse, 0, len(apps))
	for _, a := range apps {
		res = append(res, dto.AppResponse{
			ID:        a.ID,
			Name:      a.Name,
			ClientID:  a.ClientID,
			IsActive:  a.IsActive,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *AppHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid app id")
		return
	}

	if err := h.appSvc.Deactivate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to deactivate app")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

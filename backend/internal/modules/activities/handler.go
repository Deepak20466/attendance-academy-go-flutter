package activities

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type upsertRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req upsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" {
		httpapi.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	id, err := h.svc.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		httpapi.Error(w, http.StatusConflict, "an activity with this name already exists")
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid activity id")
		return
	}
	var req upsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" {
		httpapi.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := h.svc.Update(r.Context(), id, req.Name, req.Description, req.IsActive); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	onlyActive := r.URL.Query().Get("only_active") == "true"
	items, err := h.svc.List(r.Context(), actor, onlyActive)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid activity id")
		return
	}
	if !actor.HasActivity(id) {
		httpapi.Error(w, http.StatusForbidden, "you do not have access to this activity")
		return
	}
	a, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpapi.Error(w, http.StatusNotFound, "activity not found")
		return
	}
	httpapi.JSON(w, http.StatusOK, a)
}

package locations

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
	ActivityID   uuid.UUID `json:"activity_id"`
	Name         string    `json:"name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req upsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.ActivityID == uuid.Nil || req.Name == "" {
		httpapi.Error(w, http.StatusBadRequest, "activity_id and name are required")
		return
	}
	if req.RadiusMeters <= 0 {
		req.RadiusMeters = 100
	}
	id, err := h.svc.Create(r.Context(), UpsertInput{
		ActivityID: req.ActivityID, Name: req.Name,
		Latitude: req.Latitude, Longitude: req.Longitude, RadiusMeters: req.RadiusMeters,
	})
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid location id")
		return
	}
	var req upsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" {
		httpapi.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.RadiusMeters <= 0 {
		req.RadiusMeters = 100
	}
	if err := h.svc.Update(r.Context(), id, UpsertInput{
		Name: req.Name, Latitude: req.Latitude, Longitude: req.Longitude, RadiusMeters: req.RadiusMeters,
	}); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid location id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) ListByActivity(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	activityID, err := uuid.Parse(r.URL.Query().Get("activity_id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "activity_id query param is required")
		return
	}
	items, err := h.svc.ListByActivity(r.Context(), actor, activityID)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

package classes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type BatchHandler struct {
	svc *BatchService
}

func NewBatchHandler(svc *BatchService) *BatchHandler { return &BatchHandler{svc: svc} }

type batchUpsertRequest struct {
	ActivityID     uuid.UUID  `json:"activity_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	DefaultCoachID *uuid.UUID `json:"default_coach_id"`
	LocationID     *uuid.UUID `json:"location_id"`
	DaysOfWeek     []int16    `json:"days_of_week"`
	StartTime      string     `json:"start_time"`
	EndTime        string     `json:"end_time"`
	IsActive       *bool      `json:"is_active"`
}

func (h *BatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req batchUpsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" || req.ActivityID == uuid.Nil ||
		req.StartTime == "" || req.EndTime == "" || len(req.DaysOfWeek) == 0 {
		httpapi.Error(w, http.StatusBadRequest, "name, activity_id, start_time, end_time and at least one day_of_week are required")
		return
	}
	id, err := h.svc.Create(r.Context(), actor, BatchInput{
		ActivityID: req.ActivityID, Name: req.Name, Description: req.Description,
		DefaultCoachID: req.DefaultCoachID, LocationID: req.LocationID,
		DaysOfWeek: req.DaysOfWeek, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err == ErrBatchBadReference {
		httpapi.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *BatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid batch id")
		return
	}
	b, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, b)
}

func (h *BatchHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var activityFilter *uuid.UUID
	if v := r.URL.Query().Get("activity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid activity_id")
			return
		}
		activityFilter = &id
	}
	items, err := h.svc.List(r.Context(), actor, activityFilter)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

func (h *BatchHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid batch id")
		return
	}
	var req batchUpsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" || req.StartTime == "" || req.EndTime == "" || len(req.DaysOfWeek) == 0 {
		httpapi.Error(w, http.StatusBadRequest, "name, start_time, end_time and at least one day_of_week are required")
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	err = h.svc.Update(r.Context(), actor, id, BatchInput{
		Name: req.Name, Description: req.Description, DefaultCoachID: req.DefaultCoachID,
		LocationID: req.LocationID, DaysOfWeek: req.DaysOfWeek, StartTime: req.StartTime, EndTime: req.EndTime,
	}, isActive)
	if err == ErrBatchBadReference {
		httpapi.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

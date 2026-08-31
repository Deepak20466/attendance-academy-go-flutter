package attendance

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type CoachHandler struct {
	svc *CoachService
}

func NewCoachHandler(svc *CoachService) *CoachHandler { return &CoachHandler{svc: svc} }

type checkRequest struct {
	ClassID   uuid.UUID `json:"class_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func (h *CoachHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req checkRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.ClassID == uuid.Nil {
		httpapi.Error(w, http.StatusBadRequest, "class_id, latitude and longitude are required")
		return
	}
	record, err := h.svc.CheckIn(r.Context(), actor, req.ClassID, req.Latitude, req.Longitude)
	if err != nil {
		if err == ErrOutsideGeofence || err == ErrNoGeofenceConfigured {
			httpapi.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, record)
}

func (h *CoachHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req checkRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.ClassID == uuid.Nil {
		httpapi.Error(w, http.StatusBadRequest, "class_id, latitude and longitude are required")
		return
	}
	err := h.svc.CheckOut(r.Context(), actor, req.ClassID, req.Latitude, req.Longitude)
	if err != nil {
		if err == ErrOutsideGeofence || err == ErrNoGeofenceConfigured {
			httpapi.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if err == ErrNoCheckIn {
			httpapi.Error(w, http.StatusBadRequest, "you must check in before checking out")
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "checked_out"})
}

func (h *CoachHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	pg := httpapi.ParsePagination(r)

	var coachFilter *uuid.UUID
	if v := r.URL.Query().Get("coach_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid coach_id")
			return
		}
		coachFilter = &id
	}
	var f CoachAttendanceFilter
	if v := r.URL.Query().Get("date_from"); v != "" {
		d, err := time.Parse("2006-01-02", v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid date_from")
			return
		}
		f.DateFrom = &d
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		d, err := time.Parse("2006-01-02", v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid date_to")
			return
		}
		f.DateTo = &d
	}

	items, total, err := h.svc.List(r.Context(), actor, coachFilter, f, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

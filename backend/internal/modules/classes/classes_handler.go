package classes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type ClassHandler struct {
	svc *ClassService
}

func NewClassHandler(svc *ClassService) *ClassHandler { return &ClassHandler{svc: svc} }

type classCreateRequest struct {
	BatchID    uuid.UUID `json:"batch_id"`
	ActivityID uuid.UUID `json:"activity_id"`
	CoachID    uuid.UUID `json:"coach_id"`
	ClassDate  string    `json:"class_date"` // YYYY-MM-DD
	StartTime  string    `json:"start_time"` // HH:MM
	EndTime    string    `json:"end_time"`
}

func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req classCreateRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	date, err := time.Parse("2006-01-02", req.ClassDate)
	if err != nil || req.BatchID == uuid.Nil || req.ActivityID == uuid.Nil || req.CoachID == uuid.Nil ||
		req.StartTime == "" || req.EndTime == "" {
		httpapi.Error(w, http.StatusBadRequest, "batch_id, activity_id, coach_id, class_date (YYYY-MM-DD), start_time and end_time are required")
		return
	}
	id, err := h.svc.Create(r.Context(), ClassInput{
		BatchID: req.BatchID, ActivityID: req.ActivityID, CoachID: req.CoachID,
		ClassDate: date, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	switch err {
	case nil:
	case ErrClassConflict:
		httpapi.Error(w, http.StatusConflict, err.Error())
		return
	case ErrClassBadReference:
		httpapi.Error(w, http.StatusBadRequest, err.Error())
		return
	default:
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *ClassHandler) Roster(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid class id")
		return
	}
	students, err := h.svc.Roster(r.Context(), actor, id)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, students)
}

func (h *ClassHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid class id")
		return
	}
	c, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, c)
}

func (h *ClassHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	pg := httpapi.ParsePagination(r)

	q := r.URL.Query()
	var activityFilter, coachFilter *uuid.UUID
	var dateFrom, dateTo *time.Time

	if v := q.Get("activity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid activity_id")
			return
		}
		activityFilter = &id
	}
	if v := q.Get("coach_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid coach_id")
			return
		}
		coachFilter = &id
	}
	if v := q.Get("date_from"); v != "" {
		d, err := time.Parse("2006-01-02", v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid date_from")
			return
		}
		dateFrom = &d
	}
	if v := q.Get("date_to"); v != "" {
		d, err := time.Parse("2006-01-02", v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid date_to")
			return
		}
		dateTo = &d
	}

	items, total, err := h.svc.List(r.Context(), actor, activityFilter, coachFilter, dateFrom, dateTo, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

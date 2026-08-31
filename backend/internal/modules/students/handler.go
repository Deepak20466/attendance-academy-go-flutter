package students

import (
	"net/http"
	"time"

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
	BatchID       uuid.UUID  `json:"batch_id"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	GuardianName  string     `json:"guardian_name"`
	GuardianPhone string     `json:"guardian_phone"`
	Email         string     `json:"email"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	IsActive      *bool      `json:"is_active"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req upsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" || req.BatchID == uuid.Nil {
		httpapi.Error(w, http.StatusBadRequest, "name and batch_id are required")
		return
	}
	id, err := h.svc.Create(r.Context(), actor, CreateInput{
		BatchID: req.BatchID, Name: req.Name, Phone: req.Phone,
		GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
		Email: req.Email, DateOfBirth: req.DateOfBirth,
	})
	if err == ErrBatchNotFound {
		httpapi.Error(w, http.StatusBadRequest, "batch not found")
		return
	}
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid student id")
		return
	}
	student, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, student)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	pg := httpapi.ParsePagination(r)

	var activityFilter, batchFilter *uuid.UUID
	if v := r.URL.Query().Get("activity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid activity_id")
			return
		}
		activityFilter = &id
	}
	if v := r.URL.Query().Get("batch_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid batch_id")
			return
		}
		batchFilter = &id
	}

	items, total, err := h.svc.List(r.Context(), actor, activityFilter, batchFilter, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid student id")
		return
	}
	var req upsertRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" || req.BatchID == uuid.Nil {
		httpapi.Error(w, http.StatusBadRequest, "name and batch_id are required")
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	err = h.svc.Update(r.Context(), actor, id, CreateInput{
		BatchID: req.BatchID, Name: req.Name, Phone: req.Phone,
		GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
		Email: req.Email, DateOfBirth: req.DateOfBirth,
	}, isActive)
	if err == ErrBatchNotFound {
		httpapi.Error(w, http.StatusBadRequest, "batch not found")
		return
	}
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

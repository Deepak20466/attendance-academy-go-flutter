package coaches

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type Handler struct {
	svc *ServiceWithUserCreation
}

func NewHandler(svc *ServiceWithUserCreation) *Handler { return &Handler{svc: svc} }

type createCoachRequest struct {
	Name          string      `json:"name"`
	Email         string      `json:"email"`
	Phone         string      `json:"phone"`
	Password      string      `json:"password"`
	EmployeeCode  string      `json:"employee_code"`
	MonthlySalary float64     `json:"monthly_salary"`
	ActivityIDs   []uuid.UUID `json:"activity_ids"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCoachRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" || req.Email == "" ||
		len(req.Password) < 8 || req.EmployeeCode == "" {
		httpapi.Error(w, http.StatusBadRequest, "name, email, employee_code and a password of at least 8 characters are required")
		return
	}
	id, err := h.svc.CreateCoach(r.Context(), CreateCoachInput{
		Name: req.Name, Email: req.Email, Phone: req.Phone, Password: req.Password,
		EmployeeCode: req.EmployeeCode, MonthlySalary: req.MonthlySalary, ActivityIDs: req.ActivityIDs,
	})
	if err != nil {
		httpapi.Error(w, http.StatusConflict, "a coach with this email or employee code already exists")
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pg := httpapi.ParsePagination(r)
	var activityID *uuid.UUID
	if v := r.URL.Query().Get("activity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid activity_id")
			return
		}
		activityID = &id
	}
	items, total, err := h.svc.List(r.Context(), activityID, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid coach id")
		return
	}
	if err := EnsureCanViewCoach(actor, id); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpapi.Error(w, http.StatusNotFound, "coach not found")
		return
	}
	httpapi.JSON(w, http.StatusOK, c)
}

// Me returns the calling coach's own profile — the primary way the mobile
// app fetches "my assigned activities".
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	c, err := h.svc.GetSelf(r.Context(), actor.UserID)
	if err != nil {
		httpapi.Error(w, http.StatusNotFound, "coach profile not found")
		return
	}
	httpapi.JSON(w, http.StatusOK, c)
}

type setActivitiesRequest struct {
	ActivityIDs []uuid.UUID `json:"activity_ids"`
}

func (h *Handler) SetActivities(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid coach id")
		return
	}
	var req setActivitiesRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SetActivities(r.Context(), id, req.ActivityIDs); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) SetActive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid coach id")
		return
	}
	active := r.URL.Query().Get("active") != "false"
	if err := h.svc.SetActive(r.Context(), id, active); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

type updateSalaryRequest struct {
	MonthlySalary float64 `json:"monthly_salary"`
}

func (h *Handler) UpdateSalary(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid coach id")
		return
	}
	var req updateSalaryRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateSalary(r.Context(), id, req.MonthlySalary); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

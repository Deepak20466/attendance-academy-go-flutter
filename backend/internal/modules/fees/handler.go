package fees

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

type createRequest struct {
	StudentID   uuid.UUID `json:"student_id"`
	ActivityID  uuid.UUID `json:"activity_id"`
	Amount      float64   `json:"amount"`
	DueDate     string    `json:"due_date"`
	PeriodMonth int       `json:"period_month"`
	PeriodYear  int       `json:"period_year"`
	Remarks     string    `json:"remarks"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.StudentID == uuid.Nil || req.Amount <= 0 {
		httpapi.Error(w, http.StatusBadRequest, "student_id and a positive amount are required")
		return
	}
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "due_date must be YYYY-MM-DD")
		return
	}
	id, err := h.svc.Create(r.Context(), CreateInput{
		StudentID: req.StudentID, ActivityID: req.ActivityID, Amount: req.Amount,
		DueDate: dueDate, PeriodMonth: req.PeriodMonth, PeriodYear: req.PeriodYear, Remarks: req.Remarks,
	})
	if err != nil {
		httpapi.Error(w, http.StatusConflict, "a fee record already exists for this student and period")
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

type generateRequest struct {
	ActivityID  uuid.UUID `json:"activity_id"`
	Amount      float64   `json:"amount"`
	DueDate     string    `json:"due_date"`
	PeriodMonth int       `json:"period_month"`
	PeriodYear  int       `json:"period_year"`
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.ActivityID == uuid.Nil || req.Amount <= 0 {
		httpapi.Error(w, http.StatusBadRequest, "activity_id and a positive amount are required")
		return
	}
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "due_date must be YYYY-MM-DD")
		return
	}
	count, err := h.svc.GenerateForActivity(r.Context(), req.ActivityID, req.Amount, dueDate, req.PeriodMonth, req.PeriodYear)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{"created": count})
}

type markPaidRequest struct {
	PaidDate      string `json:"paid_date"`
	PaymentMethod string `json:"payment_method"`
}

func (h *Handler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid fee id")
		return
	}
	var req markPaidRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	paidDate := time.Now()
	if req.PaidDate != "" {
		paidDate, err = time.Parse("2006-01-02", req.PaidDate)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "paid_date must be YYYY-MM-DD")
			return
		}
	}
	if err := h.svc.MarkPaid(r.Context(), id, paidDate, req.PaymentMethod); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "paid"})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid fee id")
		return
	}
	fee, err := h.svc.Get(r.Context(), actor, id)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, fee)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	pg := httpapi.ParsePagination(r)
	q := r.URL.Query()

	var activityFilter, studentFilter *uuid.UUID
	if v := q.Get("activity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid activity_id")
			return
		}
		activityFilter = &id
	}
	if v := q.Get("student_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid student_id")
			return
		}
		studentFilter = &id
	}

	items, total, err := h.svc.List(r.Context(), actor, activityFilter, studentFilter, q.Get("status"), pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func (h *Handler) PendingSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.PendingSummary(r.Context())
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, summary)
}

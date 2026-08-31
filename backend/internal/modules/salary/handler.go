package salary

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

type generateRequest struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Month < 1 || req.Month > 12 || req.Year < 2000 {
		httpapi.Error(w, http.StatusBadRequest, "valid month and year are required")
		return
	}
	count, err := h.svc.GenerateForPeriod(r.Context(), req.Month, req.Year)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{"created": count})
}

func (h *Handler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid acknowledgement id")
		return
	}
	if err := h.svc.Acknowledge(r.Context(), actor, id); err != nil {
		if err == ErrAlreadyAcknowledged {
			httpapi.Error(w, http.StatusConflict, err.Error())
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	pg := httpapi.ParsePagination(r)
	items, total, err := h.svc.ListMine(r.Context(), *actor.CoachID, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func (h *Handler) ListForPeriod(w http.ResponseWriter, r *http.Request) {
	pg := httpapi.ParsePagination(r)
	month := atoiDefault(r.URL.Query().Get("month"))
	year := atoiDefault(r.URL.Query().Get("year"))
	status := r.URL.Query().Get("status")
	if month == 0 || year == 0 {
		httpapi.Error(w, http.StatusBadRequest, "month and year query params are required")
		return
	}
	items, total, err := h.svc.ListForPeriod(r.Context(), month, year, status, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func atoiDefault(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

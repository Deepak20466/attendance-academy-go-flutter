package leaves

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

type applyRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Reason    string `json:"reason"`
}

func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req applyRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Reason == "" {
		httpapi.Error(w, http.StatusBadRequest, "start_date, end_date and reason are required")
		return
	}
	start, err1 := time.Parse("2006-01-02", req.StartDate)
	end, err2 := time.Parse("2006-01-02", req.EndDate)
	if err1 != nil || err2 != nil {
		httpapi.Error(w, http.StatusBadRequest, "start_date and end_date must be YYYY-MM-DD")
		return
	}
	id, err := h.svc.Apply(r.Context(), *actor.CoachID, start, end, req.Reason)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) { h.review(w, r, true) }
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request)  { h.review(w, r, false) }

func (h *Handler) review(w http.ResponseWriter, r *http.Request, approve bool) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid leave id")
		return
	}
	if err := h.svc.Review(r.Context(), actor, id, approve); err != nil {
		if err == ErrNotPending {
			httpapi.Error(w, http.StatusConflict, err.Error())
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid leave id")
		return
	}
	if err := h.svc.Cancel(r.Context(), actor, id); err != nil {
		if err == ErrNotPending {
			httpapi.Error(w, http.StatusConflict, err.Error())
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
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

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	pg := httpapi.ParsePagination(r)
	status := r.URL.Query().Get("status")
	items, total, err := h.svc.ListAll(r.Context(), status, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

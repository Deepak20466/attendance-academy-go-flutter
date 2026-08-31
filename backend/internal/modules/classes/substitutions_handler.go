package classes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type SubstitutionHandler struct {
	svc *SubstitutionService
}

func NewSubstitutionHandler(svc *SubstitutionService) *SubstitutionHandler {
	return &SubstitutionHandler{svc: svc}
}

type substitutionCreateRequest struct {
	ClassID           uuid.UUID `json:"class_id"`
	SubstituteCoachID uuid.UUID `json:"substitute_coach_id"`
	Reason            string    `json:"reason"`
}

func (h *SubstitutionHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req substitutionCreateRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.ClassID == uuid.Nil || req.SubstituteCoachID == uuid.Nil {
		httpapi.Error(w, http.StatusBadRequest, "class_id and substitute_coach_id are required")
		return
	}
	id, err := h.svc.Create(r.Context(), actor, req.ClassID, req.SubstituteCoachID, req.Reason)
	if err != nil {
		switch err {
		case ErrSelfSubstitution:
			httpapi.Error(w, http.StatusBadRequest, err.Error())
		case ErrActiveSubExists:
			httpapi.Error(w, http.StatusConflict, err.Error())
		default:
			httpapi.HandleServiceError(w, err)
		}
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *SubstitutionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid substitution id")
		return
	}
	if err := h.svc.Cancel(r.Context(), actor, id); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *SubstitutionHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	items, err := h.svc.ListMine(r.Context(), *actor.CoachID)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

func (h *SubstitutionHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	pg := httpapi.ParsePagination(r)
	items, total, err := h.svc.ListAll(r.Context(), pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

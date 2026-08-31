package audit

import (
	"net/http"

	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
)

type Handler struct {
	svc *Repo
}

func NewHandler(repo *Repo) *Handler { return &Handler{svc: repo} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pg := httpapi.ParsePagination(r)
	q := r.URL.Query()

	f := Filter{EntityType: q.Get("entity_type")}
	if v := q.Get("entity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid entity_id")
			return
		}
		f.EntityID = &id
	}
	if v := q.Get("actor_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, http.StatusBadRequest, "invalid actor_id")
			return
		}
		f.ActorID = &id
	}

	items, total, err := h.svc.List(r.Context(), f, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

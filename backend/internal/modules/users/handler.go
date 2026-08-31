package users

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

type createAdminRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *Handler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req createAdminRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Name == "" || req.Email == "" || len(req.Password) < 8 {
		httpapi.Error(w, http.StatusBadRequest, "name, email and a password of at least 8 characters are required")
		return
	}
	id, err := h.svc.CreateAdmin(r.Context(), req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		httpapi.Error(w, http.StatusConflict, "a user with this email already exists")
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pg := httpapi.ParsePagination(r)
	roleFilter := r.URL.Query().Get("role")
	items, total, err := h.svc.List(r.Context(), roleFilter, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, false)
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, true)
}

func (h *Handler) setActive(w http.ResponseWriter, r *http.Request, active bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	// Enforced here rather than only in the UI: a client-side disabled
	// control is not a substitute for a real check, and an admin locking
	// themselves out (or being tricked into it) has no recovery path short
	// of direct DB access.
	actor := middleware.ActorFromContext(r.Context())
	if !active && id == actor.UserID {
		httpapi.Error(w, http.StatusBadRequest, "you cannot deactivate your own account")
		return
	}
	if err := h.svc.SetActive(r.Context(), id, active); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req changePasswordRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || len(req.NewPassword) < 8 {
		httpapi.Error(w, http.StatusBadRequest, "current_password and a new_password of at least 8 characters are required")
		return
	}
	if err := h.svc.ChangePassword(r.Context(), actor.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		if err == ErrWrongPassword {
			httpapi.Error(w, http.StatusBadRequest, "current password is incorrect")
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "password_updated"})
}

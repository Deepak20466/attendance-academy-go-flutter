package authapi

import (
	"errors"
	"net/http"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.Email == "" || req.Password == "" {
		httpapi.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	tokens, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrAccountDisabled) {
			httpapi.Error(w, http.StatusForbidden, "account is deactivated")
			return
		}
		httpapi.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	httpapi.JSON(w, http.StatusOK, tokens)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		httpapi.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}
	tokens, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpapi.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	httpapi.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.RefreshToken == "" {
		httpapi.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}
	_ = h.svc.Logout(r.Context(), req.RefreshToken)
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	if actor == nil {
		httpapi.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]interface{}{
		"user_id":      actor.UserID,
		"role":         actor.Role,
		"coach_id":     actor.CoachID,
		"activity_ids": actor.ActivityIDList(),
	})
}

type fcmTokenRequest struct {
	FCMToken string `json:"fcm_token"`
}

func (h *Handler) UpdateFCMToken(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req fcmTokenRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.FCMToken == "" {
		httpapi.Error(w, http.StatusBadRequest, "fcm_token is required")
		return
	}
	if err := h.svc.UpdateFCMToken(r.Context(), actor.UserID, req.FCMToken); err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

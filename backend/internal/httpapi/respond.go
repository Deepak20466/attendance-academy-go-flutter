package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"attendance-backend/internal/authz"
)

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("respond: encode error: %v", err)
		}
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

// HandleServiceError maps common service-layer sentinel errors to the
// correct HTTP status without leaking internals to the client.
func HandleServiceError(w http.ResponseWriter, err error) {
	switch {
	case err == authz.ErrForbidden:
		Error(w, http.StatusForbidden, "you do not have access to this resource")
	case err == ErrNotFound:
		Error(w, http.StatusNotFound, "resource not found")
	case err == ErrConflict:
		Error(w, http.StatusConflict, "conflict: resource already exists")
	case err == ErrValidation:
		Error(w, http.StatusBadRequest, "invalid request")
	default:
		log.Printf("internal error: %v", err)
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func DecodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

package attendance

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
	"attendance-backend/internal/middleware"
)

type StudentHandler struct {
	svc *StudentService
}

func NewStudentHandler(svc *StudentService) *StudentHandler { return &StudentHandler{svc: svc} }

type markEntryRequest struct {
	StudentID uuid.UUID `json:"student_id"`
	Status    string    `json:"status"`
	Remarks   string    `json:"remarks"`
}

type markBulkRequest struct {
	ClassID uuid.UUID          `json:"class_id"`
	Entries []markEntryRequest `json:"entries"`
}

func (h *StudentHandler) MarkBulk(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	var req markBulkRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil || req.ClassID == uuid.Nil || len(req.Entries) == 0 {
		httpapi.Error(w, http.StatusBadRequest, "class_id and at least one entry are required")
		return
	}

	entries := make([]MarkEntry, len(req.Entries))
	for i, e := range req.Entries {
		if e.StudentID == uuid.Nil || e.Status == "" {
			httpapi.Error(w, http.StatusBadRequest, "each entry requires student_id and status")
			return
		}
		entries[i] = MarkEntry{StudentID: e.StudentID, Status: e.Status, Remarks: e.Remarks}
	}

	if err := h.svc.MarkBulk(r.Context(), actor, req.ClassID, entries, actor.UserID); err != nil {
		if err == ErrInvalidStatus || err == ErrStudentNotInBatch {
			httpapi.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]string{"status": "attendance_recorded"})
}

func (h *StudentHandler) ListForClass(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	classID, err := uuid.Parse(chi.URLParam(r, "classId"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid class id")
		return
	}
	items, err := h.svc.ListForClass(r.Context(), actor, classID)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

func (h *StudentHandler) History(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	studentID, err := uuid.Parse(chi.URLParam(r, "studentId"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid student id")
		return
	}
	pg := httpapi.ParsePagination(r)
	items, total, err := h.svc.History(r.Context(), actor, studentID, pg.Limit, pg.Offset)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, httpapi.PagedResult{Data: items, Page: pg.Page, PageSize: pg.Limit, TotalCount: total})
}

func (h *StudentHandler) MonthlyPercentage(w http.ResponseWriter, r *http.Request) {
	actor := middleware.ActorFromContext(r.Context())
	studentID, err := uuid.Parse(chi.URLParam(r, "studentId"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "invalid student id")
		return
	}
	batchID, err := uuid.Parse(r.URL.Query().Get("batch_id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "batch_id query param is required")
		return
	}
	year := atoiDefault(r.URL.Query().Get("year"), 0)
	month := atoiDefault(r.URL.Query().Get("month"), 0)
	if year == 0 || month == 0 {
		httpapi.Error(w, http.StatusBadRequest, "year and month query params are required")
		return
	}

	result, err := h.svc.MonthlyPercentage(r.Context(), actor, studentID, batchID, year, month)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, result)
}

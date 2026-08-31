package analytics

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"attendance-backend/internal/httpapi"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// monthYearFromQuery rejects out-of-range values rather than passing them
// to time.Date, which silently normalizes overflow (month=13 becomes
// January of the following year) instead of erroring — that would make an
// admin's typo return a real but wrong period's numbers instead of a
// clear rejection.
func monthYearFromQuery(r *http.Request) (int, int, bool) {
	q := r.URL.Query()
	month, err1 := strconv.Atoi(q.Get("month"))
	year, err2 := strconv.Atoi(q.Get("year"))
	if err1 != nil || err2 != nil {
		now := time.Now()
		return int(now.Month()), now.Year(), q.Get("month") == "" && q.Get("year") == ""
	}
	if month < 1 || month > 12 || year < 2000 || year > 2100 {
		return 0, 0, false
	}
	return month, year, true
}

func (h *Handler) Overall(w http.ResponseWriter, r *http.Request) {
	month, year, ok := monthYearFromQuery(r)
	if !ok {
		httpapi.Error(w, http.StatusBadRequest, "invalid month/year")
		return
	}
	summary, err := h.svc.OverallSummary(r.Context(), month, year)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, summary)
}

func (h *Handler) ActivitySummary(w http.ResponseWriter, r *http.Request) {
	activityID, err := uuid.Parse(r.URL.Query().Get("activity_id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "activity_id query param is required")
		return
	}
	month, year, ok := monthYearFromQuery(r)
	if !ok {
		httpapi.Error(w, http.StatusBadRequest, "invalid month/year")
		return
	}
	summary, err := h.svc.ActivitySummary(r.Context(), activityID, month, year)
	if err != nil {
		// ActivitySummary's first query looks up the activity by id, so in
		// practice this is always "no such activity" — same convention as
		// the activities/coaches Get handlers.
		httpapi.Error(w, http.StatusNotFound, "activity not found")
		return
	}
	httpapi.JSON(w, http.StatusOK, summary)
}

func (h *Handler) PerfectAttendance(w http.ResponseWriter, r *http.Request) {
	activityID, err := uuid.Parse(r.URL.Query().Get("activity_id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "activity_id query param is required")
		return
	}
	month, year, ok := monthYearFromQuery(r)
	if !ok {
		httpapi.Error(w, http.StatusBadRequest, "invalid month/year")
		return
	}
	items, err := h.svc.PerfectAttendance(r.Context(), activityID, month, year)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

func (h *Handler) MonthlyReport(w http.ResponseWriter, r *http.Request) {
	activityID, err := uuid.Parse(r.URL.Query().Get("activity_id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "activity_id query param is required")
		return
	}
	month, year, ok := monthYearFromQuery(r)
	if !ok {
		httpapi.Error(w, http.StatusBadRequest, "invalid month/year")
		return
	}
	items, err := h.svc.MonthlyReport(r.Context(), activityID, month, year)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, items)
}

// MonthlyReportCSV exports the same monthly report as a downloadable CSV
// for the admin dashboard's "export report" action.
func (h *Handler) MonthlyReportCSV(w http.ResponseWriter, r *http.Request) {
	activityID, err := uuid.Parse(r.URL.Query().Get("activity_id"))
	if err != nil {
		httpapi.Error(w, http.StatusBadRequest, "activity_id query param is required")
		return
	}
	month, year, ok := monthYearFromQuery(r)
	if !ok {
		httpapi.Error(w, http.StatusBadRequest, "invalid month/year")
		return
	}
	items, err := h.svc.MonthlyReport(r.Context(), activityID, month, year)
	if err != nil {
		httpapi.HandleServiceError(w, err)
		return
	}

	filename := fmt.Sprintf("attendance_report_%d_%02d.csv", year, month)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Student ID", "Student Name", "Total Classes", "Present Count", "Percent Present"})
	for _, row := range items {
		_ = writer.Write([]string{
			row.StudentID.String(), row.StudentName,
			strconv.FormatInt(row.TotalClasses, 10),
			strconv.FormatInt(row.PresentCount, 10),
			strconv.FormatFloat(row.PercentPresent, 'f', 2, 64),
		})
	}
	writer.Flush()
}

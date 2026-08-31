package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type OverallSummary struct {
	TotalStudents          int64   `json:"total_students"`
	TotalCoaches           int64   `json:"total_coaches"`
	TotalActivities        int64   `json:"total_activities"`
	TotalClassesThisMonth  int64   `json:"total_classes_this_month"`
	ClassesWithAttendance  int64   `json:"classes_with_attendance"`
	AttendancePercent      float64 `json:"attendance_percent"`
	FeesCollectedThisMonth float64 `json:"fees_collected_this_month"`
	FeesPendingThisMonth   float64 `json:"fees_pending_this_month"`
	PendingLeaves          int64   `json:"pending_leaves"`
	CoachCheckInsToday     int64   `json:"coach_check_ins_today"`
	TodayClasses           int64   `json:"today_classes"`
	TodayMissingAttendance int64   `json:"today_missing_attendance"`
}

func (r *Repo) OverallSummary(ctx context.Context, month, year int) (*OverallSummary, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	var s OverallSummary
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE is_active = true`).Scan(&s.TotalStudents); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM coaches WHERE is_active = true`).Scan(&s.TotalCoaches); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM activities WHERE is_active = true`).Scan(&s.TotalActivities); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE attendance_marked)
		FROM classes WHERE class_date >= $1 AND class_date < $2
	`, from, to).Scan(&s.TotalClassesThisMonth, &s.ClassesWithAttendance); err != nil {
		return nil, err
	}
	if s.TotalClassesThisMonth > 0 {
		s.AttendancePercent = float64(s.ClassesWithAttendance) / float64(s.TotalClassesThisMonth) * 100
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount) FILTER (WHERE status = 'paid'), 0),
		       COALESCE(SUM(amount) FILTER (WHERE status IN ('pending','overdue')), 0)
		FROM fees WHERE period_month = $1 AND period_year = $2
	`, month, year).Scan(&s.FeesCollectedThisMonth, &s.FeesPendingThisMonth); err != nil {
		return nil, err
	}

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM leaves WHERE status = 'pending'`).Scan(&s.PendingLeaves); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM coach_attendance WHERE attendance_date = CURRENT_DATE AND check_in_time IS NOT NULL`).Scan(&s.CoachCheckInsToday); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM classes WHERE class_date = CURRENT_DATE`).Scan(&s.TodayClasses); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM classes WHERE class_date = CURRENT_DATE AND attendance_marked = false AND (class_date + end_time) < now()
	`).Scan(&s.TodayMissingAttendance); err != nil {
		return nil, err
	}

	return &s, nil
}

type ActivitySummary struct {
	ActivityID           uuid.UUID `json:"activity_id"`
	ActivityName         string    `json:"activity_name"`
	StudentCount         int64     `json:"student_count"`
	CoachCount           int64     `json:"coach_count"`
	ClassCount           int64     `json:"class_count"`
	PresentCount         int64     `json:"present_count"`
	AbsentCount          int64     `json:"absent_count"`
	AttendancePercent    float64   `json:"attendance_percent"`
	PerfectAttendance    int64     `json:"perfect_attendance_students"`
	FeesCollected        float64   `json:"fees_collected"`
	FeesPending          float64   `json:"fees_pending"`
	CoachAttendanceDays  int64     `json:"coach_attendance_days"`
}

func (r *Repo) ActivitySummary(ctx context.Context, activityID uuid.UUID, month, year int) (*ActivitySummary, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	var s ActivitySummary
	s.ActivityID = activityID

	if err := r.pool.QueryRow(ctx, `SELECT name FROM activities WHERE id = $1`, activityID).Scan(&s.ActivityName); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE activity_id = $1 AND is_active = true`, activityID).Scan(&s.StudentCount); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM coach_activities WHERE activity_id = $1`, activityID).Scan(&s.CoachCount); err != nil {
		return nil, err
	}

	var totalMarks int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status IN ('present','late'))
		FROM student_attendance WHERE activity_id = $1 AND marked_at >= $2 AND marked_at < $3
	`, activityID, from, to).Scan(&totalMarks, &s.PresentCount); err != nil {
		return nil, err
	}
	s.AbsentCount = totalMarks - s.PresentCount
	if totalMarks > 0 {
		s.AttendancePercent = float64(s.PresentCount) / float64(totalMarks) * 100
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM classes WHERE activity_id = $1 AND class_date >= $2 AND class_date < $3
	`, activityID, from, to).Scan(&s.ClassCount); err != nil {
		return nil, err
	}

	perfect, err := r.PerfectAttendanceStudents(ctx, activityID, month, year)
	if err != nil {
		return nil, err
	}
	s.PerfectAttendance = int64(len(perfect))

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount) FILTER (WHERE status = 'paid'), 0),
		       COALESCE(SUM(amount) FILTER (WHERE status IN ('pending','overdue')), 0)
		FROM fees WHERE activity_id = $1 AND period_month = $2 AND period_year = $3
	`, activityID, month, year).Scan(&s.FeesCollected, &s.FeesPending); err != nil {
		return nil, err
	}

	// Distinct calendar days, not attendance rows — a coach teaching three
	// classes in one day was present one day, not three.
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT ca.attendance_date) FROM coach_attendance ca
		JOIN coach_activities cact ON cact.coach_id = ca.coach_id
		WHERE cact.activity_id = $1 AND ca.attendance_date >= $2 AND ca.attendance_date < $3
	`, activityID, from, to).Scan(&s.CoachAttendanceDays); err != nil {
		return nil, err
	}

	return &s, nil
}

type StudentAttendanceSummary struct {
	StudentID      uuid.UUID `json:"student_id"`
	StudentName    string    `json:"student_name"`
	TotalClasses   int64     `json:"total_classes"`
	PresentCount   int64     `json:"present_count"`
	PercentPresent float64   `json:"percent_present"`
}

// PerfectAttendanceStudents lists every student in the activity who was
// present at every completed class held for their batch in the month —
// i.e. exactly 100% attendance, not merely "no absences recorded".
func (r *Repo) PerfectAttendanceStudents(ctx context.Context, activityID uuid.UUID, month, year int) ([]StudentAttendanceSummary, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name,
		       COUNT(DISTINCT cl.id) AS total_classes,
		       COUNT(DISTINCT cl.id) FILTER (WHERE sa.status IN ('present','late')) AS present_count
		FROM students s
		JOIN classes cl ON cl.batch_id = s.batch_id AND cl.status = 'completed'
		                AND cl.class_date >= $2 AND cl.class_date < $3
		LEFT JOIN student_attendance sa ON sa.class_id = cl.id AND sa.student_id = s.id
		WHERE s.activity_id = $1 AND s.is_active = true
		GROUP BY s.id, s.name
		HAVING COUNT(DISTINCT cl.id) > 0
		   AND COUNT(DISTINCT cl.id) = COUNT(DISTINCT cl.id) FILTER (WHERE sa.status IN ('present','late'))
		ORDER BY s.name
	`, activityID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []StudentAttendanceSummary
	for rows.Next() {
		var s StudentAttendanceSummary
		if err := rows.Scan(&s.StudentID, &s.StudentName, &s.TotalClasses, &s.PresentCount); err != nil {
			return nil, err
		}
		s.PercentPresent = 100
		out = append(out, s)
	}
	if out == nil {
		out = []StudentAttendanceSummary{}
	}
	return out, nil
}

// MonthlyStudentReport returns every student's attendance percentage for
// the activity/month — the full roster, used for the "monthly report" /
// CSV export, not just the 100% subset.
func (r *Repo) MonthlyStudentReport(ctx context.Context, activityID uuid.UUID, month, year int) ([]StudentAttendanceSummary, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name,
		       COUNT(DISTINCT cl.id) AS total_classes,
		       COUNT(DISTINCT cl.id) FILTER (WHERE sa.status IN ('present','late')) AS present_count
		FROM students s
		JOIN classes cl ON cl.batch_id = s.batch_id AND cl.status = 'completed'
		                AND cl.class_date >= $2 AND cl.class_date < $3
		LEFT JOIN student_attendance sa ON sa.class_id = cl.id AND sa.student_id = s.id
		WHERE s.activity_id = $1 AND s.is_active = true
		GROUP BY s.id, s.name
		ORDER BY s.name
	`, activityID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []StudentAttendanceSummary
	for rows.Next() {
		var s StudentAttendanceSummary
		if err := rows.Scan(&s.StudentID, &s.StudentName, &s.TotalClasses, &s.PresentCount); err != nil {
			return nil, err
		}
		if s.TotalClasses > 0 {
			s.PercentPresent = float64(s.PresentCount) / float64(s.TotalClasses) * 100
		}
		out = append(out, s)
	}
	if out == nil {
		out = []StudentAttendanceSummary{}
	}
	return out, nil
}

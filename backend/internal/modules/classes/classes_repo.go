package classes

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Class struct {
	ID                 uuid.UUID  `json:"id"`
	BatchID            uuid.UUID  `json:"batch_id"`
	ActivityID         uuid.UUID  `json:"activity_id"`
	CoachID            uuid.UUID  `json:"coach_id"`
	ClassDate          time.Time  `json:"class_date"`
	StartTime          string     `json:"start_time"`
	EndTime            string     `json:"end_time"`
	Status             string     `json:"status"`
	AttendanceMarked   bool       `json:"attendance_marked"`
	AttendanceMarkedAt *time.Time `json:"attendance_marked_at"`
}

type ClassRepo struct {
	pool *pgxpool.Pool
}

func NewClassRepo(pool *pgxpool.Pool) *ClassRepo { return &ClassRepo{pool: pool} }

type ClassInput struct {
	BatchID    uuid.UUID
	ActivityID uuid.UUID
	CoachID    uuid.UUID
	ClassDate  time.Time
	StartTime  string
	EndTime    string
}

func (r *ClassRepo) Create(ctx context.Context, in ClassInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO classes (batch_id, activity_id, coach_id, class_date, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, in.BatchID, in.ActivityID, in.CoachID, in.ClassDate, in.StartTime, in.EndTime).Scan(&id)
	return id, err
}

func (r *ClassRepo) Get(ctx context.Context, id uuid.UUID) (*Class, error) {
	var c Class
	err := r.pool.QueryRow(ctx, `
		SELECT id, batch_id, activity_id, coach_id, class_date, start_time::text, end_time::text,
		       status, attendance_marked, attendance_marked_at
		FROM classes WHERE id = $1
	`, id).Scan(&c.ID, &c.BatchID, &c.ActivityID, &c.CoachID, &c.ClassDate, &c.StartTime, &c.EndTime,
		&c.Status, &c.AttendanceMarked, &c.AttendanceMarkedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type ClassFilter struct {
	AllowedActivityIDs []uuid.UUID // nil for admin (no restriction)
	ActivityID         *uuid.UUID
	CoachID            *uuid.UUID
	DateFrom           *time.Time
	DateTo             *time.Time
}

func (r *ClassRepo) List(ctx context.Context, f ClassFilter, limit, offset int) ([]Class, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	add := func(cond string, v interface{}) {
		args = append(args, v)
		where = append(where, cond+placeholder(len(args)))
	}

	if f.AllowedActivityIDs != nil {
		add("activity_id = ANY($", f.AllowedActivityIDs)
		where[len(where)-1] += ")"
	}
	if f.ActivityID != nil {
		add("activity_id = $", *f.ActivityID)
	}
	if f.CoachID != nil {
		add("coach_id = $", *f.CoachID)
	}
	if f.DateFrom != nil {
		add("class_date >= $", *f.DateFrom)
	}
	if f.DateTo != nil {
		add("class_date <= $", *f.DateTo)
	}

	whereClause := "WHERE " + joinConds(where)

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM classes "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := `
		SELECT id, batch_id, activity_id, coach_id, class_date, start_time::text, end_time::text,
		       status, attendance_marked, attendance_marked_at
		FROM classes ` + whereClause + `
		ORDER BY class_date DESC, start_time DESC
		LIMIT $` + placeholder(len(args)-1) + ` OFFSET $` + placeholder(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.BatchID, &c.ActivityID, &c.CoachID, &c.ClassDate, &c.StartTime, &c.EndTime,
			&c.Status, &c.AttendanceMarked, &c.AttendanceMarkedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []Class{}
	}
	return out, total, nil
}

// ListForCoach returns classes the coach is either permanently assigned to
// or currently has an active substitution for — the exact set of classes
// they are authorized to act on. This is the SQL-level mirror of
// authz.AuthorizeClassAccess, used so list views never show a coach a
// class they couldn't actually open.
func (r *ClassRepo) ListForCoach(ctx context.Context, coachID uuid.UUID, dateFrom, dateTo *time.Time, limit, offset int) ([]Class, int64, error) {
	where := []string{`(
		coach_id = $1
		OR id IN (SELECT class_id FROM substitutions WHERE substitute_coach_id = $1 AND status = 'active')
	)`}
	args := []interface{}{coachID}
	if dateFrom != nil {
		args = append(args, *dateFrom)
		where = append(where, "class_date >= $"+placeholder(len(args)))
	}
	if dateTo != nil {
		args = append(args, *dateTo)
		where = append(where, "class_date <= $"+placeholder(len(args)))
	}
	whereClause := "WHERE " + joinConds(where)

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM classes "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := `
		SELECT id, batch_id, activity_id, coach_id, class_date, start_time::text, end_time::text,
		       status, attendance_marked, attendance_marked_at
		FROM classes ` + whereClause + `
		ORDER BY class_date DESC, start_time DESC
		LIMIT $` + placeholder(len(args)-1) + ` OFFSET $` + placeholder(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.BatchID, &c.ActivityID, &c.CoachID, &c.ClassDate, &c.StartTime, &c.EndTime,
			&c.Status, &c.AttendanceMarked, &c.AttendanceMarkedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []Class{}
	}
	return out, total, nil
}

type RosterStudent struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// RosterForClass returns the active students enrolled in a class's batch.
// It intentionally does not filter by activity membership — callers must
// authorize via AuthorizeClassAccess first, which already covers the
// substitute-coach case that a general activity check would otherwise
// block despite the substitute being legitimately authorized for this
// one class.
func (r *ClassRepo) RosterForClass(ctx context.Context, classID uuid.UUID) ([]RosterStudent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name
		FROM students s
		JOIN classes cl ON cl.batch_id = s.batch_id
		WHERE cl.id = $1 AND s.is_active = true
		ORDER BY s.name
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RosterStudent
	for rows.Next() {
		var s RosterStudent
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []RosterStudent{}
	}
	return out, nil
}

func (r *ClassRepo) MarkAttendanceDone(ctx context.Context, classID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE classes SET attendance_marked = true, attendance_marked_at = now(), status = 'completed'
		WHERE id = $1
	`, classID)
	return err
}

// PendingAttendance returns classes whose end_time was more than
// `minutesAfterEnd` minutes ago (on class_date), attendance not yet
// marked, and a reminder not already sent — used by the 15-minute
// reminder job. Relies on idx_classes_pending_attendance.
func (r *ClassRepo) PendingAttendance(ctx context.Context, minutesAfterEnd int) ([]Class, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, batch_id, activity_id, coach_id, class_date, start_time::text, end_time::text,
		       status, attendance_marked, attendance_marked_at
		FROM classes
		WHERE attendance_marked = false
		  AND status = 'scheduled'
		  AND reminder_sent_at IS NULL
		  AND (class_date + end_time) < (now() - make_interval(mins => $1))
	`, minutesAfterEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.BatchID, &c.ActivityID, &c.CoachID, &c.ClassDate, &c.StartTime, &c.EndTime,
			&c.Status, &c.AttendanceMarked, &c.AttendanceMarkedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *ClassRepo) MarkReminderSent(ctx context.Context, classID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE classes SET reminder_sent_at = now() WHERE id = $1`, classID)
	return err
}

// MissingAttendanceForDate is used by the end-of-day report: every class
// on the given date, past its end time, still unmarked.
func (r *ClassRepo) MissingAttendanceForDate(ctx context.Context, date time.Time) ([]Class, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, batch_id, activity_id, coach_id, class_date, start_time::text, end_time::text,
		       status, attendance_marked, attendance_marked_at
		FROM classes
		WHERE class_date = $1 AND attendance_marked = false AND (class_date + end_time) < now()
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.BatchID, &c.ActivityID, &c.CoachID, &c.ClassDate, &c.StartTime, &c.EndTime,
			&c.Status, &c.AttendanceMarked, &c.AttendanceMarkedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func joinConds(conds []string) string {
	out := conds[0]
	for _, c := range conds[1:] {
		out += " AND " + c
	}
	return out
}

package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentAttendanceRecord struct {
	ID         uuid.UUID `json:"id"`
	ClassID    uuid.UUID `json:"class_id"`
	StudentID  uuid.UUID `json:"student_id"`
	ActivityID uuid.UUID `json:"activity_id"`
	Status     string    `json:"status"`
	MarkedBy   uuid.UUID `json:"marked_by"`
	MarkedAt   time.Time `json:"marked_at"`
	Remarks    string    `json:"remarks"`
}

type StudentAttendanceRepo struct {
	pool *pgxpool.Pool
}

func NewStudentAttendanceRepo(pool *pgxpool.Pool) *StudentAttendanceRepo {
	return &StudentAttendanceRepo{pool: pool}
}

type MarkInput struct {
	ClassID    uuid.UUID
	StudentID  uuid.UUID
	ActivityID uuid.UUID
	Status     string
	Remarks    string
	MarkedBy   uuid.UUID
}

// Upsert inserts or corrects an attendance mark. The UNIQUE(class_id,
// student_id) constraint plus ON CONFLICT DO UPDATE guarantees a student
// can never end up with two attendance rows for the same class, while
// still allowing a coach to fix a mis-tap before the class window closes.
func (r *StudentAttendanceRepo) Upsert(ctx context.Context, in MarkInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO student_attendance (class_id, student_id, activity_id, status, marked_by, remarks)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (class_id, student_id)
		DO UPDATE SET status = EXCLUDED.status, remarks = EXCLUDED.remarks,
		              marked_by = EXCLUDED.marked_by, marked_at = now()
	`, in.ClassID, in.StudentID, in.ActivityID, in.Status, in.MarkedBy, in.Remarks)
	return err
}

type ClassInfo struct {
	BatchID    uuid.UUID
	ActivityID uuid.UUID
}

func (r *StudentAttendanceRepo) GetClassInfo(ctx context.Context, classID uuid.UUID) (*ClassInfo, error) {
	var ci ClassInfo
	err := r.pool.QueryRow(ctx, `SELECT batch_id, activity_id FROM classes WHERE id = $1`, classID).Scan(&ci.BatchID, &ci.ActivityID)
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

// StudentBelongsToBatch guards against a coach marking attendance for a
// student who isn't actually enrolled in the class's batch.
func (r *StudentAttendanceRepo) StudentBelongsToBatch(ctx context.Context, studentID, batchID uuid.UUID) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE id = $1 AND batch_id = $2`, studentID, batchID).Scan(&count)
	return count > 0, err
}

func (r *StudentAttendanceRepo) ListForClass(ctx context.Context, classID uuid.UUID) ([]StudentAttendanceRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, class_id, student_id, activity_id, status, marked_by, marked_at, COALESCE(remarks,'')
		FROM student_attendance WHERE class_id = $1
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []StudentAttendanceRecord
	for rows.Next() {
		var a StudentAttendanceRecord
		if err := rows.Scan(&a.ID, &a.ClassID, &a.StudentID, &a.ActivityID, &a.Status, &a.MarkedBy, &a.MarkedAt, &a.Remarks); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []StudentAttendanceRecord{}
	}
	return out, nil
}

// HistoryForStudent paginates a single student's attendance, most recent
// first — backed by idx_student_attendance_student so it stays fast even
// once the table holds millions of rows.
func (r *StudentAttendanceRepo) HistoryForStudent(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]StudentAttendanceRecord, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM student_attendance WHERE student_id = $1`, studentID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, class_id, student_id, activity_id, status, marked_by, marked_at, COALESCE(remarks,'')
		FROM student_attendance
		WHERE student_id = $1
		ORDER BY marked_at DESC
		LIMIT $2 OFFSET $3
	`, studentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []StudentAttendanceRecord
	for rows.Next() {
		var a StudentAttendanceRecord
		if err := rows.Scan(&a.ID, &a.ClassID, &a.StudentID, &a.ActivityID, &a.Status, &a.MarkedBy, &a.MarkedAt, &a.Remarks); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []StudentAttendanceRecord{}
	}
	return out, total, nil
}

type MonthlyPercentage struct {
	StudentID      uuid.UUID `json:"student_id"`
	Year           int       `json:"year"`
	Month          int       `json:"month"`
	TotalClasses   int64     `json:"total_classes"`
	PresentCount   int64     `json:"present_count"`
	AbsentCount    int64     `json:"absent_count"`
	PercentPresent float64   `json:"percent_present"`
}

// MonthlyPercentage computes attendance percentage against every completed
// class held for the student's batch in the given month — not just the
// classes the student happened to have a row for — so a student who was
// never marked at all still counts as absent for that class.
func (r *StudentAttendanceRepo) MonthlyPercentage(ctx context.Context, studentID uuid.UUID, batchID uuid.UUID, year, month int) (*MonthlyPercentage, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	var total, present int64
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT cl.id),
			COUNT(DISTINCT cl.id) FILTER (WHERE sa.status = 'present' OR sa.status = 'late')
		FROM classes cl
		LEFT JOIN student_attendance sa ON sa.class_id = cl.id AND sa.student_id = $1
		WHERE cl.batch_id = $2 AND cl.status = 'completed'
		  AND cl.class_date >= $3 AND cl.class_date < $4
	`, studentID, batchID, from, to).Scan(&total, &present)
	if err != nil {
		return nil, err
	}

	pct := 0.0
	if total > 0 {
		pct = (float64(present) / float64(total)) * 100
	}

	return &MonthlyPercentage{
		StudentID: studentID, Year: year, Month: month,
		TotalClasses: total, PresentCount: present, AbsentCount: total - present,
		PercentPresent: pct,
	}, nil
}

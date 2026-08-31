package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CoachAttendanceRecord struct {
	ID                uuid.UUID  `json:"id"`
	CoachID           uuid.UUID  `json:"coach_id"`
	ClassID           *uuid.UUID `json:"class_id"`
	AttendanceDate    time.Time  `json:"attendance_date"`
	CheckInTime       *time.Time `json:"check_in_time"`
	CheckInLat        *float64   `json:"check_in_lat"`
	CheckInLng        *float64   `json:"check_in_lng"`
	CheckInVerified   bool       `json:"check_in_verified"`
	CheckOutTime      *time.Time `json:"check_out_time"`
	CheckOutLat       *float64   `json:"check_out_lat"`
	CheckOutLng       *float64   `json:"check_out_lng"`
	CheckOutVerified  bool       `json:"check_out_verified"`
	Status            string     `json:"status"`
}

type CoachAttendanceRepo struct {
	pool *pgxpool.Pool
}

func NewCoachAttendanceRepo(pool *pgxpool.Pool) *CoachAttendanceRepo {
	return &CoachAttendanceRepo{pool: pool}
}

// LocationForClass resolves the geofence center configured for a class via
// its batch, falling back to an activity-level location only when the
// activity has exactly one — with two or more locations on the activity,
// guessing which one this class means would make geofencing meaningless
// (it could silently verify a coach at the wrong site), so that case is
// treated the same as "no geofence configured" rather than picking one.
type Location struct {
	Latitude     float64
	Longitude    float64
	RadiusMeters int
}

func (r *CoachAttendanceRepo) LocationForClass(ctx context.Context, classID uuid.UUID) (*Location, error) {
	var loc Location
	err := r.pool.QueryRow(ctx, `
		SELECT l.latitude, l.longitude, l.radius_meters
		FROM classes cl
		JOIN batches b ON b.id = cl.batch_id
		JOIN locations l ON l.id = COALESCE(b.location_id, (
			SELECT MIN(id::text)::uuid FROM locations WHERE activity_id = cl.activity_id
			GROUP BY activity_id HAVING COUNT(*) = 1
		))
		WHERE cl.id = $1
	`, classID).Scan(&loc.Latitude, &loc.Longitude, &loc.RadiusMeters)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

// CheckIn requires a concrete class_id (never NULL) so the UNIQUE(coach_id,
// class_id) constraint can do its job via ON CONFLICT — Postgres treats
// NULLs as distinct in unique constraints, which would otherwise let a
// coach create unlimited duplicate check-in rows.
func (r *CoachAttendanceRepo) CheckIn(ctx context.Context, coachID, classID uuid.UUID, lat, lng float64, verified bool) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO coach_attendance (coach_id, class_id, attendance_date, check_in_time, check_in_lat, check_in_lng, check_in_verified)
		VALUES ($1, $2, CURRENT_DATE, now(), $3, $4, $5)
		ON CONFLICT (coach_id, class_id) DO UPDATE SET
			check_in_time = now(), check_in_lat = $3, check_in_lng = $4, check_in_verified = $5
		RETURNING id
	`, coachID, classID, lat, lng, verified).Scan(&id)
	return id, err
}

func (r *CoachAttendanceRepo) CheckOut(ctx context.Context, coachID, classID uuid.UUID, lat, lng float64, verified bool) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE coach_attendance
		SET check_out_time = now(), check_out_lat = $3, check_out_lng = $4, check_out_verified = $5
		WHERE coach_id = $1 AND class_id = $2
	`, coachID, classID, lat, lng, verified)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoCheckIn
	}
	return nil
}

type CoachAttendanceFilter struct {
	CoachID  *uuid.UUID
	DateFrom *time.Time
	DateTo   *time.Time
}

func (r *CoachAttendanceRepo) List(ctx context.Context, f CoachAttendanceFilter, limit, offset int) ([]CoachAttendanceRecord, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	add := func(cond string, v interface{}) {
		args = append(args, v)
		where = append(where, cond+"$"+itoaLocal(len(args)))
	}
	if f.CoachID != nil {
		add("coach_id = ", *f.CoachID)
	}
	if f.DateFrom != nil {
		add("attendance_date >= ", *f.DateFrom)
	}
	if f.DateTo != nil {
		add("attendance_date <= ", *f.DateTo)
	}
	whereClause := "WHERE " + joinLocal(where, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM coach_attendance "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := `
		SELECT id, coach_id, class_id, attendance_date, check_in_time, check_in_lat, check_in_lng, check_in_verified,
		       check_out_time, check_out_lat, check_out_lng, check_out_verified, status
		FROM coach_attendance ` + whereClause + `
		ORDER BY attendance_date DESC, check_in_time DESC
		LIMIT $` + itoaLocal(len(args)-1) + ` OFFSET $` + itoaLocal(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []CoachAttendanceRecord
	for rows.Next() {
		var c CoachAttendanceRecord
		if err := rows.Scan(&c.ID, &c.CoachID, &c.ClassID, &c.AttendanceDate, &c.CheckInTime, &c.CheckInLat, &c.CheckInLng,
			&c.CheckInVerified, &c.CheckOutTime, &c.CheckOutLat, &c.CheckOutLng, &c.CheckOutVerified, &c.Status); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []CoachAttendanceRecord{}
	}
	return out, total, nil
}

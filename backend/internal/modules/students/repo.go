package students

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Student struct {
	ID            uuid.UUID  `json:"id"`
	BatchID       uuid.UUID  `json:"batch_id"`
	ActivityID    uuid.UUID  `json:"activity_id"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	GuardianName  string     `json:"guardian_name"`
	GuardianPhone string     `json:"guardian_phone"`
	Email         string     `json:"email"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	JoiningDate   time.Time  `json:"joining_date"`
	IsActive      bool       `json:"is_active"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type CreateInput struct {
	BatchID       uuid.UUID
	ActivityID    uuid.UUID
	Name          string
	Phone         string
	GuardianName  string
	GuardianPhone string
	Email         string
	DateOfBirth   *time.Time
}

// GetBatchActivityID resolves the activity a batch belongs to. Callers
// must use this rather than trusting a client-supplied activity_id, so a
// student's denormalized activity_id can never drift from its batch's
// real activity (which would otherwise let it slip past activity-scoped
// authorization checks).
func (r *Repo) GetBatchActivityID(ctx context.Context, batchID uuid.UUID) (uuid.UUID, error) {
	var activityID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT activity_id FROM batches WHERE id = $1`, batchID).Scan(&activityID)
	return activityID, err
}

func (r *Repo) Create(ctx context.Context, in CreateInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO students (batch_id, activity_id, name, phone, guardian_name, guardian_phone, email, date_of_birth)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, in.BatchID, in.ActivityID, in.Name, in.Phone, in.GuardianName, in.GuardianPhone, in.Email, in.DateOfBirth).Scan(&id)
	return id, err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Student, error) {
	var s Student
	err := r.pool.QueryRow(ctx, `
		SELECT id, batch_id, activity_id, name, COALESCE(phone,''), COALESCE(guardian_name,''),
		       COALESCE(guardian_phone,''), COALESCE(email,''), date_of_birth, joining_date, is_active
		FROM students WHERE id = $1
	`, id).Scan(&s.ID, &s.BatchID, &s.ActivityID, &s.Name, &s.Phone, &s.GuardianName,
		&s.GuardianPhone, &s.Email, &s.DateOfBirth, &s.JoiningDate, &s.IsActive)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// List filters by allowedActivityIDs whenever it is non-nil (i.e. the
// caller is a coach) — this is the SQL-level enforcement of activity
// privacy, independent of and in addition to the service-layer authz
// check, so a bug in one layer doesn't expose another activity's roster.
func (r *Repo) List(ctx context.Context, allowedActivityIDs []uuid.UUID, activityFilter, batchFilter *uuid.UUID, limit, offset int) ([]Student, int64, error) {
	where := []string{"s.is_active = true"}
	args := []interface{}{}
	argN := 0

	nextArg := func(v interface{}) string {
		argN++
		args = append(args, v)
		return "$" + strconv.Itoa(argN)
	}

	if allowedActivityIDs != nil {
		where = append(where, "s.activity_id = ANY("+nextArg(allowedActivityIDs)+")")
	}
	if activityFilter != nil {
		where = append(where, "s.activity_id = "+nextArg(*activityFilter))
	}
	if batchFilter != nil {
		where = append(where, "s.batch_id = "+nextArg(*batchFilter))
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students s "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitArg := nextArg(limit)
	offsetArg := nextArg(offset)
	query := `
		SELECT s.id, s.batch_id, s.activity_id, s.name, COALESCE(s.phone,''), COALESCE(s.guardian_name,''),
		       COALESCE(s.guardian_phone,''), COALESCE(s.email,''), s.date_of_birth, s.joining_date, s.is_active
		FROM students s ` + whereClause + `
		ORDER BY s.name
		LIMIT ` + limitArg + ` OFFSET ` + offsetArg

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.BatchID, &s.ActivityID, &s.Name, &s.Phone, &s.GuardianName,
			&s.GuardianPhone, &s.Email, &s.DateOfBirth, &s.JoiningDate, &s.IsActive); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []Student{}
	}
	return out, total, nil
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, in CreateInput, isActive bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE students SET batch_id = $1, name = $2, phone = $3, guardian_name = $4,
		       guardian_phone = $5, email = $6, date_of_birth = $7, is_active = $8
		WHERE id = $9
	`, in.BatchID, in.Name, in.Phone, in.GuardianName, in.GuardianPhone, in.Email, in.DateOfBirth, isActive, id)
	return err
}

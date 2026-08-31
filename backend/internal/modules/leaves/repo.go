package leaves

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Leave struct {
	ID          uuid.UUID  `json:"id"`
	CoachID     uuid.UUID  `json:"coach_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`
	ReviewedBy  *uuid.UUID `json:"reviewed_by"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Create(ctx context.Context, coachID uuid.UUID, start, end time.Time, reason string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO leaves (coach_id, start_date, end_date, reason) VALUES ($1, $2, $3, $4)
		RETURNING id
	`, coachID, start, end, reason).Scan(&id)
	return id, err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Leave, error) {
	var l Leave
	err := r.pool.QueryRow(ctx, `
		SELECT id, coach_id, start_date, end_date, reason, status, reviewed_by, reviewed_at, created_at
		FROM leaves WHERE id = $1
	`, id).Scan(&l.ID, &l.CoachID, &l.StartDate, &l.EndDate, &l.Reason, &l.Status, &l.ReviewedBy, &l.ReviewedAt, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE leaves SET status = $1, reviewed_by = $2, reviewed_at = now() WHERE id = $3
	`, status, reviewedBy, id)
	return err
}

func (r *Repo) ListForCoach(ctx context.Context, coachID uuid.UUID, limit, offset int) ([]Leave, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM leaves WHERE coach_id = $1`, coachID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, coach_id, start_date, end_date, reason, status, reviewed_by, reviewed_at, created_at
		FROM leaves WHERE coach_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, coachID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, err := scanLeaves(rows)
	return out, total, err
}

func (r *Repo) ListAll(ctx context.Context, statusFilter string, limit, offset int) ([]Leave, int64, error) {
	where := ""
	args := []interface{}{}
	if statusFilter != "" {
		where = "WHERE status = $1"
		args = append(args, statusFilter)
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM leaves "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := `
		SELECT id, coach_id, start_date, end_date, reason, status, reviewed_by, reviewed_at, created_at
		FROM leaves ` + where + ` ORDER BY created_at DESC LIMIT $` + placeholder(len(args)-1) + ` OFFSET $` + placeholder(len(args))
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, err := scanLeaves(rows)
	return out, total, err
}

func scanLeaves(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
}) ([]Leave, error) {
	var out []Leave
	for rows.Next() {
		var l Leave
		if err := rows.Scan(&l.ID, &l.CoachID, &l.StartDate, &l.EndDate, &l.Reason, &l.Status, &l.ReviewedBy, &l.ReviewedAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	if out == nil {
		out = []Leave{}
	}
	return out, nil
}

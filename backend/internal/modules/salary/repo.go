package salary

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Acknowledgement struct {
	ID              uuid.UUID  `json:"id"`
	CoachID         uuid.UUID  `json:"coach_id"`
	PeriodMonth     int        `json:"period_month"`
	PeriodYear      int        `json:"period_year"`
	Amount          *float64   `json:"amount"`
	Status          string     `json:"status"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// CreateForAllActiveCoaches generates one pending record per active coach
// for the given period, skipping coaches who already have one. Used by
// both the admin-triggered endpoint and the 10th-of-month scheduled job.
func (r *Repo) CreateForAllActiveCoaches(ctx context.Context, month, year int) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO salary_acknowledgements (coach_id, period_month, period_year, amount)
		SELECT id, $1, $2, monthly_salary FROM coaches WHERE is_active = true
		ON CONFLICT (coach_id, period_month, period_year) DO NOTHING
	`, month, year)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Acknowledgement, error) {
	var a Acknowledgement
	err := r.pool.QueryRow(ctx, `
		SELECT id, coach_id, period_month, period_year, amount, status, acknowledged_at
		FROM salary_acknowledgements WHERE id = $1
	`, id).Scan(&a.ID, &a.CoachID, &a.PeriodMonth, &a.PeriodYear, &a.Amount, &a.Status, &a.AcknowledgedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repo) Acknowledge(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE salary_acknowledgements SET status = 'acknowledged', acknowledged_at = now() WHERE id = $1
	`, id)
	return err
}

func (r *Repo) ListForCoach(ctx context.Context, coachID uuid.UUID, limit, offset int) ([]Acknowledgement, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM salary_acknowledgements WHERE coach_id = $1`, coachID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, coach_id, period_month, period_year, amount, status, acknowledged_at
		FROM salary_acknowledgements WHERE coach_id = $1
		ORDER BY period_year DESC, period_month DESC LIMIT $2 OFFSET $3
	`, coachID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, err := scanAcks(rows)
	return out, total, err
}

func (r *Repo) ListForPeriod(ctx context.Context, month, year int, statusFilter string, limit, offset int) ([]Acknowledgement, int64, error) {
	where := "WHERE period_month = $1 AND period_year = $2"
	args := []interface{}{month, year}
	if statusFilter != "" {
		args = append(args, statusFilter)
		where += " AND status = $3"
	}

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM salary_acknowledgements "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	limitIdx := len(args) - 1
	offsetIdx := len(args)
	query := `
		SELECT id, coach_id, period_month, period_year, amount, status, acknowledged_at
		FROM salary_acknowledgements ` + where + `
		ORDER BY status, coach_id
		LIMIT $` + itoa(limitIdx) + ` OFFSET $` + itoa(offsetIdx)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, err := scanAcks(rows)
	return out, total, err
}

func scanAcks(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
}) ([]Acknowledgement, error) {
	var out []Acknowledgement
	for rows.Next() {
		var a Acknowledgement
		if err := rows.Scan(&a.ID, &a.CoachID, &a.PeriodMonth, &a.PeriodYear, &a.Amount, &a.Status, &a.AcknowledgedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []Acknowledgement{}
	}
	return out, nil
}

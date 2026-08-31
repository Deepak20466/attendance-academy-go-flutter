package fees

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Fee struct {
	ID            uuid.UUID  `json:"id"`
	StudentID     uuid.UUID  `json:"student_id"`
	ActivityID    uuid.UUID  `json:"activity_id"`
	Amount        float64    `json:"amount"`
	DueDate       time.Time  `json:"due_date"`
	PaidDate      *time.Time `json:"paid_date"`
	PaymentMethod string     `json:"payment_method"`
	Status        string     `json:"status"`
	PeriodMonth   int        `json:"period_month"`
	PeriodYear    int        `json:"period_year"`
	Remarks       string     `json:"remarks"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type CreateInput struct {
	StudentID   uuid.UUID
	ActivityID  uuid.UUID
	Amount      float64
	DueDate     time.Time
	PeriodMonth int
	PeriodYear  int
	Remarks     string
}

func (r *Repo) Create(ctx context.Context, in CreateInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO fees (student_id, activity_id, amount, due_date, period_month, period_year, remarks)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, in.StudentID, in.ActivityID, in.Amount, in.DueDate, in.PeriodMonth, in.PeriodYear, in.Remarks).Scan(&id)
	return id, err
}

// GenerateForActivity creates a pending fee record for every active
// student in the activity for the given period and amount, skipping
// students who already have one for that period.
func (r *Repo) GenerateForActivity(ctx context.Context, activityID uuid.UUID, amount float64, dueDate time.Time, month, year int) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO fees (student_id, activity_id, amount, due_date, period_month, period_year)
		SELECT id, activity_id, $2, $3, $4, $5 FROM students WHERE activity_id = $1 AND is_active = true
		ON CONFLICT (student_id, period_month, period_year) DO NOTHING
	`, activityID, amount, dueDate, month, year)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Fee, error) {
	var f Fee
	err := r.pool.QueryRow(ctx, `
		SELECT id, student_id, activity_id, amount, due_date, paid_date, COALESCE(payment_method,''),
		       status, period_month, period_year, COALESCE(remarks,'')
		FROM fees WHERE id = $1
	`, id).Scan(&f.ID, &f.StudentID, &f.ActivityID, &f.Amount, &f.DueDate, &f.PaidDate, &f.PaymentMethod,
		&f.Status, &f.PeriodMonth, &f.PeriodYear, &f.Remarks)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *Repo) MarkPaid(ctx context.Context, id uuid.UUID, paidDate time.Time, method string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE fees SET status = 'paid', paid_date = $1, payment_method = $2 WHERE id = $3
	`, paidDate, method, id)
	return err
}

type ListFilter struct {
	AllowedActivityIDs []uuid.UUID
	ActivityID         *uuid.UUID
	StudentID          *uuid.UUID
	Status             string
}

func (r *Repo) List(ctx context.Context, f ListFilter, limit, offset int) ([]Fee, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	add := func(cond string, v interface{}) {
		args = append(args, v)
		where = append(where, cond+"$"+strconv.Itoa(len(args)))
	}
	if f.AllowedActivityIDs != nil {
		args = append(args, f.AllowedActivityIDs)
		where = append(where, "activity_id = ANY($"+strconv.Itoa(len(args))+")")
	}
	if f.ActivityID != nil {
		add("activity_id = ", *f.ActivityID)
	}
	if f.StudentID != nil {
		add("student_id = ", *f.StudentID)
	}
	if f.Status != "" {
		add("status = ", f.Status)
	}
	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM fees "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := `
		SELECT id, student_id, activity_id, amount, due_date, paid_date, COALESCE(payment_method,''),
		       status, period_month, period_year, COALESCE(remarks,'')
		FROM fees ` + whereClause + `
		ORDER BY due_date DESC
		LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Fee
	for rows.Next() {
		var f Fee
		if err := rows.Scan(&f.ID, &f.StudentID, &f.ActivityID, &f.Amount, &f.DueDate, &f.PaidDate, &f.PaymentMethod,
			&f.Status, &f.PeriodMonth, &f.PeriodYear, &f.Remarks); err != nil {
			return nil, 0, err
		}
		out = append(out, f)
	}
	if out == nil {
		out = []Fee{}
	}
	return out, total, nil
}

// PendingSummaryByActivity powers the end-of-month pending-fee report:
// total pending amount and count per activity.
type PendingSummary struct {
	ActivityID   uuid.UUID `json:"activity_id"`
	PendingCount int64     `json:"pending_count"`
	PendingTotal float64   `json:"pending_total"`
}

func (r *Repo) PendingSummaryByActivity(ctx context.Context) ([]PendingSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT activity_id, COUNT(*), COALESCE(SUM(amount), 0)
		FROM fees WHERE status IN ('pending', 'overdue')
		GROUP BY activity_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PendingSummary
	for rows.Next() {
		var p PendingSummary
		if err := rows.Scan(&p.ActivityID, &p.PendingCount, &p.PendingTotal); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// MarkOverdue flips any still-pending fee whose due date has passed to
// 'overdue' — called by the daily job before generating reports.
func (r *Repo) MarkOverdue(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE fees SET status = 'overdue' WHERE status = 'pending' AND due_date < CURRENT_DATE
	`)
	return err
}


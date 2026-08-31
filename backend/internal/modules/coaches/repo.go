package coaches

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Coach struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	Name           string      `json:"name"`
	Email          string      `json:"email"`
	Phone          string      `json:"phone"`
	EmployeeCode   string      `json:"employee_code"`
	JoiningDate    time.Time   `json:"joining_date"`
	MonthlySalary  float64     `json:"monthly_salary"`
	IsActive       bool        `json:"is_active"`
	ActivityIDs    []uuid.UUID `json:"activity_ids"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type CreateInput struct {
	UserID        uuid.UUID
	EmployeeCode  string
	MonthlySalary float64
	ActivityIDs   []uuid.UUID
}

// Create inserts the coach profile and its activity assignments atomically
// so a coach is never left without any authorized activity scope (or with
// a scope that doesn't match what the admin intended) if something fails
// mid-way.
func (r *Repo) Create(ctx context.Context, in CreateInput) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var coachID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO coaches (user_id, employee_code, monthly_salary)
		VALUES ($1, $2, $3) RETURNING id
	`, in.UserID, in.EmployeeCode, in.MonthlySalary).Scan(&coachID)
	if err != nil {
		return uuid.Nil, err
	}

	for _, actID := range in.ActivityIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO coach_activities (coach_id, activity_id) VALUES ($1, $2)
		`, coachID, actID); err != nil {
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return coachID, nil
}

func (r *Repo) SetActivities(ctx context.Context, coachID uuid.UUID, activityIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM coach_activities WHERE coach_id = $1`, coachID); err != nil {
		return err
	}
	for _, actID := range activityIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO coach_activities (coach_id, activity_id) VALUES ($1, $2)
		`, coachID, actID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

const selectCoachCols = `
	SELECT c.id, c.user_id, u.name, u.email, COALESCE(u.phone,''), c.employee_code,
	       c.joining_date, c.monthly_salary, c.is_active,
	       COALESCE(array_agg(ca.activity_id) FILTER (WHERE ca.activity_id IS NOT NULL), '{}')
	FROM coaches c
	JOIN users u ON u.id = c.user_id
	LEFT JOIN coach_activities ca ON ca.coach_id = c.id
`

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Coach, error) {
	query := selectCoachCols + ` WHERE c.id = $1 GROUP BY c.id, u.name, u.email, u.phone`
	var c Coach
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.EmployeeCode,
		&c.JoiningDate, &c.MonthlySalary, &c.IsActive, &c.ActivityIDs,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) GetByUserID(ctx context.Context, userID uuid.UUID) (*Coach, error) {
	query := selectCoachCols + ` WHERE c.user_id = $1 GROUP BY c.id, u.name, u.email, u.phone`
	var c Coach
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.EmployeeCode,
		&c.JoiningDate, &c.MonthlySalary, &c.IsActive, &c.ActivityIDs,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) List(ctx context.Context, activityID *uuid.UUID, limit, offset int) ([]Coach, int64, error) {
	where := ""
	args := []interface{}{}
	if activityID != nil {
		where = "WHERE ca.activity_id = $1"
		args = append(args, *activityID)
	}

	var total int64
	countQuery := `SELECT COUNT(DISTINCT c.id) FROM coaches c LEFT JOIN coach_activities ca ON ca.coach_id = c.id ` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := selectCoachCols + where + `
		GROUP BY c.id, u.name, u.email, u.phone
		ORDER BY u.name
		LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Coach
	for rows.Next() {
		var c Coach
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Email, &c.Phone, &c.EmployeeCode,
			&c.JoiningDate, &c.MonthlySalary, &c.IsActive, &c.ActivityIDs); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

func (r *Repo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE coaches SET is_active = $1 WHERE id = $2`, active, id)
	return err
}

func (r *Repo) UpdateSalary(ctx context.Context, id uuid.UUID, salary float64) error {
	_, err := r.pool.Exec(ctx, `UPDATE coaches SET monthly_salary = $1 WHERE id = $2`, salary, id)
	return err
}

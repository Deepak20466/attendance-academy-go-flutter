package classes

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Batch struct {
	ID              uuid.UUID `json:"id"`
	ActivityID      uuid.UUID `json:"activity_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DefaultCoachID  *uuid.UUID `json:"default_coach_id"`
	LocationID      *uuid.UUID `json:"location_id"`
	DaysOfWeek      []int16   `json:"days_of_week"`
	StartTime       string    `json:"start_time"`
	EndTime         string    `json:"end_time"`
	IsActive        bool      `json:"is_active"`
}

type BatchRepo struct {
	pool *pgxpool.Pool
}

func NewBatchRepo(pool *pgxpool.Pool) *BatchRepo { return &BatchRepo{pool: pool} }

type BatchInput struct {
	ActivityID     uuid.UUID
	Name           string
	Description    string
	DefaultCoachID *uuid.UUID
	LocationID     *uuid.UUID
	DaysOfWeek     []int16
	StartTime      string
	EndTime        string
}

func (r *BatchRepo) Create(ctx context.Context, in BatchInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO batches (activity_id, name, description, default_coach_id, location_id, days_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, in.ActivityID, in.Name, in.Description, in.DefaultCoachID, in.LocationID, in.DaysOfWeek, in.StartTime, in.EndTime).Scan(&id)
	return id, err
}

func (r *BatchRepo) Get(ctx context.Context, id uuid.UUID) (*Batch, error) {
	var b Batch
	err := r.pool.QueryRow(ctx, `
		SELECT id, activity_id, name, COALESCE(description,''), default_coach_id, location_id,
		       days_of_week, start_time::text, end_time::text, is_active
		FROM batches WHERE id = $1
	`, id).Scan(&b.ID, &b.ActivityID, &b.Name, &b.Description, &b.DefaultCoachID, &b.LocationID,
		&b.DaysOfWeek, &b.StartTime, &b.EndTime, &b.IsActive)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BatchRepo) ListByActivities(ctx context.Context, allowedActivityIDs []uuid.UUID, activityFilter *uuid.UUID) ([]Batch, error) {
	query := `
		SELECT id, activity_id, name, COALESCE(description,''), default_coach_id, location_id,
		       days_of_week, start_time::text, end_time::text, is_active
		FROM batches WHERE is_active = true`
	args := []interface{}{}
	if allowedActivityIDs != nil {
		args = append(args, allowedActivityIDs)
		query += " AND activity_id = ANY($1)"
	}
	if activityFilter != nil {
		args = append(args, *activityFilter)
		query += " AND activity_id = $" + placeholder(len(args))
	}
	query += " ORDER BY name"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Batch
	for rows.Next() {
		var b Batch
		if err := rows.Scan(&b.ID, &b.ActivityID, &b.Name, &b.Description, &b.DefaultCoachID, &b.LocationID,
			&b.DaysOfWeek, &b.StartTime, &b.EndTime, &b.IsActive); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if out == nil {
		out = []Batch{}
	}
	return out, nil
}

func (r *BatchRepo) Update(ctx context.Context, id uuid.UUID, in BatchInput, isActive bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE batches SET name = $1, description = $2, default_coach_id = $3, location_id = $4,
		       days_of_week = $5, start_time = $6, end_time = $7, is_active = $8
		WHERE id = $9
	`, in.Name, in.Description, in.DefaultCoachID, in.LocationID, in.DaysOfWeek, in.StartTime, in.EndTime, isActive, id)
	return err
}

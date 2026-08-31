package locations

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Location struct {
	ID           uuid.UUID `json:"id"`
	ActivityID   uuid.UUID `json:"activity_id"`
	Name         string    `json:"name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type UpsertInput struct {
	ActivityID   uuid.UUID
	Name         string
	Latitude     float64
	Longitude    float64
	RadiusMeters int
}

func (r *Repo) Create(ctx context.Context, in UpsertInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO locations (activity_id, name, latitude, longitude, radius_meters)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, in.ActivityID, in.Name, in.Latitude, in.Longitude, in.RadiusMeters).Scan(&id)
	return id, err
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, in UpsertInput) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE locations SET name = $1, latitude = $2, longitude = $3, radius_meters = $4
		WHERE id = $5
	`, in.Name, in.Latitude, in.Longitude, in.RadiusMeters, id)
	return err
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM locations WHERE id = $1`, id)
	return err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Location, error) {
	var l Location
	err := r.pool.QueryRow(ctx, `
		SELECT id, activity_id, name, latitude, longitude, radius_meters FROM locations WHERE id = $1
	`, id).Scan(&l.ID, &l.ActivityID, &l.Name, &l.Latitude, &l.Longitude, &l.RadiusMeters)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *Repo) ListByActivity(ctx context.Context, activityID uuid.UUID) ([]Location, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, activity_id, name, latitude, longitude, radius_meters
		FROM locations WHERE activity_id = $1 ORDER BY name
	`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Location
	for rows.Next() {
		var l Location
		if err := rows.Scan(&l.ID, &l.ActivityID, &l.Name, &l.Latitude, &l.Longitude, &l.RadiusMeters); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	if out == nil {
		out = []Location{}
	}
	return out, nil
}

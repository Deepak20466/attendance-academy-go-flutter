package activities

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Activity struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Create(ctx context.Context, name, description string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO activities (name, description) VALUES ($1, $2) RETURNING id
	`, name, description).Scan(&id)
	return id, err
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, name, description string, isActive bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE activities SET name = $1, description = $2, is_active = $3 WHERE id = $4
	`, name, description, isActive, id)
	return err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Activity, error) {
	var a Activity
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, COALESCE(description,''), is_active, created_at FROM activities WHERE id = $1
	`, id).Scan(&a.ID, &a.Name, &a.Description, &a.IsActive, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListAll is used only for admin views and internal id-filtering (never
// exposed unscoped to a coach caller).
func (r *Repo) ListAll(ctx context.Context, onlyActive bool) ([]Activity, error) {
	query := `SELECT id, name, COALESCE(description,''), is_active, created_at FROM activities`
	if onlyActive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// ListByIDs returns only the activities in the given id set — used to
// render a coach's own scoped activity list.
func (r *Repo) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]Activity, error) {
	if len(ids) == 0 {
		return []Activity{}, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, COALESCE(description,''), is_active, created_at
		FROM activities WHERE id = ANY($1) ORDER BY name
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

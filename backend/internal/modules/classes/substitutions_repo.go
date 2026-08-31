package classes

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Substitution struct {
	ID                uuid.UUID `json:"id"`
	ClassID           uuid.UUID `json:"class_id"`
	OriginalCoachID   uuid.UUID `json:"original_coach_id"`
	SubstituteCoachID uuid.UUID `json:"substitute_coach_id"`
	AuthorizedBy      uuid.UUID `json:"authorized_by"`
	Reason            string    `json:"reason"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

type SubstitutionRepo struct {
	pool *pgxpool.Pool
}

func NewSubstitutionRepo(pool *pgxpool.Pool) *SubstitutionRepo { return &SubstitutionRepo{pool: pool} }

func (r *SubstitutionRepo) Create(ctx context.Context, classID, originalCoachID, substituteCoachID, authorizedBy uuid.UUID, reason string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO substitutions (class_id, original_coach_id, substitute_coach_id, authorized_by, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, classID, originalCoachID, substituteCoachID, authorizedBy, reason).Scan(&id)
	return id, err
}

func (r *SubstitutionRepo) Cancel(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE substitutions SET status = 'cancelled' WHERE id = $1`, id)
	return err
}

// GetActiveForClass returns the active substitution for a class, if any,
// so callers (e.g. the reminder job) know who is actually responsible for
// the class right now.
func (r *SubstitutionRepo) GetActiveForClass(ctx context.Context, classID uuid.UUID) (*Substitution, error) {
	var s Substitution
	err := r.pool.QueryRow(ctx, `
		SELECT id, class_id, original_coach_id, substitute_coach_id, authorized_by, COALESCE(reason,''), status, created_at
		FROM substitutions WHERE class_id = $1 AND status = 'active'
	`, classID).Scan(&s.ID, &s.ClassID, &s.OriginalCoachID, &s.SubstituteCoachID, &s.AuthorizedBy, &s.Reason, &s.Status, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubstitutionRepo) ListForCoach(ctx context.Context, coachID uuid.UUID) ([]Substitution, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, class_id, original_coach_id, substitute_coach_id, authorized_by, COALESCE(reason,''), status, created_at
		FROM substitutions
		WHERE substitute_coach_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSubstitutions(rows)
}

func (r *SubstitutionRepo) ListAll(ctx context.Context, limit, offset int) ([]Substitution, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM substitutions`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, class_id, original_coach_id, substitute_coach_id, authorized_by, COALESCE(reason,''), status, created_at
		FROM substitutions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out, err := scanSubstitutions(rows)
	return out, total, err
}

func scanSubstitutions(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
}) ([]Substitution, error) {
	var out []Substitution
	for rows.Next() {
		var s Substitution
		if err := rows.Scan(&s.ID, &s.ClassID, &s.OriginalCoachID, &s.SubstituteCoachID,
			&s.AuthorizedBy, &s.Reason, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []Substitution{}
	}
	return out, nil
}

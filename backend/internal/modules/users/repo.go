package users

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	RoleName  string    `json:"role"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Create(ctx context.Context, roleName, name, email, phone, passwordHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (role_id, name, email, phone, password_hash)
		SELECT id, $2, $3, $4, $5 FROM roles WHERE name = $1
		RETURNING id
	`, roleName, name, email, phone, passwordHash).Scan(&id)
	return id, err
}

func (r *Repo) List(ctx context.Context, roleFilter string, limit, offset int) ([]User, int64, error) {
	args := []interface{}{}
	where := ""
	if roleFilter != "" {
		where = "WHERE r.name = $1"
		args = append(args, roleFilter)
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id " + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	limitPos := len(args) - 1
	offsetPos := len(args)
	query := `
		SELECT u.id, r.name, u.name, u.email, COALESCE(u.phone, ''), u.is_active, u.created_at
		FROM users u JOIN roles r ON r.id = u.role_id
		` + where + `
		ORDER BY u.created_at DESC
		LIMIT $` + strconv.Itoa(limitPos) + ` OFFSET $` + strconv.Itoa(offsetPos)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.RoleName, &u.Name, &u.Email, &u.Phone, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, nil
}

func (r *Repo) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET is_active = $1 WHERE id = $2`, active, id)
	return err
}

func (r *Repo) UpdatePasswordHash(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func (r *Repo) GetPasswordHash(ctx context.Context, id uuid.UUID) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&hash)
	return hash, err
}

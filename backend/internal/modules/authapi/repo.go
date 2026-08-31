package authapi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRecord struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	RoleName     string
	IsActive     bool
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) FindByEmail(ctx context.Context, email string) (*UserRecord, error) {
	var u UserRecord
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.password_hash, r.name, u.is_active
		FROM users u JOIN roles r ON r.id = u.role_id
		WHERE u.email = $1
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.RoleName, &u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) findByID(ctx context.Context, id uuid.UUID) (*UserRecord, error) {
	var u UserRecord
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.password_hash, r.name, u.is_active
		FROM users u JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.RoleName, &u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

type RefreshTokenRecord struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
	Revoked   bool
}

func (r *Repo) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenRecord, error) {
	var rt RefreshTokenRecord
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token_hash = $1
	`, tokenHash).Scan(&rt.UserID, &rt.ExpiresAt, &rt.Revoked)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// RevokeRefreshToken invalidates a single token (used on refresh rotation
// and logout).
func (r *Repo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *Repo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID)
	return err
}

func (r *Repo) UpdateFCMToken(ctx context.Context, userID uuid.UUID, fcmToken string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET fcm_token = $1 WHERE id = $2`, fcmToken, userID)
	return err
}

var ErrNoRows = pgx.ErrNoRows

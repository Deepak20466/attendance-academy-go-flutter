package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	IsRead    bool            `json:"is_read"`
	CreatedAt time.Time       `json:"created_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Create(ctx context.Context, userID uuid.UUID, title, body, notifType string, data map[string]string) (uuid.UUID, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, title, body, type, data, sent_at)
		VALUES ($1, $2, $3, $4, $5, now())
		RETURNING id
	`, userID, title, body, notifType, dataJSON).Scan(&id)
	return id, err
}

func (r *Repo) ListForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Notification, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, title, body, type, data, is_read, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	if out == nil {
		out = []Notification{}
	}
	return out, total, nil
}

// ErrNotFound is returned when the notification doesn't exist or belongs
// to a different user — the two cases are indistinguishable on purpose,
// so a caller probing other users' notification IDs learns nothing.
var ErrNotFound = errors.New("notification not found")

func (r *Repo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UserDeviceToken and role helpers used by the job scheduler to know who
// to notify.
func (r *Repo) FCMTokenForUser(ctx context.Context, userID uuid.UUID) (string, error) {
	var token *string
	err := r.pool.QueryRow(ctx, `SELECT fcm_token FROM users WHERE id = $1`, userID).Scan(&token)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", nil
	}
	return *token, nil
}

func (r *Repo) AdminUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id FROM users u JOIN roles r ON r.id = u.role_id WHERE r.name = 'admin' AND u.is_active = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func (r *Repo) UserIDForCoach(ctx context.Context, coachID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT user_id FROM coaches WHERE id = $1`, coachID).Scan(&userID)
	return userID, err
}

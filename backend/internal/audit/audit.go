package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger struct {
	pool *pgxpool.Pool
}

func NewLogger(pool *pgxpool.Pool) *Logger {
	return &Logger{pool: pool}
}

// Log records an audit trail entry. before/after may be nil; they are
// marshalled to JSONB as-is. Failures to write an audit log are logged by
// the caller's context but must never block the primary operation, so this
// returns an error for the caller to decide how to handle (typically just
// logged, not surfaced to the user).
func (l *Logger) Log(ctx context.Context, actorUserID *uuid.UUID, action, entityType string, entityID *uuid.UUID, before, after interface{}, ipAddress string) error {
	var beforeJSON, afterJSON []byte
	var err error
	if before != nil {
		beforeJSON, err = json.Marshal(before)
		if err != nil {
			return err
		}
	}
	if after != nil {
		afterJSON, err = json.Marshal(after)
		if err != nil {
			return err
		}
	}

	_, err = l.pool.Exec(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, before_data, after_data, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, actorUserID, action, entityType, entityID, beforeJSON, afterJSON, ipAddress)
	return err
}

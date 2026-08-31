package audit

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LogEntry struct {
	ID          uuid.UUID       `json:"id"`
	ActorUserID *uuid.UUID      `json:"actor_user_id"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    *uuid.UUID      `json:"entity_id"`
	BeforeData  json.RawMessage `json:"before_data"`
	AfterData   json.RawMessage `json:"after_data"`
	IPAddress   string          `json:"ip_address"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type Filter struct {
	EntityType string
	EntityID   *uuid.UUID
	ActorID    *uuid.UUID
}

// List is admin-only (route-gated) — the audit trail spans every activity,
// so it must never be exposed to a coach-scoped caller.
func (r *Repo) List(ctx context.Context, f Filter, limit, offset int) ([]LogEntry, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	add := func(cond string, v interface{}) {
		args = append(args, v)
		where = append(where, cond+"$"+strconv.Itoa(len(args)))
	}
	if f.EntityType != "" {
		add("entity_type = ", f.EntityType)
	}
	if f.EntityID != nil {
		add("entity_id = ", *f.EntityID)
	}
	if f.ActorID != nil {
		add("actor_user_id = ", *f.ActorID)
	}
	whereClause := "WHERE " + joinAnd(where)

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs "+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	query := `
		SELECT id, actor_user_id, action, entity_type, entity_id, before_data, after_data, COALESCE(ip_address,''), created_at
		FROM audit_logs ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.ActorUserID, &l.Action, &l.EntityType, &l.EntityID, &l.BeforeData, &l.AfterData, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, l)
	}
	if out == nil {
		out = []LogEntry{}
	}
	return out, total, nil
}

func joinAnd(parts []string) string { return strings.Join(parts, " AND ") }

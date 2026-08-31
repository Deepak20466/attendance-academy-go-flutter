package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	RoleAdmin = "admin"
	RoleCoach = "coach"
)

// Actor is the authenticated identity for the current request, loaded fresh
// from the database on every call (never trusted from the JWT or the
// client) so that role/activity changes and deactivation take effect
// immediately instead of waiting for a token to expire.
type Actor struct {
	UserID      uuid.UUID
	Role        string
	IsActive    bool
	CoachID     *uuid.UUID
	ActivityIDs map[uuid.UUID]bool
}

func (a *Actor) IsAdmin() bool { return a.Role == RoleAdmin }

func (a *Actor) HasActivity(activityID uuid.UUID) bool {
	if a.IsAdmin() {
		return true
	}
	return a.ActivityIDs[activityID]
}

func LoadActor(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*Actor, error) {
	var actor Actor
	actor.UserID = userID
	actor.ActivityIDs = map[uuid.UUID]bool{}

	err := pool.QueryRow(ctx, `
		SELECT r.name, u.is_active
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&actor.Role, &actor.IsActive)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}

	if actor.Role == RoleCoach {
		var coachID uuid.UUID
		err := pool.QueryRow(ctx, `SELECT id FROM coaches WHERE user_id = $1 AND is_active = true`, userID).Scan(&coachID)
		if err != nil {
			return nil, fmt.Errorf("load coach profile: %w", err)
		}
		actor.CoachID = &coachID

		rows, err := pool.Query(ctx, `SELECT activity_id FROM coach_activities WHERE coach_id = $1`, coachID)
		if err != nil {
			return nil, fmt.Errorf("load coach activities: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var actID uuid.UUID
			if err := rows.Scan(&actID); err != nil {
				return nil, err
			}
			actor.ActivityIDs[actID] = true
		}
	}

	return &actor, nil
}

// ActivityIDList returns the activity IDs for use in SQL `= ANY($1)` filters.
func (a *Actor) ActivityIDList() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(a.ActivityIDs))
	for id := range a.ActivityIDs {
		ids = append(ids, id)
	}
	return ids
}

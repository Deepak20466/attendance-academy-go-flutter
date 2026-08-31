package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrForbidden is returned by every guard below when the actor is not
// authorized. Handlers must map it to HTTP 403 and MUST NOT leak whether
// the underlying resource exists to an unauthorized caller.
var ErrForbidden = errors.New("forbidden")

// AuthorizeActivity is the single choke point for "can this actor touch
// data belonging to this activity". Every service method that reads or
// writes activity-scoped data (students, classes, attendance, fees, ...)
// must call this before doing anything else. Centralizing it here means a
// new endpoint can't accidentally skip the check.
func AuthorizeActivity(actor *Actor, activityID uuid.UUID) error {
	if actor.HasActivity(activityID) {
		return nil
	}
	return ErrForbidden
}

// AuthorizeClassAccess allows: admins; the coach permanently assigned to
// the class's activity; or a coach with an active substitution row for
// this exact class. This is the rule that lets a substitute mark
// attendance for one class without gaining broader activity access.
func AuthorizeClassAccess(ctx context.Context, pool *pgxpool.Pool, actor *Actor, classID uuid.UUID) error {
	if actor.IsAdmin() {
		return nil
	}
	if actor.CoachID == nil {
		return ErrForbidden
	}

	var activityID uuid.UUID
	var assignedCoachID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT activity_id, coach_id FROM classes WHERE id = $1
	`, classID).Scan(&activityID, &assignedCoachID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}

	if assignedCoachID == *actor.CoachID && actor.HasActivity(activityID) {
		return nil
	}

	var subCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM substitutions
		WHERE class_id = $1 AND substitute_coach_id = $2 AND status = 'active'
	`, classID, *actor.CoachID).Scan(&subCount)
	if err != nil {
		return err
	}
	if subCount > 0 {
		return nil
	}

	return ErrForbidden
}

// AuthorizeStudentAccess resolves the student's activity and checks it the
// same way as any other activity-scoped resource.
func AuthorizeStudentAccess(ctx context.Context, pool *pgxpool.Pool, actor *Actor, studentID uuid.UUID) error {
	if actor.IsAdmin() {
		return nil
	}
	var activityID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT activity_id FROM students WHERE id = $1`, studentID).Scan(&activityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}
	return AuthorizeActivity(actor, activityID)
}

// AuthorizeCoachSelf allows admins, or a coach acting on their own record.
func AuthorizeCoachSelf(actor *Actor, coachID uuid.UUID) error {
	if actor.IsAdmin() {
		return nil
	}
	if actor.CoachID != nil && *actor.CoachID == coachID {
		return nil
	}
	return ErrForbidden
}

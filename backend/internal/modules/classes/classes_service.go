package classes

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"attendance-backend/internal/authz"
)

// ErrClassConflict means a class already exists for this batch at this
// exact date and time. ErrClassBadReference means batch_id, activity_id or
// coach_id doesn't point at a real row — distinct outcomes that the old
// blanket "conflict" response conflated, misreporting a bad coach/batch
// reference as if the class already existed.
var (
	ErrClassConflict     = errors.New("a class already exists for this batch at this date and time")
	ErrClassBadReference = errors.New("batch_id, activity_id or coach_id does not refer to an existing record")
)

const pgForeignKeyViolationCode = "23503"

type ClassService struct {
	repo *ClassRepo
	pool *pgxpool.Pool
}

func NewClassService(repo *ClassRepo, pool *pgxpool.Pool) *ClassService {
	return &ClassService{repo: repo, pool: pool}
}

// Create is admin-only (enforced at the route level) — coaches never
// schedule classes for themselves.
func (s *ClassService) Create(ctx context.Context, in ClassInput) (uuid.UUID, error) {
	id, err := s.repo.Create(ctx, in)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolationCode:
				return uuid.Nil, ErrClassConflict
			case pgForeignKeyViolationCode:
				return uuid.Nil, ErrClassBadReference
			}
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (s *ClassService) Get(ctx context.Context, actor *authz.Actor, id uuid.UUID) (*Class, error) {
	if err := authz.AuthorizeClassAccess(ctx, s.pool, actor, id); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, id)
}

func (s *ClassService) List(ctx context.Context, actor *authz.Actor, activityFilter, coachFilter *uuid.UUID, dateFrom, dateTo *time.Time, limit, offset int) ([]Class, int64, error) {
	if actor.IsAdmin() {
		return s.repo.List(ctx, ClassFilter{
			ActivityID: activityFilter, CoachID: coachFilter, DateFrom: dateFrom, DateTo: dateTo,
		}, limit, offset)
	}
	// Coaches only ever see classes they own or are substituting for,
	// regardless of what activity/coach filters they pass.
	return s.repo.ListForCoach(ctx, *actor.CoachID, dateFrom, dateTo, limit, offset)
}

func (s *ClassService) MarkAttendanceDone(ctx context.Context, classID uuid.UUID) error {
	return s.repo.MarkAttendanceDone(ctx, classID)
}

// Roster requires only class-level authorization, not general activity
// membership, so a substitute covering exactly this one class can still
// see who to mark attendance for.
func (s *ClassService) Roster(ctx context.Context, actor *authz.Actor, classID uuid.UUID) ([]RosterStudent, error) {
	if err := authz.AuthorizeClassAccess(ctx, s.pool, actor, classID); err != nil {
		return nil, err
	}
	return s.repo.RosterForClass(ctx, classID)
}
